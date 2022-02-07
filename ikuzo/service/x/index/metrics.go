package index

import "time"

type Metrics struct {
	started time.Time
	Nats    struct {
		Published uint64
		Consumed  uint64
		Failed    uint64
	}
	Index struct {
		Successful uint64
		Failed     uint64
	}
}
