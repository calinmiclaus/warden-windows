package prison_client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPrisonClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PrisonClient Suite")
}
