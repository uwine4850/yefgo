package pyclass

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
	"github.com/uwine4850/yefgo/pyclass/memory"
	"github.com/uwine4850/yefgo/pyclass/module"
	"github.com/uwine4850/yefgo/pytypes"
)

type PyInstance struct {
	pyInit    *module.InitPython
	pyModule  pytypes.Module
	className string
	args      []interface{}
}

func NewPyInstance(pyInit *module.InitPython, pyModule pytypes.Module, className string, args ...interface{}) *PyInstance {
	return &PyInstance{pyInit: pyInit, pyModule: pyModule, className: className, args: args}
}

func (p *PyInstance) getInitArgs() *C.PyObject {
	init := C.PyTuple_New(C.long(len(p.args)))
	memory.Link.Increment()
	InitArgs(p.pyInit, pytypes.TuplePtr(init), &p.args)
	return init
}

func (p *PyInstance) Create() (pytypes.ClassInstance, error) {
	pyClass, err := module.GetPyObjectByString(pytypes.ObjectPtr((*C.PyObject)(p.pyModule)), p.className)
	if err != nil {
		return nil, err
	}
	defer memory.FreeObjectNow(pyClass)
	init := p.getInitArgs()
	defer memory.FreeObjectNow(pytypes.ObjectPtr(init))

	pyInstance := C.PyObject_CallObject((*C.PyObject)(pyClass), init)
	if pyInstance == nil {
		return nil, errors.New("failed to create instance")
	}
	memory.Link.Increment()
	return pytypes.ClassInstance(pyInstance), nil
}

func GetPyClass(name string, pyModule pytypes.Module) (pytypes.Class, error) {
	pyClass, err := module.GetPyObjectByString(pytypes.ObjectPtr((*C.PyObject)(pyModule)), name)
	if err != nil {
		return nil, err
	}
	return pytypes.Class((*C.PyObject)(pyClass)), nil
}
