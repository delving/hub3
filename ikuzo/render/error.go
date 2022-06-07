package render

import (
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// DefaultConfig is a package-level variable set to our default Logger. We do this
// because it allows you to set your own logger when no logger is supplied with
// the ErrorConfig.
var DefaultConfig = ErrorConfig{
	Log:        &log.Logger,
	StatusCode: http.StatusInternalServerError,
	Message:    "unable to handle request",
}

type ErrorConfig struct {
	Log           *zerolog.Logger
	StatusCode    int
	Message       string
	DatasetID     string
	OrgID         string
	PreventBubble bool // prevent message from being logged or sent to sentry
}

func (ec *ErrorConfig) deepCopy() *ErrorConfig {
	return &ErrorConfig{
		Log:           ec.Log,
		StatusCode:    ec.StatusCode,
		Message:       ec.Message,
		DatasetID:     ec.DatasetID,
		OrgID:         ec.OrgID,
		PreventBubble: ec.PreventBubble,
	}
}

type errorResponse struct {
	Status      string `json:"status"`
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

func Error(w http.ResponseWriter, r *http.Request, err error, cfg *ErrorConfig) {
	if cfg == nil {
		cfg = DefaultConfig.deepCopy()
	}

	if cfg.StatusCode == http.StatusNotFound {
		cfg.PreventBubble = true
	}

	msg := "unable to handle request"
	if cfg.Message != "" {
		msg = cfg.Message
	}

	if cfg.OrgID == "" {
		orgID := domain.GetOrganizationID(r)
		cfg.OrgID = orgID.String()
	}

	requestID, _ := hlog.IDFromRequest(r)

	if !cfg.PreventBubble && cfg.Log != nil {
		l := cfg.Log.Error().Err(err)
		if cfg.DatasetID != "" {
			l = l.Str("dataset_id", cfg.DatasetID)
		}

		l.Str("org_id", cfg.OrgID).Str("req_id", requestID.String()).Msg(msg)
	}

	hub := sentry.GetHubFromContext(r.Context())
	if hub != nil && !cfg.PreventBubble {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetContext("Hub3", map[string]interface{}{
				"orgID":     cfg.OrgID,
				"datasetID": cfg.DatasetID,
				"requestID": requestID,
			})
		})
		hub.CaptureException(err)
	}

	Status(r, cfg.StatusCode)

	resp := errorResponse{
		Status:      http.StatusText(cfg.StatusCode),
		Code:        cfg.StatusCode,
		Message:     err.Error(),
		Description: cfg.Message,
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	JSON(w, r, resp)
}
