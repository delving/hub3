package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/labstack/gommon/log"
)

func gafApeProxy(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Method)
	//var query string
	//switch r.Method {
	//case http.MethodGet:
	//query = r.URL.Query().Get("query")
	//case http.MethodPost:
	//query = r.FormValue("query")
	//}

	//if query == "" {
	//render.Status(r, http.StatusBadRequest)
	//render.JSON(w, r, &ErrorMessage{"Bad Request", "a value in the query param is required."})
	//return
	//}
	//log.Info(query)
	resp, statusCode, contentType, err := runGafApeQuery(r)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	render.Status(r, statusCode)
	return
}

// runGafApeQuery sends a gaf query to the gaf-endpoint specified in the configuration
func runGafApeQuery(r *http.Request) (body []byte, statusCode int, contentType string, err error) {
	gafBaseURL := "https://acpt.nationaalarchief.nl"
	fullPath := fmt.Sprintf("%s%s", gafBaseURL, r.URL.Path)
	log.Printf("path %#v", fullPath)
	req, err := http.NewRequest("POST", fullPath, r.Body)
	if err != nil {
		log.Errorf("Unable to create gaf request %s", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	//q := req.URL.Query()
	//q.Add("query", query)
	//req.URL.RawQuery = q.Encode()

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Do(req)
	if err != nil {
		log.Errorf("Error in gaf query: %s", err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("Unable to read the response body with error: %s", err)
	}
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
