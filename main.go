package python

/*
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>

void* python_library;

typedef void (*py_initialize_f)();
py_initialize_f py_initialize;

typedef void (*pyrun_simplestring_f)(const char*);
pyrun_simplestring_f pyrun_simplestring;

void map_calls() {
	py_initialize = dlsym(python_library, "Py_Initialize");
	pyrun_simplestring = dlsym(python_library, "PyRun_SimpleString");
}

void Py_Initialize() {
	py_initialize();
}

void PyRun_SimpleString(const char* input) {
	pyrun_simplestring(input);
}

int load_library(char* path) {
	python_library = dlopen(path, RTLD_NOW | RTLD_LOCAL);
	if(python_library == NULL) {
		return -1;
	}
	return 0;
}
*/
import "C"

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unsafe"
)

// ABC is abc.
const ABC = 1

var (
	errEmptyPath   = errors.New("Empty PATH")
	errLibNotFound = errors.New("Library not found")
	errLibLoad     = errors.New("Couldn't load library")

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
		// Not supported
	}
	pythonLibraryPath = filepath.Join(libDir, libName)
	if _, err := os.Stat(pythonLibraryPath); os.IsNotExist(err) {
		return errLibNotFound
	}
	return nil
}

func mapCalls() {
	C.map_calls()
}

func loadLibrary() error {
	libPath := C.CString(pythonLibraryPath)
	defer C.free(unsafe.Pointer(libPath))
	result := C.load_library(libPath)
	if result == -1 {
		return errLibLoad
	}
	return nil
}

// PyInitialize wraps a C call.
func PyInitialize() {
	C.Py_Initialize()
}

// PyRunSimpleString wraps a C call.
func PyRunSimpleString(input string) {
	s := C.CString(input)
	defer C.free(unsafe.Pointer(s))
	C.PyRun_SimpleString(s)
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
