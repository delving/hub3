package harvesting_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHarvesting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Harvesting Suite")
}
