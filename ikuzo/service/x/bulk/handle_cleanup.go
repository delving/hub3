package bulk

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
)

// CleanupResponse contains the result of an orphan cleanup operation
type CleanupResponse struct {
	Status    string `json:"status"`
	OrgID     string `json:"orgID"`
	DatasetID string `json:"datasetID"`
	Revision  int    `json:"revision"`
	Message   string `json:"message,omitempty"`
}

// DatasetCleanupResult contains the result of cleaning up a single dataset
type DatasetCleanupResult struct {
	Spec            string `json:"spec"`
	Revision        int    `json:"revision"`
	OrphanRevisions int    `json:"orphanRevisions"`
	Status          string `json:"status"`
	Error           string `json:"error,omitempty"`
}

// GlobalCleanupResponse contains the result of a global orphan cleanup operation
type GlobalCleanupResponse struct {
	Status         string                 `json:"status"`
	OrgID          string                 `json:"orgID"`
	TotalDatasets  int                    `json:"totalDatasets"`
	WithOrphans    int                    `json:"withOrphans"`
	CleanedUp      int                    `json:"cleanedUp"`
	Errors         int                    `json:"errors"`
	Skipped        int                    `json:"skipped"`
	Results        []DatasetCleanupResult `json:"results"`
}

// handleCleanupOrphans handles POST /api/admin/cleanup-orphans/{orgID}/{datasetID}
// This endpoint manually triggers orphan deletion for a dataset, bypassing the
// normal verification checks. Use this to clean up orphaned records that persist
// due to verification failures or missing tags.
func (s *Service) handleCleanupOrphans(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	datasetID := chi.URLParam(r, "datasetID")

	if orgID == "" || datasetID == "" {
		http.Error(w, "orgID and datasetID are required", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("orgID", orgID).
		Str("datasetID", datasetID).
		Msg("manual orphan cleanup requested")

	// Get the dataset
	ds, err := models.GetDataSet(orgID, datasetID)
	if err != nil {
		log.Error().Err(err).
			Str("orgID", orgID).
			Str("datasetID", datasetID).
			Msg("unable to find dataset for orphan cleanup")
		http.Error(w, "dataset not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Perform orphan deletion
	_, err = ds.DropOrphans(context.Background(), nil, nil)
	if err != nil {
		log.Error().Err(err).
			Str("orgID", orgID).
			Str("datasetID", datasetID).
			Int("revision", ds.Revision).
			Msg("failed to cleanup orphans")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CleanupResponse{
			Status:    "error",
			OrgID:     orgID,
			DatasetID: datasetID,
			Revision:  ds.Revision,
			Message:   err.Error(),
		})
		return
	}

	log.Info().
		Str("orgID", orgID).
		Str("datasetID", datasetID).
		Int("revision", ds.Revision).
		Msg("orphan cleanup completed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CleanupResponse{
		Status:    "success",
		OrgID:     orgID,
		DatasetID: datasetID,
		Revision:  ds.Revision,
		Message:   "orphan records deleted",
	})
}

// handleGlobalCleanupOrphans handles POST /api/admin/cleanup-orphans
// This endpoint scans all datasets for the organization (from request context),
// identifies those with orphaned records (multiple revisions in the index),
// and cleans them up automatically.
// It skips deleted datasets and reports results for each dataset processed.
func (s *Service) handleGlobalCleanupOrphans(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)

	if orgID == "" {
		http.Error(w, "organization not found in request context", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("orgID", string(orgID)).
		Msg("global orphan cleanup requested")

	// List all datasets for this organization
	datasets, err := models.ListDataSets(string(orgID))
	if err != nil {
		log.Error().Err(err).
			Str("orgID", string(orgID)).
			Msg("unable to list datasets for orphan cleanup")
		http.Error(w, "failed to list datasets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := GlobalCleanupResponse{
		OrgID:         string(orgID),
		TotalDatasets: len(datasets),
		Results:       []DatasetCleanupResult{},
	}

	ctx := r.Context()

	for _, ds := range datasets {
		// Skip deleted datasets
		if ds.Deleted {
			response.Skipped++
			continue
		}

		// Get index stats to check for orphan revisions
		stats, err := models.CreateDataSetStats(ctx, string(orgID), ds.Spec)
		if err != nil {
			log.Warn().Err(err).
				Str("datasetID", ds.Spec).
				Msg("unable to get stats for dataset, skipping")
			response.Skipped++
			continue
		}

		// Check if there are orphan revisions (more than 1 revision in index)
		orphanRevisions := len(stats.IndexStats.Revisions) - 1
		if orphanRevisions <= 0 {
			// No orphans, skip
			continue
		}

		// Found orphans, add to count
		response.WithOrphans++

		result := DatasetCleanupResult{
			Spec:            ds.Spec,
			Revision:        ds.Revision,
			OrphanRevisions: orphanRevisions,
		}

		// Perform cleanup
		_, cleanupErr := ds.DropOrphans(ctx, nil, nil)
		if cleanupErr != nil {
			log.Error().Err(cleanupErr).
				Str("datasetID", ds.Spec).
				Int("orphanRevisions", orphanRevisions).
				Msg("failed to cleanup orphans for dataset")
			result.Status = "error"
			result.Error = cleanupErr.Error()
			response.Errors++
		} else {
			log.Info().
				Str("datasetID", ds.Spec).
				Int("orphanRevisions", orphanRevisions).
				Int("currentRevision", ds.Revision).
				Msg("cleaned up orphans for dataset")
			result.Status = "cleaned"
			response.CleanedUp++
		}

		response.Results = append(response.Results, result)
	}

	if response.Errors > 0 {
		response.Status = "partial"
	} else if response.CleanedUp > 0 {
		response.Status = "success"
	} else {
		response.Status = "no_orphans"
	}

	log.Info().
		Str("orgID", string(orgID)).
		Int("totalDatasets", response.TotalDatasets).
		Int("withOrphans", response.WithOrphans).
		Int("cleanedUp", response.CleanedUp).
		Int("errors", response.Errors).
		Int("skipped", response.Skipped).
		Msg("global orphan cleanup completed")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
