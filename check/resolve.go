package check

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"plaid/parser"
)

type node struct {
	flag     int
	path     string
	ast      *parser.Program
	children []*node
	parents  []*node
	module   *Module
}

type graph struct {
	root  *node
	nodes map[string]*node
}

func (g *graph) resetFlags() {
	for _, n := range g.nodes {
		n.flag = 0
	}
}

// Link does some stuff
func Link(path string, ast *parser.Program, builtins ...*Module) (*Module, error) {
	order, err := resolve(path, ast)
	if err != nil {
		return nil, err
	}

	// Link ordered modules.
	for _, n := range order {
		for _, child := range n.children {
			n.module.Imports = append(n.module.Imports, child.module)
		}
	}

	// Perform type-checking on the ordered modules.
	for _, n := range order {
		Check(n.module, builtins...)
	}

	return order[len(order)-1].module, nil
}

// resolve determines if a module has any dependency cycles
func resolve(path string, ast *parser.Program) ([]*node, error) {
	n := makeNode(path, ast)
	g, err := buildGraph(n, loadDependency)
	if err != nil {
		return nil, err
	}

	if cycle := findCycle(g.root, nil); cycle != nil {
		return nil, fmt.Errorf("Dependency cycle: %s", routeToString(cycle))
	}

	order := orderDependencies(g)
	return order, nil
}

func nodesToModules(route []*node) (mods []*Module) {
	for _, n := range route {
		mod := n.module
		for _, dep := range n.children {
			if containsNode(route, dep) == false {

			}
		}
		mods = append(mods, mod)
	}
	return mods
}

func orderDependencies(g *graph) (order []*node) {
	const FlagTemp = 1
	const FlagPerm = 2

	var visit func(*node)
	visit = func(n *node) {
		if n.flag == FlagPerm {
			return
		} else if n.flag == FlagTemp {
			panic("not a DAG")
		} else {
			n.flag = FlagTemp
			for _, m := range n.children {
				visit(m)
			}
			n.flag = FlagPerm
			order = append(order, n)
		}
	}

	g.resetFlags()
	visit(g.root)
	return order
}

func routeToString(route []*node) (out string) {
	if len(route) == 0 {
		return "empty route"
	}

	for i, n := range route {
		if i == len(route)-1 {
			out += filepath.Base(n.path)
		} else {
			out += fmt.Sprintf("%s <- ", filepath.Base(n.path))
		}
	}
	return out
}

func findCycle(n *node, route []*node) (cycle []*node) {
	for _, child := range n.children {
		newRoute := append(route, n)

		if containsNode(newRoute, child) {
			return extractCycle(append(newRoute, child))
		}

		if cycle := findCycle(child, newRoute); cycle != nil {
			return cycle
		}
	}

	return nil
}

func containsNode(list []*node, goal *node) bool {
	for _, n := range list {
		if goal == n {
			return true
		}
	}

	return false
}

func extractCycle(route []*node) (cycle []*node) {
	for i := len(route) - 1; i >= 0; i-- {
		if len(cycle) > 0 && route[i] == cycle[0] {
			return append(cycle, route[i])
		}

		cycle = append(cycle, route[i])
	}

	return nil
}

func buildGraph(n *node, load func(string) (*node, error)) (g *graph, err error) {
	g = &graph{n, map[string]*node{}}
	done := map[string]*node{}
	todo := []*node{n}
	g.nodes[n.path] = n
	done[n.path] = n

	for len(todo) > 0 {
		n, todo = todo[0], todo[1:]
		for _, path := range getDependencyPaths(n) {
			if dep := done[path]; dep != nil {
				addParent(dep, n)
				addChild(n, dep)
			} else if dep := g.nodes[path]; dep != nil {
				addParent(dep, n)
				addChild(n, dep)
				addTodo(&todo, dep)
			} else if dep, err = load(path); err == nil {
				addParent(dep, n)
				addChild(n, dep)
				addTodo(&todo, dep)
				g.nodes[path] = dep
			} else {
				return nil, err
			}
		}
		addDone(done, n)
	}

	return g, nil
}

func getDependencyPaths(n *node) (paths []string) {
	for _, stmt := range n.ast.Stmts {
		if stmt, ok := stmt.(*parser.UseStmt); ok {
			dir := filepath.Dir(n.path)
			path := filepath.Join(dir, stmt.Path.Val)
			paths = append(paths, path)
		}
	}

	return paths
}

func addParent(child *node, parent *node) {
	child.parents = append(child.parents, parent)
}

func addChild(parent *node, child *node) {
	parent.children = append(parent.children, child)
}

func addTodo(todo *[]*node, n *node) {
	*todo = append(*todo, n)
}

func addDone(done map[string]*node, n *node) {
	done[n.path] = n
}

func loadDependency(path string) (n *node, err error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ast, err := parser.Parse(path, string(buf))
	if err != nil {
		return nil, err
	}

	return makeNode(path, ast), nil
}

func makeNode(path string, ast *parser.Program) *node {
	return &node{
		path: path,
		ast:  ast,
		module: &Module{
			Name: path,
			AST:  ast,
		},
	}
}
