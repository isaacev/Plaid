package lang

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func Link(filepath string, ast *RootNode, stdlib map[string]*Library) (mod *VirtualModule, errs []error) {
	// Determine an order that all dependencies can be loaded such that no
	// dependent is loaded before any of its dependencies.
	order, errs := resolve(filepath, ast, stdlib)
	if len(errs) > 0 {
		return nil, errs
	}

	// Link each module in the dependency graph to its dependencies.
	for _, n := range order {
		for _, child := range n.children {
			n.module.(*VirtualModule).imports = append(n.module.Imports(), child.module)
		}
	}

	return order[len(order)-1].module.(*VirtualModule), nil
}

type node struct {
	flag     int
	native   bool
	children []*node
	parents  []*node
	module   Module
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

// resolve determines if a module has any dependency cycles
func resolve(path string, ast *RootNode, stdlib map[string]*Library) ([]*node, []error) {
	n := makeNode(path, ast)
	g, errs := buildGraph(n, getDependencyPaths, loadDependency, stdlib)
	if len(errs) > 0 {
		return nil, errs
	}

	if cycle := findCycle(g.root, nil); cycle != nil {
		return nil, []error{fmt.Errorf("Dependency cycle: %s", routeToString(cycle))}
	}

	order := orderDependencies(g)
	return order, nil
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
			out += filepath.Base(n.module.Path())
		} else {
			out += fmt.Sprintf("%s <- ", filepath.Base(n.module.Path()))
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

// extractCycle takes a slice of nodes and--assuming the last node in the slice
// is the start of a cycle--works backward through the slice trying to find a
// repetition of the node route[len(route)-1].
func extractCycle(route []*node) (cycle []*node) {
	for i := len(route) - 1; i >= 0; i-- {
		if len(cycle) > 0 && route[i] == cycle[0] {
			return append(cycle, route[i])
		}

		cycle = append(cycle, route[i])
	}

	return nil
}

func buildGraph(n *node, branch func(*node) ([]string, []string), load func(string, string, map[string]*Library) (*node, []error), stdlib map[string]*Library) (g *graph, errs []error) {
	g = &graph{n, map[string]*node{}}
	done := map[string]*node{}
	todo := []*node{n}
	g.nodes[n.module.Path()] = n
	done[n.module.Path()] = n

	for len(todo) > 0 {
		n, todo = todo[0], todo[1:]
		literal, relative := branch(n)
		for i, path := range relative {
			if dep := done[path]; dep != nil {
				addParent(dep, n)
				addChild(n, dep)
			} else if dep := g.nodes[path]; dep != nil {
				addParent(dep, n)
				addChild(n, dep)
				addTodo(&todo, dep)
			} else if dep, errs = load(literal[i], path, stdlib); len(errs) == 0 {
				addParent(dep, n)
				addChild(n, dep)
				addTodo(&todo, dep)
				g.nodes[path] = dep
			} else {
				return nil, errs
			}
		}
		addDone(done, n)
	}

	return g, nil
}

func getDependencyPaths(n *node) (raw []string, paths []string) {
	if n.native == false {
		for _, stmt := range n.module.(*VirtualModule).ast.Stmts {
			if stmt, ok := stmt.(*UseStmt); ok {
				raw = append(raw, stmt.Path.Val)
				dir := filepath.Dir(n.module.Path())
				path := filepath.Join(dir, stmt.Path.Val)
				paths = append(paths, path)
			}
		}
	}

	return raw, paths
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
	done[n.module.Path()] = n
}

func loadDependency(literal string, path string, stdlib map[string]*Library) (n *node, errs []error) {
	if lib, ok := stdlib[literal]; ok {
		return loadLibraryDependency(lib, literal)
	}

	return loadFileDependency(path)
}

func loadLibraryDependency(lib *Library, path string) (n *node, errs []error) {
	return &node{
		native: true,
		module: &NativeModule{
			path:    path,
			scope:   lib.toScope(),
			library: lib,
		},
	}, nil
}

func loadFileDependency(path string) (n *node, errs []error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, append(errs, err)
	}

	ast, errs := Parse(path, string(buf))
	if len(errs) > 0 {
		return nil, errs
	}

	return makeNode(path, ast), nil
}

func makeNode(path string, ast *RootNode) *node {
	return &node{
		module: &VirtualModule{
			path:    path,
			ast:     ast,
			scope:   nil,
			imports: nil,
		},
	}
}
