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

package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	c "github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/mediamanager"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

// WebResourceAPIResource is the router struct for webresource data
type WebResourceAPIResource struct{}

// Routes returns the chi.Router
func (wra WebResourceAPIResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{urn}*", listWebResource)
	return r
}

func listWebResource(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, "urn")
	if strings.HasSuffix(urn, "__") {
		path := filepath.Join(c.Config.WebResource.WebResourceDir, strings.TrimPrefix(urn, "urn:"))
		log.Printf(path)
		matches, err := filepath.Glob(fmt.Sprintf("%s*", path))
		if err != nil {
			log.Printf("%v", err)
		}
		log.Printf("matches: %s", matches)
	}
	log.Printf("urn: %s", urn)
	render.JSON(w, r, `{"type": "thumbnail"}`)
	return
}

// ThumbnailResource is the router struct for Thumbnail links
type ThumbnailResource struct{}

// Routes returns the chi.Router
func (rs ThumbnailResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{orgId}/{spec}/{localId}/{size}", renderThumbnail)
	return r
}

// DeepZoomResource is the router struct for DeepZoom paths
type DeepZoomResource struct{}

// Routes returns the chi.Router
func (rs DeepZoomResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{orgId}/{spec}/{localId}.tif.dzi", renderDeepZoom)
	r.Get("/{orgId}/{spec}/{localId}.dzi", renderDeepZoom)
	r.Get("/{orgId}/{spec}/{localId}_files/{level}/{col}_{row}.{tile_format}", renderDeepZoomTiles)
	r.Get("/{orgId}/{spec}/{localId}_.tif.files/{level}/{col}_{row}.{tile_format}", renderDeepZoomTiles)
	return r
}

// ExploreResource is the router struct for DeepZoom paths
type ExploreResource struct{}

// Routes returns the chi.Router
func (rs ExploreResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, `{"type": "explore"}`)
		return
	})
	r.Get("/index", func(w http.ResponseWriter, r *http.Request) {
		err := mediamanager.IndexWebResources(bp)
		if err != nil {
			log.Printf("Unable to index webresources: %s", err)
		}
		return
	})
	return r
}

func renderThumbnail(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, `{"type": "thumbnail"}`)
	return
}

func renderDeepZoom(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, `{"type": "deepzoom tiles"}`)
	return
}

func renderDeepZoomTiles(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, `{"type": "deepzoom tiles"}`)
	return
}
