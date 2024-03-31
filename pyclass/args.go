package pyclass

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"github.com/uwine4850/yefgo/goclass"
	"github.com/uwine4850/yefgo/pyclass/memory"
	"github.com/uwine4850/yefgo/pyclass/module"
	"github.com/uwine4850/yefgo/pytypes"
	"reflect"
	"unsafe"
)

func InitArgs(pyInit *module.InitPython, pyTuple pytypes.TuplePtr, args *[]interface{}) {
	tuple := (*C.PyObject)(pyTuple)
	for i := 0; i < len(*args); i++ {
		arg := (*args)[i]
		switch reflect.TypeOf(arg).Kind() {
		case reflect.String:
			sString := C.CString(arg.(string))
			C.PyTuple_SetItem(tuple, C.long(i), C.PyUnicode_FromString(sString))
			C.free(unsafe.Pointer(sString))
		case reflect.Int:
			var intValue C.longlong = C.longlong(arg.(int))
			C.PyTuple_SetItem(tuple, C.long(i), C.PyLong_FromLongLong(intValue))
		case reflect.Float64:
			C.PyTuple_SetItem(tuple, C.long(i), C.PyFloat_FromDouble(C.double(arg.(float64))))
		case reflect.Bool:
			value := arg.(bool)
			var boolValue C.long
			if value {
				boolValue = C.long(1)
			} else {
				boolValue = C.long(0)
			}
			C.PyTuple_SetItem(tuple, C.long(i), C.PyBool_FromLong(boolValue))
		case reflect.Slice:
			pyList := C.PyList_New(0)
			memory.Link.Increment()
			pyInit.FreeObject(unsafe.Pointer(pyList))
			newSlice(pyInit, arg, unsafe.Pointer(pyList))
			C.PyTuple_SetItem(tuple, C.long(i), pyList)
		case reflect.TypeOf(&goclass.Class{}).Kind():
			class := arg.(*goclass.Class)
			C.PyTuple_SetItem(tuple, C.long(i), (*C.PyObject)(class.GetInstance()))
		default:
			panic("unhandled default case")
		}
	}
}

func newSlice(pyInit *module.InitPython, arg interface{}, _pyList unsafe.Pointer) {
	argSlice := reflect.ValueOf(arg)
	pyList := (*C.PyObject)(_pyList)
	for j := 0; j < reflect.ValueOf(arg).Len(); j++ {
		switch reflect.TypeOf(arg).Elem().Kind() {
		case reflect.String:
			sString := C.CString(argSlice.Index(j).String())
			C.PyList_Append(pyList, C.PyUnicode_FromString(sString))
			C.free(unsafe.Pointer(sString))
		case reflect.Int:
			var intValue C.longlong = C.longlong(argSlice.Index(j).Int())
			C.PyList_Append(pyList, C.PyLong_FromLongLong(intValue))
		case reflect.Float64:
			var floatValue = C.double(argSlice.Index(j).Float())
			C.PyList_Append(pyList, C.PyFloat_FromDouble(floatValue))
		case reflect.Bool:
			var value = argSlice.Index(j).Bool()
			var boolValue C.long
			if value {
				boolValue = C.long(1)
			} else {
				boolValue = C.long(0)
			}
			C.PyList_Append(pyList, C.PyBool_FromLong(boolValue))
		case reflect.Slice:
			newPyList := C.PyList_New(0)
			memory.Link.Increment()
			pyInit.FreeObject(unsafe.Pointer(newPyList))
			newSlice(pyInit, argSlice.Index(j).Interface(), unsafe.Pointer(newPyList))
			C.PyList_Append(pyList, newPyList)
		case reflect.TypeOf(&goclass.Class{}).Kind():
			class := argSlice.Index(j).Interface().(*goclass.Class)
			C.PyList_Append(pyList, (*C.PyObject)(class.GetInstance()))
		default:
			panic("unhandled default case")
		}
	}
}
