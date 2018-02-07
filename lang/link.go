package lang

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

func Link(path string, ast *AST, stdlib map[string]Module) (Module, []error) {
	var g *graph
	var order []*node
	var errs []error

	// Build a dependency graph without performing any cycle detection.
	if g, errs = connect(path, ast, stdlib); len(errs) > 0 {
		return nil, errs
	}

	// From the graph determine a linear order that all dependencies can be
	// analyzed such that all dependencies are analyzed before they are needed by
	// any dependents.
	if order, errs = flatten(g); len(errs) > 0 {
		return nil, errs
	}

	// Link each dependent module to all of its dependencies.
	for _, dependent := range order {
		for relative, dependency := range dependent.children {
			alias, err := identifierToAlias(dependency.module.Identifier())
			if err != nil {
				return nil, append(errs, err)
			}
			dependent.module.link(alias, relative, dependency.module)
		}
	}

	return order[len(order)-1].module, nil
}

type graph struct {
	root  *node
	nodes map[string]*node
}

type node struct {
	flag     int
	native   bool
	children map[string]*node
	parents  []*node
	module   Module
}

func connect(path string, ast *AST, stdlib map[string]Module) (*graph, []error) {
	n := &node{
		children: make(map[string]*node),
		module: &ModuleVirtual{
			path:      path,
			structure: ast,
		},
	}

	loadDependency := func(path string) (*node, []error) {
		if mod, ok := stdlib[path]; ok {
			// Load dependency from the standard library.
			return &node{
				children: make(map[string]*node),
				module:   mod,
			}, nil
		} else {
			// Load dependency from the file system.
			var buf []byte
			var err error
			if buf, err = ioutil.ReadFile(path); err != nil {
				return nil, []error{err}
			}

			var ast *AST
			var errs []error
			if ast, errs = Parse(path, string(buf)); len(errs) > 0 {
				return nil, errs
			}

			return &node{
				children: make(map[string]*node),
				module: &ModuleVirtual{
					path:      path,
					structure: ast,
				},
			}, nil
		}
	}

	return buildGraphFromNode(n, loadDependency)
}

func (n *node) branch() (paths []string) {
	if mod, ok := n.module.(*ModuleVirtual); ok {
		for _, stmt := range mod.structure.Stmts {
			if stmt, ok := stmt.(*UseStmt); ok {
				paths = append(paths, stmt.Path.Val)
			}
		}
	}

	return paths
}

func isFilePath(path string) bool {
	return filepath.Ext(path) == ".plaid"
}

type loadDependencyFunc func(path string) (*node, []error)

func buildGraphFromNode(n *node, load loadDependencyFunc) (*graph, []error) {
	g := &graph{}                  // Graph to track relations.
	errs := []error{}              // Collection of errors detected.
	done := make(map[string]*node) // Cache of nodes already analyzed.
	todo := []*node{}              // Nodes yet to be analyzed.

	g.root = n
	g.nodes = make(map[string]*node)
	g.nodes[n.module.Identifier()] = n
	addDone(done, n)
	addTodo(&todo, n)

	for len(todo) > 0 {
		n, todo = todo[0], todo[1:]
		for _, relative := range n.branch() {
			// If dependency path seems like a relative path to another script, then
			// convert the path to an absolute path relative to the directory path of
			// the current node.
			path := relative
			if isFilePath(path) {
				path = filepath.Join(filepath.Dir(n.module.Identifier()), path)
			}

			if dep := done[path]; dep != nil {
				// The dependency has already been fully analyzed so all that's left is
				// to link the dependency and the dependant.
				addParent(dep, n)
				addChild(n, relative, dep)
			} else if dep := g.nodes[path]; dep != nil {
				// The dependency is in the `todo` queue awaiting analysis.
				addParent(dep, n)
				addChild(n, relative, dep)
			} else if dep, errs = load(path); len(errs) == 0 {
				// The dependency is novel and thus not already in the `todo` queue so
				// load and parse the script then add it to the `todo` queue for future
				// dependency analysis.
				addParent(dep, n)
				addChild(n, relative, dep)
				addTodo(&todo, dep)
				g.nodes[path] = dep
			} else {
				return nil, errs
			}
		}

		// Mark the current node as having been fully analyzed.
		addDone(done, n)
	}

	return g, nil
}

func addParent(child *node, parent *node) {
	child.parents = append(child.parents, parent)
}

func addChild(parent *node, relative string, child *node) {
	parent.children[relative] = child
}

func addTodo(todo *[]*node, n *node) {
	*todo = append(*todo, n)
}

func addDone(done map[string]*node, n *node) {
	done[n.module.Identifier()] = n
}

func flatten(g *graph) ([]*node, []error) {
	if cycle := findCycle(g.root, nil); cycle != nil {
		err := fmt.Errorf("Dependency cycle: %s", cycle)
		return nil, []error{err}
	}

	return findOrder(g), nil
}

func findCycle(n *node, route []*node) (cycle []*node) {
	for _, child := range n.children {
		route = append(route, n)
		if containsNode(route, child) {
			return extractCycle(append(route, child))
		} else if cycle := findCycle(child, route); cycle != nil {
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

// Take a slice of nodes and--assuming the last node in the slice is the start
// of a cycle--work backward through the slice trying to find a duplicate
// reference to the node at the end of the given slice. In that case, a cycle
// exists from the last node to its duplicate reference.
func extractCycle(route []*node) (cycle []*node) {
	for i := len(route) - 1; i >= 0; i-- {
		if len(cycle) > 0 && route[i] == cycle[0] {
			return append(cycle, route[i])
		}
		cycle = append(cycle, route[i])
	}
	return nil
}

func findOrder(g *graph) []*node {
	var order []*node
	var visit func(*node)
	const FlagTemp = 1
	const FlagPerm = 2

	visit = func(n *node) {
		if n.flag == FlagPerm {
			return
		} else if n.flag == FlagTemp {
			panic("not an acyclic dependency graph")
		} else {
			n.flag = FlagTemp
			for _, m := range n.children {
				visit(m)
			}
			n.flag = FlagPerm
			order = append(order, n)
		}
	}

	// Reset all flags in the graph before trying to build order.
	for _, n := range g.nodes {
		n.flag = 0
	}
	visit(g.root)
	return order
}

func identifierToAlias(identifier string) (string, error) {
	ext := regexp.MustCompile(`([a-zA-Z]+).plaid$`)
	match := ext.FindStringSubmatch(identifier)
	if len(match) >= 2 {
		return match[1], nil
	}

	lib := regexp.MustCompile(`^[a-zA-Z]+$`)
	match = lib.FindStringSubmatch(identifier)
	if len(match) >= 1 {
		return match[0], nil
	}

	return "", fmt.Errorf("could not determine alias for '%s'", identifier)
}
