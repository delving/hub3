package imageproxy

import (
	"fmt"

	"github.com/go-chi/chi"
)

func (s *Service) Routes(pattern string, router chi.Router) {
	proxyPrefix := fmt.Sprintf("/%s/{options}/*", s.proxyPrefix)
	router.Get(proxyPrefix, s.proxyImage)
}
