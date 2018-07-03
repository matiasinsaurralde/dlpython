package python

/*
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>

void* python_lib;

typedef struct _object {} PyObject;

typedef void (*Py_Initialize_f)();
Py_Initialize_f _Py_Initialize;
void Py_Initialize() { _Py_Initialize(); };

typedef PyObject* (*PyUnicode_FromString_f)(const char*);
PyUnicode_FromString_f _PyUnicode_FromString;
PyObject* PyUnicode_FromString(const char* u) { return _PyUnicode_FromString(u); };

typedef PyObject* (*PyImport_Import_f)(PyObject*);
PyImport_Import_f _PyImport_Import;
PyObject* PyImport_Import(PyObject* m) { return _PyImport_Import(m); };

typedef PyObject* (*PyModule_GetDict_f)(PyObject*);
PyModule_GetDict_f _PyModule_GetDict;
PyObject* PyModule_GetDict(PyObject* p) { return _PyModule_GetDict(p); };

typedef PyObject* (*PyDict_GetItemString_f)(PyObject*, const char*);
PyDict_GetItemString_f _PyDict_GetItemString;
PyObject* PyDict_GetItemString(PyObject* a, const char* b) { return _PyDict_GetItemString(a,b); };

typedef PyObject* (*PyObject_CallObject_f)(PyObject*, void*);
PyObject_CallObject_f _PyObject_CallObject;
PyObject* PyObject_CallObject(PyObject* p) {
	return _PyObject_CallObject(p, NULL);
 };

 typedef void* (*PyTuple_GetItem_f)(void*, int);
 PyTuple_GetItem_f _PyTuple_GetItem;
 void* PyTuple_GetItem(void* object, int index) {
	 return _PyTuple_GetItem(object, index);
 };

 typedef char* (*PyBytes_AsString_f)(void*);
 PyBytes_AsString_f _PyBytes_AsString;
 char* PyBytes_AsString(void* object) {
	 return _PyBytes_AsString(object);
 };

typedef void (*PyRunSimpleString_f)(const char*);
PyRunSimpleString_f _PyRunSimpleString;
void PyRunSimpleString(const char* m) {
	_PyRunSimpleString(m);
};

*/
import "C"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unsafe"
)

var (
	errEmptyPath      = errors.New("Empty PATH")
	errLibNotFound    = errors.New("Library not found")
	errLibLoad        = errors.New("Couldn't load library")
	errOSNotSupported = errors.New("OS not supported")

	pythonConfigExpr = regexp.MustCompile(`python(.*)-config`)

	pythonConfigPath  string
	pythonLibraryPath string
)

// FindPythonConfig scans PATH for common python-config locations.
func FindPythonConfig(prefix string) error {
	// Not sure if this can be replaced with os.LookPath:
	paths := os.Getenv("PATH")
	if paths == "" {
		return errEmptyPath
	}
	pythonConfigBinaries := []string{}
	for _, p := range strings.Split(paths, ":") {
		files, err := ioutil.ReadDir(p)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := f.Name()
			matches := pythonConfigExpr.FindAllStringSubmatch(name, -1)
			if len(matches) > 0 {
				version := matches[0][1]
				fmt.Println("Found Python installation", version, name)
				if strings.HasPrefix(version, prefix) {
					fullPath := filepath.Join(p, name)
					pythonConfigBinaries = append(pythonConfigBinaries, fullPath)
				}
			}
		}
	}

	// Pick the first item:
	for _, p := range pythonConfigBinaries {
		pythonConfigPath = p
		break
	}
	err := getLibraryPath()
	if err != nil {
		return err
	}
	return nil
}

func getLibraryPath() error {
	out, err := exec.Command(pythonConfigPath, "--ldflags").Output()
	if err != nil {
		return err
	}
	outString := string(out)
	var libDir string
	var libName string
	for _, v := range strings.Split(outString, " ") {
		prefix := v[0:2]
		switch prefix {
		case "-L":
			libDir = strings.Replace(v, prefix, "", -1)
		case "-l":
			if strings.Contains(v, "python") {
				libName = strings.Replace(v, prefix, "", -1)
			}
		}
	}
	switch runtime.GOOS {
	case "darwin":
		libName = "lib" + libName + ".dylib"
	case "linux":
		libName = "lib" + libName + ".so"
	default:
		return errOSNotSupported
	}
	pythonLibraryPath = filepath.Join(libDir, libName)
	if _, err := os.Stat(pythonLibraryPath); os.IsNotExist(err) {
		return errLibNotFound
	}
	return nil
}

