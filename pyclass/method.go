package pyclass

/*
#include <Python.h>
#cgo pkg-config: python3
*/
import "C"
import (
	"errors"
	"fmt"
	"github.com/uwine4850/yefgo/goclass"
	"github.com/uwine4850/yefgo/pyclass/memory"
	"github.com/uwine4850/yefgo/pyclass/module"
	"github.com/uwine4850/yefgo/pytypes"
	"reflect"
	"unsafe"
)

// CallMethod calls a function.
// Depending on the argument passed, parentObject can call either a class method or just a module function.
func CallMethod(pyInit *module.InitPython, parentObject unsafe.Pointer, methodName string, args ...interface{}) (unsafe.Pointer, error) {
	pyMethodName := C.CString(methodName)
	memory.Link.Increment()
	defer memory.FreePointerNow(unsafe.Pointer(pyMethodName))
	parentObjectC := (*C.PyObject)(parentObject)
	pyMethod := C.PyObject_GetAttrString(parentObjectC, pyMethodName)
	if pyMethod == nil {
		return nil, errors.New("failed to get method")
	}
	memory.Link.Increment()
	defer memory.FreeObjectNow(pytypes.ObjectPtr(pyMethod))
	methodArgs := C.PyTuple_New(C.long(len(args)))
	memory.Link.Increment()
	defer memory.FreeObjectNow(pytypes.ObjectPtr(methodArgs))
	InitArgs(pyInit, pytypes.TuplePtr(methodArgs), &args)
	res := C.PyObject_CallObject(pyMethod, methodArgs)
	if res != C.Py_None && res != nil {
		pyInit.FreeObject(unsafe.Pointer(res))
		memory.Link.Increment()
	}
	return unsafe.Pointer(res), nil
}

func CallClassMethod(pyInit *module.InitPython, class *goclass.Class, methodName string, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(pyInit, unsafe.Pointer(class.GetClass()), methodName, args...)
}

func CallInstanceMethod(pyInit *module.InitPython, class *goclass.Class, methodName string, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(pyInit, unsafe.Pointer(class.GetInstance()), methodName, args...)
}

func CallModuleMethod(pyInit *module.InitPython, pyModule pytypes.Module, methodName string, args ...interface{}) (unsafe.Pointer, error) {
	return CallMethod(pyInit, unsafe.Pointer(pyModule), methodName, args...)
}

// MethodOutput processes the result of the CallMethod function and converts it to a golang data type.
// Places data in the output argument passed as a reference.
// IMPORTANT: the output argument must be of the type that is expected as output. If the output result is if there is
// a class, it needs to be passed to the desired golang structure that represents this class.
func MethodOutput(pyInit *module.InitPython, _res unsafe.Pointer, output interface{}) error {
	if reflect.TypeOf(output).Kind() != reflect.Pointer {
		return errors.New("output variable must be a pointer")
	}

	res := (*C.PyObject)(_res)
	if res == C.Py_None || res == nil {
		output = nil
		return nil
	}

	switch reflect.TypeOf(output).Elem().Kind() {
	case reflect.String:
		reflect.ValueOf(output).Elem().SetString(C.GoString(C.PyUnicode_AsUTF8(res)))
	case reflect.Int:
		cIntValue := C.PyLong_AsLongLong(res)
		reflect.ValueOf(output).Elem().SetInt(int64(cIntValue))
	case reflect.Float64:
		reflect.ValueOf(output).Elem().SetFloat(float64(C.PyFloat_AsDouble(res)))
	case reflect.Bool:
		reflect.ValueOf(output).Elem().SetBool(C.PyObject_IsTrue(res) != 0)
	case reflect.Slice:
		listLength := int(C.PyList_Size(res))
		var newSlice []interface{}
		err := sliceOutput(pyInit, &newSlice, _res, listLength, reflect.TypeOf(output).Elem())
		if err != nil {
			return err
		}
		outSlice := makeSliceOfType(reflect.TypeOf(output).Elem().Elem(), listLength)
		fillOutSlice(outSlice, newSlice)
		reflect.ValueOf(output).Elem().Set(outSlice)
	case reflect.Map:
		mapType := reflect.TypeOf(output).Elem()
		goMap := reflect.MakeMap(mapType)
		err := mapOutput(pyInit, &goMap, unsafe.Pointer(res))
		if err != nil {
			return err
		}
		reflect.ValueOf(output).Elem().Set(goMap)
	case reflect.Struct:
		if reflect.TypeOf(output).Elem() == reflect.TypeOf(DynamicFunc{}) {
			err := createDynamicFunc(pyInit, output.(*DynamicFunc), unsafe.Pointer(res))
			if err != nil {
				return err
			}
			return nil
		}
		instance, err := createStructFromInstance(pyInit, unsafe.Pointer(res), reflect.TypeOf(output).Elem())
		if err != nil {
			return err
		}
		reflect.ValueOf(output).Elem().Set(instance)
	default:
		return errors.New(fmt.Sprintf("unhandled output type %s", reflect.TypeOf(output).Elem().Kind().String()))
	}
	return nil
}

