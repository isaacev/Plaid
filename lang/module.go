package lang

import (
	"fmt"
	"plaid/lang/types"
	"strings"
)

type Module interface {
	fmt.Stringer
	Identifier() string
	Exports() types.Struct
	Dependencies() []Module
	IsNative() bool
	link(string, Module)
}

type ModuleNative struct {
	name    string
	library *Library
}

func (m *ModuleNative) String() string {
	var lines []string
	lines = append(lines, "---")
	lines = append(lines, "type: native")
	lines = append(lines, fmt.Sprintf("identifier: %s", m.name))

	exports := m.library.toType()
	if len(exports.Fields) > 0 {
		lines = append(lines, "exports:")
		for _, field := range exports.Fields {
			lines = append(lines, fmt.Sprintf("  - name: \"%s\"", field.Name))
			lines = append(lines, fmt.Sprintf("    type: %s", field.Type))
		}
	}

	return strings.Join(lines, "\n")
}

func (m *ModuleNative) Identifier() string {
	return m.name
}

func (m *ModuleNative) Exports() types.Struct {
	return m.library.toType()
}

func (m *ModuleNative) Dependencies() []Module {
	return nil
}

func (m *ModuleNative) IsNative() bool {
	return true
}

func (m *ModuleNative) link(string, Module) {}

type ModuleVirtual struct {
	path         string
	exports      types.Struct
	structure    *RootNode
	scope        *Scope
	dependencies map[string]Module
}

func (m *ModuleVirtual) String() string {
	var lines []string
	lines = append(lines, "---")
	lines = append(lines, "type: virtual")
	lines = append(lines, fmt.Sprintf("identifier: %s", m.path))

	if len(m.exports.Fields) > 0 {
		lines = append(lines, "exports:")
		for _, field := range m.exports.Fields {
			lines = append(lines, fmt.Sprintf("  - name: \"%s\"", field.Name))
			lines = append(lines, fmt.Sprintf("    type: %s", field.Type))
		}
	}

	if len(m.dependencies) > 0 {
		lines = append(lines, "dependencies:")
		for alias, dep := range m.dependencies {
			lines = append(lines, fmt.Sprintf("  - path: \"%s\"", dep.Identifier()))
			lines = append(lines, fmt.Sprintf("    alias: %s", alias))
		}
	}

	return strings.Join(lines, "\n")
}

func (m *ModuleVirtual) Identifier() string {
	return m.path
}

func (m *ModuleVirtual) Exports() types.Struct {
	return m.exports
}

func (m *ModuleVirtual) AddExport(name string, typ types.Type) {
	field := struct {
		Name string
		Type types.Type
	}{name, typ}
	m.exports = types.Struct{append(m.exports.Fields, field)}
}

func (m *ModuleVirtual) Dependencies() []Module {
	var deps []Module
	for _, dep := range m.dependencies {
		deps = append(deps, dep)
	}
	return deps
}

func (m *ModuleVirtual) IsNative() bool {
	return false
}

func (m *ModuleVirtual) link(name string, dep Module) {
	m.dependencies[name] = dep
}
