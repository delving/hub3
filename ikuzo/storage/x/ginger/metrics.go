package ginger

import (
	"sync"
	"time"
)

type PostHookCounter struct {
	ToIndex           int  `json:"toIndex"`
	ToDelete          int  `json:"toDelete"`
	InError           int  `json:"inError"`
	LifeTimeQueued    int  `json:"lifeTimeQueued"`
	LifeTimeProcessed int  `json:"lifeTimeProcessed"`
	IsActive          bool `json:"isActive"`
}

type PostHookGauge struct {
	Created        time.Time                   `json:"created"`
	QueueSize      int                         `json:"queueSize"`
	ActiveDatasets int                         `json:"activeDatasets"`
	Counters       map[string]*PostHookCounter `json:"counters"`
	sync.Mutex
}

func (phg *PostHookGauge) SetActive(counter *PostHookCounter) {
	if counter.LifeTimeProcessed != counter.LifeTimeQueued {
		if counter.IsActive {
			return
		}
		counter.IsActive = true
		return
	}
	if counter.IsActive {
		counter.IsActive = false
		return
	}
	return
}

func (phg *PostHookGauge) Done(ph *PostHookJob) error {
	counter, ok := phg.Counters[ph.item.DatasetID]
	if !ok {
		counter = &PostHookCounter{}
		phg.Counters[ph.item.DatasetID] = counter
	}
	phg.Lock()
	defer phg.Unlock()
	phg.QueueSize--

	switch ph.item.Deleted {
	case true:
		counter.ToDelete--
	default:
		counter.ToIndex--
	}
	counter.LifeTimeProcessed++
	phg.SetActive(counter)
	return nil
}

func (phg *PostHookGauge) Error(ph *PostHookJob) error {
	counter, ok := phg.Counters[ph.item.DatasetID]
	if !ok {
		counter = &PostHookCounter{}
		phg.Counters[ph.item.DatasetID] = counter
	}
	phg.Lock()
	defer phg.Unlock()
	switch ph.item.Deleted {
	case true:
		counter.ToDelete--
	default:
		counter.ToIndex--
	}
	counter.InError++
	counter.LifeTimeProcessed++
	phg.QueueSize--
	phg.SetActive(counter)
	return nil
}

func (phg *PostHookGauge) Queue(ph *PostHookJob) error {
	counter, ok := phg.Counters[ph.item.DatasetID]
	if !ok {
		counter = &PostHookCounter{}
		phg.Counters[ph.item.DatasetID] = counter
	}
	phg.Lock()
	defer phg.Unlock()
	counter.LifeTimeQueued++
	phg.QueueSize++
	phg.SetActive(counter)

	if ph.item.Deleted {
		counter.ToDelete++
		return nil
	}

	counter.ToIndex++
	return nil
}
