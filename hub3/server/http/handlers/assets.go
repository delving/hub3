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
	"io/ioutil"
	"log"
	"net/http"

	"github.com/delving/hub3/hub3/server/http/assets"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterStaticAssets(r chi.Router) {
	// use to register static asset routes when applicable
}

func serveHTML(w http.ResponseWriter, r *http.Request, filePath string) error {
	file, err := assets.FileSystem.Open(filePath)
	if err != nil {
		log.Printf("Unable to open file %s: %v", filePath, err)
		render.Status(r, http.StatusNotFound)
		render.PlainText(w, r, "")

		return err
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Unable to read file %s: %v", filePath, err)
		render.Status(r, http.StatusNotFound)
		render.PlainText(w, r, "")

		return err
	}

	render.HTML(w, r, string(body))

	return nil
}
