package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Metric struct {
	Name  string
	Value atomic.Int64
}

type MetricValue struct {
	Name  string
	Value interface{}
}

type Reporter interface {
	Start(ctx context.Context, opType string) error
	Update(values []MetricValue)
	Complete(values []MetricValue)
}

type Stats struct {
	metrics     []*Metric
	metricIndex map[string]*Metric
	startTime   time.Time
	lastUpdate  time.Time
	mu          sync.RWMutex
}

func NewStats() *Stats {
	return &Stats{
		metrics:     make([]*Metric, 0),
		metricIndex: make(map[string]*Metric),
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
	}
}

func (s *Stats) RegisterMetric(name string) *Metric {
	s.mu.Lock()
	defer s.mu.Unlock()
	metric := &Metric{Name: name}
	s.metrics = append(s.metrics, metric)
	s.metricIndex[name] = metric
	return metric
}

func (s *Stats) GetValues() []MetricValue {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]MetricValue, 0, len(s.metrics)*2+1)
	duration := time.Since(s.startTime).Seconds()

	for _, metric := range s.metrics {
		value := metric.Value.Load()
		values = append(values,
			MetricValue{Name: metric.Name, Value: value},
			MetricValue{Name: metric.Name + "_per_second", Value: float64(value) / duration},
		)
	}
	values = append(values, MetricValue{Name: "duration", Value: time.Since(s.startTime).Round(time.Second)})
	return values
}

type SlogReporter struct {
	logger   *slog.Logger
	interval time.Duration
}

func NewSlogReporter(logger *slog.Logger, interval time.Duration) *SlogReporter {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogReporter{
		logger:   logger,
		interval: interval,
	}
}

func (sr *SlogReporter) Start(ctx context.Context, opType string) error {
	sr.logger.Info("starting " + opType)
	return nil
}

func (sr *SlogReporter) Update(values []MetricValue) {
	args := make([]interface{}, 0, len(values)*2)
	for _, mv := range values {
		args = append(args, mv.Name, mv.Value)
	}
	sr.logger.Info("progress update", args...)
}

func (sr *SlogReporter) Complete(values []MetricValue) {
	args := make([]interface{}, 0, len(values)*2)
	for _, mv := range values {
		args = append(args, mv.Name, mv.Value)
	}
	sr.logger.Info("operation complete", args...)
}

type ConsoleReporter struct {
	interval time.Duration
}

func NewConsoleReporter(interval time.Duration) *ConsoleReporter {
	return &ConsoleReporter{interval: interval}
}

func (cr *ConsoleReporter) Start(ctx context.Context, opType string) error {
	println("Starting " + opType + "...")
	return nil
}

func (cr *ConsoleReporter) Update(values []MetricValue) {
	m := make(map[string]interface{}, len(values))
	for _, mv := range values {
		m[mv.Name] = mv.Value
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	println(string(b))
}

func (cr *ConsoleReporter) Complete(values []MetricValue) {
	println("\nFinal Statistics:")
	m := make(map[string]interface{}, len(values))
	for _, mv := range values {
		m[mv.Name] = mv.Value
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	println(string(b))
}

type ProgressReporter struct {
	interval time.Duration
}

func NewProgressReporter(interval time.Duration) *ProgressReporter {
	return &ProgressReporter{interval: interval}
}

func (pr *ProgressReporter) Start(ctx context.Context, opType string) error {
	fmt.Printf("Starting %s...\n", opType)
	return nil
}

func (pr *ProgressReporter) Update(values []MetricValue) {
	// Build progress string
	var parts []string
	for _, mv := range values {
		switch mv.Name {
		case "duration":
			parts = append(parts, fmt.Sprintf("time: %v", mv.Value))
		case "processed":
			parts = append(parts, fmt.Sprintf("processed: %v", mv.Value))
		case "processed_per_second":
			parts = append(parts, fmt.Sprintf("rate: %.1f/s", mv.Value))
		case "errors":
			if val, ok := mv.Value.(int64); ok && val > 0 {
				parts = append(parts, fmt.Sprintf("errors: %v", mv.Value))
			}
		}
	}
	
	// Write to same line using carriage return
	fmt.Printf("\r%s", strings.Join(parts, " | "))
}

func (pr *ProgressReporter) Complete(values []MetricValue) {
	// Move to new line and print final stats
	fmt.Printf("\n")
	var parts []string
	for _, mv := range values {
		parts = append(parts, fmt.Sprintf("%s: %v", mv.Name, mv.Value))
	}
	fmt.Printf("Complete: %s\n", strings.Join(parts, " | "))
}

type MultiReporter struct {
	reporters []Reporter
}

func NewMultiReporter(reporters ...Reporter) *MultiReporter {
	return &MultiReporter{reporters: reporters}
}

func (mr *MultiReporter) Start(ctx context.Context, opType string) error {
	for _, r := range mr.reporters {
		if err := r.Start(ctx, opType); err != nil {
			return err
		}
	}
	return nil
}

func (mr *MultiReporter) Update(values []MetricValue) {
	for _, r := range mr.reporters {
		r.Update(values)
	}
}

func (mr *MultiReporter) Complete(values []MetricValue) {
	for _, r := range mr.reporters {
		r.Complete(values)
	}
}
