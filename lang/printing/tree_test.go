package printing

import (
	"testing"
)

type dummy struct {
	name  string
	leafs []dummy
}

func (d dummy) String() string {
	return d.name
}

func (d dummy) StringerChildren() (s []StringerTree) {
	for _, l := range d.leafs {
		s = append(s, l)
	}
	return s
}

func TestPrettyTree(t *testing.T) {
	test := func(root StringerTree, exp string) {
		got := PrettyTree(root)
		expectString(t, got, exp)
	}

	d := dummy{name: "d"}
	c := dummy{name: "c", leafs: []dummy{d}}
	b := dummy{name: "b"}
	a := dummy{name: "a", leafs: []dummy{b, c}}

	test(d, "╭─\n┤ d\n╰─")
	test(a, "╭─\n┤ a\n│ ╭─\n├─┤ b\n│ ╰─\n│ ╭─\n╰─┤ c\n  │ ╭─\n  ╰─┤ d\n    ╰─")
}

func TestBrace(t *testing.T) {
	test := func(source string, close bool, exp string) {
		got := brace(source, close)
		expectString(t, got, exp)
	}

	test("", true, "╭─\n┤ \n╰─")
	test("", false, "╭─\n┤ ")
	test("foo\nbar\nbaz", true, "╭─\n┤ foo\n│ bar\n│ baz\n╰─")
	test("foo\nbar\nbaz", false, "╭─\n┤ foo\n│ bar\n│ baz")
}

func TestSplit(t *testing.T) {
	expectStringSlice(t, split(""), []string{""})
	expectStringSlice(t, split("\n"), []string{"", ""})
	expectStringSlice(t, split("foo\nbar\nbaz"), []string{"foo", "bar", "baz"})
}

func TestJoin(t *testing.T) {
	expectString(t, join(nil), "")
	expectString(t, join([]string{""}), "")
	expectString(t, join([]string{"", ""}), "\n")
	expectString(t, join([]string{"foo", "bar", "baz"}), "foo\nbar\nbaz")
}

func TestIndentAndAttach(t *testing.T) {
	test := func(source string, connector string, rest string, exp string) {
		got := indentAndAttach(source, connector, rest)
		expectString(t, got, exp)
	}

	test("foo\nbar\nbaz", "╰", " ", "│ foo\n╰─bar\n  baz")
	test("foo\nbar\nbaz", "├", "│", "│ foo\n├─bar\n│ baz")
}

func expectStringSlice(t *testing.T, got []string, exp []string) {
	if len(got) != len(exp) {
		t.Errorf("Expected %d strings, got %d strings", len(exp), len(got))
	} else {
		for i := 0; i < len(exp); i++ {
			expectString(t, got[i], exp[i])
		}
	}
}

func expectString(t *testing.T, got string, exp string) {
	if exp != got {
		t.Errorf("Expected '%s', got '%s'", exp, got)
	}
}
