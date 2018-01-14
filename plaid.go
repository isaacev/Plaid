package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"plaid/lang"
	"plaid/lib"
)

func main() {
	if len(os.Args[1:]) >= 1 {
		if errs := run(os.Args[1]); len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			os.Exit(1)
		}
	}
}

func read(filename string) (src string, errs []error) {
	if contents, err := ioutil.ReadFile(filename); err == nil {
		return string(contents), nil
	} else {
		return "", []error{err}
	}
}

func run(filename string) (errs []error) {
	var src string
	var ast *lang.RootNode
	var mod *lang.VirtualModule

	fmt.Println("\n=== SOURCE CODE")
	if src, errs = read(filename); len(errs) > 0 {
		return errs
	} else {
		fmt.Println(src)
	}

	fmt.Println("\n=== SYNTAX TREE")
	if ast, errs = lang.Parse(filename, src); len(errs) > 0 {
		return errs
	} else {
		fmt.Println(ast)
	}

	if mod, errs = lang.Link(filename, ast); len(errs) > 0 {
		return errs
	}

	fmt.Println("\n=== MODULE")
	if _, errs = lang.Check(mod, lib.IO, lib.Conv); len(errs) > 0 {
		return errs
	} else {
		fmt.Println(mod)
	}

	return nil
}
