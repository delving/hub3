package ead

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

var (
	ErrTaskNotFound         = errors.New("task not found")
	ErrTaskAlreadySubmitted = errors.New("task already submitted")
)

type ProcessingState string

const (
	StateSubmitted             ProcessingState = "submitted source EAD"
	StatePending                               = "pending processing"
	StateProcessingDescription                 = "processing description"
	StateProcessingMetsFiles                   = "processing METS files"
	StateProcessingInventories                 = "processing and indexing inventories"
	StateInError                               = "stopped processing with error"
	StateCancelled                             = "cancelled processing"
	StateFinished                              = "finished processing EAD"
)

type Transition struct {
	State       ProcessingState   `json:"state"`
	Started     time.Time         `json:"started"`
	Finished    time.Time         `json:"finished"`
	Metrics     map[string]uint64 `json:"metrics,omitempty"`
	Duration    time.Duration     `json:"duration"`
	DurationFmt string            `json:"durationFmt"`
}

type Task struct {
	ID          string `json:"id"`
	Meta        *Meta
	InState     ProcessingState `json:"inState"`
	ErrorMsg    string          `json:"errorMsg"`
	Transitions []*Transition   `json:"transitions"`
	Interrupted bool
	s           *Service
	ctx         context.Context
	cancel      context.CancelFunc
}

func (t *Task) finishState() *Transition {
	last := t.Transitions[len(t.Transitions)-1]
	last.Finished = time.Now()
	last.Duration = last.Finished.Sub(last.Started)
	last.DurationFmt = last.Duration.String()

	return last
}

func (t *Task) isActive() bool {
	inActiveStates := []ProcessingState{StateInError, StateCancelled, StateFinished}
	for _, state := range inActiveStates {
		if state == t.InState {
			return false
		}
	}

	return true
}

func (t *Task) finishTask() {
	last := t.Transitions[len(t.Transitions)-1]
	last.Finished = time.Now()

	var startProcessing *Transition

	for _, transition := range t.Transitions {
		if transition.State == StatePending {
			startProcessing = transition
			break
		}
	}

	t.Meta.ProcessingDuration = last.Finished.Sub(startProcessing.Finished)
	t.Meta.ProcessingDurationFmt = t.Meta.ProcessingDuration.String()

	log.Info().Str("datasetID", t.Meta.DatasetID).Dur("processing", t.Meta.ProcessingDuration).
		Int("inventories", int(t.Meta.Clevels)).
		Int("metsFiles", int(t.Meta.DaoLinks)).
		Int("recordsPublished", int(t.Meta.RecordsPublished)).
		Int("digitalObjects", int(t.Meta.DigitalObjects)).
		Msg("finished processing")

	t.moveState(StateFinished)
	t.finishState()
}

func (t *Task) moveState(state ProcessingState) {
	current := t.finishState()
	log.Info().Str("datasetID", t.Meta.DatasetID).Str("taskID", t.ID).
		Str("oldState", string(t.InState)).Str("newState", string(state)).Dur("dur", current.Duration).Msg("EAD state transition")

	t.InState = state
	t.Transitions = append(t.Transitions, &Transition{State: state, Started: time.Now()})
}

func (t *Task) finishWithError(err error) error {
	t.finishState()
	log.Error().Err(err).Str("orgID", t.Meta.OrgID).Str("datasetID", t.Meta.DatasetID).
		Str("taskState", string(t.InState)).Msg("stopped EAD task with error")
	t.moveState(StateInError)
	t.ErrorMsg = err.Error()

	atomic.AddUint64(&t.s.m.Failed, 1)

	return err
}

func (t *Task) currentTransition() *Transition {
	return t.Transitions[len(t.Transitions)-1]
}

func (t *Task) Next() {
	switch t.InState {
	case StateSubmitted:
		t.moveState(StatePending)
	case StatePending:
		t.moveState(StateProcessingDescription)
	case StateProcessingDescription:
		t.moveState(StateProcessingInventories)
	case StateProcessingInventories:
		t.moveState(StateFinished)
		t.finishTask()
	case StateInError:
		t.finishState()
	case StateCancelled:
		t.finishState()
		atomic.AddUint64(&t.s.m.Canceled, 1)
	}
}

func (s *Service) NewTask(meta *Meta) (*Task, error) {
	task := &Task{
		ID:      xid.New().String(),
		s:       s,
		Meta:    meta,
		InState: StateSubmitted,
	}

	entry := &Transition{
		State:   StateSubmitted,
		Started: time.Now(),
		Metrics: map[string]uint64{
			"fileSize": meta.FileSize,
		},
	}
	task.Transitions = append(task.Transitions, entry)
	task.Next()

	task.ctx, task.cancel = context.WithCancel(context.Background())

	if _, err := s.findTask("", meta.DatasetID, true); !errors.Is(err, ErrTaskNotFound) {
		return nil, ErrTaskAlreadySubmitted
	}

	s.rw.Lock()
	s.tasks[task.ID] = task
	s.rw.Unlock()

	return task, nil
}
