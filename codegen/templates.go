package codegen

const pkgInit = `package {{.PkgName}}

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"github.com/uwine4850/yefgo/goclass"
	"github.com/uwine4850/yefgo/pyclass"
	"github.com/uwine4850/yefgo/pyclass/module"
	"github.com/uwine4850/yefgo/pytypes"
	"unsafe"
	{{.Imports}}
)
`

const classinit = `
type {{.StructName}} struct {
	goclass.Class
	PyInit *module.InitPython
}

func (n {{.StructName}}) New(pyInit *module.InitPython, pyModule pytypes.Module, {{.Args}}) ({{.StructName}}, error) {
	n.PyInit = pyInit
	instance := pyclass.NewPyInstance(n.PyInit, pyModule, "{{.StructName}}", {{.ArgsForFunc}})
	newInstance, err := instance.Create()
	if err != nil {
		return {{.StructName}}{}, err
	}
	n.PyInit.FreeObject(unsafe.Pointer(newInstance))
	n.SetInstance(newInstance)

	class, err := pyclass.GetPyClass("{{.StructName}}", pyModule)
	if err != nil {
		return {{.StructName}}{}, err
	}
	n.PyInit.FreeObject(unsafe.Pointer(class))
	n.SetClass(class)
	n.SetPyModule(pyModule)
	return n, nil
}
`

const funcInstanceCall = `pyclass.CallInstanceMethod(n.PyInit, &n.Class, "{{.PyFuncName}}", {{.ArgsForFunc}})`

const funcClassCall = `pyclass.CallClassMethod(n.PyInit, &n.Class, "{{.PyFuncName}}", {{.ArgsForFunc}})`

const funcInit = `
func (n {{.StructName}}) {{.GoFuncName}}({{.Args}}) error {
	_, err := {{.FuncCall}}
	if err != nil {
		return err
	}
	return nil
}
`

const funcWithOutputInit = `
func (n {{.StructName}}) {{.GoFuncName}}({{.Args}}) (*{{.OutputType}}, error) {
	res, err := {{.FuncCall}}
	if err != nil {
		return nil, err
	}
	var output {{.OutputType}}
	err = pyclass.MethodOutput(n.PyInit, res, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
`

const moduleFunc = `
func {{.Name}}(init *module.InitPython, pyModule pytypes.Module, {{.Args}}) error {
	_, err := pyclass.CallModuleMethod(init, pyModule, "{{.PyFuncName}}", {{.ArgsForFunc}})
	if err != nil {
		return err
	}
	return nil
}
`

const moduleFuncWithOutput = `
func {{.Name}}(init *module.InitPython, pyModule pytypes.Module, {{.Args}}) (*{{.OutputType}}, error)  {
	res, err := pyclass.CallModuleMethod(init, pyModule, "{{.PyFuncName}}", {{.ArgsForFunc}})
	if err != nil {
		return nil, err
	}
	var output {{.OutputType}}
	err = pyclass.MethodOutput(init, res, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
`
