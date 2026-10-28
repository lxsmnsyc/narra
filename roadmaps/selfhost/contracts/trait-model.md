---
name: trait-model
version: 2
status: frozen
defined_by: traits
---

# Contract: trait-model

## Purpose

How traits and their implementations are represented, so maps (`Hashable`), equality (`Equatable`) and later nodes can ask whether a type implements a trait and which C function to call.

## Interface

```go
type SelfType = TypeParamType // one per trait (Trait.Self)

type Trait struct {
	Name   string
	Funcs  []*TraitFunc
	Self   *SelfType
	Decl   *TraitDecl
	Std    bool
	Module string // declaring module, "" for the prelude
}

type TraitFunc struct {
	Name   string
	Params []Type // Self appears as *SelfType
	Result Type
	Decl   *FuncDecl
}

func (c *Checker) stdTrait(name string) *Trait

func (c *Checker) implements(t Type, trait *Trait) bool
// implFunc returns the user implementation, or nil when the compiler provides it
// (automatic Equatable and Hashable, or derives).
func (c *Checker) implFunc(t Type, trait *Trait, name string) *FuncDecl

// C function names for the standard traits. Arguments are borrowed.
func eqFunc(t Type) string   // bool f(T a, T b)
func hashFunc(t Type) string // uint64_t f(T v)
func cmpFunc(t Type) string  // int f(T a, T b), negative, zero or positive
func showFunc(t Type) string // lstr* f(T v), returns an owned string

// Inside the emitter, use these. Each returns the same name and makes sure
// the generated definition is emitted.
func (e *Emitter) useEq(t Type) string
func (e *Emitter) useHash(t Type) string
func (e *Emitter) useCmp(t Type) string
func (e *Emitter) useShow(t Type) string
```

- The standard traits and `enum Ordering { Less, Equal, Greater }` are in the Narra prelude: `Equatable.equals(a: Self, b: Self): bool`, `Hashable.hash(value: Self): int`, `Ordered.compare(a: Self, b: Self): Ordering`, `Show.show(value: Self): string`.
- `Program.Impls` maps a type and `"Trait.function"` to the user's implementation. The generated functions call it when present.
- An `implement` block must be in the module that declares the type. `export trait` and `import { Trait } from "..."` work; qualified trait names do not yet.

C names:

| Type | eq | hash | cmp | show |
| --- | --- | --- | --- | --- |
| `int` | `narra_eq_int` | `narra_hash_int` | `narra_cmp_int` | `lstr_from_int` |
| `f64` | `narra_eq_f64` | none | `narra_cmp_f64` | `lstr_from_f64` |
| `bool`, `char` | `narra_eq_bool`, `narra_eq_char` | `narra_hash_bool`, `narra_hash_char` | `narra_cmp_bool`, `narra_cmp_char` | `narra_show_bool`, `narra_show_char` |
| `string` | `lstr_eq` | `lstr_hash` | `lstr_cmp` | `narra_show_string` |
| arrays, structs, enums, instances | `<mangle(t)>__eq` | `<mangle(t)>__hash` | `<mangle(t)>__cmp` | `<mangle(t)>__show` |

`mangle` is the existing function in `types.go`, so `Pair<int, string>` in module `app/a` gives `S_app_a__Pair__int__string__eq`. Prelude types have no module prefix (`E_Ordering__show`). A function type mangles as `F<n>__<params>__<result>`, so `(int) => int` is `F1__int__int`.

Implementation functions are `FuncDecl`s in the type's module named `<Type>__0<cModule(trait module)><Trait>__<fn>`, so `funcCName` gives `f_app_geo__Square__0app_shape__Area__area` or `f_app_geo__Square__0Show__show`. Trait names and trait function names follow the module-graph v2 segment rule (no `__`, no leading or trailing `_`), which keeps every name injective.

Runtime helpers in `runtime.h` besides the names above: `narra_hash_mix`, `narra_ordering`, `narra_quote`, `lstr_append_c`, `lstr_append_owned`.

## Stub

maps can build before traits lands by calling the primitive names above for `int` and `string` keys. The `narra_eq_int`, `narra_hash_int` and `lstr_hash` helpers go in `runtime.h` with whichever node lands first.

## Changes

- v1: initial.
- v2 (frozen 2026-10-28, from the traits change request): `__eq`-style names, emitter `use*` methods, `SelfType`, extra `Trait` fields, `Program.Impls`, prelude traits and `Ordering`, implementation naming rule, `FuncType` mangle, per-module `implement` rule.
