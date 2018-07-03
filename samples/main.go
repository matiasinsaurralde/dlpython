package main

import (
	"os"

	"github.com/matiasinsaurralde/go-python-dyn"
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
}
