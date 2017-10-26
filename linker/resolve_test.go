package linker

import (
	"fmt"
	"plaid/parser"
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

func TestRouteToString(t *testing.T) {
	n1 := &node{path: "n1"}
	n2 := &node{path: "n2"}
	n3 := &node{path: "n3"}
	n4 := &node{path: "n4"}
	r1 := []*node{n1, n2, n3, n4}
	r2 := []*node{n1}
	r3 := []*node{}

	expectString(t, routeToString(r1), "n1 <- n2 <- n3 <- n4")
	expectString(t, routeToString(r2), "n1")
	expectString(t, routeToString(r3), "empty route")
}

func TestContainsNode(t *testing.T) {
	n1 := &node{path: "foo"}
	n2 := &node{path: "foo"}
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
		switch n.path {
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

	load := func(path string) (*node, error) {
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
			return nil, fmt.Errorf("random error")
		}
	}

	g, err := buildGraph(n0, branch, load)
	expectNoError(t, err)
	expectChildren(t, g, "n0", "n1", "n2")
	expectChildren(t, g, "n1", "n2", "n3")
	expectChildren(t, g, "n2")
	expectChildren(t, g, "n3", "n0")
	expectParents(t, g, "n0", "n3")
	expectParents(t, g, "n2", "n0", "n1")

	_, err = buildGraph(n4, branch, load)
	expectAnError(t, err, "random error")
}

func TestGetDependencyPaths(t *testing.T) {
	n := &node{path: "a/b/c.plaid", ast: &parser.Program{Stmts: []parser.Stmt{
		&parser.UseStmt{Path: &parser.StringExpr{Val: "foo"}},
		&parser.UseStmt{Path: &parser.StringExpr{Val: "../bar"}},
		&parser.ExprStmt{},
		&parser.UseStmt{Path: &parser.StringExpr{Val: "baz/quux"}},
	}}}

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
	n := &node{}
	addDone(done, n)
	if done[""] != n {
		t.Errorf("expected node to be stored in 'done', found nothing")
	}
}

func TestMakeNode(t *testing.T) {
	path := "foo bar"
	ast := &parser.Program{}
	n := makeNode(path, ast)

	expectBool(t, n.path == path, true)
	expectBool(t, n.ast == ast, true)
	expectBool(t, n.module.Name == path, true)
	expectBool(t, n.module.AST == ast, true)
}

func testNode(path string, children ...*node) *node {
	return &node{
		path:     path,
		children: children,
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}
}

func expectAnError(t *testing.T, err error, exp string) {
	if err == nil {
		t.Fatalf("Expected an error '%s'", exp)
	} else if err.Error() != exp {
		t.Fatalf("Expected error '%s', got '%s'", exp, err)
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

		if n.children[i].path != p {
			t.Errorf("Expected child #%d of %s to be named '%s', was named '%s'", i, path, p, n.children[i].path)
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

		if n.parents[i].path != p {
			t.Errorf("Expected parent #%d of %s to be named '%s', was named '%s'", i, path, p, n.parents[i].path)
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
