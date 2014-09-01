package prison_client_test

import (
	"bufio"
	"fmt"
	"github.com/mattn/go-ole"
	PrisonClient "github.com/uhurusoftware/warden-windows/prison_client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strings"
)

var _ = Describe("Container", func() {
	BeforeEach(func() {
		ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	})

	Describe("CreateContainer", func() {
		Context("when no error occurs", func() {
			It("should create the container", func() {
				container, err := PrisonClient.CreateContainer()

				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("GetContainerId", func() {
		Context("when no error occurs", func() {
			It("should get the container id", func() {
				container, err := PrisonClient.CreateContainer()

				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())
				Expect(container.Id()).ShouldNot(BeEmpty())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("HomePath", func() {
		Context("when set homepath succeeds", func() {
			It("should set the right homepath", func() {
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				filename := "test"
				container.SetHomePath(filename)
				Expect(container.GetHomePath()).To(Equal(filename))

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("MemoryLimitBytes", func() {
		Context("when set memory limit bytes succeeds", func() {
			It("should set the right memory limit bytes", func() {
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				var memoryLimit int64 = 128
				container.SetMemoryLimitBytes(memoryLimit)
				Expect(container.GetMemoryLimitBytes()).To(Equal(memoryLimit))

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("DiskLimitBytes", func() {
		Context("when set disk limit bytes succeeds", func() {
			It("should set the right disk limit bytes", func() {
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				var diskLimit int64 = 40000000
				container.SetDiskLimitBytes(diskLimit)
				Expect(container.GetDiskLimitBytes()).To(Equal(diskLimit))

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("NetworkPort", func() {
		Context("when set network port succeeds", func() {
			It("should set the right network port", func() {
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				port := 9000
				container.SetNetworkPort(port)
				Expect(container.GetNetworkPort()).To(Equal(uint32(port)))

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("Run", func() {
		Context("when no error occurs", func() {
			It("should run the container", func() {
				defer GinkgoRecover()
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				containerId := container.Id()
				containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
				os.MkdirAll(containerPath, 0700)
				container.SetHomePath(containerPath)

				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				cri.SetFilename("c:\\Windows\\system32\\cmd.exe")
				cri.SetArguments("/c echo test")

				reader, err := cri.StdoutPipe()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Lockdown()
				Expect(err).ShouldNot(HaveOccurred())

				process, err := container.Run(cri)
				Expect(err).ShouldNot(HaveOccurred())

				go func() {
					outputOk := false
					r := bufio.NewReader(reader)

					for {
						x, _, err := r.ReadLine()
						if err != nil {
							break
						}
						outputOk = (outputOk || strings.EqualFold(string(x), "test"))
					}

					Expect(outputOk).To(BeTrue())
				}()

				exitCode, err := process.Wait()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(exitCode).To(BeZero())

				err = process.Release()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Stop()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Destroy()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("should run the container with stdin input", func() {
				defer GinkgoRecover()
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				containerId := container.Id()
				containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
				os.MkdirAll(containerPath, 0700)
				container.SetHomePath(containerPath)

				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				cri.SetFilename("c:\\Windows\\system32\\cmd.exe")
				cri.SetArguments("/c more")

				writer, err := cri.StdinPipe()
				Expect(err).ShouldNot(HaveOccurred())

				reader, err := cri.StdoutPipe()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Lockdown()
				Expect(err).ShouldNot(HaveOccurred())

				process, err := container.Run(cri)
				Expect(err).ShouldNot(HaveOccurred())

				go func() {
					defer GinkgoRecover()

					outputOk := false
					r := bufio.NewReader(reader)

					w := bufio.NewWriter(writer)
					w.WriteString("from stdin")
					w.Flush()
					writer.Close()

					for {
						x, _, err := r.ReadLine()
						if err != nil {
							break
						}
						outputOk = (outputOk || strings.EqualFold(string(x), "from stdin"))
					}

					Expect(outputOk).To(BeTrue())
				}()

				exitCode, err := process.Wait()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(exitCode).To(BeZero())

				err = process.Release()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Stop()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Destroy()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("should return the right exit code", func() {
				defer GinkgoRecover()
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				containerId := container.Id()
				containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
				os.MkdirAll(containerPath, 0700)
				container.SetHomePath(containerPath)

				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				cri.SetFilename("c:\\Windows\\system32\\cmd.exe")
				cri.SetArguments("/c exit 123")

				err = container.Lockdown()
				Expect(err).ShouldNot(HaveOccurred())

				process, err := container.Run(cri)
				Expect(err).ShouldNot(HaveOccurred())

				exitCode, err := process.Wait()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(exitCode).To(Equal(123))

				err = process.Release()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Destroy()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when error occurs", func() {
			It("should run the container and put error in stderr", func() {
				defer GinkgoRecover()
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				containerId := container.Id()
				containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
				os.MkdirAll(containerPath, 0700)
				container.SetHomePath(containerPath)

				cri, err := PrisonClient.CreateContainerRunInfo()
				Expect(err).ShouldNot(HaveOccurred())
				cri.SetFilename("c:\\Windows\\system32\\cmd.exe")
				cri.SetArguments("/c test")

				reader, err := cri.StderrPipe()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Lockdown()
				Expect(err).ShouldNot(HaveOccurred())

				process, err := container.Run(cri)
				Expect(err).ShouldNot(HaveOccurred())

				go func() {
					outputOk := false
					r := bufio.NewReader(reader)

					for {
						x, _, err := r.ReadLine()
						if err != nil {
							break
						}
						outputOk = (outputOk || strings.Contains(string(x), "is not recognized as an internal or external command"))
					}

					Expect(outputOk).To(BeTrue())
				}()

				exitCode, err := process.Wait()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(exitCode).To(Equal(1))

				err = process.Release()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Stop()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Destroy()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("Lockdown", func() {
		Context("when no error occurs", func() {
			It("should lockdown the container", func() {
				defer GinkgoRecover()
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				containerId := container.Id()
				containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
				os.MkdirAll(containerPath, 0700)
				container.SetHomePath(containerPath)

				err = container.Lockdown()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container.IsLockedDown()).To(BeTrue())

				err = container.Destroy()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		It("should error prison already locked", func() {
			defer GinkgoRecover()
			container, err := PrisonClient.CreateContainer()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(container).ShouldNot(BeNil())

			containerId := container.Id()
			containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
			os.MkdirAll(containerPath, 0700)
			container.SetHomePath(containerPath)

			err = container.Lockdown()
			Expect(err).ShouldNot(HaveOccurred())

			err = container.Lockdown()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("This prison is already locked."))

			err = container.Release()
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should error prison has to be locked", func() {
			defer GinkgoRecover()
			container, err := PrisonClient.CreateContainer()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(container).ShouldNot(BeNil())

			containerId := container.Id()
			containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
			os.MkdirAll(containerPath, 0700)
			container.SetHomePath(containerPath)

			cri, err := PrisonClient.CreateContainerRunInfo()
			Expect(err).ShouldNot(HaveOccurred())

			process, err := container.Run(cri)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("This prison has to be locked before you can use it."))
			Expect(process).Should(BeNil())

			err = container.Release()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Describe("Release", func() {
		Context("when error occurs", func() {
			It("should error container already released", func() {
				defer GinkgoRecover()
				container, err := PrisonClient.CreateContainer()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(container).ShouldNot(BeNil())

				containerId := container.Id()
				containerPath := fmt.Sprintf("c:\\prison_tests\\%s", containerId)
				os.MkdirAll(containerPath, 0700)
				container.SetHomePath(containerPath)

				err = container.Lockdown()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).ShouldNot(HaveOccurred())

				err = container.Release()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("Container is already released"))
			})
		})
	})
})
