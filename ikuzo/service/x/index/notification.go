// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// NotificationConfig configures the webhook notification system
type NotificationConfig struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
	Timeout  int    `json:"timeout"` // timeout in seconds
	APIKey   string `json:"apiKey"`  // API key for Bearer token authentication
}

// IndexingNotification represents a notification about indexing status
type IndexingNotification struct {
	Type      string    `json:"type"` // "success", "warning", "error"
	OrgID     string    `json:"orgID"`
	DatasetID string    `json:"datasetID"`
	Revision  int       `json:"revision"`
	Timestamp time.Time `json:"timestamp"`

	// Success fields
	RecordsIndexed  int `json:"recordsIndexed,omitempty"`
	RecordsExpected int `json:"recordsExpected,omitempty"`
	OrphansDeleted  int `json:"orphansDeleted,omitempty"`

	// Error fields
	Errors []IndexingError `json:"errors,omitempty"`

	// Warning/Error message
	Message string `json:"message,omitempty"`
}

// IndexingError represents a single indexing error
type IndexingError struct {
	DocumentID string `json:"documentId"`
	ErrorType  string `json:"errorType"`
	Reason     string `json:"reason"`
}

// Notifier handles sending notifications to configured endpoints
type Notifier struct {
	config NotificationConfig
	client *http.Client
}

// NewNotifier creates a new Notifier with the given configuration
func NewNotifier(config NotificationConfig) *Notifier {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10
	}

	return &Notifier{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// Send sends a notification to the configured endpoint
func (n *Notifier) Send(ctx context.Context, notification *IndexingNotification) error {
	if n == nil || !n.config.Enabled || n.config.Endpoint == "" {
		return nil
	}

	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("unable to marshal notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set Authorization header if API key is configured
	if n.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+n.config.APIKey)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification endpoint returned error status: %d", resp.StatusCode)
	}

	log.Debug().
		Str("type", notification.Type).
		Str("datasetID", notification.DatasetID).
		Int("revision", notification.Revision).
		Msg("notification sent successfully")

	return nil
}

// SendSuccess sends a success notification
func (n *Notifier) SendSuccess(ctx context.Context, orgID, datasetID string, revision, recordsIndexed, recordsExpected, orphansDeleted int) {
	notification := &IndexingNotification{
		Type:            "success",
		OrgID:           orgID,
		DatasetID:       datasetID,
		Revision:        revision,
		Timestamp:       time.Now(),
		RecordsIndexed:  recordsIndexed,
		RecordsExpected: recordsExpected,
		OrphansDeleted:  orphansDeleted,
	}

	if err := n.Send(ctx, notification); err != nil {
		log.Error().Err(err).
			Str("datasetID", datasetID).
			Msg("failed to send success notification")
	}
}

// SendWarning sends a warning notification (e.g., verification failed)
func (n *Notifier) SendWarning(ctx context.Context, orgID, datasetID string, revision int, message string, recordsExpected, recordsActual int) {
	notification := &IndexingNotification{
		Type:            "warning",
		OrgID:           orgID,
		DatasetID:       datasetID,
		Revision:        revision,
		Timestamp:       time.Now(),
		Message:         message,
		RecordsExpected: recordsExpected,
		RecordsIndexed:  recordsActual,
	}

	if err := n.Send(ctx, notification); err != nil {
		log.Error().Err(err).
			Str("datasetID", datasetID).
			Msg("failed to send warning notification")
	}
}

// SendError sends an error notification with indexing errors
func (n *Notifier) SendError(ctx context.Context, orgID, datasetID string, revision int, message string, errors []IndexingError) {
	notification := &IndexingNotification{
		Type:      "error",
		OrgID:     orgID,
		DatasetID: datasetID,
		Revision:  revision,
		Timestamp: time.Now(),
		Message:   message,
		Errors:    errors,
	}

	if err := n.Send(ctx, notification); err != nil {
		log.Error().Err(err).
			Str("datasetID", datasetID).
			Msg("failed to send error notification")
	}
}
