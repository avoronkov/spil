package main

import (
	"fmt"
	"io"
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
		for _, bigint := range []bool{false, true} {
			name := test
			if bigint {
				name += "-big"
			}
			t.Run(name, func(t *testing.T) {
				dir, file := filepath.Split(test)
				output := filepath.Join(dir, "out"+file[2:])
				checkInterpreter(t, test, output, false, bigint)
			})
		}
	}
}

func TestBuiltin(t *testing.T) {
	inputs, err := filepath.Glob("examples/builtin.*")
	if err != nil {
		panic(err)
	}
	for _, test := range inputs {
		for _, bigint := range []bool{false, true} {
			name := test
			if bigint {
				name += "-big"
			}
			t.Run(name, func(t *testing.T) {
				dir, file := filepath.Split(test)
				output := filepath.Join(dir, "output."+file)
				checkInterpreter(t, test, output, true, bigint)
			})
		}
	}
}

func checkInterpreter(t *testing.T, input, output string, builtin, bigint bool) {
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

	builtinDir := ""
	if builtin {
		builtinDir = "./builtin"
	}

	buffer := &strings.Builder{}
	in := NewInterpreter(buffer, builtinDir)
	in.UseBigInt(bigint)

	inputPath, err := filepath.Abs(input)
	if err != nil {
		t.Fatalf("Abs(%v) failed: %v", input, err)
	}
	if err := run(in, inputPath, fin); err != nil {
		t.Fatalf("Interpreter Run() failed: %v", err)
	}

	if act, exp := buffer.String(), string(expData); act != exp {
		t.Errorf("Incorrect output for %v: expected %q, actual %q", input, exp, act)
	}
}

func run(i *Interpret, file string, input io.Reader) error {
	if err := i.Parse(file, input); err != nil {
		return err
	}
	if err := i.Check(); err != nil {
		return fmt.Errorf("Check failed: %v", err)
	}
	return i.Run()
}
