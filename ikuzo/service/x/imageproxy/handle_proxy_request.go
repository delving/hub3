package imageproxy

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// create handler fuction to serve the proxied images
func (s *Service) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	targetURL := chi.URLParam(r, "*")
	if strings.HasPrefix(targetURL, "http") {
		if !strings.HasPrefix(targetURL, "https://") && !strings.HasPrefix(targetURL, "http://") {
			targetURL = strings.ReplaceAll(targetURL, ":/", "://")
		}
	}

	options := chi.URLParam(r, "options")

	allowed, err := s.domainAllowed(targetURL)
	if err != nil {
		s.m.IncError()
		s.log.Error().Err(err).Str("url", targetURL).Msg("unable to check allowed domains")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if !allowed {
		s.m.IncRejectDomain()
		s.log.Error().Err(err).Str("url", targetURL).Msg("domain not allowed")
		http.Error(w, "domain is not allowed", http.StatusForbidden)

		return
	}

	allowed = s.reffererAllowed(r.Referer())
	if !allowed {
		s.m.IncRejectReferrer()
		s.log.Error().Err(err).Str("url", targetURL).Str("referrer", html.EscapeString(r.Referer())).Msg("domain not allowed")
		http.Error(w, fmt.Sprintf("referrer not allowed: %s", html.EscapeString(r.Referer())), http.StatusForbidden)

		return
	}

	req, err := NewRequest(
		targetURL,
		SetRawQueryString(r.URL.RawQuery),
		SetTransform(options),
		SetService(s),
		SetEnableTransform(s.enableResize),
	)
	if err != nil {
		s.log.Error().Err(err).Str("url", targetURL).Msg("unable to create proxy request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.m.IncError()

		return
	}

	switch req.TransformOptions {
	case "explain":
		explain := fmt.Sprintf("%s => %s", req.SourceURL, req.downloadedSourcePath())
		render.PlainText(w, r, explain)

		return
	case "metrics":
		render.JSON(w, r, s.m)
		return
	case "remove", "recache":
		removeErr := req.Remove()
		if removeErr != nil {
			s.log.Error().Err(removeErr).Str("url", targetURL).Msgf("unable to removed cached items; %s", removeErr)
			http.Error(w, removeErr.Error(), http.StatusInternalServerError)

			return
		}

		s.m.IncRemoved()

		if req.TransformOptions != "recache" {
			removed := fmt.Sprintf("Removed: %s => %s", req.SourceURL, req.downloadedSourcePath())
			render.PlainText(w, r, removed)

			return
		}

		req.TransformOptions = "raw"
	case "request":
		render.JSON(w, r, req)
		return
	}

	if len(s.refuselist) != 0 {
		for _, uri := range s.refuselist {
			if strings.Contains(req.SourceURL, uri) {
				http.Error(w, "not found", http.StatusNotFound)
				s.m.IncRejectURI()

				return
			}
		}
	}

	var buf bytes.Buffer

	err = s.Do(r.Context(), req, &buf)
	if err != nil {
		s.log.Error().Err(err).Str("url", req.SourceURL).Msg("unable to make proxy request")

		if s.defaultImagePath != "" {
			data, errRead := os.ReadFile(s.defaultImagePath)
			if errRead != nil {
				s.log.Error().Err(err).Str("url", req.SourceURL).Msg("could not read default image")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, errFprint := fmt.Fprintf(&buf, "%s", data)
			if errFprint != nil {
				s.log.Error().Err(err).Str("url", req.SourceURL).Msg("could not add data to buffer??")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if errors.Is(err, ErrRemoteResourceNotFound) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			io.Copy(w, &buf)
			return
		}

		if errors.Is(err, ErrRemoteResourceNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Cache-Control", "public,max-age=259200")
	r.Header.Set("Cache-Type", string(req.CacheType))
	r.Header.Set("Cache-Url", req.SourceURL)

	if _, err := io.Copy(w, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
