package lang

import (
	"fmt"
	"testing"
)

func TestGraphResetFlags(t *testing.T) {
	n1 := &node{flag: 5}
	n2 := &node{}

	g := &graph{n1, map[string]*node{}}
	g.nodes["n1"] = n1
	g.nodes["n2"] = n1
	g.resetFlags()

	expectInt(t, n1.flag, 0)
	expectInt(t, n2.flag, 0)
}

func TestOrderDependencies(t *testing.T) {
	good := func(g *graph, exp string) {
		order := orderDependencies(g)
		got := routeToString(order)
		expectString(t, got, exp)
	}

	bad := func(g *graph, exp string) {
		defer func() {
			if got := recover(); got == nil {
				t.Errorf("Expected failure when ordering cycle")
			} else if got != exp {
				t.Errorf("Expected panic '%s', got '%s'", exp, got)
			}
		}()
		orderDependencies(g)
	}

	n6 := testNode("n6")
	n5 := testNode("n5", n6)
	n4 := testNode("n4", n5, n6)
	n3 := testNode("n3", n4)
	n2 := testNode("n2", n3)
	n1 := testNode("n1", n2, n3)
	g := &graph{
		root: n1,
		nodes: map[string]*node{
			"n1": n1,
			"n2": n2,
			"n3": n3,
			"n4": n4,
			"n5": n5,
			"n6": n6,
		},
	}

	good(g, "n6 <- n5 <- n4 <- n3 <- n2 <- n1")

	n6.children = append(n6.children, n1)
	bad(g, "not a DAG")
}

func TestRouteToString(t *testing.T) {
	n1 := &node{module: &VirtualModule{name: "n1"}}
	n2 := &node{module: &VirtualModule{name: "n2"}}
	n3 := &node{module: &VirtualModule{name: "n3"}}
	n4 := &node{module: &VirtualModule{name: "n4"}}
	r1 := []*node{n1, n2, n3, n4}
	r2 := []*node{n1}
	r3 := []*node{}

	expectString(t, routeToString(r1), "n1 <- n2 <- n3 <- n4")
	expectString(t, routeToString(r2), "n1")
	expectString(t, routeToString(r3), "empty route")
}

func TestFindCycle(t *testing.T) {
	expectNoCycle := func(n *node) {
		got := findCycle(n, nil)
		if got != nil {
			t.Errorf("Expected to detect no cycle, got %s", routeToString(got))
		}
	}

	expectCycle := func(n *node, exp string) {
		got := findCycle(n, nil)
		if got == nil {
			t.Errorf("Expected to detect cycle, got nothing")
		} else {
			expectString(t, routeToString(got), exp)
		}
	}

	n5 := testNode("n5")
	n4 := testNode("n4", n5)
	n3 := testNode("n3", n4)
	n2 := testNode("n2", n3)
	n1 := testNode("n1", n2)
	expectNoCycle(n1)

	n5 = testNode("n5")
	n4 = testNode("n4", n5)
	n3 = testNode("n3", n4)
	n2 = testNode("n2", n3)
	n1 = testNode("n1", n2)
	n4.children = append(n4.children, n2)
	expectCycle(n1, "n2 <- n4 <- n3 <- n2")
}

func TestContainsNode(t *testing.T) {
	n1 := &node{module: &VirtualModule{name: "foo"}}
	n2 := &node{module: &VirtualModule{name: "foo"}}
	l := []*node{n1}

	expectBool(t, containsNode(l, n1), true)
	expectBool(t, containsNode(l, n2), false)
}

func TestExtractCycle(t *testing.T) {
	good := func(exp string, route ...*node) {
		expectRoute(t, extractCycle(route), exp)
	}

	n1 := testNode("n1")
	n2 := testNode("n2")
	n3 := testNode("n3")
	n4 := testNode("n4")
	n5 := testNode("n5")

	good("empty route")
	good("empty route", n1)
	good("n1 <- n1", n1, n1)
	good("n1 <- n3 <- n2 <- n1", n1, n2, n3, n1)
	good("n1 <- n3 <- n2 <- n1", n4, n1, n2, n3, n1)
	good("n1 <- n2 <- n1", n4, n5, n1, n2, n1)
	good("empty route", n4, n5, n1, n2, n1, n3)
}

