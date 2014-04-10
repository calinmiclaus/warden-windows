package payload_muxer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPayload_muxer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PayloadMuxer Suite")
}
