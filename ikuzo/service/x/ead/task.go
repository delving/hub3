package ead

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/transition"
	"github.com/rs/xid"
)

type Task struct {
	ID             string
	OrganizationID string
	DataSetID      string
	Started        time.Time
	Finished       time.Time
	s              *Service
	ctx            context.Context
	cancel         context.CancelFunc
	sm             *transition.StateMachine
	transition.Transition
}

func (s *Service) NewTask() *Task {
	task := &Task{
		ID:      xid.New().String(),
		s:       s,
		Started: time.Now(),
	}

	task.ctx, task.cancel = context.WithCancel(context.Background())

	s.rw.Lock()
	// TODO(kiivihal): where to check if task is already in progress
	s.tasks[task.ID] = task
	s.rw.Unlock()

	return task
}

func (t *Task) getStateMachine() *transition.StateMachine {
	var eadStateMachine = transition.New(&Task{})

	eadStateMachine.Initial("started")
	eadStateMachine.State("savedEAD").
		Enter(func(order interface{}, tx *gorm.DB) error {
			// To get order object use 'task.(*Task)'
			// business logic here
			return nil
		}).
		Exit(func(order interface{}, tx *gorm.DB) error {
			// business logic here
			return nil
		})

	eadStateMachine.State("savedMETS")
	eadStateMachine.State("savedDescription")
	eadStateMachine.State("clevelsProduced")
	eadStateMachine.State("clevelsProcessed")
	eadStateMachine.State("done")
	eadStateMachine.State("cancelled")

	eadStateMachine.Event("upload").To("savedEAD").From("started")

	return eadStateMachine
}
