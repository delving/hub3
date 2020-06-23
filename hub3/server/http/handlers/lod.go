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
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterLOD(r chi.Router) {
	if c.Config.LOD.SingleEndpoint != "" {
		r.Get(fmt.Sprintf("/{path:%s}/*", c.Config.LOD.SingleEndpoint), RenderLODResource)
	} else {
		r.Get(fmt.Sprintf("/{path:%s}/*", c.Config.LOD.RDF), RenderLODResource)
		r.Get(fmt.Sprintf("/{path:%s}/*", c.Config.LOD.Resource), RenderLODResource)
		r.Get(
			//fmt.Sprintf("/{path:%s}/*", config.Config.LOD.HTML), RenderLODResource)
			fmt.Sprintf("/{path:%s}/*", c.Config.LOD.HTML), func(w http.ResponseWriter, r *http.Request) {
				render.PlainText(w, r, `{"type": "rdf html endpoint"}`)
				return
			})
	}
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
			log.Printf("Redirecting to %s", lodKey)
			http.Redirect(w, r, lodKey, 302)
			return
		}
		lodKey = strings.TrimSuffix(lodKey, ".nt")

	}

	fr := fragments.NewFragmentRequest()
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
	var buffer bytes.Buffer
	for _, frag := range frags {
		buffer.WriteString(fmt.Sprintln(frag.Triple))
	}
	w.Write(buffer.Bytes())
	return
}
