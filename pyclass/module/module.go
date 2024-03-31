package module

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
	"github.com/uwine4850/yefgo/pyclass/memory"
	"github.com/uwine4850/yefgo/pytypes"
	"unsafe"
)

type InitPython struct {
	mustFreeObject  []unsafe.Pointer
	mustFreePointer []unsafe.Pointer
}

func (ip *InitPython) Initialize() {
	C.Py_Initialize()
}

func (ip *InitPython) Finalize() {
	ip.FreeAll()
	C.Py_Finalize()
}

func (ip *InitPython) GetPyModule(name string) (pytypes.Module, error) {
	pyModuleName := C.CString(name)
	defer C.free(unsafe.Pointer(pyModuleName))

	pyModule := C.PyImport_ImportModule(pyModuleName)
	if pyModule == nil {
		return nil, errors.New("failed to import Python module")
	}
	ip.FreeObject(unsafe.Pointer(pyModule))
	memory.Link.Increment()
	return pytypes.Module(pyModule), nil
}

func (ip *InitPython) FreeObject(obj unsafe.Pointer) {
	ip.mustFreeObject = append(ip.mustFreeObject, obj)
}

func (ip *InitPython) FreeAll() {
	for i := 0; i < len(ip.mustFreeObject); i++ {
		C.Py_DecRef((*C.PyObject)(ip.mustFreeObject[i]))
		memory.Link.Decrement()
	}
	if memory.Link.Get() != 0 {
		panic("the number of references to RAM is not 0")
	}
}

func GetPyObjectByString(obj pytypes.ObjectPtr, name string) (pytypes.ObjectPtr, error) {
	nameStr := C.CString(name)
	memory.Link.Increment()
	defer memory.FreePointerNow(unsafe.Pointer(nameStr))

	pyObj := C.PyObject_GetAttrString((*C.PyObject)(obj), nameStr)
	if pyObj == nil {
		return nil, errors.New("failed to get object")
	}
	memory.Link.Increment()
	return pytypes.ObjectPtr(pyObj), nil
}
