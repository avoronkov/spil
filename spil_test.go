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
			buffer := &strings.Builder{}
			in := NewInterpreter(fin, buffer)
			if err := in.Run(); err != nil {
				t.Fatalf("Interpreter Run() failed: %v", err)
			}

			expData, err := ioutil.ReadFile(output)
			if err != nil {
				t.Fatalf("Reading output file failed: %v", err)
			}
			if act, exp := buffer.String(), string(expData); act != exp {
				t.Errorf("Incorrect output for %v: expected %q, actual %q", test, exp, act)
			}
		})
	}
}
