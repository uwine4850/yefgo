package goclass

import (
	"github.com/uwine4850/yefgo/pytypes"
	"unsafe"
)

type Class struct {
	instance unsafe.Pointer
	class    unsafe.Pointer
	pyModule pytypes.PyModule
}

func (p *Class) SetInstance(instance unsafe.Pointer) {
	p.instance = instance
}

func (p *Class) SetClass(instance unsafe.Pointer) {
	p.class = instance
}

func (p *Class) SetPyModule(pyModule pytypes.PyModule) {
	p.pyModule = pyModule
}

func (p *Class) GetInstance() unsafe.Pointer {
	return p.instance
}

func (p *Class) GetClass() unsafe.Pointer {
	return p.class
}

func (p *Class) GetPyModule() pytypes.PyModule {
	return p.pyModule
}
