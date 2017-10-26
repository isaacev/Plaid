package typechecker

import "testing"

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
