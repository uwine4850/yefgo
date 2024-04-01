package pyclass

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"fmt"
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
		case reflect.Map:
			pyDict := C.PyDict_New()
			memory.Link.Increment()
			pyInit.FreeObject(unsafe.Pointer(pyDict))
			newMap(pyInit, reflect.ValueOf(arg).MapRange(), unsafe.Pointer(pyDict))
			C.PyTuple_SetItem(tuple, C.long(i), pyDict)
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
		case reflect.Map:
			pyDict := C.PyDict_New()
			memory.Link.Increment()
			pyInit.FreeObject(unsafe.Pointer(pyDict))
			newMap(pyInit, argSlice.Index(j).MapRange(), unsafe.Pointer(pyDict))
			C.PyList_Append(pyList, pyDict)
		case reflect.TypeOf(&goclass.Class{}).Kind():
			class := argSlice.Index(j).Interface().(*goclass.Class)
			C.PyList_Append(pyList, (*C.PyObject)(class.GetInstance()))
		default:
			panic("unhandled default case")
		}
	}
}

func newMap(pyInit *module.InitPython, arg *reflect.MapIter, pyDict unsafe.Pointer) {
	pyDictC := (*C.PyObject)(pyDict)
	for arg.Next() {
		pyKey := (*C.PyObject)(getMapCObject(pyInit, arg.Key()))
		pyValue := (*C.PyObject)(getMapCObject(pyInit, arg.Value()))
		C.PyDict_SetItem(pyDictC, pyKey, pyValue)
		memory.FreeObjectNow(pytypes.ObjectPtr(pyKey))
		memory.FreeObjectNow(pytypes.ObjectPtr(pyValue))
	}
}

func getMapCObject(pyInit *module.InitPython, _val reflect.Value) pytypes.ObjectPtr {
	var pyVal pytypes.ObjectPtr
	var val reflect.Value
	if _val.Type().Kind() == reflect.Ptr {
		val = _val.Elem()
	} else {
		val = _val
	}
	switch val.Type().Kind() {
	case reflect.String:
		cVal := C.CString(val.String())
		pyVal = pytypes.ObjectPtr(C.PyUnicode_FromString(cVal))
		C.free(unsafe.Pointer(cVal))
	case reflect.Int:
		pyVal = pytypes.ObjectPtr(C.PyLong_FromLongLong(C.longlong(val.Int())))
	case reflect.Float64:
		pyVal = pytypes.ObjectPtr(C.PyFloat_FromDouble(C.double(val.Float())))
	case reflect.Bool:
		var value = val.Bool()
		var boolValue C.long
		if value {
			boolValue = C.long(1)
		} else {
			boolValue = C.long(0)
		}
		pyVal = pytypes.ObjectPtr(C.PyBool_FromLong(boolValue))
	case reflect.Slice:
		pyList := C.PyList_New(0)
		memory.Link.Increment()
		pyInit.FreeObject(unsafe.Pointer(pyList))
		newSlice(pyInit, val.Interface(), unsafe.Pointer(pyList))
		pyVal = pytypes.ObjectPtr(pyList)
	case reflect.Map:
		pyDict := C.PyDict_New()
		newMap(pyInit, val.MapRange(), unsafe.Pointer(pyDict))
		pyVal = pytypes.ObjectPtr(pyDict)
	case reflect.Struct:
		class := val.Interface().(goclass.Class)
		pyVal = pytypes.ObjectPtr(class.GetInstance())
	default:
		panic(fmt.Sprintf("unhandled map type %s", val.Type().Kind()))
	}
	memory.Link.Increment()
	return pyVal
}
