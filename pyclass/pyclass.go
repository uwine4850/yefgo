package pyclass

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
	"github.com/uwine4850/yefgo/pymethod/pyargs"
	"github.com/uwine4850/yefgo/pytypes"
	"unsafe"
)

type FreeMemory struct {
}

func (f FreeMemory) FreeObject(obj *C.PyObject) {
	C.Py_DecRef(obj)
}

func (f FreeMemory) FreePointer(ptr unsafe.Pointer) {
	C.free(ptr)
}

type PyInstance struct {
	FreeMemory
	pyModule  pytypes.PyModule
	pyClass   unsafe.Pointer
	className string
	args      []interface{}
}

func NewPyInstance(pyModule pytypes.PyModule, className string, args ...interface{}) *PyInstance {
	return &PyInstance{pyModule: pyModule, className: className, args: args}
}

func (p *PyInstance) getInitArgs() *C.PyObject {
	init := C.PyTuple_New(C.long(len(p.args)))
	pyargs.InitArgs(unsafe.Pointer(init), &p.args)
	return init
}

func (p *PyInstance) Create() (unsafe.Pointer, error) {
	className := C.CString(p.className)
	defer p.FreePointer(unsafe.Pointer(className))

	pyClass := C.PyObject_GetAttrString((*C.PyObject)(p.pyModule), className)
	if pyClass == nil {
		return nil, errors.New("failed to get class")
	}
	defer p.FreeObject(pyClass)

	init := p.getInitArgs()
	defer p.FreeObject(init)

	pyNamInstance := C.PyObject_CallObject(pyClass, init)
	if pyNamInstance == nil {
		return nil, errors.New("failed to create Nam instance")
	}
	return unsafe.Pointer(pyNamInstance), nil
}

func GetPyClass(name string, pyModule pytypes.PyModule) (unsafe.Pointer, error) {
	className := C.CString(name)
	defer C.free(unsafe.Pointer(className))

	pyClass := C.PyObject_GetAttrString((*C.PyObject)(pyModule), className)
	if pyClass == nil {
		return nil, errors.New("failed to get class")
	}
	return unsafe.Pointer(pyClass), nil
}
