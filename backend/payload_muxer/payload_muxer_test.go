package payload_muxer_test

import (
	"encoding/json"
	"io"

	"github.com/cloudfoundry-incubator/garden/backend"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/warden-windows/backend/messages"
	. "github.com/cloudfoundry-incubator/warden-windows/backend/payload_muxer"
)

var _ = Describe("PayloadMuxer", func() {
	var muxer PayloadMuxer

	BeforeEach(func() {
		muxer = New()
	})

	Context("when a payload appears on the stream", func() {
		var stream io.Writer

		BeforeEach(func() {
			streamR, streamW := io.Pipe()

			stream = streamW

			muxer.SetSource(streamR)
		})

		It("sends it over the returned channel", func() {
			processStream := make(chan backend.ProcessStream, 1000)

			muxer.Subscribe(42, processStream)

			encoder := json.NewEncoder(stream)

			err := encoder.Encode(&messages.ProcessPayload{
				ProcessID: 42,
				Source:    backend.ProcessStreamSourceStdout,
				Data:      []byte("stdout data for 42"),
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = encoder.Encode(&messages.ProcessPayload{
				ProcessID: 43,
				Source:    backend.ProcessStreamSourceStdout,
				Data:      []byte("stdout data for 43"),
			})
			Ω(err).ShouldNot(HaveOccurred())

			err = encoder.Encode(&messages.ProcessPayload{
				ProcessID: 42,
				Source:    backend.ProcessStreamSourceStderr,
				Data:      []byte("stderr data for 42"),
			})
			Ω(err).ShouldNot(HaveOccurred())

			var payload backend.ProcessStream
			Eventually(processStream).Should(Receive(&payload))
			Ω(payload.Source).Should(Equal(backend.ProcessStreamSourceStdout))
			Ω(string(payload.Data)).Should(Equal("stdout data for 42"))

			Eventually(processStream).Should(Receive(&payload))
			Ω(payload.Source).Should(Equal(backend.ProcessStreamSourceStderr))
			Ω(string(payload.Data)).Should(Equal("stderr data for 42"))
		})

		Context("but no subscribers can consume it", func() {
			It("does not block", func() {
				processStream := make(chan backend.ProcessStream)

				muxer.Subscribe(42, processStream)

				encoder := json.NewEncoder(stream)

				err := encoder.Encode(&messages.ProcessPayload{
					ProcessID: 42,
					Source:    backend.ProcessStreamSourceStdout,
					Data:      []byte("stdout data for 42"),
				})
				Ω(err).ShouldNot(HaveOccurred())

				status := uint32(123)

				err = encoder.Encode(&messages.ProcessPayload{
					ProcessID:  42,
					ExitStatus: &status,
				})
				Ω(err).ShouldNot(HaveOccurred())

				Eventually(processStream).Should(BeClosed())
			})
		})

		Context("when an exit status is received", func() {
			It("closes the channel and unsubscribes it", func() {
				processStream := make(chan backend.ProcessStream, 1000)

				muxer.Subscribe(42, processStream)

				encoder := json.NewEncoder(stream)

				status42 := uint32(142)
				status43 := uint32(143)

				err := encoder.Encode(&messages.ProcessPayload{
					ProcessID:  43,
					ExitStatus: &status43,
				})
				Ω(err).ShouldNot(HaveOccurred())

				err = encoder.Encode(&messages.ProcessPayload{
					ProcessID:  42,
					ExitStatus: &status42,
				})
				Ω(err).ShouldNot(HaveOccurred())

				var payload backend.ProcessStream

				Eventually(processStream).Should(Receive(&payload))
				Ω(payload.ExitStatus).ShouldNot(BeNil())
				Ω(*payload.ExitStatus).Should(Equal(uint32(142)))

				Eventually(processStream).Should(BeClosed())

				err = encoder.Encode(&messages.ProcessPayload{
					ProcessID: 42,
					Source:    backend.ProcessStreamSourceStdout,
					Data:      []byte("stdout data for 42"),
				})
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
