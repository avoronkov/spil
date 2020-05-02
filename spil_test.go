package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamples(t *testing.T) {
	inputs, err := filepath.Glob("examples/ex.*")
	if err != nil {
		panic(err)
	}

	for _, test := range inputs {
		t.Run(test, func(t *testing.T) {
			dir, file := filepath.Split(test)
			output := filepath.Join(dir, "out"+file[2:])
			fin, err := os.Open(test)
			if err != nil {
				t.Fatalf("Cannot open input file: %v", err)
			}
			defer fin.Close()

			expData, err := ioutil.ReadFile(output)
			if os.IsNotExist(err) {
				t.Skipf("No output file for %v found", test)
			}
			if err != nil {
				t.Fatalf("Reading output file failed: %v", err)
			}

			buffer := &strings.Builder{}
			in := NewInterpreter(buffer)
			if err := in.Run(fin); err != nil {
				t.Fatalf("Interpreter Run() failed: %v", err)
			}

			if act, exp := buffer.String(), string(expData); act != exp {
				t.Errorf("Incorrect output for %v: expected %q, actual %q", test, exp, act)
			}
		})
	}
}

func TestBuiltin(t *testing.T) {
	inputs, err := filepath.Glob("examples/builtin.*")
	if err != nil {
		panic(err)
	}
	for _, test := range inputs {
		t.Run(test, func(t *testing.T) {
			dir, file := filepath.Split(test)
			output := filepath.Join(dir, "output."+file)
			checkInterpreter(t, test, output, true)
		})
	}
}

func checkInterpreter(t *testing.T, input, output string, builtin bool) {
	fin, err := os.Open(input)
	if err != nil {
		t.Fatalf("Cannot open input file: %v", err)
	}
	defer fin.Close()

	expData, err := ioutil.ReadFile(output)
	if os.IsNotExist(err) {
		t.Skipf("No output file for %v found: %v", input, output)
	}
	if err != nil {
		t.Fatalf("Reading output file failed: %v", err)
	}

	buffer := &strings.Builder{}
	in := NewInterpreter(buffer)
	if builtin {
		if err := in.LoadBuiltin("./builtin"); err != nil {
			t.Fatal(err)
		}
	}
	if err := in.Run(fin); err != nil {
		t.Fatalf("Interpreter Run() failed: %v", err)
	}

	if act, exp := buffer.String(), string(expData); act != exp {
		t.Errorf("Incorrect output for %v: expected %q, actual %q", input, exp, act)
	}

}
