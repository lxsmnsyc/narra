---
name: trait-model
version: 1
status: draft
defined_by: traits
---

# Contract: trait-model

## Purpose

How traits and their implementations are represented, so maps (`Hashable`), equality (`Equatable`) and later nodes can ask whether a type implements a trait and which C function to call.

## Interface

```go
type Trait struct {
	Name  string
	Funcs []*TraitFunc
}

type TraitFunc struct {
	Name   string
	Params []Type // Self appears as *SelfType
	Result Type
}

func (c *Checker) implements(t Type, trait *Trait) bool
// implFunc returns the user implementation, or nil when the compiler provides it
// (automatic Equatable and Hashable, or derives).
func (c *Checker) implFunc(t Type, trait *Trait, name string) *FuncDecl

// C function names for the standard traits. Arguments are borrowed.
func eqFunc(t Type) string   // bool f(T a, T b)
func hashFunc(t Type) string // uint64_t f(T v)
func cmpFunc(t Type) string  // int f(T a, T b), negative, zero or positive
func showFunc(t Type) string // lstr* f(T v), returns an owned string
```

C names:

| Type | eq | hash | cmp | show |
| --- | --- | --- | --- | --- |
| `int` | `narra_eq_int` | `narra_hash_int` | `narra_cmp_int` | `lstr_from_int` |
| `f64` | `narra_eq_f64` | none | `narra_cmp_f64` | `lstr_from_f64` |
| `bool`, `char` | `narra_eq_bool`, `narra_eq_char` | `narra_hash_bool`, `narra_hash_char` | `narra_cmp_bool`, `narra_cmp_char` | `narra_show_bool`, `narra_show_char` |
| `string` | `lstr_eq` | `lstr_hash` | `lstr_cmp` | `narra_show_string` |
| arrays, structs, enums, instances | `<mangle(t)>_eq` | `<mangle(t)>_hash` | `<mangle(t)>_cmp` | `<mangle(t)>_show` |

`mangle` is the existing function in `types.go`, so `Pair<int, string>` gives `S_Pair__int__string_eq`.

## Stub

maps can build before traits lands by calling the primitive names above for `int` and `string` keys. The `narra_eq_int`, `narra_hash_int` and `lstr_hash` helpers go in `runtime.h` with whichever node lands first.

## Changes

- v1: initial.
