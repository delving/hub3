package render

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// StatusCtxKey is a context key to record a future HTTP response status code.
var StatusCtxKey = &contextKey{"Status"}

// Status sets a HTTP response status code hint into request context at any point
// during the request life-cycle. Before the Responder sends its response header
// it will check the StatusCtxKey
func Status(r *http.Request, status int) {
	*r = *r.WithContext(context.WithValue(r.Context(), StatusCtxKey, status))
}

// PlainText writes a string to the response, setting the Content-Type as
// text/plain.
func PlainText(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write([]byte(v)) //nolint:errcheck
}

// Data writes raw bytes to the response, setting the Content-Type as
// application/octet-stream.
func Data(w http.ResponseWriter, r *http.Request, v []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write(v) //nolint:errcheck
}

// HTML writes a string to the response, setting the Content-Type as text/html.
func HTML(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write([]byte(v)) //nolint:errcheck
}

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json.
func JSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		Error(w, r, err, &ErrorConfig{
			StatusCode: http.StatusInternalServerError,
		})

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write(buf.Bytes()) //nolint:errcheck
}

// XML marshals 'v' to JSON, setting the Content-Type as application/xml. It
// will automatically prepend a generic XML header (see encoding/xml.Header) if
// one is not found in the first 100 bytes of 'v'.
func XML(w http.ResponseWriter, r *http.Request, v interface{}) {
	b, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	// Try to find <?xml header in first 100 bytes (just in case there're some XML comments).
	findHeaderUntil := len(b)
	if findHeaderUntil > 100 {
		findHeaderUntil = 100
	}

	if !bytes.Contains(b[:findHeaderUntil], []byte("<?xml")) {
		// No header found. Print it out first.
		w.Write([]byte(xml.Header)) //nolint:errcheck
	}

	w.Write(b) //nolint:errcheck
}

// NoContent returns a HTTP 204 "No Content" response.
func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Turtle writes a string to the response, setting the Content-Type as
// text/turtle.
func Turtle(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "text/turtle; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write([]byte(v)) //nolint:errcheck
}

// Ntriples writes a string to the response, setting the Content-Type as
// application/n-triples.
func NTriples(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "application/n-triples; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write([]byte(v)) //nolint:errcheck
}

// JSONLD writes a string to the response, setting the Content-Type as
// application/ld+json.
func JSONLD(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "application/ld+json; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write([]byte(v)) //nolint:errcheck
}

// RDFXML writes a string to the response, setting the Content-Type as
// application/rdf+xml.
func RDFXML(w http.ResponseWriter, r *http.Request, v string) {
	w.Header().Set("Content-Type", "application/rdf+xml; charset=utf-8")

	if status, ok := r.Context().Value(StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}

	w.Write([]byte(v)) //nolint:errcheck
}
