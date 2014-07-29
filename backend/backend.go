package backend

// // importing C will increase the stack size. this is useful
// // if loading a .NET COM object, beacuse the CLR requires a larger stack
import "C"

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/garden/warden"
	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/mattn/go-ole"
	//"github.com/mattn/go-ole/oleutil"
)

type Backend struct {
	containerBinaryPath string
	containerRootPath   string

	runner command_runner.CommandRunner

	containerIDs <-chan string

	containers      map[string]warden.Container
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

		containers:      make(map[string]warden.Container),
		containersMutex: new(sync.RWMutex),
	}
}

func (backend *Backend) Start() error {
	errr := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	return errr
}

func (backend *Backend) Stop() {
	log.Println("Stop")
}

func (backend *Backend) GraceTime(container warden.Container) time.Duration {
	log.Println("GraceTime")
	return time.Duration(100 * time.Second)
}

func (backend *Backend) Capacity() (warden.Capacity, error) {
	log.Println("TODO Capacity")
	return warden.Capacity{MemoryInBytes: 1024 * 1024 * 1024 * 100, DiskInBytes: 1024 * 1024 * 1024 * 100, MaxContainers: 1000}, nil
}

func (backend *Backend) Create(spec warden.ContainerSpec) (warden.Container, error) {
	log.Println("Create")

	id := <-backend.containerIDs

	handle := id
	if spec.Handle != "" {
		handle = spec.Handle
	}

	container := NewContainer(id, handle, backend.containerRootPath, backend.runner)

	backend.containersMutex.Lock()
	backend.containers[handle] = container
	backend.containersMutex.Unlock()

	return container, nil
}

func (backend *Backend) Ping() error {
	return nil
}

func (backend *Backend) Destroy(handle string) error {
	log.Println("Destroying container with handle: ", handle)
	curContainer, ok := backend.containers[handle]
	if ok == false {
		return nil
	}
	curContainer.Stop(true)
	// urContainer.Destroy()

	delete(backend.containers, handle)
	return nil
	return errors.New("not implemented")
}

func (backend *Backend) Containers(warden.Properties) (containers []warden.Container, err error) {
	backend.containersMutex.RLock()
	defer backend.containersMutex.RUnlock()

	for _, container := range backend.containers {
		containers = append(containers, container)
	}

	return containers, nil
}

func (backend *Backend) Lookup(handle string) (warden.Container, error) {
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
