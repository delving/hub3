package hub3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHub3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hub3 Suite")
}
