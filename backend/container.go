package backend

import (
	"runtime"
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
	"strings"
	//"net/rpc/jsonrpc"
	"net"
	"os"
	"os/exec"
	"time"

	//"github.com/cloudfoundry-incubator/warden-windows/backend/messages"

	"github.com/cloudfoundry-incubator/garden/warden"
	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/mattn/go-ole"
	"github.com/mattn/go-ole/oleutil"
	"github.com/natefinch/npipe"
)

type Container struct {
	id     string
	handle string

	rootPath string

	runner command_runner.CommandRunner

	rpc           *rpc.Client
	pids          map[int]*ole.IDispatch
	lastNetInPort uint32
	prison        *ole.IDispatch
	runMutex      *sync.Mutex
}

func NewContainer(
	id, handle string,
	rootPath string,
	runner command_runner.CommandRunner,
) *Container {
	iUcontainer, errr := oleutil.CreateObject("Uhuru.Prison.ComWrapper.Container")
	if errr != nil {
		log.Println(errr)
	}
	defer iUcontainer.Release()

	iDcontainer, errr := iUcontainer.QueryInterface(ole.IID_IDispatch)
	if errr != nil {
		log.Fatal(errr)
	}

	return &Container{
		id:     id,
		handle: handle,

		rootPath: rootPath,

		runner:   runner,
		pids:     make(map[int]*ole.IDispatch),
		prison:   iDcontainer,
		runMutex: &sync.Mutex{},
	}
}

