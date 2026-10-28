---
name: builtins
version: 1
status: frozen
defined_by: split-compiler
---

# Contract: builtins

## Purpose

How standard library functions implemented by the compiler are declared, checked and emitted, so each std module lives in its own file instead of edits to shared switches.

## Interface

```go
// The existing Builtin struct grows to:
type Builtin struct {
	Name    string   // "Str.slice", used in messages
	Params  []Type   // nil when Check handles the arguments
	Mutable []bool   // per parameter; nil means none are mutable
	Result  Type
	Effects []string // "alloc", "io", "unsafe"; set by every node that adds a builtin
	// Check is optional. Generic builtins such as Array.push and Array.filled
	// check their own arguments and return the result type.
	Check func(c *Checker, call *CallExpr, want Type) Type
	// Emit returns C code and whether it is an owned temporary.
	// where is the C string literal for the call's source position, used in panics.
	Emit func(e *Emitter, call *CallExpr, where string) (code string, owned bool)
}

type Module struct {
	Path    string              // "std/string"
	Funcs   map[string]*Builtin // exported names
	Runtime string              // C helpers, emitted once when the module is imported
}

func registerModule(m *Module)                    // import * as X from "std/..."
func registerPrelude(name string, b *Builtin)     // print, assert, panic
func registerNamespace(ns string, funcs map[string]*Builtin) // Array.*, Range.*
```

## Behavior

- Imports of `std/*` paths resolve through `registerModule`. Prelude names and the `Array` and `Range` namespaces need no import.
- `for ... of` over `Range` and `Array.values` stays special-cased in the statement checker and emitter. It is not routed through the registry.
- `Runtime` C is emitted after `runtime.h` and before user code, once per imported module. Prelude helpers stay in `runtime.h`.

## Changes

- v1: initial.
