package task

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type health struct {
	taskName string
	logger   *zerolog.Logger
}

type healthMsg struct {
	Label string
}

func (h *health) scheduleTask(scheduler *asynq.Scheduler) error {
	t, err := h.newHealthTask()
	if err != nil {
		return err
	}

	entryID, err := scheduler.Register(
		"@every 30s",
		t,
		asynq.TaskID("mytaskid"),
		asynq.Retention(15*time.Second),
		asynq.Unique(15*time.Second),
	)
	if err != nil {
		return err
	}
	log.Info().Str("entryID", entryID).Msg("scheduled health ping for workers")

	return nil
}

func (h *health) newHealthTask() (*asynq.Task, error) {
	payload, err := json.Marshal(healthMsg{Label: "my health message"})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(h.taskName, payload), nil
}

func (h *health) handleTask(ctx context.Context, t *asynq.Task) error {
	var msg healthMsg
	if err := json.Unmarshal(t.Payload(), &msg); err != nil {
		return err
	}
	log.Info().Str("label", msg.Label).Msg("health ping message")

	return nil
}
