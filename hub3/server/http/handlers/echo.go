package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"reflect"
	"sort"
	"unsafe"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/go-chi/render"
	elastic "github.com/olivere/elastic/v7"
)

type EchoSearchRequest struct {
	echoOn         string
	r              *http.Request
	searchRequest  *fragments.SearchRequest
	searchService  *elastic.SearchService
	searchResponse *elastic.SearchResult
	ScrollPager    *fragments.ScrollPager
}

func NewEchoSearchRequest(
	r *http.Request,
	searchRequest *fragments.SearchRequest,
	searchService *elastic.SearchService,
	searchResult *elastic.SearchResult,

) *EchoSearchRequest {
	echoRequest := r.URL.Query().Get("echo")
	return &EchoSearchRequest{
		r:              r,
		echoOn:         echoRequest,
		searchRequest:  searchRequest,
		searchService:  searchService,
		searchResponse: searchResult,
	}
}

func (e *EchoSearchRequest) HasEcho() bool {
	return e.echoOn != ""
}

func (e *EchoSearchRequest) RenderEcho(w http.ResponseWriter) error {
	if !e.HasEcho() {
		return nil
	}

	switch e.echoOn {
	case "es":
		query, err := e.searchRequest.ElasticQuery()
		if err != nil {
			return err
		}
		source, _ := query.Source()
		render.JSON(w, e.r, source)
	case "searchRequest":
		render.JSON(w, e.r, e.searchRequest)
	case "options":
		options := []string{
			"es", "searchRequest", "options", "searchService", "searchResponse", "request",
			"nextScrollID", "searchAfter", "previousScrollID",
		}
		sort.Strings(options)
		render.JSON(w, e.r, options)
	case "searchService":
		ss := reflect.ValueOf(e.searchService).Elem().FieldByName("searchSource")
		src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)
		srcMap, err := src.Source()
		if err != nil {
			return fmt.Errorf("unable to decode SearchSource: %w", err)
		}
		render.JSON(w, e.r, srcMap)
	case "searchResponse":
		render.JSON(w, e.r, e.searchResponse)
	case "request":
		dump, err := httputil.DumpRequest(e.r, true)
		if err != nil {
			msg := fmt.Sprintf("Unable to dump request: %s", err)
			log.Print(msg)
			return fmt.Errorf("unable to dump http request: %w", err)
		}

		render.PlainText(w, e.r, string(dump))
	case "previousScrollID", "nextScrollID", "searchAfter":
		if e.ScrollPager == nil {
			return fmt.Errorf("%s cannot be used with a nil fragments.ScrollPager", e.echoOn)
		}
		sr, err := fragments.SearchRequestFromHex(e.ScrollPager.NextScrollID)
		if err != nil {
			return fmt.Errorf("unable to decode next scrollID: %w", err)
		}
		if e.echoOn == "previousScrollID" {
			sr, err = fragments.SearchRequestFromHex(e.ScrollPager.PreviousScrollID)
			if err != nil {
				return fmt.Errorf("unable to decode previous scrollID: %w", err)
			}
		}

		if e.echoOn != "searchAfter" {
			render.JSON(w, e.r, sr)
			return nil
		}

		sa, err := sr.DecodeSearchAfter()
		if err != nil {
			return fmt.Errorf("unable to decode next SearchAfter; %w", err)
		}
		render.JSON(w, e.r, sa)
	default:
		return fmt.Errorf("unknown echoType: %s", e.echoOn)
	}

	return nil
}
