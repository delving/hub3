package ead

import (
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

func (s *Service) findAvailableTask() *Task {
	tasks := []*Task{}

	for _, task := range s.tasks {
		if task.InState == StatePending || task.Interrupted {
			tasks = append(tasks, task)
		}
	}

	if len(tasks) == 0 {
		return nil
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].currentTransition().Started.After(tasks[j].currentTransition().Started)
	})

	log.Info().Str("svc", "eadProcessor").Int("availableTasks", len(tasks)).Msg("returning first available task for processing")

	return tasks[0]
}

func (s *Service) Tasks(w http.ResponseWriter, r *http.Request) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	// TODO(kiivihal): add option to filter by datasetID

	render.JSON(w, r, s.tasks)
}

func (s *Service) findTask(orgID, datasetID string, filterActive bool) (*Task, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	for _, t := range s.tasks {
		// TODO(kiivihal): add filter for orgID later
		_ = orgID

		if t.Meta.DatasetID == datasetID {
			if filterActive && !t.isActive() {
				continue
			}

			return t, nil
		}
	}

	return nil, ErrTaskNotFound
}

func (s *Service) GetTask(w http.ResponseWriter, r *http.Request) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	id := chi.URLParam(r, "id")

	task, ok := s.tasks[id]
	if !ok {
		http.Error(w, "unknown task", http.StatusNotFound)
		return
	}

	render.JSON(w, r, task)
}

func (s *Service) CancelTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.rw.Lock()
	defer s.rw.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		http.Error(w, "unknown task", http.StatusNotFound)
		return
	}

	task.moveState(StateCanceled)

	task.log().Info().Msg("canceling running ead task")
	task.cancel()

	task.Next()
	// TODO(kiivihal): do we delete or keep it
	// delete(s.tasks, id)

	w.WriteHeader(http.StatusNoContent)
}
