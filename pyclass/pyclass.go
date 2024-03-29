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
	pyInit    *InitPython
	pyModule  pytypes.Module
	className string
	args      []interface{}
}

func NewPyInstance(pyInit *InitPython, pyModule pytypes.Module, className string, args ...interface{}) *PyInstance {
	return &PyInstance{pyInit: pyInit, pyModule: pyModule, className: className, args: args}
}

func (p *PyInstance) getInitArgs() *C.PyObject {
	init := C.PyTuple_New(C.long(len(p.args)))
	pyargs.InitArgs(pytypes.TuplePtr(init), &p.args)
	return init
}

func (p *PyInstance) Create() (pytypes.ClassInstance, error) {
	pyClass, err := GetPyObjectByString(pytypes.ObjectPtr((*C.PyObject)(p.pyModule)), p.className)
	if err != nil {
		return nil, err
	}
	defer p.FreeObject((*C.PyObject)(pyClass))
	init := p.getInitArgs()
	defer p.FreeObject(init)

	pyInstance := C.PyObject_CallObject((*C.PyObject)(pyClass), init)
	if pyInstance == nil {
		return nil, errors.New("failed to create instance")
	}
	p.pyInit.FreeObject(unsafe.Pointer(pyInstance))
	return pytypes.ClassInstance(pyInstance), nil
}

func GetPyClass(name string, pyModule pytypes.Module) (pytypes.Class, error) {
	pyClass, err := GetPyObjectByString(pytypes.ObjectPtr((*C.PyObject)(pyModule)), name)
	if err != nil {
		return nil, err
	}
	return pytypes.Class((*C.PyObject)(pyClass)), nil
}
