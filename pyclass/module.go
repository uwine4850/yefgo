package pyclass

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
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

func (ip *InitPython) GetPyModule(name string) (pytypes.PyModule, error) {
	pyModuleName := C.CString(name)
	defer C.free(unsafe.Pointer(pyModuleName))

	pyModule := C.PyImport_ImportModule(pyModuleName)
	if pyModule == nil {
		return nil, errors.New("failed to import Python module")
	}
	ip.FreeObject(unsafe.Pointer(pyModule))
	return pytypes.PyModule(pyModule), nil
}

func (ip *InitPython) FreeObject(obj unsafe.Pointer) {
	ip.mustFreeObject = append(ip.mustFreeObject, obj)
}

func (ip *InitPython) FreePointer(ptr unsafe.Pointer) {
	ip.mustFreePointer = append(ip.mustFreePointer, ptr)
}

func (ip *InitPython) FreeAll() {
	for i := 0; i < len(ip.mustFreeObject); i++ {
		C.Py_DecRef((*C.PyObject)(ip.mustFreeObject[i]))
	}
	for i := 0; i < len(ip.mustFreePointer); i++ {
		C.free(ip.mustFreePointer[i])
	}
}
