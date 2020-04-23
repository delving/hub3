package handlers

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/delving/hub3/hub3/server/http/assets"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func RegisterStaticAssets(r chi.Router) {}

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
