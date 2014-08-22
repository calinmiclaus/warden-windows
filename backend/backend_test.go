package backend_test

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudfoundry-incubator/cf-lager"
	WardenClient "github.com/cloudfoundry-incubator/garden/client"
	WardenConnection "github.com/cloudfoundry-incubator/garden/client/connection"
	Garden "github.com/cloudfoundry-incubator/garden/server"
	Warden "github.com/cloudfoundry-incubator/garden/warden"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/uhurusoftware/warden-windows/backend"
)

var _ = Describe("Backend", func() {
	var (
		s   *Garden.WardenServer
		err error
	)

	BeforeSuite(func() {
		defer GinkgoRecover()
		backend := New("c:\\warden", nil)
		logger := cf_lager.New("warden-winodws")
		serverLogger := logger.Session("garden")
		s = Garden.New("tcp", "0.0.0.0:9877", 0, backend, serverLogger)

		err = s.Start()
		Ω(err).ShouldNot(HaveOccurred())

	})

	AfterSuite(func() {
		defer GinkgoRecover()
		s.Stop()

	})

	Describe("connect to containers", func() {
		var (
			connection WardenConnection.Connection
			client     Warden.Client
			container  Warden.Container
			tempDir    string
		)

		BeforeEach(func() {
			defer GinkgoRecover()
			connection = WardenConnection.New("tcp", "127.0.0.1:9877")
			client = WardenClient.New(connection)
			containerSpec := Warden.ContainerSpec{}
			container, err = client.Create(containerSpec)
			Ω(err).ShouldNot(HaveOccurred())

			tempDir = rand_str(10)
			err = os.MkdirAll(tempDir, 0666)
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			defer GinkgoRecover()
			client.Destroy(container.Handle())
			os.RemoveAll(tempDir)
		})

		It("should run ping inside container", func() {
			stdinRead, _ := io.Pipe()
			stdoutRead, stdoutWrite := io.Pipe()
			stderrRead, stderrWrite := io.Pipe()
			processIo := Warden.ProcessIO{Stdin: stdinRead, Stdout: stdoutWrite, Stderr: stderrWrite}

			processSpec := Warden.ProcessSpec{Path: "C:\\Windows\\System32\\ping.exe", Args: []string{"127.0.0.1"}}
			process, err := container.Run(processSpec, processIo)
			Ω(err).ShouldNot(HaveOccurred())

			go func() {
				r := bufio.NewReader(stdoutRead)

				for {
					x, _, err := r.ReadLine()
					if err != nil {
						Ω(err).ShouldNot(HaveOccurred())
					}
					fmt.Println(string(x))
				}
			}()

			go func() {
				r := bufio.NewReader(stderrRead)

				for {
					x, _, err := r.ReadLine()
					if err != nil {
						Ω(err).ShouldNot(HaveOccurred())
					}
					fmt.Println(string(x))
				}
			}()

			process.Wait()

		})

		It("should run cmd and echo something", func() {
			stdinRead, stdinWrite := io.Pipe()
			stdoutRead, stdoutWrite := io.Pipe()
			stderrRead, stderrWrite := io.Pipe()
			processIo := Warden.ProcessIO{Stdin: stdinRead, Stdout: stdoutWrite, Stderr: stderrWrite}
			processSpec := Warden.ProcessSpec{Path: "C:\\Windows\\System32\\cmd.exe"}
			_, err := container.Run(processSpec, processIo)
			Ω(err).ShouldNot(HaveOccurred())

			w := bufio.NewWriter(stdinWrite)
			fmt.Fprint(w, "echo ###test###\n")
			w.Flush()

			go func() {
				defer GinkgoRecover()
				defer container.Stop(true)
				r := bufio.NewReader(stderrRead)

				for {
					x, _, err := r.ReadLine()
					if err != nil {
						Ω(err).ShouldNot(HaveOccurred())
					}
					fmt.Println(string(x))
				}
			}()

			timeout := make(chan bool)
			success := make(chan bool)

			go func() {
				time.Sleep(5 * time.Second)
				timeout <- true
			}()

			r := bufio.NewReader(stdoutRead)

			go func() {
				for {

					x, _, err := r.ReadLine()
					if err != nil {
						Ω(err).ShouldNot(HaveOccurred())
					}
					if x != nil {
						result := string(x)
						if result == "###test###" {
							success <- true
						}
					}

					fmt.Println(string(x))
				}
			}()

			for {
				select {
				case <-timeout:
					{
						Fail("timeout")
					}
				case <-success:
					{
						return
					}
				default:
				}
			}
		})

		It("should stream in", func() {
			destPath := "streamintest"
			testString := "Hello cartridge"
			reader, err := GetTarStreamFromString(testString, tempDir, destPath)
			Ω(err).ShouldNot(HaveOccurred())
			err = container.StreamIn(destPath, reader)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should stream out", func() {
			destPath := "streamouttest"
			testString := "Hello cartridge"

			reader, err := GetTarStreamFromString(testString, tempDir, destPath)
			Ω(err).ShouldNot(HaveOccurred())
			err = container.StreamIn(destPath, reader)
			Ω(err).ShouldNot(HaveOccurred())

			reader, err = container.StreamOut(destPath)
			result, err := GetStringFromTarStream(reader, tempDir, destPath)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal(testString))

		})
	})
})

func GetTarStreamFromString(text string, tempDir string, filename string) (io.Reader, error) {

	fmt.Println(tempDir)
	f, err := os.Create(filepath.Join(tempDir, filename))
	if err != nil {
		return nil, err
	}
	_, err = io.WriteString(f, text)
	if err != nil {
		return nil, err
	}
	f.Close()

	workingDir := tempDir
	compressArg := filename

	tarRead, tarWrite := io.Pipe()

	tarPath := "C:\\Program Files (x86)\\Git\\bin\\tar.exe"
	cmdPath := "C:\\Windows\\System32\\cmd.exe"

	cmd := &exec.Cmd{
		Path: cmdPath,
		// Dir:  workingDir,
		Args: []string{
			"/c",
			tarPath,
			"cf",
			"-",
			"-C",
			workingDir,
			compressArg,
		},
		Stdout: tarWrite,
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		tarWrite.Close()
	}()

	return tarRead, nil
}

func GetStringFromTarStream(source io.Reader, tempDir string, filename string) (string, error) {
	absDestPath := filepath.Join(tempDir, "result")

	err := os.MkdirAll(absDestPath, 0777)
	if err != nil {
		return "", err
	}

	tarPath := "C:\\Program Files (x86)\\Git\\bin\\tar.exe"
	cmdPath := "C:\\Windows\\System32\\cmd.exe"

	cmd := &exec.Cmd{
		Path: cmdPath,
		Dir:  absDestPath,
		Args: []string{
			"/c",
			tarPath,
			"xf",
			"-",
			"-C",
			"./",
		},
		Stdin: source,
	}

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadFile(filepath.Join(absDestPath, filename, filename))
	if err != nil {
		return "", err
	}
	value := string(buf)

	return value, nil
}

func rand_str(str_size int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
