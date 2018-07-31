package python

/*
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>

void* python_lib;
typedef struct _pyobject {} PyObject;
typedef struct _pythreadstate {} PyThreadState;
typedef struct _pygilstate {} PyGILState_STATE;


typedef PyObject* (*PyObject_GetAttr_f)(PyObject*, PyObject*);
PyObject_GetAttr_f PyObject_GetAttr_ptr;
PyObject* PyObject_GetAttr(PyObject* arg0, PyObject* arg1) { return PyObject_GetAttr_ptr(arg0, arg1); };

typedef char* (*PyBytes_AsString_f)(PyObject*);
PyBytes_AsString_f PyBytes_AsString_ptr;
char* PyBytes_AsString(PyObject* arg0) { return PyBytes_AsString_ptr(arg0); };

typedef PyObject* (*PyUnicode_FromString_f)(char*);
PyUnicode_FromString_f PyUnicode_FromString_ptr;
PyObject* PyUnicode_FromString(char* u) { return PyUnicode_FromString_ptr(u); };

typedef long int (*PyLong_AsLong_f)(PyObject*);
PyLong_AsLong_f PyLong_AsLong_ptr;
long int PyLong_AsLong(PyObject* arg0) { return PyLong_AsLong_ptr(arg0); };

typedef PyObject* (*PyDict_GetItemString_f)(PyObject*, char*);
PyDict_GetItemString_f PyDict_GetItemString_ptr;
PyObject* PyDict_GetItemString(PyObject* dp, char* key) { return PyDict_GetItemString_ptr(dp, key); };

typedef PyObject* (*PyModule_GetDict_f)(PyObject*);
PyModule_GetDict_f PyModule_GetDict_ptr;
PyObject* PyModule_GetDict(PyObject* arg0) { return PyModule_GetDict_ptr(arg0); };

typedef void (*PyErr_Print_f)();
PyErr_Print_f PyErr_Print_ptr;
void PyErr_Print() { return PyErr_Print_ptr(); };

typedef void (*Py_Initialize_f)();
Py_Initialize_f Py_Initialize_ptr;
void Py_Initialize() { return Py_Initialize_ptr(); };

typedef int (*Py_IsInitialized_f)();
Py_IsInitialized_f Py_IsInitialized_ptr;
int Py_IsInitialized() { return Py_IsInitialized_ptr(); };

typedef PyThreadState* (*PyEval_SaveThread_f)();
PyEval_SaveThread_f PyEval_SaveThread_ptr;
PyThreadState* PyEval_SaveThread() { return PyEval_SaveThread_ptr(); };

typedef void (*PyEval_InitThreads_f)();
PyEval_InitThreads_f PyEval_InitThreads_ptr;
void PyEval_InitThreads() { return PyEval_InitThreads_ptr(); };

typedef PyObject* (*PyImport_Import_f)(PyObject*);
PyImport_Import_f PyImport_Import_ptr;
PyObject* PyImport_Import(PyObject* name) { return PyImport_Import_ptr(name); };

typedef PyObject* (*PyObject_CallObject_f)(PyObject*, PyObject*);
PyObject_CallObject_f PyObject_CallObject_ptr;
PyObject* PyObject_CallObject(PyObject* callable_object, PyObject* args) { return PyObject_CallObject_ptr(callable_object, args); };

*/
import "C"
import "unsafe"

type dummyPtr unsafe.Pointer

func PyObject_GetAttr(arg0 *C.PyObject, arg1 *C.PyObject) *C.PyObject {
	return C.PyObject_GetAttr(arg0, arg1)
}

func PyBytes_AsString(arg0 *C.PyObject) *C.char {
	return C.PyBytes_AsString(arg0)
}

func PyUnicode_FromString(u *C.char) *C.PyObject {
	return C.PyUnicode_FromString(u)
}

func PyLong_AsLong(arg0 *C.PyObject) C.long {
	return C.PyLong_AsLong(arg0)
}

func PyDict_GetItemString(dp *C.PyObject, key *C.char) *C.PyObject {
	return C.PyDict_GetItemString(dp, key)
}

func PyModule_GetDict(arg0 *C.PyObject) *C.PyObject {
	return C.PyModule_GetDict(arg0)
}

func PyErr_Print() {
	C.PyErr_Print()
}

func Py_Initialize() {
	C.Py_Initialize()
}

func Py_IsInitialized() C.int {
	return C.Py_IsInitialized()
}

