package backend

import (
	"io"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"time"

	"github.com/cloudfoundry-incubator/warden-windows/backend/messages"

	"github.com/cloudfoundry-incubator/garden/warden"
	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/cloudfoundry-incubator/warden-windows/backend/iorpc"
	"github.com/cloudfoundry-incubator/warden-windows/backend/linecodec"
	"github.com/cloudfoundry-incubator/warden-windows/backend/payload_muxer"
)

type Container struct {
	id     string
	handle string

	rootPath string

	runner command_runner.CommandRunner
	muxer  payload_muxer.Muxer

	rpc *rpc.Client
}

func NewContainer(
	id, handle string,
	rootPath string,
	runner command_runner.CommandRunner,
	muxer payload_muxer.Muxer,
) *Container {
	return &Container{
		id:     id,
		handle: handle,

		rootPath: rootPath,

		runner: runner,
		muxer:  muxer,
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
	daemon := &exec.Cmd{
		Path: containerBinaryPath,
		Args: []string{"--handle", container.handle, "--rootPath", container.rootPath},
	}

	stdin, err := daemon.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := daemon.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := daemon.StderrPipe()
	if err != nil {
		return err
	}

	conn := iorpc.New(stdin, stdout)

	container.rpc = rpc.NewClientWithCodec(
		linecodec.New(
			stdin,
			jsonrpc.NewClientCodec(conn),
		),
	)

	container.muxer.SetSource(stderr)

	err = container.runner.Start(daemon)
	if err != nil {
		return err
	}

	return nil
}

func (container *Container) Stop(kill bool) error {
	return container.rpc.Call(
		"Container.Stop",
		&messages.StopRequest{Kill: kill},
		&messages.StopResponse{},
	)
}

func (container *Container) Info() (warden.ContainerInfo, error) {
	log.Println("TODO Info")
	return warden.ContainerInfo{}, nil
}

func (container *Container) StreamIn(dstPath string) (io.WriteCloser, error) {
	log.Println("TODO StreamIn")
	return nil, nil
}

func (container *Container) StreamOut(srcPath string) (io.Reader, error) {
	log.Println("TODO StreamOut")
	return nil, nil
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

func (container *Container) Run(spec warden.ProcessSpec) (uint32, <-chan warden.ProcessStream, error) {
	var response messages.RunResponse

	err := container.rpc.Call(
		"Container.Run",
		&messages.RunRequest{
			Script:     spec.Script,
			Privileged: spec.Privileged,
		},
		&response,
	)
	if err != nil {
		return 0, nil, err
	}

	stream := make(chan warden.ProcessStream, 1000)

	container.muxer.Subscribe(response.ProcessID, stream)

	return response.ProcessID, stream, nil
}

func (container *Container) Attach(processID uint32) (<-chan warden.ProcessStream, error) {
	stream := make(chan warden.ProcessStream, 1000)

	container.muxer.Subscribe(processID, stream)

	return stream, nil
}

func (container *Container) NetIn(hostPort uint32, containerPort uint32) (uint32, uint32, error) {
	log.Println("TODO NetIn")
	return hostPort, containerPort, nil
}

func (container *Container) NetOut(network string, port uint32) error {
	log.Println("TODO NetOut")
	return nil
}