func TestBuildGraph(t *testing.T) {
	branch := func(n *node) []string {
		switch n.module.name {
		case "n0":
			return []string{"n1", "n2"}
		case "n1":
			return []string{"n2", "n3"}
		case "n3":
			return []string{"n0"}
		case "n4":
			return []string{"n5"}
		default:
			return []string{}
		}
	}

	n0 := testNode("n0")
	n1 := testNode("n1")
	n2 := testNode("n2")
	n3 := testNode("n3")
	n4 := testNode("n4")

	load := func(path string) (*node, []error) {
		switch path {
		case "n0":
			return n0, nil
		case "n1":
			return n1, nil
		case "n2":
			return n2, nil
		case "n3":
			return n3, nil
		case "n4":
			return n4, nil
		default:
			return nil, []error{fmt.Errorf("random error")}
		}
	}

	g, errs := buildGraph(n0, branch, load)
	for _, err := range errs {
		expectNoError(t, err)
	}

	expectChildren(t, g, "n0", "n1", "n2")
	expectChildren(t, g, "n1", "n2", "n3")
	expectChildren(t, g, "n2")
	expectChildren(t, g, "n3", "n0")
	expectParents(t, g, "n0", "n3")
	expectParents(t, g, "n2", "n0", "n1")

	_, errs = buildGraph(n4, branch, load)
	var err error = nil
	if len(errs) > 0 {
		err = errs[0]
	}
	expectAnError(t, err, "random error")
}

func TestGetDependencyPaths(t *testing.T) {
	n := &node{module: &VirtualModule{name: "a/b/c.plaid", ast: &RootNode{Stmts: []Stmt{
		&UseStmt{Path: &StringExpr{Val: "foo"}},
		&UseStmt{Path: &StringExpr{Val: "../bar"}},
		&ExprStmt{},
		&UseStmt{Path: &StringExpr{Val: "baz/quux"}},
	}}}}

	got := getDependencyPaths(n)
	expectString(t, got[0], "a/b/foo")
	expectString(t, got[1], "a/bar")
	expectString(t, got[2], "a/b/baz/quux")
}

func TestAddParent(t *testing.T) {
	p := &node{}
	c := &node{}
	addParent(c, p)
	if c.parents[0] != p {
		t.Errorf("expected node to be stored in 'c.parents', found nothing")
	}
}

func TestAddChild(t *testing.T) {
	p := &node{}
	c := &node{}
	addChild(p, c)
	if p.children[0] != c {
		t.Errorf("expected node to be stored in 'p.children', found nothing")
	}
}

func TestAddTodo(t *testing.T) {
	todo := []*node{}
	n := &node{}
	addTodo(&todo, n)
	if todo[0] != n {
		t.Errorf("expected node to be stored in 'todo', found nothing")
	}
}

func TestAddDone(t *testing.T) {
	done := map[string]*node{}
	n := &node{module: &VirtualModule{}}
	addDone(done, n)
	if done[""] != n {
		t.Errorf("expected node to be stored in 'done', found nothing")
	}
}

func TestMakeNode(t *testing.T) {
	path := "foo bar"
	ast := &RootNode{}
	n := makeNode(path, ast)

	expectBool(t, n.module.name == path, true)
	expectBool(t, n.module.ast == ast, true)
	expectBool(t, n.module.String() == path, true)
	// expectBool(t, n.module.AST == ast, true)
}

func testNode(path string, children ...*node) *node {
	return &node{
		module:   &VirtualModule{name: path},
		children: children,
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}
}

func expectChildren(t *testing.T, g *graph, path string, children ...string) {
	n := g.nodes[path]
	if n == nil {
		t.Errorf("Expected '%s', found no matching node", path)
	}

	if len(n.children) != len(children) {
		t.Errorf("Expected %d children of %s, got %d", len(children), path, len(n.children))
	}

	for i, p := range children {
		if len(n.children) <= i {
			t.Errorf("Expected at least %d children of %s, found %d", i+1, path, len(n.children))
		}

		if n.children[i].module.name != p {
			t.Errorf("Expected child #%d of %s to be named '%s', was named '%s'", i, path, p, n.children[i].module.name)
		}
	}
}

func expectParents(t *testing.T, g *graph, path string, parents ...string) {
	n := g.nodes[path]
	if n == nil {
		t.Errorf("Expected '%s', found no matching node", path)
	}

	if len(n.parents) != len(parents) {
		t.Errorf("Expected %d parents of %s, got %d", len(parents), path, len(n.parents))
	}

	for i, p := range parents {
		if len(n.parents) <= i {
			t.Errorf("Expected at least %d parents of %s, found %d", i+1, path, len(n.parents))
		}

		if n.parents[i].module.name != p {
			t.Errorf("Expected parent #%d of %s to be named '%s', was named '%s'", i, path, p, n.parents[i].module.name)
		}
	}
}

func expectRoute(t *testing.T, got []*node, exp string) {
	if routeToString(got) != exp {
		t.Errorf("Expected '%s', got %s", routeToString(got), exp)
	}
}

func expectInt(t *testing.T, got int, exp int) {
	if exp != got {
		t.Errorf("Expected %d, got %d", exp, got)
	}
}
