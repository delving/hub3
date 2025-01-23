package rdf

import (
	"fmt"
	"hash/fnv"
	"log/slog"
)

// Option defines a functional option for configuring the RDFXMLDecoder
type Option func(*rdfXMLDecoder)

// WithSeed sets the seed for ID generation
func WithSeed(seed string) Option {
	return func(d *rdfXMLDecoder) {
		d.seed = seed
	}
}

// WithIDStrategy sets the ID generation strategy
func WithIDStrategy(strategy IDGenerationStrategy) Option {
	return func(d *rdfXMLDecoder) {
		switch strategy {
		case IncrementalIDs:
			d.idGenerator = &IncrementalIDGenerator{}
		case DeterministicIDs:
			d.idGenerator = &PathBasedIDGenerator{decoder: d}
		}
	}
}

// BlankNodeIDGenerator defines the interface for generating blank node IDs
type BlankNodeIDGenerator interface {
	GenerateID() string
}

// IncrementalIDGenerator generates sequential blank node IDs
type IncrementalIDGenerator struct {
	counter int
}

func (g *IncrementalIDGenerator) GenerateID() string {
	id := fmt.Sprintf("_:b%d", g.counter)
	g.counter++
	return id
}

// PathBasedIDGenerator generates deterministic IDs based on XML path
type PathBasedIDGenerator struct {
	decoder *rdfXMLDecoder
}

func (g *PathBasedIDGenerator) GenerateID() string {
	if len(g.decoder.ctx.Path) == 0 {
		slog.Warn("generating ID with empty path", "seed", g.decoder.seed)
	}

	hash := fnv.New64a()

	pathLen := fmt.Sprintf("%d", len(g.decoder.ctx.Path))
	hash.Write([]byte(pathLen))

	for i, p := range g.decoder.ctx.Path {
		posMarker := fmt.Sprintf(":%d:", i)
		hash.Write([]byte(posMarker))
		hash.Write([]byte(p))
	}

	hash.Write([]byte(g.decoder.seed))
	return fmt.Sprintf("_:b%x", hash.Sum64())
}

// IDGenerationStrategy defines the type of ID generation to use
type IDGenerationStrategy int

const (
	IncrementalIDs IDGenerationStrategy = iota
	DeterministicIDs
)