func mapCalls() {
	CPyInitializeSym := C.CString("Py_Initialize")
	defer C.free(unsafe.Pointer(CPyInitializeSym))
	C._Py_Initialize = C.Py_Initialize_f(C.dlsym(C.python_lib, CPyInitializeSym))

	CPyUnicodeFromStringSym := C.CString("PyUnicode_FromString")
	defer C.free(unsafe.Pointer(CPyUnicodeFromStringSym))
	C._PyUnicode_FromString = C.PyUnicode_FromString_f(C.dlsym(C.python_lib, CPyUnicodeFromStringSym))

	CPyImportImportSym := C.CString("PyImport_Import")
	defer C.free(unsafe.Pointer(CPyImportImportSym))
	C._PyImport_Import = C.PyImport_Import_f(C.dlsym(C.python_lib, CPyImportImportSym))

	CPyModuleGetDict := C.CString("PyModule_GetDict")
	defer C.free(unsafe.Pointer(CPyModuleGetDict))
	C._PyModule_GetDict = C.PyModule_GetDict_f(C.dlsym(C.python_lib, CPyModuleGetDict))

	CPyDictGetItemString := C.CString("PyDict_GetItemString")
	defer C.free(unsafe.Pointer(CPyDictGetItemString))
	C._PyDict_GetItemString = C.PyDict_GetItemString_f(C.dlsym(C.python_lib, CPyDictGetItemString))

	CPyObjectCallObject := C.CString("PyObject_CallObject")
	defer C.free(unsafe.Pointer(CPyObjectCallObject))
	C._PyObject_CallObject = C.PyObject_CallObject_f(C.dlsym(C.python_lib, CPyObjectCallObject))

	CPyRunSimpleStringSym := C.CString("PyRun_SimpleString")
	defer C.free(unsafe.Pointer(CPyRunSimpleStringSym))
	C._PyRunSimpleString = C.PyRunSimpleString_f(C.dlsym(C.python_lib, CPyRunSimpleStringSym))
}

// PyInitialize is a wrapper.
func PyInitialize() {
	C.Py_Initialize()
}

func PyUnicodeFromString(input string) *C.PyObject {
	Cinput := C.CString(input)
	ptr := C.PyUnicode_FromString(Cinput)
	return ptr
}

func PyImportImport(moduleName *C.PyObject) *C.PyObject {
	return C.PyImport_Import(moduleName)
}

func PyModuleGetDict(p *C.PyObject) *C.PyObject {
	return C.PyModule_GetDict(p)
}

func PyDictGetItemString(o *C.PyObject, b string) *C.PyObject {
	cstr := C.CString(b)
	defer C.free(unsafe.Pointer(cstr))
	return C.PyDict_GetItemString(o, cstr)
}

func PyObjectCallObject(o *C.PyObject) {
	C.PyObject_CallObject(o)
}

func PyTupleGetItem(o unsafe.Pointer, i int) unsafe.Pointer {
	ci := C.int(i)
	return C.PyTuple_GetItem(o, ci)
}

func PyRunSimpleString(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.PyRunSimpleString(cs)
}

func loadLibrary() error {
	libPath := C.CString(pythonLibraryPath)
	defer C.free(unsafe.Pointer(libPath))
	C.python_lib = C.dlopen(libPath, C.RTLD_NOW|C.RTLD_GLOBAL)
	if C.python_lib == nil {
		return errLibLoad
	}
	return nil
}

// Init will initialize the Python runtime.
func Init() error {
	// Try to load the library:
	err := loadLibrary()
	if err != nil {
		return err
	}
	// Map API calls and initialize runtime:
	mapCalls()
	PyInitialize()
	return nil
}
