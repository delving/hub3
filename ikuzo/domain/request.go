package domain

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
)

func URLParam(r *http.Request, key string) string {
	return sanitizeParam(chi.URLParam(r, key))
}

func sanitizeParam(param string) string {
	param = filepath.Base(param)

	if param == "." {
		return ""
	}

	return param
}
