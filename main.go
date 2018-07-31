package python

/*
#cgo LDFLAGS: -ldl

#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>
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
	var libDir, libName string
	splits := strings.Split(outString, " ")
	for _, v := range splits {
		if len(v) <= 2 {
			continue
		}
		prefix := v[0:2]
		switch prefix {
		case "-L":
			if libDir == "" {
				libDir = strings.Replace(v, prefix, "", -1)
			}
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

var libPath *C.char

// Init will initialize the Python runtime.
func Init() error {
	// Set the library path:
	libPath = C.CString(pythonLibraryPath)
	defer C.free(unsafe.Pointer(libPath))

	// Map API calls and initialize runtime:
	mapCalls()
	return nil
}
