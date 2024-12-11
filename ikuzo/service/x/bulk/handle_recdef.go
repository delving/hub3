package bulk

import (
	"net/http"

	"github.com/delving/hub3/ikuzo/render"
)

func (s *Service) handleRecDefs(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, s.recdef.List())
}
