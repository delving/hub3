package resource

import (
	"testing"

	"github.com/matryer/is"
)

func Test_randStringBytesMask(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)

	idSize := 11

	unique := map[string]bool{}

	count := 500
	for i := 0; i < count; i++ {
		got := randStringBytesMask(idSize)
		is.True(len(got) == idSize)
		is.True(got != "")

		_, ok := unique[got]
		is.True(!ok) // each key should be unique

		unique[got] = true
	}

	is.Equal(len(unique), count) // map should contain all unique ids
}

func Test_NewAnonNode(t *testing.T) {
	// nolint:gocritic
	is := is.New(t)

	anon := NewAnonNode()
	is.Equal(len(anon.RawValue()), AnonNodeIDSize)

	AnonNodeIDSize = 11
	customAnon := NewAnonNode()

	is.True(AnonNodeIDSize == 11)
	is.True(len(customAnon.RawValue()) == AnonNodeIDSize)
}