// sliceOutput processing the slice output.
// outputType - slice data type.
func sliceOutput(pyInit *module.InitPython, slicePtr interface{}, _res unsafe.Pointer, listLength int, outputType reflect.Type) error {
	if reflect.TypeOf(slicePtr).Kind() != reflect.Pointer {
		panic("slicePtr variable must be a pointer")
	}

	res := (*C.PyObject)(_res)
	var tempSlice []interface{}

	for i := 0; i < listLength; i++ {
		elem := C.PyList_GetItem(res, C.long(i))
		switch outputType.Elem().Kind() {
		case reflect.String:
			tempSlice = append(tempSlice, C.GoString(C.PyUnicode_AsUTF8(elem)))
		case reflect.Int:
			tempSlice = append(tempSlice, int(C.PyLong_AsLongLong(elem)))
		case reflect.Float64:
			tempSlice = append(tempSlice, float64(C.PyFloat_AsDouble(elem)))
		case reflect.Bool:
			tempSlice = append(tempSlice, C.PyObject_IsTrue(elem) != 0)
		case reflect.Slice:
			listLength := int(C.PyList_Size(elem))
			var newSlice1 []interface{}
			err := sliceOutput(pyInit, &newSlice1, unsafe.Pointer(elem), listLength, outputType.Elem())
			if err != nil {
				return err
			}
			tempSlice = append(tempSlice, newSlice1)
		case reflect.Map:
			mapType := outputType.Elem()
			goMap := reflect.MakeMap(mapType)
			err := mapOutput(pyInit, &goMap, unsafe.Pointer(elem))
			if err != nil {
				return err
			}
			tempSlice = append(tempSlice, goMap)
		case reflect.Struct:
			instance, err := createStructFromInstance(pyInit, unsafe.Pointer(elem), outputType.Elem())
			if err != nil {
				return err
			}
			tempSlice = reflect.Append(reflect.ValueOf(tempSlice), reflect.ValueOf(instance)).Interface().([]interface{})
		default:
			return errors.New(fmt.Sprintf("unhandled slice type %s", outputType.Elem().Kind().String()))
		}
	}
	reflect.ValueOf(slicePtr).Elem().Set(reflect.ValueOf(tempSlice))
	return nil
}

// fillOutSlice populates a new slice with data from a slice of type interface.
// For example: outSlice is of type []string, and _newSlice is []interface. This method will move the data from
// _newSlice to outSlice, thus the data will now be of type []string, rather than []interface.
func fillOutSlice(outSlice reflect.Value, _newSlice []interface{}) {
	for i := 0; i < len(_newSlice); i++ {
		if outSlice.Len() == 0 {
			ofType := makeSliceOfType(reflect.TypeOf(outSlice.Interface()).Elem(), len(_newSlice))
			outSlice.Set(ofType)
		}
		if reflect.TypeOf(_newSlice[i]).Kind() != reflect.Slice {
			var val reflect.Value
			if reflect.TypeOf(_newSlice[i]).Kind() == reflect.Struct {
				val = _newSlice[i].(reflect.Value)
			} else {
				val = reflect.ValueOf(_newSlice[i])
			}
			outSlice.Index(i).Set(val)
		} else {
			n := _newSlice[i].([]interface{})
			fillOutSlice(outSlice.Index(i), n)
		}
	}
}

func makeSliceOfType(k reflect.Type, length int) reflect.Value {
	sliceType := reflect.SliceOf(k)
	slice := reflect.MakeSlice(sliceType, length, length)
	return slice
}

// mapOutput processing map output.
func mapOutput(pyInit *module.InitPython, goMap *reflect.Value, _res unsafe.Pointer) error {
	res := (*C.PyObject)(_res)
	pyKeys := C.PyDict_Keys(res)
	keysLen := int(C.PyList_Size(pyKeys))
	for i := 0; i < keysLen; i++ {
		key := C.PyList_GetItem(pyKeys, C.long(i))
		pyValue := C.PyDict_GetItem(res, key)
		goKey, err := getMapOutputCObject(pyInit, pytypes.ObjectPtr(key), goMap)
		if err != nil {
			return err
		}
		goVal, err := getMapOutputCObject(pyInit, pytypes.ObjectPtr(pyValue), goMap)
		if err != nil {
			return err
		}
		goMap.SetMapIndex(goKey, goVal)
	}
	return nil
}

