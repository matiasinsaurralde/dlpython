dlpython
==

[![Build Status](https://travis-ci.org/matiasinsaurralde/dlpython.svg?branch=master)](https://travis-ci.org/matiasinsaurralde/dlpython)

Dynamic Python bindings for [Golang](https://golang.org).


## Sample usage

```go
package main

import (
	"os"

	python "github.com/matiasinsaurralde/dlpython"
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
    python.PyRunSimpleString("print(1)")
}

```