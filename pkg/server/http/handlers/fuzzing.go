package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterFuzzer(r chi.Router) {
	r.Post("/api/index/fuzzed", generateFuzzed)
}

func generateFuzzed(w http.ResponseWriter, r *http.Request) {
	in, _, err := r.FormFile("file")
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}
	spec := r.FormValue("spec")
	number := r.FormValue("number")
	baseURI := r.FormValue("baseURI")
	subjectType := r.FormValue("rootType")
	n, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recDef, err := fragments.NewRecDef(in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fz, err := fragments.NewFuzzer(recDef)
	fz.BaseURL = baseURI
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	records, err := fz.CreateRecords(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	typeLabel, err := c.Config.NameSpaceMap.GetSearchLabel(subjectType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	actions := []string{}
	for idx, rec := range records {
		hubID := fmt.Sprintf("%s_%s_%d", c.Config.OrgID, spec, idx)
		action := &hub3.BulkAction{
			HubID:         hubID,
			OrgID:         c.Config.OrgID,
			LocalID:       fmt.Sprintf("%d", idx),
			Spec:          spec,
			NamedGraphURI: fmt.Sprintf("%s/graph", fz.NewURI(typeLabel, idx)),
			Action:        "index",
			GraphMimeType: "application/ld+json",
			SubjectType:   subjectType,
			RecordType:    "mdr",
			Graph:         rec,
		}
		bytes, err := json.Marshal(action)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			log.Printf("Unable to create Bulkactions: %s\n", err.Error())
			render.PlainText(w, r, err.Error())
			return
		}
		actions = append(actions, string(bytes))
	}
	render.PlainText(w, r, strings.Join(actions, "\n"))
	//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	return
}
