package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var version = "0.1.2"

var (
	trace     bool
	bigint    bool
	stat      bool
	check     bool
	ver       bool
	pluginDir string
)

func init() {
	flag.BoolVar(&trace, "trace", false, "trace function calls")
	flag.BoolVar(&trace, "t", false, "trace function calls (shorthand)")

	flag.BoolVar(&bigint, "big", false, "use big math")
	flag.BoolVar(&bigint, "b", false, "use big math (shorthand)")

	flag.BoolVar(&stat, "stat", false, "dump statistics after program exit")
	flag.BoolVar(&stat, "s", false, "dump statistics after program exit (shorthand)")

	flag.BoolVar(&check, "check", false, "make parsing and typechecking only")
	flag.BoolVar(&check, "c", false, "make parsing and typechecking only (shorthand)")

	flag.StringVar(&pluginDir, "plugin-dir", "", "plugins directory")
	flag.StringVar(&pluginDir, "p", "", "plugins directory (shorthand)")

	flag.BoolVar(&ver, "version", false, "show version")
	flag.BoolVar(&ver, "v", false, "show version")
}

func doMain() int {
	flag.Parse()
	if ver {
		showVersion()
		return 0
	}
	if !trace {
		log.SetOutput(ioutil.Discard)
	}

	in := NewInterpreter(os.Stdout)
	in.UseBigInt(bigint)
	in.PluginDir = pluginDir
	in.IncludeDirs = []string{in.PluginDir}

	var file string
	var input io.Reader
	if len(flag.Args()) >= 1 {
		fname := flag.Arg(0)
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		defer f.Close()
		input = f
		file, err = filepath.Abs(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot determine absolute path for %q: %e", file, err)
			file = fname
		}
	} else {
		input = os.Stdin
		file = "__stdin__"
	}

	if err := in.Parse(file, input); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	if errs := in.Check(); len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		return 1
	}

	if check {
		return 0
	}

	if err := in.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	if stat {
		in.Stat()
	}
	return 0
}

func main() {
	os.Exit(doMain())
}

func showVersion() {
	fmt.Printf("SPIL version %v\n", version)
}
