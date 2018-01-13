package lang

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func Link(filepath string, ast *RootNode) (mod Module, errs []error) {
	// Determine an order that all dependencies can be loaded such that no
	// dependent is loaded before any of its dependencies.
	order, errs := resolve(filepath, ast)
	if len(errs) > 0 {
		return nil, errs
	}

	// Sanity check: the last module in the order must be the same program passed
	// via the `ast` parameter.
	if order[len(order)-1].module.ast != ast {
		panic("root module not loaded last")
	}

	// Link each module in the dependency graph to its dependencies.
	for _, n := range order {
		for _, child := range n.children {
			n.module.imports = append(n.module.imports, child.module)
		}
	}

	return order[len(order)-1].module, nil
}

type node struct {
	flag     int
	children []*node
	parents  []*node
	module   *VirtualModule
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

// link does some stuff
func link(path string, ast *RootNode, builtins ...Module) (Module, []error) {
	order, errs := resolve(path, ast)
	if len(errs) > 0 {
		return nil, errs
	}

	// Link ordered modules.
	for _, n := range order {
		for _, child := range n.children {
			n.module.imports = append(n.module.imports, child.module)
		}
	}

	return order[len(order)-1].module, nil
}

// resolve determines if a module has any dependency cycles
func resolve(path string, ast *RootNode) ([]*node, []error) {
	n := makeNode(path, ast)
	g, errs := buildGraph(n, getDependencyPaths, loadDependency)
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
			out += filepath.Base(n.module.name)
		} else {
			out += fmt.Sprintf("%s <- ", filepath.Base(n.module.name))
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

func buildGraph(n *node, branch func(*node) []string, load func(string) (*node, []error)) (g *graph, errs []error) {
	g = &graph{n, map[string]*node{}}
	done := map[string]*node{}
	todo := []*node{n}
	g.nodes[n.module.name] = n
	done[n.module.name] = n

	for len(todo) > 0 {
		n, todo = todo[0], todo[1:]
		for _, path := range branch(n) {
			if dep := done[path]; dep != nil {
				addParent(dep, n)
				addChild(n, dep)
			} else if dep := g.nodes[path]; dep != nil {
				addParent(dep, n)
				addChild(n, dep)
				addTodo(&todo, dep)
			} else if dep, errs = load(path); len(errs) == 0 {
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

func getDependencyPaths(n *node) (paths []string) {
	for _, stmt := range n.module.ast.Stmts {
		if stmt, ok := stmt.(*UseStmt); ok {
			dir := filepath.Dir(n.module.name)
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
	done[n.module.name] = n
}

func loadDependency(path string) (n *node, errs []error) {
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
			name:    path,
			ast:     ast,
			scope:   nil,
			imports: nil,
		},
	}
}
