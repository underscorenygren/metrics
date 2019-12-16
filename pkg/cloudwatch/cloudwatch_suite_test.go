package cloudwatch_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCloudwatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloudwatch Suite")
}
