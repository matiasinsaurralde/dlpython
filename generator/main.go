package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xlab/c-for-go/parser"
	"github.com/xlab/c-for-go/translator"
)

const defaultHeader = `
/*
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>

void* python_lib;
typedef struct _pyobject {} PyObject;
typedef struct _pythreadstate {} PyThreadState;
typedef struct _pygilstate {} PyGILState_STATE;


`

// Pkg is the main package data structure.
type Pkg struct {
	Name string
	Fns  []FnWrapper
}

// WriteTo writes the package contents to w.
func (p *Pkg) WriteTo(w io.Writer) (int64, error) {
	// Write header:
	s := bytes.NewBufferString("package ")
	s.WriteString(p.Name + "\n")
	s.WriteString(defaultHeader)

	// Append generated C code:
	for _, fn := range p.Fns {
		s.WriteString(fn.ccImpl)
	}

	// End cgo block:
	s.WriteString("*/\n")
	s.WriteString("import \"C\"\n")
	s.WriteString("import \"unsafe\"\n")

	// Add a dummy "unsafe.Pointer" so that the package builds even when "unsafe" isn't in use:
	s.WriteString("type dummyPtr unsafe.Pointer\n")

	// Append generated Go code:
	for _, fn := range p.Fns {
		s.WriteString(fn.goImpl)
	}

	s.WriteString("func mapCalls() {\n")

	// Add dlopen call:
	s.WriteString("\tC.python_lib = C.dlopen(libPath, C.RTLD_NOW|C.RTLD_GLOBAL)\n")

	for _, fn := range p.Fns {
		sym := "s_" + fn.fnName
		symDecl := fmt.Sprintf("\t%s := C.CString(\"%s\")\n", sym, fn.fnName)
		s.WriteString(symDecl)

		deferFree := fmt.Sprintf("\tdefer C.free(unsafe.Pointer(%s))\n", sym)
		s.WriteString(deferFree)

		binding := fmt.Sprintf(
			"\tC.%s =  C.%s(C.dlsym(C.python_lib, %s))\n",
			fn.fnPtrName,
			fn.fnTypeName,
			sym,
		)
		s.WriteString(binding)
		s.WriteString("\n")
	}

	s.WriteString("}\n")

	// Finally write everything to w:
	return s.WriteTo(w)
}

// FnWrapper wraps different pieces of code related to a particular function.
type FnWrapper struct {
	decl   *translator.CDecl
	ccImpl string
	goImpl string

	ccParams map[string]string

	fnPtrName  string
	fnName     string
	fnTypeName string
}

func (f *FnWrapper) getGoType(input string) string {
	switch input {
	case "PyObject*":
		return "*C.PyObject"
	case "PyThreadState*":
		return "*C.PyThreadState"
	case "PyGILState_STATE":
		return "C.PyGILState_STATE"
	case "int":
		return "C.int"
	case "long int":
		return "C.long"
	case "void*":
		return "unsafe.Pointer"
	case "char*":
		return "*C.char"
	default:
		return "unsafe.Pointer"
	}
}

func (f *FnWrapper) getCCType(input string) string {
	switch input {
	case "PyObject*":
		return "PyObject*"
	case "int":
		return "int"
	case "char*":
		return "char*"
	default:
		return "void*"
	}
}

func (f *FnWrapper) buildCCParams() {
	if f.ccParams == nil {
		f.ccParams = make(map[string]string)
	}
	spec := f.decl.Spec.(*translator.CFunctionSpec)
	for _, p := range spec.Params {
		switch p.Spec.(type) {
		case *translator.CStructSpec:
			structSpec := p.Spec.(*translator.CStructSpec)
			f.ccParams[p.Name] = f.getCCType(structSpec.String())
		case *translator.CTypeSpec:
			typeSpec := p.Spec.(*translator.CTypeSpec)
			f.ccParams[p.Name] = f.getCCType(typeSpec.String())
		}
	}
}

