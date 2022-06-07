package elasticsearch

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	"github.com/olivere/elastic/v7"
)

type Debug struct{}

// SearchService returns the search source (json send to elasticsearch) as JSON []byte
func (d Debug) SearchService(s *elastic.SearchService) ([]byte, error) {
	ss := reflect.ValueOf(s).Elem().FieldByName("searchSource")
	src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)

	srcMap, err := src.Source()
	if err != nil {
		return nil, fmt.Errorf("unable to decode SearchSource: %w", err)
	}

	return json.Marshal(srcMap)
}

func logSearchService(s *elastic.SearchService) {
	ss, err := Debug{}.SearchService(s)
	log.Printf("search service: %s; err -> %#v", ss, err)
}
