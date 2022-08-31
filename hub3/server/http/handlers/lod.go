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

var lodPathRoute = "/{path:%s}/*"

func RegisterLOD(r chi.Router) {
	if c.Config.LOD.SingleEndpoint != "" {
		r.Get(fmt.Sprintf(lodPathRoute, c.Config.LOD.SingleEndpoint), RenderLODResource)
	} else {
		r.Get(fmt.Sprintf(lodPathRoute, c.Config.LOD.RDF), RenderLODResource)
		r.Get(fmt.Sprintf(lodPathRoute, c.Config.LOD.Resource), RenderLODResource)
		r.Get(
			// fmt.Sprintf(lodPathRoute, config.Config.LOD.HTML), RenderLODResource)
			fmt.Sprintf(lodPathRoute, c.Config.LOD.HTML), func(w http.ResponseWriter, r *http.Request) {
				render.PlainText(w, r, `{"type": "rdf html endpoint"}`)
			})
	}

	redirects := []string{
		idPrefix, resourcePrefix, docPrefix, dataPrefix, defPrefix,
	}

	for _, prefix := range redirects {
		r.Get(fmt.Sprintf("/%s/*", prefix), lodRedirect)
	}

	r.Get("/resource", lodResolver())
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

func lodResolver() http.HandlerFunc {
	acceptedLodFormats := map[string]string{
		"turtle":    "text/turtle",
		"json-ld":   "application/ld+json",
		"n-triples": "application/n-triples",
		"n-quads":   "application/n-quads",
		"trig":      "application/trig",
		"rdfxml":    "application/rdf+xml",
	}

	acceptHeaders := []string{}
	acceptedMimeTypes := map[string]string{}

	for queryParam, mimetype := range acceptedLodFormats {
		acceptHeaders = append(acceptHeaders, mimetype)
		acceptedMimeTypes[mimetype] = queryParam
	}

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
		query := fmt.Sprintf("describe <%s>", iri)

		resp, statusCode, contentType, err := runSparqlQuery(orgID.String(), query, acceptMimeType)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.PlainText(w, r, string(resp))
			return
		}
		w.Header().Set(contentTypeKey, contentType)

		formats := []string{}
		for _, f := range acceptedLodFormats {
			formats = append(formats, f)
		}

		w.Header().Add("Accept", strings.Join(formats, ", "))

		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO(kiivihal): add support for 404 not found. Now always 200 is returned
		render.Status(r, statusCode)
	})
}

// RenderLODResource returns a list of matching fragments
// for a LOD resource. This mimicks a SPARQL describe request
func RenderLODResource(w http.ResponseWriter, r *http.Request) {
	lodKey := r.URL.Path

	if c.Config.LOD.SingleEndpoint == "" {
		resourcePrefix := fmt.Sprintf("/%s", c.Config.LOD.Resource)
		if strings.HasPrefix(lodKey, resourcePrefix) {
			// todo for  now only support  RDF data
			lodKey = strings.Replace(lodKey, c.Config.LOD.Resource, c.Config.LOD.RDF, 1)
			http.Redirect(w, r, lodKey, 302)
			return
		}

		lodKey = strings.Replace(lodKey, c.Config.LOD.RDF, c.Config.LOD.Resource, 1)
		lodKey = strings.Replace(lodKey, c.Config.LOD.HTML, c.Config.LOD.Resource, 1)
	} else {
		// for now only support nt as format
		if !strings.HasSuffix(lodKey, ".nt") {
			lodKey = fmt.Sprintf("%s.nt", strings.TrimSuffix(lodKey, "/"))
			log.Printf("Redirecting to %s", domain.LogUserInput(lodKey))
			http.Redirect(w, r, lodKey, 302)
			return
		}
		lodKey = strings.TrimSuffix(lodKey, ".nt")
	}

	orgID := domain.GetOrganizationID(r)

	fr := fragments.NewFragmentRequest(orgID.String())
	fr.LodKey = lodKey
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

	w.Header().Set("Content-Type", "text/n-triples")
	for _, frag := range frags {
		fmt.Fprintln(w, frag.Triple)
	}

	return
}
