package main

import (
	"fmt"

	"github.com/matiasinsaurralde/go-python-dyn"
)

func main() {
	fmt.Println(python.ABC)
	err := python.FindPythonConfig("3")
	if err != nil {
		panic(err)
	}
	err = python.Init()
	if err != nil {
		panic(err)
	}
	python.PyRunSimpleString(`print("Hello from Python")`)
}
