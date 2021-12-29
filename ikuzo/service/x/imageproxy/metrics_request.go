package imageproxy

import "sync/atomic"

type RequestMetrics struct {
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
	Removed            uint64
	BytesServed        uint64
	// Canceled      uint64
	// AlreadyQueued uint64
}

func (s *Service) RequestMetrics() RequestMetrics {
	return s.m
}

func (m *RequestMetrics) IncBytesServed(size int64) {
	atomic.AddUint64(&m.BytesServed, uint64(size))
}

func (m *RequestMetrics) IncSource() {
	atomic.AddUint64(&m.Source, 1)
}

func (m *RequestMetrics) IncCache() {
	atomic.AddUint64(&m.Cache, 1)
}

func (m *RequestMetrics) IncLruCache() {
	atomic.AddUint64(&m.LruCache, 1)
}

func (m *RequestMetrics) IncRemoteRequestError() {
	atomic.AddUint64(&m.RemoteRequestError, 1)
}

func (m *RequestMetrics) IncRemoved() {
	atomic.AddUint64(&m.Removed, 1)
}

func (m *RequestMetrics) IncError() {
	atomic.AddUint64(&m.Error, 1)
}

func (m *RequestMetrics) IncRejectURI() {
	atomic.AddUint64(&m.RejectURI, 1)
}

func (m *RequestMetrics) IncRejectDomain() {
	atomic.AddUint64(&m.RejectDomain, 1)
}

func (m *RequestMetrics) IncRejectReferrer() {
	atomic.AddUint64(&m.RejectReferrer, 1)
}

func (m *RequestMetrics) IncResize() {
	atomic.AddUint64(&m.Resize, 1)
}

func (m *RequestMetrics) IncDeepZoom() {
	atomic.AddUint64(&m.DeepZoom, 1)
}

// func (m *Metrics) IncAlreadyQueued() {
// atomic.AddUint64(&m.AlreadyQueued, 1)
// }
