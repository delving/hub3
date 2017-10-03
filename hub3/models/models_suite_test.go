package models

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Models Suite")
}

var _ = BeforeSuite(func() {
	Expect("test.db").ToNot(BeAnExistingFile())
	orm = newDB("test")
})

var _ = AfterSuite(func() {
	os.Remove("test.db")
})
