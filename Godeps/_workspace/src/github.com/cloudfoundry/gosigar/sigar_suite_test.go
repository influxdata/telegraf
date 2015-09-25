package sigar_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGosigar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gosigar Suite")
}
