package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plaid/codegen"
	"plaid/debug"
	"plaid/libs"
	"plaid/linker"
	"plaid/parser"
	"plaid/typechecker"
	"plaid/vm"
)

func main() {
	showAST := flag.Bool("ast", false, "output abstract syntax tree")
	showDeps := flag.Bool("deps", false, "output resolved dependency tree")
	showCheck := flag.Bool("check", false, "output type checker results")
	showIR := flag.Bool("ir", false, "output intermediate representation")
	showBC := flag.Bool("bytecode", false, "output bytecode")
	showOut := flag.Bool("out", false, "run program and print output")
	flag.Parse()

	for _, filename := range flag.Args() {
		processFile(filename, *showAST, *showDeps, *showCheck, *showIR, *showBC, *showOut)
	}
}

func processFile(filename string, showAST bool, showDeps bool, showCheck bool, showIR bool, showBC bool, showOut bool) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	src := string(buf)
	ast, err := parser.Parse(filename, src)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if showAST {
		fmt.Println(ast.String())
	}

	abs, _ := filepath.Abs(filename)
	mod, err := linker.Link(abs, ast)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if showDeps {
		fmt.Println(mod.Imports)
	}

	if showCheck || showIR || showBC || showOut {
		typechecker.Check(mod,
			linker.ConvertModule(libs.IO),
			linker.ConvertModule(libs.Conv))
		if mod.Scope.HasErrors() {
			for _, err := range mod.Scope.GetErrors() {
				fmt.Println(err)
			}
			os.Exit(1)
		} else if showCheck {
			fmt.Println(debug.PrettyTree(mod.Scope))
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
