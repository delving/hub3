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

package bulk

import "fmt"

// BulkError contains detailed information about a bulk processing failure.
// It preserves the original error detail while providing context about which
// record failed and at what stage.
type BulkError struct {
	HubID     string `json:"hubId,omitempty"`
	OrgID     string `json:"orgId,omitempty"`
	DatasetID string `json:"datasetId,omitempty"`
	LocalID   string `json:"localId,omitempty"`
	Stage     string `json:"stage"`             // Processing stage: "validation", "rdf_parse", "indexing", etc.
	Message   string `json:"message"`           // Human-readable error message
	Detail    string `json:"detail,omitempty"`  // Original error detail (e.g., "invalid language tag: unexpected character: ' '")
	cause     error  // Underlying error (not serialized)
}

// Error implements the error interface.
func (e *BulkError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Stage, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Stage, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *BulkError) Unwrap() error {
	return e.cause
}

// NewBulkError creates a new BulkError with the given parameters.
func NewBulkError(req *Request, stage, message string, cause error) *BulkError {
	be := &BulkError{
		Stage:   stage,
		Message: message,
		cause:   cause,
	}

	if req != nil {
		be.HubID = req.HubID
		be.OrgID = req.OrgID
		be.DatasetID = req.DatasetID
		be.LocalID = req.LocalID
	}

	if cause != nil {
		be.Detail = cause.Error()
	}

	return be
}

// ErrorResponse is the structured JSON response returned when bulk processing fails.
type ErrorResponse struct {
	Status    string `json:"status"`
	Error     string `json:"error"`
	Detail    string `json:"detail,omitempty"`
	HubID     string `json:"hubId,omitempty"`
	OrgID     string `json:"orgId,omitempty"`
	DatasetID string `json:"datasetId,omitempty"`
	Stage     string `json:"stage,omitempty"`
	Stats     *Stats `json:"stats,omitempty"`
}

// NewErrorResponse creates an ErrorResponse from an error.
// If the error is a BulkError, it extracts the detailed information.
func NewErrorResponse(err error, stats *Stats) *ErrorResponse {
	resp := &ErrorResponse{
		Status: "error",
		Error:  err.Error(),
		Stats:  stats,
	}

	// Extract details from BulkError if available
	if bulkErr, ok := err.(*BulkError); ok {
		resp.Detail = bulkErr.Detail
		resp.HubID = bulkErr.HubID
		resp.OrgID = bulkErr.OrgID
		resp.DatasetID = bulkErr.DatasetID
		resp.Stage = bulkErr.Stage
		// Use the cleaner message without the detail suffix
		resp.Error = fmt.Sprintf("%s: %s", bulkErr.Stage, bulkErr.Message)
	}

	return resp
}
