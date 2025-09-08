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
	if s.ts == nil {
		return fmt.Errorf("task service is nil")
	}
	
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

	store, err := s.GetStore(p.Meta.OrgID.String())
	if err != nil {
		return fmt.Errorf("unable to get store: %w", err)
	}

	return store.Save(pids...)
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
		Meta:       p.Meta,
	}

	if p.Deleted {
		pid.Tombstone = true
	}

	pid.inferType()

	return pid
}

func NewSavePIDTask(pid *PID) (*asynq.Task, error) {
	p := pid.asPayload()
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(pidSavetype, payload), nil
}
