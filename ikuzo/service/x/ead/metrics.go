package ead

import "sync/atomic"

type Metrics struct {
	Submitted     uint64
	Started       uint64
	Failed        uint64
	Finished      uint64
	Canceled      uint64
	AlreadyQueued uint64
}

func (m *Metrics) IncSubmitted() {
	atomic.AddUint64(&m.Submitted, 1)
}

func (m *Metrics) IncStarted() {
	atomic.AddUint64(&m.Started, 1)
}

func (m *Metrics) IncFailed() {
	atomic.AddUint64(&m.Failed, 1)
}

func (m *Metrics) IncFinished() {
	atomic.AddUint64(&m.Finished, 1)
}

func (m *Metrics) IncCancelled() {
	atomic.AddUint64(&m.Canceled, 1)
}

func (m *Metrics) IncAlreadyQueued() {
	atomic.AddUint64(&m.AlreadyQueued, 1)
}
