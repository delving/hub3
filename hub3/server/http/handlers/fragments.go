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
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
	"github.com/go-chi/chi"
)

func RegisterLinkedDataFragments(router chi.Router) {
	router.Get("/api/fragments", listFragments)
	router.Get("/fragments/{spec}", listFragments)
	router.Get("/fragments", listFragments)
}

// listFragments returns a list of matching fragments
// See for more info: http://linkeddatafragments.org/
func listFragments(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	fr := fragments.NewFragmentRequest(orgID.String())

	spec := chi.URLParam(r, "spec")
	if spec != "" {
		fr.Spec = spec
	}
	err := fr.ParseQueryString(r.URL.Query())
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusBadRequest,
			Message:    "Unable to list fragments",
		})
		return
	}

	frags, totalFrags, err := fr.Find(r.Context(), index.ESClient())

	switch fr.Echo {
	case "raw":
		render.JSON(w, r, frags)
		return
	case "es":
		src, err := fr.BuildQuery().Source()
		if err != nil {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode: http.StatusBadRequest,
				Message:    "Unable to get the query source",
			})
			return
		}
		render.JSON(w, r, src)
		return
	case "searchResponse":
		res, err := fr.Do(r.Context(), index.ESClient())
		if err != nil {
			render.Error(w, r, err, &render.ErrorConfig{
				StatusCode: http.StatusBadRequest,
				Message:    "Unable to dump search response",
			})
			return
		}
		render.JSON(w, r, res)
		return
	case "request":
		dump, dumpErr := httputil.DumpRequest(r, true)
		if dumpErr != nil {
			render.Error(w, r, dumpErr, &render.ErrorConfig{
				StatusCode: http.StatusBadRequest,
				Message:    "Unable to dump request",
			})
			return
		}

		render.PlainText(w, r, string(dump))
		return
	}

	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
			Message:    "No fragments for query were found.",
		})
		return
	}

	// if len(frags) == 0 {
	// log.Printf("Unable to list fragments because of: %#v", err)
	// render.JSON(w, r, APIErrorMessage{
	// HTTPStatus: http.StatusNotFound,
	// Message:    fmt.Sprint("No fragments for query were found."),
	// Error:      err,
	// })
	// return
	// }

	w.Header().Add("FRAG_COUNT", strconv.Itoa(int(totalFrags)))

	// Add hyperMediaControls
	hmd := fragments.NewHyperMediaDataSet(r, totalFrags, fr)
	controls, err := hmd.CreateControls()
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			StatusCode: http.StatusNotFound,
			Message:    "unable to create media controls",
		})
		return
	}

	w.Header().Set("Content-Type", "text/turtle; charset=utf-8")

	w.Write(controls)

	for _, frag := range frags {
		fmt.Fprintln(w, frag.Triple)
	}
	render.Turtle(w, r, "")
}
