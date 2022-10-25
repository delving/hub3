// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"net/url"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

const (
	idPrefix       = "id"
	resourcePrefix = "resource"
	docPrefix      = "doc"
	dataPrefix     = "data"
	defPrefix      = "def"
)

func RegisterLOD(r chi.Router) {
	redirects := []string{
		idPrefix, resourcePrefix, docPrefix, dataPrefix, defPrefix,
	}

	for _, prefix := range redirects {
		r.Get(fmt.Sprintf("/%s/*", prefix), lodRedirect)
	}

	resolver := sparqlLodResolver
	if strings.EqualFold(c.Config.LOD.Store, "fragments") {
		resolver = fragmentsLodResolver
	}

	r.Get("/resource", resolver())
}

func rewriteLodPrefixes(path string) string {
	parts := strings.Split(path, "/")
	switch parts[1] {
	case idPrefix:
		parts[1] = docPrefix
	case resourcePrefix:
		parts[1] = dataPrefix
	}

	return strings.Join(parts, "/")
}

func getResolveURL(r *http.Request) string {
	path := r.URL.Path
	hostDomain := c.Config.RDF.BaseURL

	return fmt.Sprintf("%s%s", hostDomain, rewriteLodPrefixes(path))
}

func lodRedirect(w http.ResponseWriter, r *http.Request) {
	sourceURI := getResolveURL(r)
	resolveURI := fmt.Sprintf("/resource?uri=%s", sourceURI)
	http.Redirect(w, r, resolveURI, http.StatusFound)
}

func getSparqlSubject(iri, fragment string) (string, error) {
	uri, err := url.Parse(iri)
	if err != nil {
		return "", err
	}

	parts := strings.Split(uri.Path, "/")
	if len(parts) > 1 {
		switch parts[1] {
		case docPrefix:
			parts[1] = idPrefix
		case dataPrefix:
			parts[1] = resourcePrefix
		}
	}

	// # in a uri must be percent-encoded `%23` to be picked up by the resolver
	// otherwise it is stripped by the browser or the URL parser.
	path := strings.Join(parts, "/")
	if fragment != "" {
		path += "#" + fragment
	}

	if uri.Fragment != "" {
		path += "#" + uri.Fragment
	}

	if uri.Scheme == "" || uri.Host == "" {
		return "", fmt.Errorf("invalid uri or subject parameter: %q", uri)
	}

	return fmt.Sprintf("%s://%s%s", uri.Scheme, uri.Host, path), nil
}

func sparqlLodResolver() http.HandlerFunc {
	acceptedLodFormats := map[string]string{
		"turtle":    "text/turtle",
		"json-ld":   "application/ld+json",
		"n-triples": "application/n-triples",
		"n-quads":   "application/n-quads",
		"trig":      "application/trig",
		"rdfxml":    "application/rdf+xml",
	}

	acceptedMimeTypes := map[string]string{}

	for queryParam, mimetype := range acceptedLodFormats {
		acceptedMimeTypes[mimetype] = queryParam
	}

	formats := []string{}
	for _, f := range acceptedLodFormats {
		formats = append(formats, f)
	}

	minimalValidResponseSize := 6 // some  formats give back empty data but not zero bytes

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		if uri == "" {
			uri = r.URL.Query().Get("subject")
		}
		iri, err := getSparqlSubject(uri, r.URL.Fragment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		acceptMimeType := "text/turtle"

		if r.URL.Query().Has("format") {
			format := r.URL.Query().Get("format")
			mimetype, ok := acceptedLodFormats[format]
			if ok {
				acceptMimeType = mimetype
			}
		} else {
			accept := r.Header.Get("Accept")
			_, ok := acceptedMimeTypes[accept]
			if ok {
				acceptMimeType = accept
			}
		}

		orgID := domain.GetOrganizationID(r)

		var (
			resp        []byte
			statusCode  int
			contentType string
		)

		queries := []string{iri, uri}
		for _, q := range queries {
			query := fmt.Sprintf("describe <%s>", q)

			resp, statusCode, contentType, err = runSparqlQuery(orgID.String(), query, acceptMimeType)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.PlainText(w, r, string(resp))
				return
			}

			if len(resp) > minimalValidResponseSize {
				break
			}
		}

		if len(resp) < minimalValidResponseSize {
			statusCode = http.StatusNotFound
		}

		w.Header().Add("Accept", strings.Join(formats, ", "))
		w.Header().Set(contentTypeKey, contentType)
		w.WriteHeader(statusCode)

		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// RenderLODResource returns a list of matching fragments
// for a LOD resource. This mimicks a SPARQL describe request

func fragmentsLodResolver() http.HandlerFunc {
	// acceptedLodFormats := map[string]string{
	// "turtle": "text/turtle",
	// }
	// _ = acceptedLodFormats
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("uri")
		if uri == "" {
			uri = r.URL.Query().Get("subject")
		}

		orgID := domain.GetOrganizationID(r)
		fr := fragments.NewFragmentRequest(orgID.String())
		fr.Subject = []string{}

		if uri != "" {
			iri, err := getSparqlSubject(uri, r.URL.Fragment)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if uri != "" {
				fr.Subject = append(fr.Subject, uri)
			}

			if iri != uri {
				fr.Subject = append(fr.Subject, iri)
			}
		}

		if r.URL.Query().Has("graph") {
			fr.Graph = r.URL.Query().Get("graph")
		}

		if fr.Graph == "" && len(fr.Subject) == 0 {
			http.Error(w, fmt.Errorf("graph or uri parameter is required").Error(), http.StatusBadRequest)
			return
		}

		frags, _, err := fr.Find(r.Context(), index.ESClient())
		if err != nil || len(frags) == 0 {
			w.WriteHeader(http.StatusNotFound)

			if err != nil {
				log.Printf("Unable to list fragments because of: %s", err)
				return
			}

			log.Printf("Unable to find fragments")
			return
		}

		w.Header().Add("Accept", "text/turtle")
		w.Header().Set("Content-Type", "text/turtle")
		for _, frag := range frags {
			fmt.Fprintln(w, frag.Triple)
		}
	})
}
