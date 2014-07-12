package backend

import (
	"strconv"
	"syscall"
	//"fmt"
	"bufio"
	"io"
	"log"
	"net/rpc"
	"path"
	"path/filepath"
	"strings"
	//"net/rpc/jsonrpc"
	"net"
	"os"
	"os/exec"
	"time"

	//"github.com/cloudfoundry-incubator/warden-windows/backend/messages"

	"github.com/cloudfoundry-incubator/garden/warden"
	"github.com/cloudfoundry/gunk/command_runner"

	//"github.com/cloudfoundry-incubator/warden-windows/backend/iorpc"
	//"github.com/cloudfoundry-incubator/warden-windows/backend/linecodec"
)

type Container struct {
	id     string
	handle string

	rootPath string

	runner command_runner.CommandRunner

	rpc           *rpc.Client
	pids          map[int]*exec.Cmd
	lastNetInPort uint32
}

func NewContainer(
	id, handle string,
	rootPath string,
	runner command_runner.CommandRunner,
) *Container {
	return &Container{
		id:     id,
		handle: handle,

		rootPath: rootPath,

		runner: runner,
		pids:   make(map[int]*exec.Cmd),
	}
}

func (container *Container) ID() string {
	return container.id
}

func (container *Container) Handle() string {
	return container.handle
}

func (container *Container) GraceTime() time.Duration {
	log.Println("TODO GraceTime")
	return 0
}

func (container *Container) Start(containerBinaryPath string) error {
	//daemon := &exec.Cmd{
	//	Path: containerBinaryPath,
	//	Args: []string{"--handle", container.handle, "--rootPath", container.rootPath},
	//}

	//stdin, err := daemon.StdinPipe()
	//if err != nil {
	//	return err
	//}

	//stdout, err := daemon.StdoutPipe()
	//if err != nil {
	//	return err
	//}

	//stderr, err := daemon.StderrPipe()
	//if err != nil {
	//	return err
	//}

	//conn := iorpc.New(stdin, stdout)

	//container.rpc = rpc.NewClientWithCodec(
	//	linecodec.New(
	//		stdin,
	//		jsonrpc.NewClientCodec(conn),
	//	),
	//)

	//container.muxer.SetSource(stderr)

	//err = container.runner.Start(daemon)
	//if err != nil {
	//	return err
	//}

	log.Println("TODO Start")
	return nil
}

func (container *Container) Stop(kill bool) error {
	//return container.rpc.Call(
	//	"Container.Stop",
	//	&messages.StopRequest{Kill: kill},
	//	&messages.StopResponse{},
	//)
	log.Println("TODO Stop")

	containers := container.pids
	container.pids = make(map[int]*exec.Cmd)

	for pid := range containers {
		log.Println("Stopping: ", strconv.Itoa(pid))

		stopPath := "C:\\Users\\stefan.schneider\\gopath\\src\\dispatch-ctrl-c\\dispatch-ctrl-c.exe"

		cmd := &exec.Cmd{
			Path: stopPath,
			Args: []string{
				stopPath, strconv.Itoa(pid), "1",
			},
			SysProcAttr: &syscall.SysProcAttr{
				// setting this flag will not stop the warden process when sending console ctrl envets to child processes
				// note: syscall.CREATE_NEW_PROCESS_GROUP is useless :/ 0x00000200
				// CREATE_NEW_CONSOLE 0x00000010
				// CREATE_NO_WINDOW 0x08000000
				// DETACHED_PROCESS 0x00000008
				CreationFlags: 0x00000010,
			},
		}

		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
		}
		log.Println(string(out))
	}
	return nil
}

func (container *Container) Info() (warden.ContainerInfo, error) {
	log.Println("TODO Info")
	return warden.ContainerInfo{Events: []string{}, ProcessIDs: []uint32{}, MappedPorts: []warden.PortMapping{}}, nil
}

func (container *Container) StreamIn(dstPath string, source io.Reader) error {
	log.Println("StreamIn dstPath:", dstPath)

	absDestPath := path.Join(container.rootPath, container.handle, dstPath)
	log.Println("Stremming in to file: ", absDestPath)

	//err := os.MkdirAll(path.Dir(absDestPath), 0777)
	err := os.MkdirAll(absDestPath, 0777)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}

	return err
}

