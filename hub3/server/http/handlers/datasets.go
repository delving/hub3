// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/asdine/storm"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

var (
	specRoute = "/{spec}"
)

func RegisterDatasets(router chi.Router) {
	r := chi.NewRouter()

	// datasets
	r.Get("/", listDataSets)
	r.Get("/histogram", listDataSetHistogram)
	r.Post("/", createDataSet)
	r.Get(specRoute, getDataSet)
	r.Get("/{spec}/stats", getDataSetStats)
	// later change to update dataset
	r.Post(specRoute, createDataSet)
	r.Delete(specRoute, deleteDataset)

	router.Mount("/api/datasets", r)
}

func listDataSetHistogram(w http.ResponseWriter, r *http.Request) {
	buckets, err := models.NewDataSetHistogram()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, buckets)
}

// listDataSets returns a list of all public datasets
func listDataSets(w http.ResponseWriter, r *http.Request) {
	sets, err := models.ListDataSets()
	if err != nil {
		log.Printf("Unable to list datasets because of: %s", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusInternalServerError,
			Message:    fmt.Sprint("Unable to list datasets was not found"),
			Error:      err,
		})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, sets)
	return
}

// getDataSetStats returns a dataset when found or a 404
func getDataSetStats(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	log.Printf("Get stats for spec %s", spec)
	stats, err := models.CreateDataSetStats(r.Context(), spec)
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", chi.URLParam(r, "spec")),
				Error:      err,
			})
			return
		}
		status := http.StatusInternalServerError
		render.Status(r, status)
		log.Printf("Unable to create dataset stats: %#v", err)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: status,
			Message:    fmt.Sprintf("Can't create stats for %s", spec),
			Error:      err,
		})
		return
	}
	render.JSON(w, r, stats)

	return
}

// getDataSet returns a dataset when found or a 404
func getDataSet(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	ds, err := models.GetDataSet(spec)
	if err != nil {
		if err == storm.ErrNotFound {
			log.Printf("Unable to retrieve a dataset: %s", err)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusNotFound,
				Message:    fmt.Sprintf("%s was not found", spec),
				Error:      err,
			})
			return
		}
		status := http.StatusInternalServerError
		render.Status(r, status)
		log.Printf("Unable to get dataset: %s", spec)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: status,
			Message:    fmt.Sprintf("Can't create stats for %s", spec),
			Error:      err,
		})
		return
	}

	render.JSON(w, r, ds)
	return
}

func deleteDataset(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")

	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		render.Status(r, http.StatusNotFound)
		log.Printf("Dataset is not found: %s", spec)
		return
	}
	ok, err := ds.DropAll(r.Context(), wp)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		log.Printf("Unable to delete request because: %s", err)
		return
	}

	if !ok {
		render.Status(r, http.StatusBadRequest)
		log.Printf("Unable to delete request because: %s", err)
		return
	}
	log.Printf("Dataset is deleted: %s", spec)
	render.Status(r, http.StatusAccepted)
	return
}

// createDataSet creates a new dataset.
func createDataSet(w http.ResponseWriter, r *http.Request) {
	spec := r.FormValue("spec")
	if spec == "" {
		spec = chi.URLParam(r, "spec")
	}
	if spec == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIErrorMessage{
			HTTPStatus: http.StatusBadRequest,
			Message:    fmt.Sprintln("spec can't be empty."),
			Error:      nil,
		})
		return
	}
	fmt.Printf("spec is %s", spec)
	ds, err := models.GetDataSet(spec)
	if err == storm.ErrNotFound {
		var created bool
		ds, created, err = models.CreateDataSet(spec)
		if created {
			err = fragments.SaveDataSet(spec, BulkProcessor())
		}
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, APIErrorMessage{
				HTTPStatus: http.StatusBadRequest,
				Message:    fmt.Sprintf("Unable to create dataset for %s", spec),
				Error:      nil,
			})
			log.Printf("Unable to create dataset for %s.\n", spec)
			return
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, ds)
		return
	}
	render.Status(r, http.StatusNotModified)
	render.JSON(w, r, ds)
	return
}
