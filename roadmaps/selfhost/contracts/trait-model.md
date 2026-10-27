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
| `int` | `lang_eq_int` | `lang_hash_int` | `lang_cmp_int` | `lstr_from_int` |
| `f64` | `lang_eq_f64` | none | `lang_cmp_f64` | `lstr_from_f64` |
| `bool`, `char` | `lang_eq_bool`, `lang_eq_char` | `lang_hash_bool`, `lang_hash_char` | `lang_cmp_bool`, `lang_cmp_char` | `lang_show_bool`, `lang_show_char` |
| `string` | `lstr_eq` | `lstr_hash` | `lstr_cmp` | `lang_show_string` |
| arrays, structs, enums, instances | `<mangle(t)>_eq` | `<mangle(t)>_hash` | `<mangle(t)>_cmp` | `<mangle(t)>_show` |

`mangle` is the existing function in `types.go`, so `Pair<int, string>` gives `S_Pair__int__string_eq`.

## Stub

maps can build before traits lands by calling the primitive names above for `int` and `string` keys. The `lang_eq_int`, `lang_hash_int` and `lstr_hash` helpers go in `runtime.h` with whichever node lands first.

## Changes

- v1: initial.