func (container *Container) StreamOut(srcPath string) (io.ReadCloser, error) {
	log.Println("StreamOut srcPath:", srcPath)

	containerPath := path.Join(container.rootPath, container.handle)

	workingDir := filepath.Dir(srcPath)
	compressArg := filepath.Base(srcPath)
	if strings.HasSuffix(srcPath, "/") {
		workingDir = srcPath
		compressArg = "."
	}

	workingDir = path.Join(containerPath, workingDir)

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

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		tarWrite.Close()
	}()

	return tarRead, nil
}

func (container *Container) LimitBandwidth(limits warden.BandwidthLimits) error {
	log.Println("TODO LimitBandwidth")
	return nil
}

func (container *Container) CurrentBandwidthLimits() (warden.BandwidthLimits, error) {
	log.Println("TODO CurrentBandwidthLimits")
	return warden.BandwidthLimits{}, nil
}

func (container *Container) LimitDisk(limits warden.DiskLimits) error {
	log.Println("TODO LimitDisk")
	return nil
}

func (container *Container) CurrentDiskLimits() (warden.DiskLimits, error) {
	log.Println("TODO CurrentDiskLimits")
	return warden.DiskLimits{}, nil
}

func (container *Container) LimitMemory(limits warden.MemoryLimits) error {
	log.Println("TODO LimitMemory")
	return nil
}

func (container *Container) CurrentMemoryLimits() (warden.MemoryLimits, error) {
	log.Println("TODO CurrentMemoryLimits")
	return warden.MemoryLimits{}, nil
}

func (container *Container) LimitCPU(limits warden.CPULimits) error {
	log.Println("TODO LimitCPU")
	return nil
}

func (container *Container) CurrentCPULimits() (warden.CPULimits, error) {
	log.Println("TODO CurrentCPULimits")
	return warden.CPULimits{}, nil
}

type wprocess struct {
	p *exec.Cmd
}

func newWprocess(p *exec.Cmd) *wprocess {
	return &wprocess{p: p}
}

func (wp *wprocess) ID() uint32 {
	return uint32(wp.p.Process.Pid)
}

func (wp *wprocess) Wait() (int, error) {
	err := wp.p.Wait()
	exitStatus := uint32(0)

	if err != nil {
		exiterr, _ := err.(*exec.ExitError)

		// The program has exited with an exit code != 0

		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			exitStatus = uint32(status.ExitStatus())
		}
	}
	return int(exitStatus), nil
}

