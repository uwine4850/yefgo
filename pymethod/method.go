package pymethod

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
	"github.com/uwine4850/yefgo/goclass"
	"github.com/uwine4850/yefgo/pyclass"
	"github.com/uwine4850/yefgo/pyclass/memory"
	"github.com/uwine4850/yefgo/pymethod/pyargs"
	"github.com/uwine4850/yefgo/pytypes"
	"reflect"
	"unsafe"
)

func CallMethod(pyInit *pyclass.InitPython, methodName string, classOrInstance unsafe.Pointer, args ...interface{}) (unsafe.Pointer, error) {
	pyMethodName := C.CString(methodName)
	memory.Link.Increment()
	defer memory.FreePointerNow(unsafe.Pointer(pyMethodName))
	classOrInstanceObj := (*C.PyObject)(classOrInstance)
	pyMethod := C.PyObject_GetAttrString(classOrInstanceObj, pyMethodName)
	if pyMethod == nil {
		return nil, errors.New("failed to get method")
	}
	memory.Link.Increment()
	defer memory.FreeObjectNow(pytypes.ObjectPtr(pyMethod))
	methodArgs := C.PyTuple_New(C.long(len(args)))
	memory.Link.Increment()
	defer memory.FreeObjectNow(pytypes.ObjectPtr(methodArgs))
	pyargs.InitArgs(pytypes.TuplePtr(methodArgs), &args)
	res := C.PyObject_CallObject(pyMethod, methodArgs)
	if res != C.Py_None && res != nil {
		pyInit.FreeObject(unsafe.Pointer(res))
		memory.Link.Increment()
	}
	return unsafe.Pointer(res), nil
}

func CallClassMethod(pyInit *pyclass.InitPython, methodName string, class *goclass.Class, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(pyInit, methodName, unsafe.Pointer(class.GetClass()), args...)
}

func CallInstanceMethod(pyInit *pyclass.InitPython, methodName string, class *goclass.Class, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(pyInit, methodName, unsafe.Pointer(class.GetInstance()), args...)
}

func MethodOutput(pyInit *pyclass.InitPython, _res unsafe.Pointer, output interface{}) {
	if reflect.TypeOf(output).Kind() != reflect.Pointer {
		panic("AAA")
	}
	res := (*C.PyObject)(_res)
	if res == C.Py_None || res == nil {
		output = nil
		return
	}

	switch reflect.TypeOf(output).Elem().Kind() {
	case reflect.String:
		reflect.ValueOf(output).Elem().SetString(C.GoString(C.PyUnicode_AsUTF8(res)))
	case reflect.Int:
		cIntValue := C.PyLong_AsLongLong(res)
		reflect.ValueOf(output).Elem().SetInt(int64(cIntValue))
	case reflect.Struct:
		t := reflect.TypeOf(output).Elem()
		newStruct := reflect.New(t).Elem()
		class := goclass.Class{}
		class.SetInstance(pytypes.ClassInstance(res))
		moduleName, err := getPyModuleNameFromInstance(res)
		if err != nil {
			panic(err)
		}

		newPyModule, err := pyInit.GetPyModule(moduleName)
		if err != nil {
			panic(err)
		}
		createClass, err := pyclass.GetPyClass(newStruct.Type().Name(), newPyModule)
		if err != nil {
			panic(err)
		}
		pyInit.FreeObject(unsafe.Pointer(createClass))
		class.SetClass(createClass)
		class.SetPyModule(newPyModule)
		newStruct.FieldByName("Class").Set(reflect.ValueOf(class))
		reflect.ValueOf(output).Elem().Set(newStruct)
	default:
		panic("unhandled default case")
	}
}

func getPyModuleNameFromInstance(instance *C.PyObject) (string, error) {
	pyModuleAttrName := C.CString("__module__")
	defer C.free(unsafe.Pointer(pyModuleAttrName))

	pyModuleAttr := C.PyObject_GetAttrString(instance, pyModuleAttrName)
	if pyModuleAttr == nil {
		return "", errors.New("failed to get module name attribute")
	}
	defer C.Py_DecRef(pyModuleAttr)

	pyModuleName := C.PyUnicode_AsUTF8(pyModuleAttr)
	if pyModuleName == nil {
		return "", errors.New("failed to convert module name to string")
	}
	return C.GoString(pyModuleName), nil
}
