// github.com/delving/hub3/tools/cmd/source_analyzer/processor/random_sample.go
package processor

import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

// RandomSample maintains a random sample of strings using reservoir sampling
type RandomSample struct {
	size   int
	count  int
	sample []string
	random *rand.Rand
	mu     sync.Mutex
}

// NewRandomSample creates a new random sample collector
func NewRandomSample(size int) *RandomSample {
	return &RandomSample{
		size:   size,
		sample: make([]string, 0, size),
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Record adds a string to the sample using reservoir sampling
func (r *RandomSample) Record(value string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++

	// If we haven't filled the reservoir yet, just append
	if len(r.sample) < r.size {
		r.sample = append(r.sample, value)
		return
	}

	// Use reservoir sampling algorithm
	// Generate random number k in range [0, count)
	k := r.random.Intn(r.count)

	// If k is within the reservoir range, replace that element
	if k < r.size {
		r.sample[k] = value
	}
}

// Values returns the current sample values in sorted order
func (r *RandomSample) Values() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy of the sample
	result := make([]string, len(r.sample))
	copy(result, r.sample)

	// Sort the copy
	sort.Strings(result)

	return result
}

// Size returns the target sample size
func (r *RandomSample) Size() int {
	return r.size
}

// Count returns the total number of items processed
func (r *RandomSample) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

// Current returns the current number of items in the sample
func (r *RandomSample) Current() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sample)
}

// IsFull returns true if the sample has reached its target size
func (r *RandomSample) IsFull() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sample) >= r.size
}

// Clear empties the sample
func (r *RandomSample) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sample = r.sample[:0]
	r.count = 0
}
