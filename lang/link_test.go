package lang

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func buildTestingScripts(root string, files map[string]string) (rootPath string, tmpDir string) {
	dir, err := ioutil.TempDir("", "plaid-test")
	if err != nil {
		log.Fatal(err)
	}

	for name, contents := range files {
		tmpFile := filepath.Join(dir, name)
		if name == root {
			rootPath = tmpFile
		}
		ioutil.WriteFile(tmpFile, []byte(contents), 0666)
	}

	return rootPath, dir
}

func destroyTestingScripts(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestLink(t *testing.T) {
	read := func(filepath string) *RootNode {
		src, err := ioutil.ReadFile(filepath)
		if err != nil {
			t.Fatalf("failed to read testing file: '%s'", filepath)
		}

		ast, errs := Parse(filepath, string(src))
		if len(errs) > 0 {
			t.Fatalf("failed to parse testing file: '%s'", filepath)
		}

		return ast
	}

	good := func(start string, modBase string, importBases []string) {
		t.Helper()
		ast := read(start)
		mod, errs := Link(start, ast, nil)

		if len(errs) > 0 {
			t.Fatal(errs[0])
		} else {
			expectSame(t, filepath.Base(mod.Path()), modBase)
			if len(mod.Imports()) != len(importBases) {
				t.Errorf("Expected %d imports, got %d", len(importBases), len(mod.Imports()))
			} else {
				for i, imp := range importBases {
					expectSame(t, filepath.Base(mod.Imports()[i].Path()), imp)
				}
			}
		}
	}

	bad := func(start string, exp string) {
		t.Helper()
		ast := read(start)
		var err error = nil
		_, errs := Link(start, ast, nil)
		if len(errs) > 0 {
			err = errs[0]
		}

		expectAnError(t, err, exp)
	}

	start, tmpDir := buildTestingScripts("b.plaid", map[string]string{
		"a.plaid": `let x := 100;`,
		"b.plaid": `use "a.plaid";`,
	})
	defer destroyTestingScripts(tmpDir)
	good(start, "b.plaid", []string{"a.plaid"})

	start, tmpDir = buildTestingScripts("d.plaid", map[string]string{
		"c.plaid": `let x :=`,
		"d.plaid": `use "c.plaid";`,
	})
	defer destroyTestingScripts(tmpDir)
	bad(start, fmt.Sprintf("%s/c.plaid(1:8) unexpected symbol", tmpDir))
}

func TestResolve(t *testing.T) {
	start, tmpDir := buildTestingScripts("e.plaid", map[string]string{
		"a.plaid": "let x := 100;",
		"b.plaid": `use "a.plaid";`,
		"c.plaid": `use "a.plaid";`,
		"d.plaid": `use "b.plaid";`,
		"e.plaid": `use "a.plaid"; use "b.plaid"; use "c.plaid"; use "d.plaid";`,
	})
	defer destroyTestingScripts(tmpDir)

	read := func(filepath string) *RootNode {
		src, err := ioutil.ReadFile(filepath)
		if err != nil {
			t.Fatalf("failed to read testing file: '%s'", filepath)
		}

		ast, errs := Parse(filepath, string(src))
		if len(errs) > 0 {
			t.Fatalf("failed to parse testing file: '%s'", filepath)
		}

		return ast
	}

	good := func(start string, exp []string) {
		t.Helper()
		ast := read(start)
		if order, errs := resolve(start, ast, nil); len(errs) > 0 {
			t.Fatal(errs[0])
		} else if len(order) != len(exp) {
			t.Errorf("expected %d loaded dependencies, got %d", len(exp), len(order))
		} else {
			for i, expDep := range exp {
				got := filepath.Base(order[i].module.Path())
				if expDep != got {
					t.Errorf("expected moduel %d to be '%s', got '%s'", i, expDep, got)
				}
			}
		}
	}

	good(start, []string{
		"a.plaid",
		"b.plaid",
		"c.plaid",
		"d.plaid",
		"e.plaid",
	})

	bad := func(start string, exp string) {
		t.Helper()
		ast := read(start)
		var err error = nil
		_, errs := resolve(start, ast, nil)
		if len(errs) > 0 {
			err = errs[0]
		}

		expectAnError(t, err, exp)
	}

	start, tmpDir = buildTestingScripts("h.plaid", map[string]string{
		"f.plaid": `use "h.plaid";`,
		"g.plaid": `use "f.plaid";`,
		"h.plaid": `use "g.plaid";`,
		"i.plaid": `use "h.plaid";`,
	})
	defer destroyTestingScripts(tmpDir)
	bad(start, "Dependency cycle: h.plaid <- f.plaid <- g.plaid <- h.plaid")

	start, tmpDir = buildTestingScripts("k.plaid", map[string]string{
		"j.plaid": `let x :=`,
		"k.plaid": `use "j.plaid";`,
	})
	defer destroyTestingScripts(tmpDir)
	bad(start, fmt.Sprintf("%s/j.plaid(1:8) unexpected symbol", tmpDir))

	start, tmpDir = buildTestingScripts("l.plaid", map[string]string{
		"l.plaid": `use "m.plaid";`,
	})
	defer destroyTestingScripts(tmpDir)
	bad(start, fmt.Sprintf("open %s/m.plaid: no such file or directory", tmpDir))
}

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
		t.Helper()
		order := orderDependencies(g)
		got := routeToString(order)
		expectString(t, got, exp)
	}

	bad := func(g *graph, exp string) {
		t.Helper()
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
	n1 := &node{module: &VirtualModule{path: "n1"}}
	n2 := &node{module: &VirtualModule{path: "n2"}}
	n3 := &node{module: &VirtualModule{path: "n3"}}
	n4 := &node{module: &VirtualModule{path: "n4"}}
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
	n1 := &node{module: &VirtualModule{path: "foo"}}
	n2 := &node{module: &VirtualModule{path: "foo"}}
	l := []*node{n1}

	expectBool(t, containsNode(l, n1), true)
	expectBool(t, containsNode(l, n2), false)
}

