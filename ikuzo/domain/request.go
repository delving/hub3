package domain

import (
	"html"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

func URLParam(r *http.Request, key string) string {
	return SanitizeParam(chi.URLParam(r, key))
}

func SanitizeParam(param string) string {
	param = filepath.Base(param)

	if param == "." {
		return ""
	}

	return param
}

// LogUserInput properly escapes user input from query or URL parameters for logging
//
// The goal is to prevent XSS and other attacks via user-input in the logging
func LogUserInput(text string) string {
	return html.EscapeString(
		strings.Replace(
			strings.Replace(text, "\r", "", -1),
			"\n",
			"",
			-1,
		),
	)
}
