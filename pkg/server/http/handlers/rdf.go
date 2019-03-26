package handlers

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gorilla/schema"
)

func RegisterRDF(r chi.Router) {
	// RDF upload endpoint
	r.Post("/api/rdf/source", rdfUpload)
}

type rdfUploadForm struct {
	Spec          string `json:"spec"`
	RecordType    string `json:"recordType"`
	TypePredicate string `json:"typePredicate"`
	IDSplitter    string `json:"idSplitter"`
}

func (ruf *rdfUploadForm) isValid() error {
	if ruf.Spec == "" {
		return fmt.Errorf("spec param is required")
	}
	if ruf.RecordType == "" {
		return fmt.Errorf("recordType param is required")
	}
	if ruf.TypePredicate == "" {
		return fmt.Errorf("typePredicate param is required")
	}
	if ruf.IDSplitter == "" {
		return fmt.Errorf("idSplitter param is required")
	}
	return nil
}

var decoder = schema.NewDecoder()

func rdfUpload(w http.ResponseWriter, r *http.Request) {
	in, header, err := r.FormFile("rdf")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer in.Close()

	var reader io.Reader
	contentType := strings.Split(header.Header.Get("Content-Type"), ";")[0]
	switch contentType {
	case "application/gzip":
		reader, err = gzip.NewReader(in)
		if err != nil {
			log.Printf("Unable to create gzip reader: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case "text/turtle":
		reader = in
	default:
		log.Println("only text/turtle is supported at the moment")
		http.Error(
			w,
			fmt.Sprintf("only text/turtle is suppurted at the moment: %s", contentType),
			http.StatusBadRequest,
		)
		return
	}

	var form rdfUploadForm
	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		log.Printf("Unable to decode form %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = form.isValid()
	if err != nil {
		log.Printf("form is not valid; %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ds, created, err := models.GetOrCreateDataSet(form.Spec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", form.Spec)
		render.PlainText(w, r, err.Error())
		return
	}
	if created {
		err = fragments.SaveDataSet(form.Spec, bulkProcessor())
		if err != nil {
			log.Printf("Unable to Save DataSet Fragment for %s\n", form.Spec)
			if err != nil {
				render.PlainText(w, r, err.Error())
				return
			}
		}
	}

	ds, err = ds.IncrementRevision()
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	upl := fragments.NewRDFUploader(
		c.Config.OrgID,
		form.Spec,
		form.RecordType,
		form.TypePredicate,
		form.IDSplitter,
		ds.Revision,
	)

	go func() {
		log.Print("Start creating resource map")
		_, err := upl.Parse(reader)
		if err != nil {
			log.Printf("Can't read turtle file: %v", err)
			return
		}
		log.Printf("Start saving fragments.")
		processedFragments, err := upl.IndexFragments(bulkProcessor())
		if err != nil {
			log.Printf("Can't save fragments: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Saved %d fragments for %s", processedFragments, upl.Spec)
		processed, err := upl.SaveFragmentGraphs(bulkProcessor())
		if err != nil {
			log.Printf("Can't save records: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Saved %d records for %s", processed, upl.Spec)
		ds.DropOrphans(context.Background(), bulkProcessor(), nil)
	}()

	render.Status(r, http.StatusCreated)
	render.PlainText(w, r, "ok")
	return
}
