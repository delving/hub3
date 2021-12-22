package imageproxy

import "sync/atomic"

type Metrics struct {
	Source             uint64
	Cache              uint64
	LruCache           uint64
	RemoteRequestError uint64
	RejectDomain       uint64
	RejectReferrer     uint64
	RejectURI          uint64
	Resize             uint64
	DeepZoom           uint64
	Error              uint64
	// Canceled      uint64
	// AlreadyQueued uint64
}

func (s *Service) Metrics() Metrics {
	return s.m
}

func (m *Metrics) IncSource() {
	atomic.AddUint64(&m.Source, 1)
}

func (m *Metrics) IncCache() {
	atomic.AddUint64(&m.Cache, 1)
}

func (m *Metrics) IncLruCache() {
	atomic.AddUint64(&m.LruCache, 1)
}

func (m *Metrics) IncRemoteRequestError() {
	atomic.AddUint64(&m.RemoteRequestError, 1)
}

func (m *Metrics) IncError() {
	atomic.AddUint64(&m.Error, 1)
}

func (m *Metrics) IncRejectURI() {
	atomic.AddUint64(&m.RejectURI, 1)
}

func (m *Metrics) IncRejectDomain() {
	atomic.AddUint64(&m.RejectDomain, 1)
}

func (m *Metrics) IncRejectReferrer() {
	atomic.AddUint64(&m.RejectReferrer, 1)
}

func (m *Metrics) IncResize() {
	atomic.AddUint64(&m.Resize, 1)
}

func (m *Metrics) IncDeepZoom() {
	atomic.AddUint64(&m.DeepZoom, 1)
}

// func (m *Metrics) IncAlreadyQueued() {
// atomic.AddUint64(&m.AlreadyQueued, 1)
// }
