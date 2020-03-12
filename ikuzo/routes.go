package ikuzo

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/ikuzo/internal/assets"
	mw "github.com/go-chi/chi/middleware"
)

// DefaultMiddleware are the default functions applied to the global router.
func DefaultMiddleware() []func(http.Handler) http.Handler {
	handlers := []func(http.Handler) http.Handler{
		mw.Heartbeat("/ping"),
		mw.StripSlashes,
	}

	return handlers
}

// routes are the default routes for ikuzo.
// These can be overwritten using SetRouters.
//
// All handlers should have lazy initialization, so when they are not called
// no connections should be initialized.
func (s *server) routes() {
	s.router.Get("/", s.handleIndex())

	s.fileServer("/static", assets.FileSystem)
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func (s *server) fileServer(path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		log.Error().
			Str("path", path).
			Msg("FileServer does not permit URL parameters. (fileserver: disabled)")

		return
	}

	fsServer := http.FileServer(root)

	fs := http.StripPrefix(path, fsServer)

	path += "*"

	s.router.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
