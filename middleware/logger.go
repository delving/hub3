package middleware

import (
	"net/http"
	"net/url"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// RequestLogger creates a middleware chain for request logging
func RequestLogger(log *zerolog.Logger) func(next http.Handler) http.Handler {
	c := alice.New()

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(*log))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to those handler, all our logs will come with some pre-populated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Dict("params", LogParamsAsDict(r.URL.Query())).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	// Here is your final handler
	return c.Then
}

// LogParamsAsDict logs the request params as a zerolog.Dict.
func LogParamsAsDict(params url.Values) *zerolog.Event {
	dict := zerolog.Dict()

	for key, values := range params {
		arr := zerolog.Arr()

		var nonEmpty bool

		for _, v := range values {
			if v != "" {
				arr = arr.Str(v)

				if !nonEmpty {
					nonEmpty = true
				}
			}
		}

		if nonEmpty {
			dict = dict.Array(key, arr)
		}
	}

	return dict
}
