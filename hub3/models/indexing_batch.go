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

package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/asdine/storm/q"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// BatchStatus represents the lifecycle state of an indexing batch
type BatchStatus string

const (
	BatchStatusPending   BatchStatus = "pending"
	BatchStatusIndexing  BatchStatus = "indexing"
	BatchStatusVerifying BatchStatus = "verifying"
	BatchStatusCompleted BatchStatus = "completed"
	BatchStatusFailed    BatchStatus = "failed"
)

// IndexingError represents a single indexing error
type IndexingError struct {
	DocumentID string `json:"documentId"`
	ErrorType  string `json:"errorType"`
	Reason     string `json:"reason"`
}

// IndexingBatch tracks the state of an indexing session
// This provides persistent tracking of bulk indexing operations, allowing:
// - Progress visibility during indexing
// - Recovery after server restart
// - Verification that all records were indexed
// - Error aggregation and debugging
type IndexingBatch struct {
	ID        string `storm:"id"` // Unique batch ID (UUID)
	OrgID     string `storm:"index"`
	DatasetID string `storm:"index"`
	Revision  int    `storm:"index"`

	// State
	Status      BatchStatus `json:"status"`
	StartedAt   time.Time   `json:"startedAt"`
	CompletedAt time.Time   `json:"completedAt,omitempty"`

	// Counts (use atomic operations when updating)
	ExpectedCount  int `json:"expectedCount"`  // Set when client knows upfront
	SubmittedCount int `json:"submittedCount"` // Incremented as records sent
	IndexedCount   int `json:"indexedCount"`   // Confirmed in ES
	FailedCount    int `json:"failedCount"`

	// Errors (limited to prevent unbounded growth)
	Errors    []IndexingError `json:"errors,omitempty"`
	MaxErrors int             `json:"maxErrors,omitempty"` // Max errors to store (default 100)

	// Verification
	Verified   bool      `json:"verified"`
	VerifiedAt time.Time `json:"verifiedAt,omitempty"`

	// Internal: mutex for thread-safe updates
	mu sync.RWMutex `storm:"-" json:"-"`
}

// CreateBatch creates a new IndexingBatch and persists it to BoltDB
func CreateBatch(orgID, datasetID string, revision int, expectedCount int) (*IndexingBatch, error) {
	batch := &IndexingBatch{
		ID:            uuid.New().String(),
		OrgID:         orgID,
		DatasetID:     datasetID,
		Revision:      revision,
		Status:        BatchStatusIndexing,
		StartedAt:     time.Now(),
		ExpectedCount: expectedCount,
		MaxErrors:     100, // Default max errors to store
	}

	if err := ORM().Save(batch); err != nil {
		return nil, fmt.Errorf("unable to save indexing batch: %w", err)
	}

	log.Debug().
		Str("batchID", batch.ID).
		Str("datasetID", datasetID).
		Int("revision", revision).
		Int("expectedCount", expectedCount).
		Msg("created indexing batch")

	return batch, nil
}

// GetBatch retrieves an IndexingBatch by its ID
func GetBatch(batchID string) (*IndexingBatch, error) {
	var batch IndexingBatch
	if err := ORM().One("ID", batchID, &batch); err != nil {
		return nil, fmt.Errorf("unable to find batch %s: %w", batchID, err)
	}
	return &batch, nil
}

// GetActiveBatch retrieves the currently active (indexing) batch for a dataset
func GetActiveBatch(orgID, datasetID string) (*IndexingBatch, error) {
	var batches []IndexingBatch

	// Find batches that are either indexing or verifying
	query := ORM().Select(
		q.And(
			q.Eq("OrgID", orgID),
			q.Eq("DatasetID", datasetID),
			q.Or(
				q.Eq("Status", BatchStatusIndexing),
				q.Eq("Status", BatchStatusVerifying),
			),
		),
	)

	if err := query.Find(&batches); err != nil {
		return nil, fmt.Errorf("unable to query active batches: %w", err)
	}

	if len(batches) == 0 {
		return nil, nil
	}

	// Return the most recent one if multiple exist
	latest := &batches[0]
	for i := range batches {
		if batches[i].StartedAt.After(latest.StartedAt) {
			latest = &batches[i]
		}
	}

	return latest, nil
}

// GetBatchByRevision retrieves a batch by dataset and revision
func GetBatchByRevision(orgID, datasetID string, revision int) (*IndexingBatch, error) {
	var batches []IndexingBatch

	query := ORM().Select(
		q.And(
			q.Eq("OrgID", orgID),
			q.Eq("DatasetID", datasetID),
			q.Eq("Revision", revision),
		),
	)

	if err := query.Find(&batches); err != nil {
		return nil, fmt.Errorf("unable to query batches by revision: %w", err)
	}

	if len(batches) == 0 {
		return nil, nil
	}

	// Return the most recent one
	latest := &batches[0]
	for i := range batches {
		if batches[i].StartedAt.After(latest.StartedAt) {
			latest = &batches[i]
		}
	}

	return latest, nil
}

