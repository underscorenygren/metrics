package blackhole_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBlackhole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blackhole Suite")
}
