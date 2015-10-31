package reflector_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoReflector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoReflector Suite")
}