// IncrementSubmitted increments the submitted count (thread-safe)
func (b *IndexingBatch) IncrementSubmitted(count int) error {
	b.mu.Lock()
	b.SubmittedCount += count
	b.mu.Unlock()

	// Persist periodically (every 100 records to reduce I/O)
	if b.SubmittedCount%100 == 0 {
		return b.save()
	}
	return nil
}

// IncrementIndexed increments the indexed count (thread-safe)
func (b *IndexingBatch) IncrementIndexed(count int) error {
	b.mu.Lock()
	b.IndexedCount += count
	b.mu.Unlock()

	// Persist periodically
	if b.IndexedCount%100 == 0 {
		return b.save()
	}
	return nil
}

// IncrementFailed increments the failed count and optionally stores the error
func (b *IndexingBatch) IncrementFailed(err IndexingError) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.FailedCount++

	// Store error if under limit
	if b.MaxErrors == 0 {
		b.MaxErrors = 100
	}
	if len(b.Errors) < b.MaxErrors {
		b.Errors = append(b.Errors, err)
	}

	// Persist on every error (errors are important)
	return b.save()
}

// SetStatus updates the batch status
func (b *IndexingBatch) SetStatus(status BatchStatus) error {
	b.mu.Lock()
	b.Status = status
	if status == BatchStatusCompleted || status == BatchStatusFailed {
		b.CompletedAt = time.Now()
	}
	b.mu.Unlock()

	return b.save()
}

// Complete marks the batch as completed
func (b *IndexingBatch) Complete() error {
	return b.SetStatus(BatchStatusCompleted)
}

// Fail marks the batch as failed with an optional reason
func (b *IndexingBatch) Fail(reason string) error {
	b.mu.Lock()
	b.Status = BatchStatusFailed
	b.CompletedAt = time.Now()
	if reason != "" && len(b.Errors) < b.MaxErrors {
		b.Errors = append(b.Errors, IndexingError{
			ErrorType: "batch_failure",
			Reason:    reason,
		})
	}
	b.mu.Unlock()

	return b.save()
}

// SetVerified marks the batch as verified
func (b *IndexingBatch) SetVerified(verified bool) error {
	b.mu.Lock()
	b.Verified = verified
	b.VerifiedAt = time.Now()
	b.mu.Unlock()

	return b.save()
}

// StartVerifying transitions the batch to verifying state
func (b *IndexingBatch) StartVerifying() error {
	return b.SetStatus(BatchStatusVerifying)
}

// Progress returns the current progress as a percentage (0-100)
func (b *IndexingBatch) Progress() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := b.ExpectedCount
	if total == 0 {
		total = b.SubmittedCount
	}
	if total == 0 {
		return 0
	}

	return float64(b.IndexedCount+b.FailedCount) / float64(total) * 100
}

// IsComplete returns true if all submitted records have been processed
func (b *IndexingBatch) IsComplete() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	processed := b.IndexedCount + b.FailedCount
	return processed >= b.SubmittedCount && b.SubmittedCount > 0
}

// save persists the batch to BoltDB
func (b *IndexingBatch) save() error {
	if err := ORM().Save(b); err != nil {
		log.Error().Err(err).Str("batchID", b.ID).Msg("failed to save batch")
		return fmt.Errorf("unable to save batch: %w", err)
	}
	return nil
}

// Flush saves any pending changes immediately
func (b *IndexingBatch) Flush() error {
	return b.save()
}

// CleanupOldBatches removes batches older than the specified duration
func CleanupOldBatches(maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)

	var batches []IndexingBatch
	if err := ORM().All(&batches); err != nil {
		return 0, fmt.Errorf("unable to query batches: %w", err)
	}

	deleted := 0
	for _, batch := range batches {
		// Only delete completed/failed batches older than cutoff
		if (batch.Status == BatchStatusCompleted || batch.Status == BatchStatusFailed) &&
			batch.CompletedAt.Before(cutoff) {
			if err := ORM().DeleteStruct(&batch); err != nil {
				log.Warn().Err(err).Str("batchID", batch.ID).Msg("failed to delete old batch")
				continue
			}
			deleted++
		}
	}

	if deleted > 0 {
		log.Info().Int("deleted", deleted).Dur("maxAge", maxAge).Msg("cleaned up old batches")
	}

	return deleted, nil
}
