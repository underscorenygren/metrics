package failsink_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFailsink(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Failsink Suite")
}
