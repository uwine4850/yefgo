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
	"github.com/uwine4850/yefgo/pymethod/pyargs"
	"github.com/uwine4850/yefgo/pytypes"
	"reflect"
	"unsafe"
)

func CallMethod(methodName string, classOrInstance unsafe.Pointer, args ...interface{}) (unsafe.Pointer, error) {
	pyMethodName := C.CString(methodName)
	defer C.free(unsafe.Pointer(pyMethodName))
	classOrInstanceObj := (*C.PyObject)(classOrInstance)
	pyMethod := C.PyObject_GetAttrString(classOrInstanceObj, pyMethodName)
	if pyMethod == nil {
		return nil, errors.New("failed to get method")
	}
	defer C.Py_DecRef(pyMethod)
	methodArgs := C.PyTuple_New(C.long(len(args)))
	defer C.Py_DecRef(methodArgs)
	pyargs.InitArgs(unsafe.Pointer(methodArgs), &args)
	res := C.PyObject_CallObject(pyMethod, methodArgs)
	return unsafe.Pointer(res), nil
}

func CallClassMethod(methodName string, class *goclass.Class, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(methodName, class.GetClass(), args...)
}

func CallInstanceMethod(methodName string, class *goclass.Class, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(methodName, class.GetInstance(), args...)
}

func MethodOutput(pyInit *pyclass.InitPython, _res unsafe.Pointer, output interface{}, pyModule pytypes.PyModule) {
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
		C.Py_DecRef(res)
	case reflect.Int:
		cIntValue := C.PyLong_AsLongLong(res)
		reflect.ValueOf(output).Elem().SetInt(int64(cIntValue))
		C.Py_DecRef(res)
	case reflect.Struct:
		t := reflect.TypeOf(output).Elem()
		newStruct := reflect.New(t).Elem()
		class := goclass.Class{}
		class.SetInstance(unsafe.Pointer(res))
		createClass, err := pyclass.GetPyClass(newStruct.Type().Name(), pyModule)
		if err != nil {
			panic(err)
		}
		class.SetClass(createClass)
		class.SetPyModule(pyModule)
		pyInit.FreeObject(_res)
		pyInit.FreeObject(createClass)
		newStruct.FieldByName("Class").Set(reflect.ValueOf(class))
		reflect.ValueOf(output).Elem().Set(newStruct)
	default:
		panic("unhandled default case")
	}
}
