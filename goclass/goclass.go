package goclass

import (
	"github.com/uwine4850/yefgo/pytypes"
)

// Class a structure that describes a Python class.
// instance - Python class instance.
// pyClass - Python class.
// pyModule - the module to which the class belongs.
type Class struct {
	instance pytypes.ClassInstance
	pyClass  pytypes.Class
	pyModule pytypes.Module
}

func (p *Class) SetInstance(instance pytypes.ClassInstance) {
	p.instance = instance
}

func (p *Class) SetClass(pyClass pytypes.Class) {
	p.pyClass = pyClass
}

func (p *Class) SetPyModule(pyModule pytypes.Module) {
	p.pyModule = pyModule
}

func (p *Class) GetInstance() pytypes.ClassInstance {
	return p.instance
}

func (p *Class) GetClass() pytypes.Class {
	return p.pyClass
}

func (p *Class) GetPyModule() pytypes.Module {
	return p.pyModule
}