func TestExtractCycle(t *testing.T) {
	good := func(exp string, route ...*node) {
		t.Helper()
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
	branch := func(n *node) ([]string, []string) {
		switch n.module.Path() {
		case "n0":
			return []string{"n1", "n2"}, []string{"n1", "n2"}
		case "n1":
			return []string{"n2", "n3"}, []string{"n2", "n3"}
		case "n3":
			return []string{"n0"}, []string{"n0"}
		case "n4":
			return []string{"n5"}, []string{"n5"}
		default:
			return []string{}, []string{}
		}
	}

	n0 := testNode("n0")
	n1 := testNode("n1")
	n2 := testNode("n2")
	n3 := testNode("n3")
	n4 := testNode("n4")

	load := func(_ string, path string, _ map[string]*Library) (*node, []error) {
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

	g, errs := buildGraph(n0, branch, load, nil)
	for _, err := range errs {
		expectNoError(t, err)
	}

	expectChildren(t, g, "n0", "n1", "n2")
	expectChildren(t, g, "n1", "n2", "n3")
	expectChildren(t, g, "n2")
	expectChildren(t, g, "n3", "n0")
	expectParents(t, g, "n0", "n3")
	expectParents(t, g, "n2", "n0", "n1")

	_, errs = buildGraph(n4, branch, load, nil)
	var err error = nil
	if len(errs) > 0 {
		err = errs[0]
	}
	expectAnError(t, err, "random error")
}

func TestGetDependencyPaths(t *testing.T) {
	n := &node{module: &VirtualModule{path: "a/b/c.plaid", ast: &RootNode{Stmts: []Stmt{
		&UseStmt{Path: &StringExpr{Val: "foo"}},
		&UseStmt{Path: &StringExpr{Val: "../bar"}},
		&ExprStmt{},
		&UseStmt{Path: &StringExpr{Val: "baz/quux"}},
	}}}}

	_, got := getDependencyPaths(n)
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

	expectBool(t, n.module.Path() == path, true)
	expectBool(t, n.module.(*VirtualModule).ast == ast, true)
}

func testNode(path string, children ...*node) *node {
	return &node{
		module:   &VirtualModule{path: path},
		children: children,
	}
}

func expectNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got '%s'", err)
	}
}

func expectChildren(t *testing.T, g *graph, path string, children ...string) {
	t.Helper()
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

		if n.children[i].module.Path() != p {
			t.Errorf("Expected child #%d of %s to be named '%s', was named '%s'", i, path, p, n.children[i].module.Path())
		}
	}
}

func expectParents(t *testing.T, g *graph, path string, parents ...string) {
	t.Helper()
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

		if n.parents[i].module.Path() != p {
			t.Errorf("Expected parent #%d of %s to be named '%s', was named '%s'", i, path, p, n.parents[i].module.Path())
		}
	}
}

func expectRoute(t *testing.T, got []*node, exp string) {
	t.Helper()
	if routeToString(got) != exp {
		t.Errorf("Expected '%s', got %s", routeToString(got), exp)
	}
}

func expectInt(t *testing.T, got int, exp int) {
	t.Helper()
	if exp != got {
		t.Errorf("Expected %d, got %d", exp, got)
	}
}