// Generate walks through the declaration spec and generates code.
func (f *FnWrapper) Generate() error {
	// Cast the declaration spec to CFunctionSpec:
	spec := f.decl.Spec.(*translator.CFunctionSpec)
	if spec == nil {
		return nil
	}

	f.fnName = spec.String()
	f.fnPtrName = f.fnName + "_ptr"
	f.fnTypeName = f.fnName + "_f"
	paramsLength := len(spec.Params) - 1
	var notImpl bool
	var fnReturnType string
	if spec.Return == nil {
		fnReturnType = "void"
	} else {
		fnReturnType = spec.Return.String()
	}

	// Start building the C code:
	f.buildCCParams()
	ccCode := bytes.NewBufferString("typedef ")
	ccCode.WriteString(fnReturnType)

	// TODO: add args
	ccCode.WriteString(" (*" + f.fnTypeName + ")")
	ccCode.WriteString("(")
	// ccCode.WriteString(f.buildCCArgs(paramsLength, false))
	i := 0
	for _, t := range f.ccParams {
		ccCode.WriteString(t)
		if i < paramsLength {
			ccCode.WriteString(", ")
		}
		i++
	}
	ccCode.WriteString(");\n")

	ccCode.WriteString(f.fnTypeName + " " + f.fnPtrName + ";\n")
	// Write function wrapper:
	ccCode.WriteString(fnReturnType + " " + f.fnName + "(")
	i = 0
	for p, t := range f.ccParams {
		ccCode.WriteString(t + " " + p)
		if i < paramsLength {
			ccCode.WriteString(", ")
		}
		i++
	}
	// ccCode.WriteString(f.buildCCArgs(paramsLength, true))
	ccCode.WriteString(") ")
	ccCode.WriteString("{ return " + f.fnPtrName + "(")
	i = 0
	for p := range f.ccParams {
		ccCode.WriteString(p)
		if i < paramsLength {
			ccCode.WriteString(", ")
		}
		i++
	}
	ccCode.WriteString("); };\n")

	ccCode.WriteRune('\n')

	// Start building the Go function code:
	goCode := bytes.NewBufferString("func ")
	goCode.WriteString(spec.CGoName())

	// Build the args:
	goCode.WriteString("(")
	args := []string{}
	for i, p := range spec.Params {
		switch p.Spec.(type) {
		case *translator.CStructSpec:
			args = append(args, p.Name)
			structSpec := p.Spec.(*translator.CStructSpec)
			goCode.WriteString(p.Name)
			goCode.WriteString(" ")
			goCode.WriteString(f.getGoType(structSpec.String()))
		case *translator.CTypeSpec:
			args = append(args, p.Name)
			typeSpec := p.Spec.(*translator.CTypeSpec)
			goCode.WriteString(p.Name)
			goCode.WriteString(" ")
			goCode.WriteString(f.getGoType(typeSpec.String()))
		default:
			// Not handled
			notImpl = true
			break
		}
		if i < paramsLength {
			goCode.WriteString(", ")
		}
	}

	// TODO: display error
	if notImpl {
		f.goImpl = ""
		f.ccImpl = ""
		return nil
	}
	goCode.WriteString(") ")

	// Handle Go return value:
	if spec.Return != nil {
		goCode.WriteString(f.getGoType(spec.Return.String()))
	}

	// Build the function body:
	goCode.WriteString(" {\n")
	goCode.WriteRune('\t')

	if spec.Return == nil {
		goCode.WriteString("C." + f.fnName + "(")
	} else {
		goCode.WriteString("return C." + f.fnName + "(")
	}

	// Add args:
	goCode.WriteString(strings.Join(args, ","))
	goCode.WriteString(")")
	goCode.WriteRune('\n')
	goCode.WriteString("}\n")
	goCode.WriteRune('\n')

	f.ccImpl = ccCode.String()
	f.goImpl = goCode.String()

	return nil
}

func main() {
	// gcc -print-search-dirs -h
	cfg := &parser.Config{
		IncludePaths: []string{
			"/usr/include",
			"/usr/local/include",
			// "/Users/matias/.gvm/pkgsets/go1.10/global/src/github.com/matiasinsaurralde/go-python-dyn/c-for-go/headers",
			"/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/lib/clang/9.0.0/include",
		},
		SourcesPaths: []string{
			"/Users/matias/.gvm/pkgsets/go1.10/global/src/github.com/matiasinsaurralde/dlpython/c-for-go/headers/Python.h",
		},
		// Arch:    "x86_64",
		Defines: map[string]interface{}{
			// "LONG_BIT": "64",
		},
	}
	unit, err := parser.ParseWith(cfg)
	if err != nil {
		panic(err)
	}
	if unit != nil {
	}

	translatorCfg := &translator.Config{}
	tl, err := translator.New(translatorCfg)
	tl.Learn(unit)

	declares := tl.Declares()

	pkg := &Pkg{
		Name: "python",
		Fns:  make([]FnWrapper, 0),
	}

	whitelist := []string{
		"Py_Initialize",
		"Py_IsInitialized",
		"PyEval_InitThreads",
		"PyEval_SaveThread",
		"PyUnicode_FromString",
		"PyImport_Import",
		"PyModule_GetDict",
		"PyErr_Print",
		"PyDict_GetItemString", // Investigate args order
		"PyGILState_Ensure",
		"PyGILState_Release",
		"PyObject_GetAttr",
		"PyObject_CallObject",
		"PyBytes_AsString",
		"PyLong_AsLong",
		// "PyTuple_Pack",
	}

	for _, decl := range declares {
		switch decl.Spec.Kind() {
		case translator.FunctionKind:
			// Check if the function is part of the whitelist:
			var found bool
			for _, v := range whitelist {
				if v == decl.Name {
					found = true
					break
				}
			}
			if !found {
				continue
			}

			// Construct the function wrapper and generate the binding code:
			f := FnWrapper{decl: decl}
			if err := f.Generate(); err != nil {
				panic(err)
			}
			pkg.Fns = append(pkg.Fns, f)
		}
	}

	pkg.WriteTo(os.Stdout)
}
