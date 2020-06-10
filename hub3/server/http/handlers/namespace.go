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
	"net/http"

	c "github.com/delving/hub3/config"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterNamespace(router chi.Router) {
	router.Get("/api/namespaces", listNameSpaces)
}

// listNameSpaces list all currently defined NameSpace object
func listNameSpaces(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, c.Config.NameSpaceMap.ByPrefix())
	//render.JSON(w, r, c.Config.NameSpaces)
	return
}