func getMapOutputCObject(pyInit *module.InitPython, pyObject pytypes.ObjectPtr, outputMap *reflect.Value) (reflect.Value, error) {
	var value reflect.Value
	cPyObject := (*C.PyObject)(unsafe.Pointer(pyObject))
	pyType := C.GoString(C.PyUnicode_AsUTF8(C.PyObject_GetAttrString(C.PyObject_Type(cPyObject), C.CString("__name__"))))
	switch pyType {
	case "str":
		value = reflect.ValueOf(C.GoString(C.PyUnicode_AsUTF8(cPyObject)))
	case "int":
		value = reflect.ValueOf(int(C.PyLong_AsLongLong(cPyObject)))
	case "float":
		value = reflect.ValueOf(float64(C.PyFloat_AsDouble(cPyObject)))
	case "bool":
		value = reflect.ValueOf(C.PyObject_IsTrue(cPyObject) != 0)
	case "list":
		listLength := int(C.PyList_Size(cPyObject))
		var newSlice []interface{}
		err := sliceOutput(pyInit, &newSlice, unsafe.Pointer(cPyObject), listLength, outputMap.Type().Elem())
		if err != nil {
			panic(err)
		}
		convertedSlice := makeSliceOfType(outputMap.Type().Elem().Elem(), len(newSlice))
		fillOutSlice(convertedSlice, newSlice)
		value = convertedSlice
	case "dict":
		mapType := outputMap.Type().Elem()
		goMap := reflect.MakeMap(mapType)
		err := mapOutput(pyInit, &goMap, unsafe.Pointer(cPyObject))
		if err != nil {
			return reflect.Value{}, err
		}
		value = goMap
	default:
		// Handle struct
		if outputMap.Type().Elem().Kind() == reflect.Struct {
			instance, err := createStructFromInstance(pyInit, unsafe.Pointer(cPyObject), outputMap.Type().Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			value = instance
		}
	}
	return value, nil
}

// createStructFromInstance creating a structure that represents the passed instance of the python class.
func createStructFromInstance(pyInit *module.InitPython, instance unsafe.Pointer, itype reflect.Type) (reflect.Value, error) {
	instanceC := (*C.PyObject)(instance)
	newStruct := reflect.New(itype).Elem()
	class := goclass.Class{}
	class.SetInstance(pytypes.ClassInstance(instanceC))
	moduleName, err := getPyModuleNameFromInstance(instanceC)
	if err != nil {
		return reflect.Value{}, err
	}

	newPyModule, err := pyInit.GetPyModule(moduleName)
	if err != nil {
		return reflect.Value{}, err
	}
	createClass, err := GetPyClass(newStruct.Type().Name(), newPyModule)
	if err != nil {
		return reflect.Value{}, err
	}
	pyInit.FreeObject(unsafe.Pointer(createClass))
	class.SetClass(createClass)
	class.SetPyModule(newPyModule)
	newStruct.FieldByName("Class").Set(reflect.ValueOf(class))
	newStruct.FieldByName("PyInit").Set(reflect.ValueOf(pyInit))
	return newStruct, nil
}

// getPyModuleNameFromInstance returns the name of the module for an instance of the class.
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

type DynamicFunc struct {
	PyInit   *module.InitPython
	PyModule pytypes.Module
	Name     string
}

func (fn *DynamicFunc) Call(args ...interface{}) (*int, error) {
	res, err := CallModuleMethod(fn.PyInit, fn.PyModule, fn.Name, args...)
	if err != nil {
		return nil, err
	}
	var output int
	err = MethodOutput(fn.PyInit, res, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func getResFuncName(resFunc unsafe.Pointer) (string, error) {
	fn := (*C.PyObject)(resFunc)
	funcAttrName := C.CString("__name__")
	defer C.free(unsafe.Pointer(funcAttrName))

	pyFuncAttr := C.PyObject_GetAttrString(fn, funcAttrName)
	if pyFuncAttr == nil {
		return "", errors.New("failed to get func name attribute")
	}
	defer C.Py_DecRef(pyFuncAttr)

	pyFuncName := C.PyUnicode_AsUTF8(pyFuncAttr)
	if pyFuncName == nil {
		return "", errors.New("failed to convert module name to string")
	}
	return C.GoString(pyFuncName), nil
}

func createDynamicFunc(pyInit *module.InitPython, output *DynamicFunc, resFunc unsafe.Pointer) error {
	fn := (*C.PyObject)(resFunc)
	funcName, err := getResFuncName(resFunc)
	if err != nil {
		return err
	}
	moduleName, err := getPyModuleNameFromInstance(fn)
	if err != nil {
		return err
	}
	pyModule, err := pyInit.GetPyModule(moduleName)
	if err != nil {
		return err
	}
	output.PyInit = pyInit
	output.PyModule = pyModule
	output.Name = funcName
	return nil
}
