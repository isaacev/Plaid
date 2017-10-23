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

// Resolve determines if a module has any dependency cycles
func Resolve(path string, ast *parser.Program) {
	n := &node{path: path, ast: ast}
	g, err := buildGraph(n, loadDependency)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Analyze dependency graph
	for path, n := range g.nodes {
		fmt.Println(filepath.Base(path))
		if len(n.parents) == 0 {
			fmt.Println("  no parents")
		} else {
			for _, p := range n.parents {
				fmt.Printf("  <- %s\n", filepath.Base(p.path))
			}
		}
		if len(n.children) == 0 {
			fmt.Println("  no children ")
		} else {
			for _, c := range n.children {
				fmt.Printf("  -> %s\n", filepath.Base(c.path))
			}
		}
	}

	// Analyze any dependency cycles
	if cycle := findCycle(g.root, nil); cycle == nil {
		fmt.Println("order:")
		printRoute(orderDependencies(g))
	} else {
		fmt.Println("cycle:")
		printRoute(cycle)
	}
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

func printRoute(route []*node) {
	if len(route) == 0 {
		fmt.Println("empty route")
	} else {
		for i, n := range route {
			if i == len(route)-1 {
				fmt.Println(filepath.Base(n.path))
			} else {
				fmt.Printf("%s <- ", filepath.Base(n.path))
			}
		}
	}
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

	return &node{path: path, ast: ast}, nil
}