func (container *Container) Destroy() error {
	defer container.prison.Release()

	isLocked, errr := oleutil.CallMethod(container.prison, "IsLockedDown")
	if errr != nil {
		log.Println(errr)
	}
	blocked := isLocked.Value().(bool)

	if blocked == true {
		_, errr = oleutil.CallMethod(container.prison, "Destroy")
		if errr != nil {
			log.Println(errr)
			return errr
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

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)

	log.Println("Stop with kill", kill)
	_, errr := oleutil.CallMethod(container.prison, "Stop")
	if errr != nil {
		log.Println(errr)
	}

	if kill == true {
		container.Destroy()
	}

	return errr

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

type wprocess struct {
	pt *ole.IDispatch
}

func newWprocess(pt *ole.IDispatch) *wprocess {
	return &wprocess{pt: pt}
}

func (wp *wprocess) ID() uint32 {
	iDpid, errr := oleutil.CallMethod(wp.pt, "GetPid")
	if errr != nil {
		log.Println(errr)
	}

	return uint32(iDpid.Value().(int64))
}

func (wp *wprocess) Wait() (int, error) {
	_, errr := oleutil.CallMethod(wp.pt, "Wait")
	if errr != nil {
		log.Println(errr)
	}

	exitCode, errr := oleutil.CallMethod(wp.pt, "GetExitCode")
	if errr != nil {
		log.Println(errr)
	}

	return int(exitCode.Value().(int64)), nil
}

func (container *Container) Run(spec warden.ProcessSpec, pio warden.ProcessIO) (warden.Process, error) {
	container.runMutex.Lock()
	defer container.runMutex.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)

	log.Println("Run command: ", spec.Path, spec.Args, spec.Dir, spec.Privileged, spec.Env)

	cmdPath := "C:\\Windows\\System32\\cmd.exe"
	rootPath := path.Join(container.rootPath, container.handle)
	strings.Replace(rootPath, "/", "\\", -1)

	spec.Dir = path.Join(rootPath, spec.Dir)
	spec.Path = path.Join(rootPath, spec.Path)

	envs := spec.Env
	// TOTD: remove this (HATCK?!) port overriding
	// after somebody cleans up this hardcoded values: https://github.com/cloudfoundry-incubator/app-manager/blob/master/start_message_builder/start_message_builder.go#L182
	envs = append(envs, "NETIN_PORT="+strconv.FormatUint(uint64(container.lastNetInPort), 10))

	IUcri, errr := oleutil.CreateObject("Uhuru.Prison.ComWrapper.ContainerRunInfo")
	if errr != nil {
		log.Println(errr)
		return nil, errr
	}
	defer IUcri.Release()

	cri, errr := IUcri.QueryInterface(ole.IID_IDispatch)
	if errr != nil {
		log.Println(errr)
		return nil, errr
	}
	defer cri.Release()

	for _, env := range envs {
		spltiEnv := strings.SplitN(env, "=", 2)

		_, errr = oleutil.CallMethod(cri, "AddEnvironemntVariable", spltiEnv[0], spltiEnv[1])
		if errr != nil {
			log.Println(errr)
			return nil, errr
		}
	}

	concatArgs := ""
	for _, v := range spec.Args {
		concatArgs = concatArgs + " " + v
	}

	spec.Path = strings.Replace(spec.Path, "/", "\\", -1)
	log.Println("Filename ", spec.Path, "Arguments: ", concatArgs)
	concatArgs = " /c " + spec.Path + " " + concatArgs

	//oleutil.PutProperty(cri, "Filename", spec.Path)
	oleutil.PutProperty(cri, "Filename", cmdPath)
	oleutil.PutProperty(cri, "Arguments", concatArgs)

	//command := &exec.Cmd{
	//	Path: cmdPath,
	//	// Dir:  rootPath,
	//	Dir: spec.Dir,
	//	Env: envs,
	//	Args: append(
	//		[]string{
	//			"/c",
	//			"< nul", // safety: < nul will prevent the "Terminate batch job" prompt
	//			spec.Path,
	//		},
	//		spec.Args...,
	//	),
	//}

	//// https://github.com/jnwhiteh/golang/blob/master/src/pkg/syscall/syscall_windows.go#L434
	//inp, _ := command.StdinPipe()
	//errp, _ := command.StderrPipe()
	//outp, _ := command.StdoutPipe()

	//go func() {
	//	io.Copy(inp, pio.Stdin)
	//}()

	//go func() {
	//	io.Copy(pio.Stdout, outp)
	//}()

	//go func() {
	//	io.Copy(pio.Stderr, errp)
	//}()

	stdinPipe := oleutil.MustCallMethod(cri, "RedirectStdin", true).ToString()
	stdoutPipe := oleutil.MustCallMethod(cri, "RedirectStdout", true).ToString()
	stderrPipe := oleutil.MustCallMethod(cri, "RedirectStderr", true).ToString()

	go func() {
		conn, err := npipe.Dial(`\\.\pipe\` + stdoutPipe)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Connected to pipe ", conn)
		io.Copy(pio.Stdout, conn)
	}()

	go func() {
		conn, err := npipe.Dial(`\\.\pipe\` + stderrPipe)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Connected to pipe ", conn)
		io.Copy(pio.Stdout, conn)
	}()

	go func() {

		conn, err := npipe.Dial(`\\.\pipe\` + stdinPipe)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Connected to pipe ", conn)
		io.Copy(conn, pio.Stdin)
		conn.Close()
	}()

	//go ListenOnPipe(`\\.\pipe\` + stdoutPipe)
	//go ListenOnPipe(`\\.\pipe\` + stderrPipe)
	//go WriteOnPipe(`\\.\pipe\`+stdinPipe, "tralalal\nsdfg\nsdfg")

	isLocked, errr := oleutil.CallMethod(container.prison, "IsLockedDown")
	if errr != nil {
		log.Println(errr)
	}
	blocked := isLocked.Value().(bool)

	if blocked == false {

		oleutil.PutProperty(container.prison, "HomePath", rootPath)
		// oleutil.PutProperty(container, "MemoryLimitBytes", 1024*1024*300)

		log.Println("Locking down...")
		_, errr = oleutil.CallMethod(container.prison, "Lockdown")
		if errr != nil {
			log.Println(errr)
			return nil, errr
		}
		log.Println("Locked down.")
	}

	log.Println("Running process...")
	ptrackerRes, errr := oleutil.CallMethod(container.prison, "Run", cri)
	if errr != nil {
		log.Println(errr)
		return nil, errr
	}
	ptracker := ptrackerRes.ToIDispatch()

	//err := command.Start()
	//if err != nil {
	//	log.Println(err)
	//}

	iDpid, errr := oleutil.CallMethod(ptracker, "GetPid")
	if errr != nil {
		log.Println(errr)
	}
	pid := int(iDpid.Value().(int64))
	container.pids[pid] = ptracker

	//go func() {
	//	command.Process.Wait()
	//	delete(container.pids, pid)
	//}()

	return newWprocess(ptracker), nil
}

func (container *Container) Attach(processID uint32, pio warden.ProcessIO) (warden.Process, error) {
	log.Println("Attaching to: ", processID)

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