func (container *Container) Run(spec warden.ProcessSpec, pio warden.ProcessIO) (warden.Process, error) {
	//var response messages.RunResponse

	//err := container.rpc.Call(
	//	"Container.Run",
	//	&messages.RunRequest{
	//		Script:     spec.Script,
	//		Privileged: spec.Privileged,
	//	},
	//	&response,
	//)
	//if err != nil {
	//	return 0, nil, err
	//}

	//stream := make(chan warden.ProcessStream, 1000)

	//container.muxer.Subscribe(response.ProcessID, stream)

	//return response.ProcessID, stream, nil

	log.Println("Run command: ", spec.Path, spec.Args, spec.Dir, spec.Privileged, spec.Env)

	cmdPath := "C:\\Windows\\System32\\cmd.exe"
	rootPath := path.Join(container.rootPath, container.handle)

	spec.Dir = path.Join(rootPath, spec.Dir)
	spec.Path = path.Join(rootPath, spec.Path)

	envs := os.Environ()
	envs = append(envs, spec.Env...)
	// TOTD: remove this (HATCK?!) port overriding
	// after somebody cleans up this hardcoded values: https://github.com/cloudfoundry-incubator/app-manager/blob/master/start_message_builder/start_message_builder.go#L182
	envs = append(envs, "NETIN_PORT="+strconv.FormatUint(uint64(container.lastNetInPort), 10))

	command := &exec.Cmd{
		Path: cmdPath,
		// Dir:  rootPath,
		Dir: spec.Dir,
		Env: envs,
		Args: append(
			[]string{
				"/c",
				"< nul", // safety: < nul will prevent the "Terminate batch job" prompt
				spec.Path,
			},
			spec.Args...,
		),
		SysProcAttr: &syscall.SysProcAttr{
			// setting this flag will not stop the warden process when sending console ctrl envets to child processes
			// note: syscall.CREATE_NEW_PROCESS_GROUP is useless :/ 0x00000200
			// CREATE_NEW_CONSOLE 0x00000010
			// CREATE_NO_WINDOW 0x08000000
			// DETACHED_PROCESS 0x00000008
			CreationFlags: 0x00000010,
		},
		// Stdin: source,
	}

	// https://github.com/jnwhiteh/golang/blob/master/src/pkg/syscall/syscall_windows.go#L434
	errp, _ := command.StderrPipe()
	outp, _ := command.StdoutPipe()
	berrp := bufio.NewScanner(errp)
	boutp := bufio.NewScanner(outp)

	go func() {
		for berrp.Scan() {
			pio.Stdout.Write([]byte(berrp.Text() + "\n"))
			//stream <- warden.ProcessStream{
			//	Source: warden.ProcessStreamSourceStderr,
			//	Data:   []byte(berrp.Text() + "\n"),
			//	//ExitStatus: &exitStatus,
			//}

		}
	}()

	go func() {
		for boutp.Scan() {
			pio.Stderr.Write([]byte(berrp.Text() + "\n"))
			//stream <- warden.ProcessStream{
			//	Source: warden.ProcessStreamSourceStdout,
			//	Data:   []byte(boutp.Text() + "\n"),
			//	//ExitStatus: &exitStatus,
			//}

		}
	}()

	err := command.Start()
	if err != nil {
		log.Println(err)
	}

	pid := command.Process.Pid
	container.pids[pid] = command

	//go func() {
	//	err := command.Wait()
	//	exitStatus := uint32(0)

	//	if err != nil {
	//		exiterr, _ := err.(*exec.ExitError)

	//		// The program has exited with an exit code != 0

	//		// This works on both Unix and Windows. Although package
	//		// syscall is generally platform dependent, WaitStatus is
	//		// defined for both Unix and Windows and in both cases has
	//		// an ExitStatus() method with the same signature.
	//		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
	//			exitStatus = uint32(status.ExitStatus())
	//		}
	//	}

	//	//if _, ok := container.pids[pid]; ok {
	//	delete(container.pids, pid)
	//	exitStatus = uint32(0) //hack
	//	stream <- warden.ProcessStream{
	//		// Source:     ProcessStreamSourceInvalid,
	//		// Data:       nil,
	//		ExitStatus: &exitStatus,
	//	}
	//	log.Println("Sending exitStatus ", exitStatus, " for pid ", pid)
	//	//}
	//}()

	return newWprocess(command), nil
	// return uint32(pid), stream, nil
}

func (container *Container) Attach(processID uint32, pio warden.ProcessIO) (warden.Process, error) {
	log.Println("TODO Attach", processID)
	//stream := make(chan warden.ProcessStream, 1000)

	//container.muxer.Subscribe(processID, stream)

	//return stream, nil

	// stream := make(chan warden.ProcessStream, 1000)
	cmd := container.pids[int(processID)]
	var res *wprocess
	if cmd != nil {
		res = newWprocess(cmd)
	}
	return res, nil
}

func (container *Container) NetIn(hostPort uint32, containerPort uint32) (uint32, uint32, error) {
	log.Println("TODO NetIn", hostPort, containerPort)
	freePort := freeTcp4Port()
	container.lastNetInPort = freePort
	return freePort, freePort, nil
	//return hostPort, containerPort, nil
}

func (container *Container) NetOut(network string, port uint32) error {
	log.Println("TODO NetOut", network, port)
	return nil
}

func freeTcp4Port() uint32 {
	l, _ := net.Listen("tcp4", ":0")
	defer l.Close()
	freePort := strings.Split(l.Addr().String(), ":")[1]
	ret, _ := strconv.ParseUint(freePort, 10, 32)
	return uint32(ret)
}
