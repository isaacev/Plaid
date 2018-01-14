package printing

import (
	"fmt"
	"strings"
)

const (
	pipeVert  = "│"
	pipeHoriz = "─"
	cornerNW  = "╭"
	cornerNE  = "╮"
	cornerSW  = "╰"
	cornerSE  = "╯"
	teeN      = "┴"
	teeS      = "┬"
	teeW      = "┤"
	teeE      = "├"
)

// StringerTree describes the node of a tree that can have 0 or more children
// each of whom defines a `String() string` function and may have its own
// children
type StringerTree interface {
	fmt.Stringer
	StringerChildren() []StringerTree
}

// PrettyTree takes a StringTree and returns a string of the form:
// ╭
// ┤ (leaf 1).String()
// │ ╭
// ├─┤ (leaf 1.1).String()
// │ │ ╭
// │ ╰─┤ (leaf 1.1.1).String()
// │   ╰
// │ ╭
// ╰─┤ (leaf 1.2).String()
//   ╰
func PrettyTree(root StringerTree) (out string) {
	children := root.StringerChildren()
	total := len(children)
	out += brace(root.String(), total == 0)
	for i, child := range children {
		if i < total-1 {
			out += "\n" + indentAndAttach(PrettyTree(child), teeE, pipeVert)
		} else {
			out += "\n" + indentAndAttach(PrettyTree(child), cornerSW, " ")
		}
	}
	return out
}

func brace(source string, close bool) string {
	var lines []string
	lines = append(lines, cornerNW+pipeHoriz)
	for i, line := range split(source) {
		if i == 0 {
			lines = append(lines, teeW+" "+line)
		} else {
			lines = append(lines, pipeVert+" "+line)
		}
	}
	if close {
		lines = append(lines, cornerSW+pipeHoriz)
	}
	return join(lines)
}

func split(source string) []string {
	return strings.Split(source, "\n")
}

func join(source []string) string {
	return strings.Join(source, "\n")
}

func indentAndAttach(source string, connector string, rest string) string {
	lines := split(source)
	for i, line := range lines {
		if i == 0 {
			lines[i] = pipeVert + " " + line
		} else if i == 1 {
			lines[i] = connector + pipeHoriz + line
		} else {
			lines[i] = rest + " " + line
		}
	}
	return join(lines)
}
