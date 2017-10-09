package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"plaid/check"
	"plaid/codegen"
	"plaid/parser"
	"plaid/vm"
)

func main() {
	showAST := flag.Bool("ast", false, "output abstract syntax tree")
	showIR := flag.Bool("ir", false, "output intermediate representation")
	showBC := flag.Bool("bytecode", false, "output bytecode")
	showOut := flag.Bool("out", false, "run program and print output")
	flag.Parse()

	for _, filename := range flag.Args() {
		processFile(filename, *showAST, *showIR, *showBC, *showOut)
	}
}

func processFile(filename string, showAST bool, showIR bool, showBC bool, showOut bool) {
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

	scope := check.Check(ast, nil)
	if len(scope.Errors()) > 0 {
		for i, err := range scope.Errors() {
			fmt.Printf("%4d %s\n", i, err)
		}
		os.Exit(1)
	}

	ir := codegen.Transform(ast)

	if showIR {
		fmt.Println(ir.String())
	}

	mod := codegen.Generate(ir)

	if showBC {
		fmt.Println(mod.Main.String())
	}

	if showOut {
		vm.Run(mod.Main)
	}
}
