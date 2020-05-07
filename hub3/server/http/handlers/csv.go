package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterCSV(r chi.Router) {
	// CSV upload endpoint
	r.Post("/api/rdf/csv", csvUpload)
	r.Delete("/api/rdf/csv", csvDelete)
}

func csvDelete(w http.ResponseWriter, r *http.Request) {
	conv := fragments.NewCSVConvertor()
	conv.DefaultSpec = r.FormValue("defaultSpec")

	if conv.DefaultSpec == "" {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "defaultSpec is a required field")
		return
	}

	ds, _, err := models.GetOrCreateDataSet(conv.DefaultSpec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", conv.DefaultSpec)
		render.PlainText(w, r, err.Error())
		return
	}
	_, err = ds.DropRecords(ctx, wp)
	if err != nil {
		log.Printf("Unable to delete all fragments for %s: %s", conv.DefaultSpec, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}

	render.Status(r, http.StatusNoContent)
	return
}

func csvUpload(w http.ResponseWriter, r *http.Request) {
	in, _, err := r.FormFile("csv")
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	conv := fragments.NewCSVConvertor()
	conv.InputFile = in
	conv.SubjectColumn = r.FormValue("subjectColumn")
	conv.SubjectClass = r.FormValue("subjectClass")
	conv.SubjectURIBase = r.FormValue("subjectURIBase")
	conv.Separator = r.FormValue("separator")
	conv.PredicateURIBase = r.FormValue("predicateURIBase")
	conv.SubjectColumn = r.FormValue("subjectColumn")
	conv.ObjectResourceColumns = strings.Split(r.FormValue("objectResourceColumns"), ",")
	conv.ObjectIntegerColumns = strings.Split(r.FormValue("objectIntegerColumns"), ",")
	conv.ObjectURIFormat = r.FormValue("objectURIFormat")
	conv.DefaultSpec = r.FormValue("defaultSpec")
	conv.ThumbnailURIBase = r.FormValue("thumbnailURIBase")
	conv.ThumbnailColumn = r.FormValue("thumbnailColumn")
	conv.ManifestColumn = r.FormValue("manifestColumn")
	conv.ManifestURIBase = r.FormValue("manifestURIBase")
	conv.ManifestLocale = r.FormValue("manifestLocale")

	if conv.Separator == "" {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "Separator is a required field. When ';' is the separator you can escape it as '%3B'")
		return
	}

	ds, created, err := models.GetOrCreateDataSet(conv.DefaultSpec)
	if err != nil {
		log.Printf("Unable to get DataSet for %s\n", conv.DefaultSpec)
		render.PlainText(w, r, err.Error())
		return
	}
	if created {
		err = fragments.SaveDataSet(conv.DefaultSpec, BulkProcessor())
		if err != nil {
			log.Printf("Unable to Save DataSet Fragment for %s\n", conv.DefaultSpec)
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

	triplesCreated, rowsSeen, err := conv.IndexFragments(NewOldBulkProcessor(), ds.Revision)
	conv.RowsProcessed = rowsSeen
	conv.TriplesCreated = triplesCreated
	log.Printf("Processed %d csv rows\n", rowsSeen)
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	_, err = ds.DropOrphans(ctx, BulkProcessor(), wp)
	if err != nil {
		render.PlainText(w, r, err.Error())
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, conv)
	return
}
