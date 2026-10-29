package nde

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/delving/hub3/hub3/models"
)

const narthexSyncInterval = time.Hour

// syncNarthexLoop syncs the dataset descriptions from Narthex at startup and then every narthexSyncInterval.
func (s *Service) syncNarthexLoop(ctx context.Context) {
	ticker := time.NewTicker(narthexSyncInterval)
	defer ticker.Stop()

	for {
		if err := s.syncNarthex(); err != nil {
			log.Printf("unable to sync nde datasets with narthex: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// syncNarthex stores the dataset descriptions from the Narthex triple store as the
// dataset description, which createDataSet uses as the name of the dataset.
func (s *Service) syncNarthex() error {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	for _, orgID := range s.orgIDs {
		infos, err := models.GetNDEInfo(orgID)
		if err != nil {
			return fmt.Errorf("unable to get nde info for %q: %w", orgID, err)
		}

		for _, info := range infos {
			ds, err := models.GetDataSet(orgID, info.Spec)
			if err != nil || ds.Description == info.Description {
				continue
			}

			ds.Description = info.Description
			if err := ds.Save(); err != nil {
				log.Printf("unable to save nde description for %q: %v", info.Spec, err)
			}
		}
	}

	return nil
}

func (s *Service) HandleNarthexSync(w http.ResponseWriter, r *http.Request) {
	if err := s.syncNarthex(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
