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
	"strings"

	"github.com/asdine/storm"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/go-chi/chi"
)

var specRoute = "/{spec}"

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
	r.Delete(specRoute, DeleteDataset)

	router.Mount("/api/datasets", r)
}

func listDataSetHistogram(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)

	buckets, err := models.NewDataSetHistogram(orgID.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, buckets)
}

// listDataSets returns a list of all public datasets
func listDataSets(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)

	sets, err := models.ListDataSets(orgID.String())
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to list datasets",
		})

		return
	}

	if strings.EqualFold(r.URL.Query().Get("applyPreSave"), "true") {
		total := len(sets)

		for idx, ds := range sets {
			if err := ds.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if idx%250 == 0 && idx > 0 {
				log.Printf("presaved %d of %d", idx, total)
			}
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, sets)
}

// getDataSetStats returns a dataset when found or a 404
func getDataSetStats(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	spec := chi.URLParam(r, "spec")
	log.Printf("Get stats for spec %s", domain.LogUserInput(spec))
	stats, err := models.CreateDataSetStats(r.Context(), string(orgID), spec)
	if err != nil {
		if err == storm.ErrNotFound {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("Unable to retrieve dataset: %s", spec),
			})
			return
		}

		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Unable to create dataset stats: %s", spec),
		})

		return
	}

	render.JSON(w, r, stats)
}

// getDataSet returns a dataset when found or a 404
func getDataSet(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	spec := chi.URLParam(r, "spec")
	ds, err := models.GetDataSet(orgID.String(), spec)
	if err != nil {
		if err == storm.ErrNotFound {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("Unable to retrieve dataset: %s", spec),
			})

			return
		}

		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Unable to get dataset: %s", spec),
		})

		return
	}

	render.JSON(w, r, ds)
}

func DeleteDataset(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	spec := chi.URLParam(r, "spec")
	err := models.DeleteDataSet(orgID.String(), spec, r.Context())
	if err != nil {
		if err == storm.ErrNotFound {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("Unable to retrieve dataset: %s", spec),
			})

			return
		}
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("Unable to delete dataset: %s", spec),
		})
		return
	}

	msg := fmt.Sprintf("Dataset is deleted: %s", spec)
	render.Status(r, http.StatusAccepted)
	render.PlainText(w, r, msg)
}

// createDataSet creates a new dataset.
func createDataSet(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)

	spec := r.FormValue("spec")
	if spec == "" {
		spec = chi.URLParam(r, "spec")
	}

	if spec == "" {
		render.Error(w, r, fmt.Errorf("dataset id cannot be empty"), &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
		})

		return
	}

	ds, err := models.GetDataSet(string(orgID), spec)
	if err == storm.ErrNotFound {
		var created bool
		ds, created, err = models.CreateDataSet(string(orgID), spec)
		if created {
			err = fragments.SaveDataSet(string(orgID), spec, nil)
		}
		if err != nil {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("unable to create dataset for %q", spec),
			})
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
