package main

import (
	"os"

	"github.com/matiasinsaurralde/dlpython"
)

func main() {
	cwd, _ := os.Getwd()
	os.Setenv("PYTHONPATH", cwd)

	err := python.FindPythonConfig("3.5")
	if err != nil {
		panic(err)
	}
	err = python.Init()
	if err != nil {
		panic(err)
	}

	moduleName := python.PyUnicodeFromString("hello")
	module := python.PyImportImport(moduleName)
	dict := python.PyModuleGetDict(module)
	fn := python.PyDictGetItemString(dict, "myfn")
	python.PyObjectCallObject(fn)
	python.PyRunSimpleString("import google.protobuf.internal.api_implementation")
	python.PyRunSimpleString("import cffi")
	python.PyRunSimpleString("print(google.protobuf.internal.api_implementation.Type())")
}
