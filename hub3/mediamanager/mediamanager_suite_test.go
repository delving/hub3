package mediamanager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMediamanager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mediamanager Suite")
}
