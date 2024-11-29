package pid

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	pidSavetype = "pid:save_record"
)

func (s *Service) registerTaskHandlers() error {
	s.log.Info().Msg("scheduling sync task handlers")
	s.ts.RegisterWorkerFunc(pidSavetype, s.HandleSavePidTask)
	return nil
}

func (s *Service) HandleSavePidTask(ctx context.Context, t *asynq.Task) error {
	var p SavePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	pids, err := s.process(p)
	if err != nil {
		return fmt.Errorf("failed to process pid: %v", err)
	}

	return s.store.Save(pids...)
}

type SavePayload struct {
	RDFSubject string // the rdf subject that is linked to the PID
	PID        string // the external PID
	IsShownAt  string // the IsShownAt URL that is linked to the RDFSubject
	Type       Type   // the type of the PID. When left empty we will try to infer the type from the PID
	Meta
	Deleted bool // if the PID is deleted
}

func (p SavePayload) asPID() PID {
	pid := PID{
		ID:         p.RDFSubject,
		ExternalID: p.PID,
		ModifiedAt: time.Now(),
		IsShownAt:  p.IsShownAt,
		Meta:       Meta{},
	}

	if p.Deleted {
		pid.Tombstone = true
	}

	pid.inferType()

	return pid
}

func NewSavePIDTask(p SavePayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(pidSavetype, payload), nil
}
