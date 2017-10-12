package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"plaid/check"
	"plaid/codegen"
	"plaid/debug"
	"plaid/libs"
	"plaid/parser"
	"plaid/vm"
)

func main() {
	showAST := flag.Bool("ast", false, "output abstract syntax tree")
	showCheck := flag.Bool("check", false, "output type checker results")
	showIR := flag.Bool("ir", false, "output intermediate representation")
	showBC := flag.Bool("bytecode", false, "output bytecode")
	showOut := flag.Bool("out", false, "run program and print output")
	flag.Parse()

	for _, filename := range flag.Args() {
		processFile(filename, *showAST, *showCheck, *showIR, *showBC, *showOut)
	}
}

func processFile(filename string, showAST bool, showCheck bool, showIR bool, showBC bool, showOut bool) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	src := string(buf)
	ast, err := parser.Parse(src)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if showAST {
		fmt.Println(ast.String())
	}

	if showCheck || showIR || showBC || showOut {
		scope := check.Check(ast, libs.IO, libs.Conv)
		if len(scope.Errors()) > 0 {
			for i, err := range scope.Errors() {
				fmt.Printf("%4d %s\n", i, err)
			}
			os.Exit(1)
		} else if showCheck {
			fmt.Println(debug.PrettyTree(scope))
		}

		if showIR || showBC || showOut {
			ir := codegen.Transform(ast, libs.IO, libs.Conv)

			if showIR {
				fmt.Println(ir.String())
			}

			if showBC || showOut {
				mod := codegen.Generate(ir)

				if showBC {
					fmt.Println(debug.PrettyTree(mod.Root))
				}

				if showOut {
					vm.Eval(mod)
				}
			}
		}
	}
}
