package pytypes

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import "unsafe"

type Module unsafe.Pointer
type Class unsafe.Pointer
type ClassInstance unsafe.Pointer
type TuplePtr unsafe.Pointer
type ObjectPtr unsafe.Pointer
