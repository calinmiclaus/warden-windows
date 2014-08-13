package prison_client

import (
	"errors"
	"github.com/mattn/go-ole"
	"github.com/mattn/go-ole/oleutil"
	"log"
	"runtime"
)

type ContainerManager struct {
	cm *ole.IDispatch
}

func NewContainerManager(cm *ole.IDispatch) *ContainerManager {
	cm.AddRef() // call addref because we clone the pointr
	ret := &ContainerManager{cm: cm}
	runtime.SetFinalizer(ret, finalizeContainerManager)
	return ret
}

func finalizeContainerManager(t *ContainerManager) {
	if t.cm != nil {
		lastRefCount := t.cm.Release()

		log.Println("ContainerManager ref count after finalizer: ", lastRefCount)
		t.cm = nil
	}
}

func (t *ContainerManager) Release() error {
	if t.cm != nil {
		t.cm.Release()
		t.cm = nil
		return nil
	} else {
		return errors.New("ContainerManager is already released")
	}
}

func (t *ContainerManager) ListContainerIds() []string {
	containerIds, err := oleutil.CallMethod(t.cm, "ListContainerIds")
	if err != nil {
		log.Fatal(err)
	}
	defer containerIds.Clear()

	return containerIds.ToArray().ToStringArray()
}
