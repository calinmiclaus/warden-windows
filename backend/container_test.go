package backend_test

import (
	"errors"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"reflect"

	"github.com/cloudfoundry-incubator/garden/backend"
	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	. "github.com/cloudfoundry/gunk/command_runner/fake_command_runner/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/warden-windows/backend"
	"github.com/cloudfoundry-incubator/warden-windows/backend/iorpc"
	"github.com/cloudfoundry-incubator/warden-windows/backend/messages"
	"github.com/cloudfoundry-incubator/warden-windows/backend/payload_muxer/fake_muxer"
)

type FakeContainerServer struct {
	WhenStopping func(request *messages.StopRequest, response *messages.StopResponse) error
	WhenRunning  func(request *messages.RunRequest, response *messages.RunResponse) error
}

func (server FakeContainerServer) Stop(request *messages.StopRequest, response *messages.StopResponse) error {
	defer GinkgoRecover()

	if server.WhenStopping != nil {
		return server.WhenStopping(request, response)
	}

	return nil
}

func (server FakeContainerServer) Run(request *messages.RunRequest, response *messages.RunResponse) error {
	defer GinkgoRecover()

	if server.WhenRunning != nil {
		return server.WhenRunning(request, response)
	}

	return nil
}

var _ = Describe("Container", func() {
	var runner *fake_command_runner.FakeCommandRunner
	var muxer *fake_muxer.FakeMuxer

	var container *Container

	BeforeEach(func() {
		runner = fake_command_runner.New()
		muxer = fake_muxer.New()
		container = NewContainer("some-id", "some-handle", runner, muxer)
	})

	Describe("Start", func() {
		It("spawns a daemon with the correct handle", func() {
			err := container.Start()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(runner).Should(HaveStartedExecuting(
				fake_command_runner.CommandSpec{
					Path: "DAEMON_PATH",
					Args: []string{"--handle", "some-handle"},
				},
			))
		})

		It("hooks into the daemon's stdin/stdout", func() {
			runner.WhenRunning(
				fake_command_runner.CommandSpec{
					Path: "DAEMON_PATH",
					Args: []string{"--handle", "some-handle"},
				}, func(cmd *exec.Cmd) error {
					Ω(cmd.Stdin).ShouldNot(BeNil())
					Ω(cmd.Stdout).ShouldNot(BeNil())
					Ω(cmd.Stderr).ShouldNot(BeNil())
					return nil
				},
			)

			err := container.Start()
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when spawning the daemon fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				runner.WhenRunning(
					fake_command_runner.CommandSpec{
						Path: "DAEMON_PATH",
						Args: []string{"--handle", "some-handle"},
					}, func(cmd *exec.Cmd) error {
						return disaster
					},
				)
			})

			It("returns the error", func() {
				err := container.Start()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("after starting", func() {
		var containerServer *FakeContainerServer
		var server *rpc.Server

		var processPayloadStream io.Writer

		BeforeEach(func() {
			containerServer = new(FakeContainerServer)

			server = rpc.NewServer()
			server.RegisterName("Container", containerServer)

			runner.WhenRunning(
				fake_command_runner.CommandSpec{
					Path: "DAEMON_PATH",
					Args: []string{"--handle", "some-handle"},
				}, func(cmd *exec.Cmd) error {
					go server.ServeCodec(jsonrpc.NewServerCodec(iorpc.New(cmd.Stdout.(io.WriteCloser), cmd.Stdin.(io.ReadCloser))))

					processPayloadStream = cmd.Stderr

					return nil
				},
			)

			err := container.Start()
			Ω(err).ShouldNot(HaveOccurred())
		})

		Describe("Stop", func() {
			It("sends a Stop message to the container", func() {
				calledStop := false

				containerServer.WhenStopping = func(request *messages.StopRequest, response *messages.StopResponse) error {
					calledStop = true
					return nil
				}

				err := container.Stop(false)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(calledStop).Should(BeTrue())
			})

			Context("when told to kill", func() {
				It("sends the message with Kill true", func() {
					containerServer.WhenStopping = func(request *messages.StopRequest, response *messages.StopResponse) error {
						Ω(request.Kill).Should(BeTrue())

						return nil
					}

					err := container.Stop(true)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			Context("when the RPC fails", func() {
				disaster := errors.New("oh no!")

				BeforeEach(func() {
					containerServer.WhenStopping = func(request *messages.StopRequest, response *messages.StopResponse) error {
						return disaster
					}
				})

				It("returns an error", func() {
					err := container.Stop(true)
					Ω(err.Error()).Should(Equal(disaster.Error()))
				})
			})
		})

		Describe("Run", func() {
			It("sends a Run message to the container", func() {
				calledRun := false

				containerServer.WhenRunning = func(request *messages.RunRequest, response *messages.RunResponse) error {
					calledRun = true
					response.ProcessID = 42
					return nil
				}

				pid, _, err := container.Run(backend.ProcessSpec{
					Script:     "rm -rf /",
					Privileged: true,
				})
				Ω(err).ShouldNot(HaveOccurred())

				Ω(calledRun).Should(BeTrue())
				Ω(pid).Should(Equal(uint32(42)))
			})

			It("sends the script and privileged along", func() {
				containerServer.WhenRunning = func(request *messages.RunRequest, response *messages.RunResponse) error {
					Ω(request.Script).Should(Equal("rm -rf /"))
					Ω(request.Privileged).Should(BeTrue())
					return nil
				}

				_, _, err := container.Run(backend.ProcessSpec{
					Script:     "rm -rf /",
					Privileged: true,
				})
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("subscribes to the process id's payloads", func() {
				processID, stream, err := container.Run(backend.ProcessSpec{
					Script:     "rm -rf /",
					Privileged: true,
				})
				Ω(err).ShouldNot(HaveOccurred())

				subscribers := muxer.Subscribers(processID)
				Ω(subscribers).Should(HaveLen(1))

				writeEnd := subscribers[0]

				// cannot assert equality as they have different types (<-chan vs chan<-)
				Ω(reflect.ValueOf(writeEnd).Pointer()).Should(Equal(reflect.ValueOf(stream).Pointer()))
			})

			Context("when process payloads show up over the channel", func() {
				It("writes them to the stream", func() {
					processID, stream, err := container.Run(backend.ProcessSpec{
						Script:     "rm -rf /",
						Privileged: true,
					})
					Ω(err).ShouldNot(HaveOccurred())

					subscribers := muxer.Subscribers(processID)
					Ω(subscribers).Should(HaveLen(1))

					writeEnd := subscribers[0]

					go func() {
						writeEnd <- backend.ProcessStream{
							Source: backend.ProcessStreamSourceStdout,
							Data:   []byte("stdout data for 42"),
						}

						writeEnd <- backend.ProcessStream{
							Source: backend.ProcessStreamSourceStderr,
							Data:   []byte("stderr data for 42"),
						}

						exitStatus := uint32(142)

						writeEnd <- backend.ProcessStream{
							ExitStatus: &exitStatus,
						}
					}()

					var payload backend.ProcessStream
					Eventually(stream).Should(Receive(&payload))
					Ω(payload.Source).Should(Equal(backend.ProcessStreamSourceStdout))
					Ω(string(payload.Data)).Should(Equal("stdout data for 42"))

					Eventually(stream).Should(Receive(&payload))
					Ω(payload.Source).Should(Equal(backend.ProcessStreamSourceStderr))
					Ω(string(payload.Data)).Should(Equal("stderr data for 42"))

					Eventually(stream).Should(Receive(&payload))
					Ω(payload.ExitStatus).ShouldNot(BeNil())
					Ω(*payload.ExitStatus).Should(Equal(uint32(142)))
				})
			})
		})
	})
})
