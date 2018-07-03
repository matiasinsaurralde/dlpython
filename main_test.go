package python

import (
	"testing"
)

func TestFindPythonConfig(t *testing.T) {
	err := FindPythonConfig("0.0")
	if err == nil {
		t.Fatal("Should fail when loading a nonexistent Python version")
	}
	err = FindPythonConfig("3")
	if err != nil {
		t.Fatal("Couldn't find Python 3.x")
	}
}

func TestInit(t *testing.T) {
	err := FindPythonConfig("3")
	if err != nil {
		t.Fatal("Couldn't find Python 3.x")
	}
	err = Init()
	if err != nil {
		t.Fatal("Couldn't load Python runtime")
	}
	moduleName := PyUnicodeFromString("json")
	if moduleName == nil {
		t.Fatal("Couldn't initialize test Python string")
	}
	jsonModule := PyImportImport(moduleName)
	if jsonModule == nil {
		t.Fatal("Couldn't load json module")
	}
}
