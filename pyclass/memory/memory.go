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

// FreeObjectNow frees a Python object from memory.
func FreeObjectNow(obj pytypes.ObjectPtr) {
	C.Py_DecRef((*C.PyObject)(obj))
	Link.Decrement()
}

// FreePointerNow frees the value pointer.
func FreePointerNow(ptr unsafe.Pointer) {
	C.free(ptr)
	Link.Decrement()
}

// memoryLink Counter of active memory allocation references.
// When allocating memory you need to use Increment(), when freeing it - Decrement().
type memoryLink int

func (l *memoryLink) Increment() {
	*l++
}

func (l *memoryLink) Decrement() {
	*l--
}

// Get number of links.
func (l *memoryLink) Get() int {
	return int(*l)
}

var Link = memoryLink(0)
