package ead_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEad(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ead Suite")
}
