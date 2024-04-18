package module

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
	"fmt"
	"github.com/uwine4850/yefgo/pyclass/memory"
	"github.com/uwine4850/yefgo/pytypes"
	"strconv"
	"unsafe"
)

// InitPython initializes Python to start working.
// Used to manage global data like modules and clean up allocated memory.
type InitPython struct {
	mustFreeObject  []unsafe.Pointer
	mustFreePointer []unsafe.Pointer
}

// Initialize Python initialization. This method should always be called first.
func (ip *InitPython) Initialize() {
	C.Py_Initialize()
}

// Finalize Quit Python. Clears allocated memory.
func (ip *InitPython) Finalize() {
	ip.FreeAll()
	C.Py_Finalize()
}

// GetPyModule get the module by name.
func (ip *InitPython) GetPyModule(name string) (pytypes.Module, error) {
	pyModuleName := C.CString(name)
	defer C.free(unsafe.Pointer(pyModuleName))

	pyModule := C.PyImport_ImportModule(pyModuleName)
	if pyModule == nil {
		return nil, errors.New("failed to import Python module")
	}
	ip.FreeObject(unsafe.Pointer(pyModule))
	memory.Link.Increment()
	return pytypes.Module(pyModule), nil
}

// FreeObject queues a reference to the allocated memory to be freed at the end of Python execution.
// Important: Memory will only be freed during the Finalize() call.
func (ip *InitPython) FreeObject(obj unsafe.Pointer) {
	ip.mustFreeObject = append(ip.mustFreeObject, obj)
}

// FreeAll frees all allocated memory.
func (ip *InitPython) FreeAll() {
	for i := 0; i < len(ip.mustFreeObject); i++ {
		C.Py_DecRef((*C.PyObject)(ip.mustFreeObject[i]))
		memory.Link.Decrement()
	}
	if memory.Link.Get() != 0 {
		panic(fmt.Sprintf("the number of RAM accesses is %s, not 0", strconv.Itoa(memory.Link.Get())))
	}
}

// GetPyObjectByString retrieves an object by name from another object.
// For example, a class from a module, or a method from a class.
func GetPyObjectByString(obj pytypes.ObjectPtr, name string) (pytypes.ObjectPtr, error) {
	nameStr := C.CString(name)
	memory.Link.Increment()
	defer memory.FreePointerNow(unsafe.Pointer(nameStr))

	pyObj := C.PyObject_GetAttrString((*C.PyObject)(obj), nameStr)
	if pyObj == nil {
		return nil, errors.New("failed to get object")
	}
	memory.Link.Increment()
	return pytypes.ObjectPtr(pyObj), nil
}
