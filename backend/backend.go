package backend

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/garden/backend"
	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/cloudfoundry-incubator/warden-windows/backend/payload_muxer"
)

type Backend struct {
	containerBinaryPath string
	containerRootPath   string

	runner command_runner.CommandRunner

	containerIDs <-chan string

	containers      map[string]backend.Container
	containersMutex *sync.RWMutex
}

type UnknownHandleError struct {
	Handle string
}

func (e UnknownHandleError) Error() string {
	return "unknown handle: " + e.Handle
}

func New(
	containerBinaryPath string,
	containerRootPath string,
	runner command_runner.CommandRunner,
) *Backend {
	containerIDs := make(chan string)

	go generateContainerIDs(containerIDs)

	return &Backend{
		containerBinaryPath: containerBinaryPath,
		containerRootPath:   containerRootPath,

		runner: runner,

		containerIDs: containerIDs,

		containers:      make(map[string]backend.Container),
		containersMutex: new(sync.RWMutex),
	}
}

func (backend *Backend) Start() error {
	return nil
}

func (backend *Backend) Stop() {
	log.Println("Stop")
}

func (backend *Backend) Create(spec backend.ContainerSpec) (backend.Container, error) {
	log.Println("Create")

	id := <-backend.containerIDs

	handle := id
	if spec.Handle != "" {
		handle = spec.Handle
	}

	container := NewContainer(id, handle, backend.containerRootPath, backend.runner, payload_muxer.New())

	backend.containersMutex.Lock()
	backend.containers[handle] = container
	backend.containersMutex.Unlock()

	err := container.Start(backend.containerBinaryPath)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (backend *Backend) Destroy(handle string) error {
	return errors.New("not implemented")
}

func (backend *Backend) Containers() (containers []backend.Container, err error) {
	backend.containersMutex.RLock()
	defer backend.containersMutex.RUnlock()

	for _, container := range backend.containers {
		containers = append(containers, container)
	}

	return containers, nil
}

func (backend *Backend) Lookup(handle string) (backend.Container, error) {
	backend.containersMutex.RLock()
	defer backend.containersMutex.RUnlock()

	container, found := backend.containers[handle]
	if !found {
		return nil, UnknownHandleError{handle}
	}

	return container, nil
}

func generateContainerIDs(ids chan<- string) string {
	for containerNum := time.Now().UnixNano(); ; containerNum++ {
		containerID := []byte{}

		var i uint
		for i = 0; i < 11; i++ {
			containerID = strconv.AppendInt(
				containerID,
				(containerNum>>(55-(i+1)*5))&31,
				32,
			)
		}

		ids <- string(containerID)
	}
}
