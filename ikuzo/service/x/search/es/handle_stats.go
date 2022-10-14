package es

import (
	"fmt"
	"log"
	"net/http"

	"github.com/delving/hub3/ikuzo/render"
)

func (s *Service) searchLabelStats(w http.ResponseWriter, r *http.Request) {
	res, err := s.client.GetResourceEntryStats("searchLabel", r)
	if err != nil {
		render.Error(w, r, err, &render.ErrorConfig{
			Log:        &s.log,
			StatusCode: http.StatusBadRequest,
		})

		log.Print("Unable to get statistics for searchLabels")
		render.PlainText(w, r, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits.Value)
	render.JSON(w, r, res)
}

func (s *Service) predicateStats(w http.ResponseWriter, r *http.Request) {
	res, err := s.client.GetResourceEntryStats("predicate", r)
	if err != nil {
		log.Print("Unable to get statistics for predicate")
		render.PlainText(w, r, err.Error())
		render.Status(r, http.StatusBadRequest)
		return
	}
	fmt.Printf("total hits: %d\n", res.Hits.TotalHits.Value)
	render.JSON(w, r, res)
}
