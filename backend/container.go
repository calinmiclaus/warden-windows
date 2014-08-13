package backend

import "C"

import (
	"strconv"
	"sync"
	//"syscall"
	//"fmt"
	//"bufio"
	"io"
	"log"
	"net/rpc"
	"path"
	"path/filepath"

	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/garden/warden"
	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/UhuruSoftware/warden-windows/prison_client"
)

type Container struct {
	id     string
	handle string

	rootPath string

	runner command_runner.CommandRunner

	rpc           *rpc.Client
	pids          map[int]*prison_client.ProcessTracker
	lastNetInPort uint32
	prison        *prison_client.Container
	runMutex      *sync.Mutex
}

func NewContainer(
	id, handle string,
	rootPath string,
	runner command_runner.CommandRunner,
) *Container {
	container, err := prison_client.CreateContainer()
	if err != nil {
		log.Fatal(err)
	}

	return &Container{
		id:     id,
		handle: handle,

		rootPath: rootPath,

		runner:   runner,
		pids:     make(map[int]*prison_client.ProcessTracker),
		prison:   container,
		runMutex: &sync.Mutex{},
	}
}

func (container *Container) Destroy() error {
	defer container.prison.Release()

	blocked := container.prison.IsLockedDown()

	if blocked == true {

		log.Println("Invoking destory on prison")
		err := container.prison.Destroy()
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("Container destoryed: ", container.id)
	}
	return nil
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

func (container *Container) Stop(kill bool) error {
	container.runMutex.Lock()
	defer container.runMutex.Unlock()

	//ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)

	blocked := container.prison.IsLockedDown()

	if blocked == true {
		log.Println("Stop with kill", kill)
		err := container.prison.Stop()
		if err != nil {
			log.Println(err)
		}

		if kill == true {
			container.Destroy()
		}

		return err
	}

	//containers := container.pids
	//container.pids = make(map[int]*exec.Cmd)

	//for pid := range containers {
	//	log.Println("Stopping: ", strconv.Itoa(pid))

	//	stopPath := "C:\\Users\\stefan.schneider\\gopath\\src\\dispatch-ctrl-c\\dispatch-ctrl-c.exe"

	//	cmd := &exec.Cmd{
	//		Path: stopPath,
	//		Args: []string{
	//			stopPath, strconv.Itoa(pid), "1",
	//		},
	//		SysProcAttr: &syscall.SysProcAttr{
	//			// setting this flag will not stop the warden process when sending console ctrl envets to child processes
	//			// note: syscall.CREATE_NEW_PROCESS_GROUP is useless :/ 0x00000200
	//			// CREATE_NEW_CONSOLE 0x00000010
	//			// CREATE_NO_WINDOW 0x08000000
	//			// DETACHED_PROCESS 0x00000008
	//			CreationFlags: 0x00000010,
	//		},
	//	}

	//	out, err := cmd.Output()
	//	if err != nil {
	//		log.Println(err)
	//		log.Println(string(out))
	//	}
	//	//log.Println(string(out))
	//}
	return nil
}

func (container *Container) Info() (warden.ContainerInfo, error) {
	log.Println("TODO Info")
	return warden.ContainerInfo{Events: []string{"party"}, ProcessIDs: []uint32{}, MappedPorts: []warden.PortMapping{}}, nil
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

/*
type wprocess struct {
	pt *ole.IDispatch
}

func newWprocess(pt *ole.IDispatch) *wprocess {
	pt.AddRef() // call addref because we clone the pointr
	ret := &wprocess{pt: pt}
	runtime.SetFinalizer(ret, func(t *wprocess) { log.Println("ProcessTracker ref count after finalizer: ", t.pt.Release()) })
	return ret
}

func (wp *wprocess) ID() uint32 {
	iDpid, errr := oleutil.CallMethod(wp.pt, "GetPid")
	if errr != nil {
		log.Println(errr)
	}
	defer iDpid.Clear()

	return uint32(iDpid.Value().(int64))
}

func (wp *wprocess) Wait() (int, error) {
	_, errr := oleutil.CallMethod(wp.pt, "Wait")
	if errr != nil {
		log.Println(errr)
		return 0, errr
	}

	exitCode, errr := oleutil.CallMethod(wp.pt, "GetExitCode")
	if errr != nil {
		log.Println(errr)
		return 0, errr
	}
	defer exitCode.Clear()

	return int(exitCode.Value().(int64)), nil
}

func (wp *wprocess) SetTTY(warden.TTYSpec) error {
	log.Println("TODO SetTTY")
	return nil
}
*/

func (container *Container) Run(spec warden.ProcessSpec, pio warden.ProcessIO) (warden.Process, error) {
	container.runMutex.Lock()
	defer container.runMutex.Unlock()
	// ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)

	log.Println("Run command: ", spec.Path, spec.Args, spec.Dir, spec.Privileged, spec.Env)

	cmdPath := "C:\\Windows\\System32\\cmd.exe"
	rootPath := path.Join(container.rootPath, container.handle)
	strings.Replace(rootPath, "/", "\\", -1)

	spec.Dir = path.Join(rootPath, spec.Dir)
	spec.Path = path.Join(rootPath, spec.Path)

	envs := spec.Env
	// TOTD: remove this (HACK?!) port overriding
	// after somebody cleans up this hardcoded values: https://github.com/cloudfoundry-incubator/app-manager/blob/master/start_message_builder/start_message_builder.go#L182
	envs = append(envs, "NETIN_PORT="+strconv.FormatUint(uint64(container.lastNetInPort), 10))

	cri, err := prison_client.CreateContainerRunInfo()
	defer cri.Release()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, env := range envs {
		spltiEnv := strings.SplitN(env, "=", 2)
		cri.AddEnvironemntVariable(spltiEnv[0], spltiEnv[1])
	}

	concatArgs := ""
	for _, v := range spec.Args {
		concatArgs = concatArgs + " " + v
	}

	spec.Path = strings.Replace(spec.Path, "/", "\\", -1)

	concatArgs = " /c " + spec.Path + " " + concatArgs
	log.Println("Filename ", spec.Path, "Arguments: ", concatArgs, "Concat Args: ", concatArgs)

	cri.SetFilename(cmdPath)
	cri.SetArguments(concatArgs)

	stdinWriter, err := cri.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdoutReader, err := cri.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrReader, err := cri.StderrPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		log.Println("Streaming stdout ", stdoutReader)

		io.Copy(pio.Stdout, stdoutReader)
		stdoutReader.Close()

		log.Println("Stdout pipe closed", stdoutReader)
	}()

	go func() {
		log.Println("Streaming stderr ", stderrReader)

		io.Copy(pio.Stderr, stderrReader)
		stderrReader.Close()

		log.Println("Stderr pipe closed", stderrReader)
	}()

	go func() {
		log.Println("Streaming stdin ", stdinWriter)

		io.Copy(stdinWriter, pio.Stdin)
		stdinWriter.Close()

		log.Println("Stdin pipe closed", stdinWriter)
	}()

	blocked := container.prison.IsLockedDown()

	if blocked == false {

		container.prison.SetHomePath(rootPath)
		// oleutil.PutProperty(container, "MemoryLimitBytes", 1024*1024*300)

		log.Println("Locking down...")
		err = container.prison.Lockdown()
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println("Locked down.")
	}

	log.Println("Running process...")
	pt, err := container.prison.Run(cri)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	//pid := pt.ID()
	//container.pids[pid] = pt
	//go func() {
	//	pt.Wait()
	//	delete(container.pids, pid)
	//}()

	return pt, nil
}

func (container *Container) Attach(processID uint32, pio warden.ProcessIO) (warden.Process, error) {
	log.Println("Attaching to: ", processID)

	cmd := container.pids[int(processID)]

	return cmd, nil
}

func (container *Container) NetIn(hostPort uint32, containerPort uint32) (uint32, uint32, error) {
	log.Println("TODO NetIn", hostPort, containerPort)
	freePort := freeTcp4Port()
	container.lastNetInPort = freePort
	return freePort, containerPort, nil
	//return freePort, freePort, nil
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