func PyEval_SaveThread() *C.PyThreadState {
	return C.PyEval_SaveThread()
}

func PyEval_InitThreads() {
	C.PyEval_InitThreads()
}

func PyImport_Import(name *C.PyObject) *C.PyObject {
	return C.PyImport_Import(name)
}

func PyObject_CallObject(callable_object *C.PyObject, args *C.PyObject) *C.PyObject {
	return C.PyObject_CallObject(callable_object, args)
}

func mapCalls() {
	C.python_lib = C.dlopen(libPath, C.RTLD_NOW|C.RTLD_GLOBAL)
	s_PyObject_GetAttr := C.CString("PyObject_GetAttr")
	defer C.free(unsafe.Pointer(s_PyObject_GetAttr))
	C.PyObject_GetAttr_ptr = C.PyObject_GetAttr_f(C.dlsym(C.python_lib, s_PyObject_GetAttr))

	s_PyBytes_AsString := C.CString("PyBytes_AsString")
	defer C.free(unsafe.Pointer(s_PyBytes_AsString))
	C.PyBytes_AsString_ptr = C.PyBytes_AsString_f(C.dlsym(C.python_lib, s_PyBytes_AsString))

	s_PyUnicode_FromString := C.CString("PyUnicode_FromString")
	defer C.free(unsafe.Pointer(s_PyUnicode_FromString))
	C.PyUnicode_FromString_ptr = C.PyUnicode_FromString_f(C.dlsym(C.python_lib, s_PyUnicode_FromString))

	s_PyLong_AsLong := C.CString("PyLong_AsLong")
	defer C.free(unsafe.Pointer(s_PyLong_AsLong))
	C.PyLong_AsLong_ptr = C.PyLong_AsLong_f(C.dlsym(C.python_lib, s_PyLong_AsLong))

	s_PyDict_GetItemString := C.CString("PyDict_GetItemString")
	defer C.free(unsafe.Pointer(s_PyDict_GetItemString))
	C.PyDict_GetItemString_ptr = C.PyDict_GetItemString_f(C.dlsym(C.python_lib, s_PyDict_GetItemString))

	s_PyModule_GetDict := C.CString("PyModule_GetDict")
	defer C.free(unsafe.Pointer(s_PyModule_GetDict))
	C.PyModule_GetDict_ptr = C.PyModule_GetDict_f(C.dlsym(C.python_lib, s_PyModule_GetDict))

	s_PyErr_Print := C.CString("PyErr_Print")
	defer C.free(unsafe.Pointer(s_PyErr_Print))
	C.PyErr_Print_ptr = C.PyErr_Print_f(C.dlsym(C.python_lib, s_PyErr_Print))

	s_Py_Initialize := C.CString("Py_Initialize")
	defer C.free(unsafe.Pointer(s_Py_Initialize))
	C.Py_Initialize_ptr = C.Py_Initialize_f(C.dlsym(C.python_lib, s_Py_Initialize))

	s_Py_IsInitialized := C.CString("Py_IsInitialized")
	defer C.free(unsafe.Pointer(s_Py_IsInitialized))
	C.Py_IsInitialized_ptr = C.Py_IsInitialized_f(C.dlsym(C.python_lib, s_Py_IsInitialized))

	s_PyEval_SaveThread := C.CString("PyEval_SaveThread")
	defer C.free(unsafe.Pointer(s_PyEval_SaveThread))
	C.PyEval_SaveThread_ptr = C.PyEval_SaveThread_f(C.dlsym(C.python_lib, s_PyEval_SaveThread))

	s_PyEval_InitThreads := C.CString("PyEval_InitThreads")
	defer C.free(unsafe.Pointer(s_PyEval_InitThreads))
	C.PyEval_InitThreads_ptr = C.PyEval_InitThreads_f(C.dlsym(C.python_lib, s_PyEval_InitThreads))

	s_PyImport_Import := C.CString("PyImport_Import")
	defer C.free(unsafe.Pointer(s_PyImport_Import))
	C.PyImport_Import_ptr = C.PyImport_Import_f(C.dlsym(C.python_lib, s_PyImport_Import))

	s_PyObject_CallObject := C.CString("PyObject_CallObject")
	defer C.free(unsafe.Pointer(s_PyObject_CallObject))
	C.PyObject_CallObject_ptr = C.PyObject_CallObject_f(C.dlsym(C.python_lib, s_PyObject_CallObject))

}
