package backend

import (
	"github.com/cloudfoundry-incubator/warden-windows/backend/messages"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"time"

	"github.com/cloudfoundry-incubator/garden/backend"
	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/cloudfoundry-incubator/warden-windows/backend/iorpc"
	"github.com/cloudfoundry-incubator/warden-windows/backend/payload_muxer"
)

type Container struct {
	id     string
	handle string

	runner command_runner.CommandRunner
	muxer  payload_muxer.Muxer

	rpc *rpc.Client
}

func NewContainer(id, handle string, runner command_runner.CommandRunner, muxer payload_muxer.Muxer) *Container {
	return &Container{
		id:     id,
		handle: handle,

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

func (container *Container) Start() error {
	daemon := &exec.Cmd{
		Path: "DAEMON_PATH",
		Args: []string{"--handle", container.handle},
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

	err = container.runner.Start(daemon)
	if err != nil {
		return err
	}

	container.rpc = jsonrpc.NewClient(iorpc.New(stdin, stdout))

	container.muxer.SetSource(stderr)

	return nil
}

func (container *Container) Stop(kill bool) error {
	return container.rpc.Call(
		"Container.Stop",
		&messages.StopRequest{Kill: kill},
		&messages.StopResponse{},
	)
}

func (container *Container) Info() (backend.ContainerInfo, error) {
	log.Println("TODO Info")
	return backend.ContainerInfo{}, nil
}

func (container *Container) CopyIn(src, dst string) error {
	log.Println("TODO CopyIn")
	return nil
}

func (container *Container) CopyOut(src, dst, owner string) error {
	log.Println("TODO CopyOut")
	return nil
}

func (container *Container) LimitBandwidth(limits backend.BandwidthLimits) error {
	log.Println("TODO LimitBandwidth")
	return nil
}

func (container *Container) CurrentBandwidthLimits() (backend.BandwidthLimits, error) {
	log.Println("TODO CurrentBandwidthLimits")
	return backend.BandwidthLimits{}, nil
}

func (container *Container) LimitDisk(limits backend.DiskLimits) error {
	log.Println("TODO LimitDisk")
	return nil
}

func (container *Container) CurrentDiskLimits() (backend.DiskLimits, error) {
	log.Println("TODO CurrentDiskLimits")
	return backend.DiskLimits{}, nil
}

func (container *Container) LimitMemory(limits backend.MemoryLimits) error {
	log.Println("TODO LimitMemory")
	return nil
}

func (container *Container) CurrentMemoryLimits() (backend.MemoryLimits, error) {
	log.Println("TODO CurrentMemoryLimits")
	return backend.MemoryLimits{}, nil
}

func (container *Container) LimitCPU(limits backend.CPULimits) error {
	log.Println("TODO LimitCPU")
	return nil
}

func (container *Container) CurrentCPULimits() (backend.CPULimits, error) {
	log.Println("TODO CurrentCPULimits")
	return backend.CPULimits{}, nil
}

type ProcessPayloadDispatcher struct {
	stream chan<- backend.ProcessStream
}

func (container *Container) Run(spec backend.ProcessSpec) (uint32, <-chan backend.ProcessStream, error) {
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

	stream := make(chan backend.ProcessStream, 1000)

	container.muxer.Subscribe(response.ProcessID, stream)

	return response.ProcessID, stream, nil
}

func (container *Container) Attach(processID uint32) (<-chan backend.ProcessStream, error) {
	stream := make(chan backend.ProcessStream, 1000)

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
