package pyargs

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"github.com/uwine4850/yefgo/goclass"
	"reflect"
	"unsafe"
)

func InitArgs(pyTuple unsafe.Pointer, args *[]interface{}) {
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
			doubleValue := C.double(arg.(float64))
			C.PyTuple_SetItem(tuple, C.long(i), C.PyFloat_FromDouble(doubleValue))
		case reflect.Bool:
			value := arg.(bool)
			var boolValue C.long
			if value {
				boolValue = C.long(1)
			} else {
				boolValue = C.long(0)
			}
			C.PyTuple_SetItem(tuple, C.long(i), C.PyBool_FromLong(boolValue))
		case reflect.TypeOf(&goclass.Class{}).Kind():
			class := arg.(*goclass.Class)
			C.PyTuple_SetItem(tuple, C.long(i), (*C.PyObject)(class.GetInstance()))
		default:
			panic("unhandled default case")
		}
	}
}
