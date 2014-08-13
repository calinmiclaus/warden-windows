package prison_client

import (
	"errors"
	"fmt"
	"github.com/mattn/go-ole"
	"github.com/mattn/go-ole/oleutil"
	"log"
	"runtime"
)

type Container struct {
	cont *ole.IDispatch
}

func newContainer(cont *ole.IDispatch) *Container {
	ret := &Container{cont: cont}
	return ret
}

func CreateContainer() (*Container, error) {
	IUc, err := oleutil.CreateObject("Uhuru.Prison.ComWrapper.Container")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer IUc.Release()

	container, err := IUc.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ret := newContainer(container)
	runtime.SetFinalizer(ret, finalizeContainer)

	return ret, nil
}

func finalizeContainer(t *Container) {
	if t.cont != nil {
		lastRefCount := t.cont.Release()

		log.Println("Container ref count after finalizer: ", lastRefCount)
		t.cont = nil
	}
}

func (t *Container) Release() error {
	if t.cont != nil {
		t.cont.Release()
		t.cont = nil
		return nil
	} else {
		return errors.New("Container is already released")
	}
}

func (t *Container) Id() uint32 {
	id, err := oleutil.GetProperty(t.cont, "Id")
	if err != nil {
		log.Fatal(err)
	}
	defer id.Clear()

	return uint32(id.Value().(int64))
}

func (t *Container) GetHomePath() string {
	path, err := oleutil.GetProperty(t.cont, "HomePath")
	if err != nil {
		log.Fatal(err)
	}
	defer path.Clear()
	return path.ToString()
}

func (t *Container) SetHomePath(value string) {
	_, err := oleutil.PutProperty(t.cont, "HomePath", value)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *Container) GetMemoryLimitBytes() int64 {
	value, err := oleutil.GetProperty(t.cont, "MemoryLimitBytes")
	if err != nil {
		log.Fatal(err)
	}
	defer value.Clear()

	return value.Value().(int64)
}

func (t *Container) SetMemoryLimitBytes(value int64) {
	_, err := oleutil.PutProperty(t.cont, "MemoryLimitBytes", value)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *Container) GetDiskLimitBytes() int64 {
	value, err := oleutil.GetProperty(t.cont, "DiskLimitBytes")
	if err != nil {
		log.Fatal(err)
	}
	defer value.Clear()

	return value.Value().(int64)
}

func (t *Container) SetDiskLimitBytes(value int64) {
	_, err := oleutil.PutProperty(t.cont, "DiskLimitBytes", value)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *Container) GetNetworkPort() uint32 {
	value, err := oleutil.GetProperty(t.cont, "NetworkPort")
	if err != nil {
		log.Fatal(err)
	}
	defer value.Clear()

	return uint32(value.Value().(int64))
}

func (t *Container) SetNetworkPort(value int) {
	_, err := oleutil.PutProperty(t.cont, "NetworkPort", value)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *Container) IsLockedDown() bool {
	isLocked, err := oleutil.CallMethod(t.cont, "IsLockedDown")
	if err != nil {
		log.Fatal(err)
	}
	defer isLocked.Clear()
	return isLocked.Value().(bool)
}

func (t *Container) Lockdown() error {
	_, err := oleutil.CallMethod(t.cont, "Lockdown")
	if err != nil {
		oleerr := err.(*ole.OleError)
		// S_FALSE           = 0x00000001

		fmt.Println(err.Error())
		fmt.Println(oleerr.String())
		return err
	}
	return nil
}

func (t *Container) Run(runInfo *ContainerRunInfo) (*ProcessTracker, error) {
	iDcri, _ := runInfo.GetIDispatch()
	defer iDcri.Release()
	ptrackerRes, err := oleutil.CallMethod(t.cont, "Run", iDcri)

	if err != nil {
		return nil, err
	}
	defer ptrackerRes.Clear()
	ptracker := ptrackerRes.ToIDispatch()
	ptracker.AddRef() // ToIDispatch does not incease ref count
	defer ptracker.Release()

	pt := NewProcessTracker(ptracker)

	return pt, nil
}

func (t *Container) Stop() error {
	_, err := oleutil.CallMethod(t.cont, "Stop")
	if err != nil {
		return err
	}
	return nil
}

func (t *Container) Destroy() error {
	_, err := oleutil.CallMethod(t.cont, "Destroy")
	if err != nil {
		return err
	}
	return nil
}
