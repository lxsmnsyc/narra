package narra

import (
	"fmt"
	"sort"
	"strings"
)

// Builtin describes a standard function the compiler implements directly.
// See roadmaps/selfhost/contracts/builtins.md.
type Builtin struct {
	Name    string // "Str.slice", used in messages and as the key the emitter looks up
	Params  []Type // nil when Check handles the arguments
	Mutable []bool // per parameter; nil means none are mutable
	Result  Type
	Effects []string // "alloc", "io", "unsafe"; set by every node that adds a builtin
	// Check is optional. Generic builtins such as Array.push and Array.filled
	// check their own arguments and return the result type.
	Check func(c *Checker, call *CallExpr, want Type) Type
	// Emit returns C code and whether it is an owned temporary.
	// where is the C string literal for the call's source position, used in panics.
	Emit func(e *Emitter, call *CallExpr, where string) (code string, owned bool)
}

// Module is a standard library module such as "std/string".
type Module struct {
	Path    string              // "std/string"
	Funcs   map[string]*Builtin // exported names
	Runtime string              // C helpers, emitted once when the module is imported
}

var (
	stdModules        = map[string]*Module{}
	preludeFuncs      = map[string]*Builtin{} // print, assert, panic
	namespaceFuncs    = map[string]*Builtin{} // "Array.push", "Range.upTo"
	namespaces        = map[string]bool{}     // "Array", "Range"
	builtinsByName    = map[string]*Builtin{} // every builtin by Builtin.Name
	registeredModules []string
)

// registerModule makes `import * as X from "std/..."` resolve to m.
func registerModule(m *Module) {
	stdModules[m.Path] = m
	registeredModules = append(registeredModules, m.Path)
	for _, b := range m.Funcs {
		registerByName(b)
	}
}

// registerPrelude adds a function that needs no import, such as print.
func registerPrelude(name string, b *Builtin) {
	if b.Name == "" {
		b.Name = name
	}
	preludeFuncs[name] = b
	registerByName(b)
}

// registerNamespace adds functions called as ns.name, such as Array.push.
func registerNamespace(ns string, funcs map[string]*Builtin) {
	namespaces[ns] = true
	for n, b := range funcs {
		if b.Name == "" {
			b.Name = ns + "." + n
		}
		namespaceFuncs[ns+"."+n] = b
		registerByName(b)
	}
}

func registerByName(b *Builtin) {
	if _, dup := builtinsByName[b.Name]; dup {
		panic("duplicate builtin " + b.Name)
	}
	builtinsByName[b.Name] = b
}

// moduleList names the registered modules for error messages.
func moduleList() string {
	paths := append([]string(nil), registeredModules...)
	sort.Strings(paths)
	if len(paths) < 2 {
		return strings.Join(paths, "")
	}
	return fmt.Sprintf("%s and %s", strings.Join(paths[:len(paths)-1], ", "), paths[len(paths)-1])
}
