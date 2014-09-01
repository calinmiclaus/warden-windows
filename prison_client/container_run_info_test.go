package prison_client_test

import (
	"github.com/mattn/go-ole"
	"github.com/mattn/go-ole/oleutil"
	"github.com/natefinch/npipe"
	PrisonClient "github.com/uhurusoftware/warden-windows/prison_client"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainerRunInfo", func() {
	BeforeEach(func() {
		ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	})

	Describe("CreateContainerRunInfo", func() {
		Context("when no error occurs", func() {
			It("should create the run info container", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()

				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("EnvironemntVariables", func() {
		Context("when set environemnt variables succeeds", func() {
			It("should set the right environemnt variables", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				envName := "foo"
				envValue := "bar"
				cri.AddEnvironemntVariable(envName, envValue)

				oleDispatch, err := cri.GetIDispatch()
				Expect(err).ShouldNot(HaveOccurred())

				envList, err := oleutil.GetProperty(oleDispatch, "ListEnvironemntVariableKeys")
				Expect(err).ShouldNot(HaveOccurred())

				existKey := false
				for _, b := range envList.ToArray().ToStringArray() {
					if b == envName {
						existKey = true
					}
				}
				Expect(existKey).To(BeTrue())

				envKeyValue, err := oleutil.GetProperty(oleDispatch, "GetEnvironmentVariableFromKey", envName)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(envKeyValue.ToString()).To(Equal(envValue))

				oleDispatch.Release()

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("Filename", func() {
		Context("when set filename succeeds", func() {
			It("should set the right filename", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				filename := "test"
				cri.SetFilename(filename)
				Expect(cri.GetFilename()).To(Equal(filename))

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("CurrentDirectory", func() {
		Context("when set filename succeeds", func() {
			It("should set the right filename", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				cd := "test"
				cri.SetCurrentDirectory(cd)
				Expect(cri.GetCurrentDirectory()).To(Equal(cd))

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("Arguments", func() {
		Context("when set arguments succeeds", func() {
			It("should set the right arguments", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				arguments := "arg1"
				cri.SetArguments(arguments)
				Expect(cri.GetArguments()).To(Equal(arguments))

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("StreamingIn", func() {
		Context("when streaming succeeds", func() {
			It("shoul read Stdin pipe", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				reader, err := cri.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				pipeConn, castOk := reader.(*npipe.PipeConn)
				Expect(castOk).To(Equal(true))

				pipeName := pipeConn.LocalAddr().String()
				Expect(pipeConn.RemoteAddr().String()).To(Equal(pipeName))
				Expect(strings.HasPrefix(pipeName, `\\.\pipe\`)).To(BeTrue())
				Expect(strings.HasSuffix(pipeName, "stdin")).To(BeTrue())

				err = pipeConn.Close()
				Expect(err).ShouldNot(HaveOccurred())

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("StreamingOut", func() {
		Context("when streaming succeeds", func() {
			It("shoul read Stdout pipe", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				reader, err := cri.StdoutPipe()
				Expect(err).ShouldNot(HaveOccurred())

				pipeConn, castOk := reader.(*npipe.PipeConn)
				Expect(castOk).To(Equal(true))

				pipeName := pipeConn.LocalAddr().String()
				Expect(pipeConn.RemoteAddr().String()).To(Equal(pipeName))
				Expect(strings.HasPrefix(pipeName, `\\.\pipe\`)).To(BeTrue())
				Expect(strings.HasSuffix(pipeName, "stdout")).To(BeTrue())

				err = pipeConn.Close()
				Expect(err).ShouldNot(HaveOccurred())

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("StreamingError", func() {
		Context("when streaming succeeds", func() {
			It("shoul read Stderr pipe", func() {
				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cri).ShouldNot(BeNil())

				reader, err := cri.StderrPipe()
				Expect(err).ShouldNot(HaveOccurred())

				pipeConn, castOk := reader.(*npipe.PipeConn)
				Expect(castOk).To(Equal(true))

				pipeName := pipeConn.LocalAddr().String()
				Expect(pipeConn.RemoteAddr().String()).To(Equal(pipeName))
				Expect(strings.HasPrefix(pipeName, `\\.\pipe\`)).To(BeTrue())
				Expect(strings.HasSuffix(pipeName, "stderr")).To(BeTrue())

				err = pipeConn.Close()
				Expect(err).ShouldNot(HaveOccurred())

				err = cri.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
