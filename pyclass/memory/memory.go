package memory

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"github.com/uwine4850/yefgo/pytypes"
	"unsafe"
)

func FreeObjectNow(obj pytypes.ObjectPtr) {
	C.Py_DecRef((*C.PyObject)(obj))
	Link.Decrement()
}

func FreePointerNow(ptr unsafe.Pointer) {
	C.free(ptr)
	Link.Decrement()
}

type memoryLink int

func (l *memoryLink) Increment() {
	*l++
}

func (l *memoryLink) Decrement() {
	*l--
}

func (l *memoryLink) Get() int {
	return int(*l)
}

var Link = memoryLink(0)
