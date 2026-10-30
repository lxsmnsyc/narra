# Change request: trait-model v1 to v2

- From node: traits
- Status: accepted

## Problem

1. The name functions `eqFunc`, `hashFunc`, `cmpFunc` and `showFunc` are package functions that return a name. For arrays, structs and enums the C function is generated, so the emitter must know which ones a program uses. A package function cannot record that, so a caller (maps) that only calls `eqFunc(t)` gets a name with no definition and the C fails to compile.
2. After modules (module-graph v2), the v1 C name `<mangle(t)>_eq` can equal the C name of a type: struct `Point_eq` in the same module gives `S_app__Point_eq`, the same as the eq function of `Point`. C typedef and function names share one namespace.
3. The contract does not say which module may implement a trait, or how traits cross modules.
4. The contract does not say what `Self` is beyond `*SelfType`, how `Ordered.compare` returns (`Ordering`), or where the standard traits live.

## Proposed change

Change the generated names in the table from `<mangle(t)>_eq`, `_hash`, `_cmp`, `_show` to `<mangle(t)>__eq`, `__hash`, `__cmp`, `__show`. Example: `Pair<int, string>` in module `app/a` gives `S_app_a__Pair__int__string__eq`. No output of `mangle` is `eq`, `hash`, `cmp` or `show`, and type names contain no `__`, so these never equal a type's or a generic instance's C name. Primitive names are unchanged. `E_Ordering__show` keeps no module prefix (prelude).

Functions of `implement` blocks are `FuncDecl`s with `Module` set to the type's module and `Name` from `implFuncName`: `<Type>__0<cModule(trait module)><Trait>__<fn>`. Through `funcCName`, that gives `f_app_geo__Square__0app_shape__Area__area`, or `f_app_geo__Square__0Show__show` for a prelude trait. The `0` segment cannot start an identifier or a `mangle` output, so it never equals a generic instance name.

Naming rule that keeps these names injective: trait names and trait function names must pass `cSafeSegment`, the module-graph v2 rule for path segments: no `__`, and no leading or trailing `_`. Type names already contain no `__`. The type name ends at the first `__0`, because the part after it always starts with `0`, never `_`. The remaining `__`-separated segments are then all non-empty and free of edge `_`, so the last one is the function, the one before it the trait, and any before that the trait's module. Without the rule, trait `T_` with function `f` and trait `T` with function `_f` both gave `X__0T___f` (review of PR #3). `testdata/errors/traits_name_underscore.nra` pins the error.

`mangle` gains a case for `FuncType`: `F<n>` then the parameter types and the result, joined by `__` like type arguments. `(int) => int` is `F1__int__int`. This keeps `Box<(int) => int>` and `Box<(bool) => int>` apart.

Modules: an `implement` block must be in the module that declares the type (SPEC section 4, D5). `export trait` exports a trait, and `import { Trait } from "..."` imports it. `Trait.Module` holds the declaring module ("" for the prelude).

Keep the rest of v1, and add:

```go
// Emitter methods. Each returns the same name as the package function and
// makes sure the definition is emitted. Use these inside the emitter.
func (e *Emitter) useEq(t Type) string
func (e *Emitter) useHash(t Type) string
func (e *Emitter) useCmp(t Type) string
func (e *Emitter) useShow(t Type) string

// SelfType is an alias of TypeParamType, one per trait (Trait.Self).
type SelfType = TypeParamType

// Extra fields, additive:
type Trait struct { ...; Self *SelfType; Decl *TraitDecl; Std bool; Module string }
type TraitFunc struct { ...; Decl *FuncDecl }

// Look up a standard trait for implements().
func (c *Checker) stdTrait(name string) *Trait
```

- The standard traits and `enum Ordering { Less, Equal, Greater }` are declared in the Narra prelude in `check.go`. Signatures: `Equatable.equals(a: Self, b: Self): bool`, `Hashable.hash(value: Self): int`, `Ordered.compare(a: Self, b: Self): Ordering`, `Show.show(value: Self): string`.
- `Program.Impls` maps a type and `"Trait.function"` to the user's implementation `FuncDecl`. The generated `<mangle(t)>_eq` and friends call it when present.
- Runtime helpers added to `runtime.h` besides the v1 names: `narra_hash_mix`, `narra_ordering` (cmp result to `Ordering` tag), `lstr_append_c`, `lstr_append_owned`.

## Affected nodes

maps (uses `Hashable`, `eqFunc`, `hashFunc`). Any later node that emits `==` or `Show` through the trait functions.

## Workaround in place

Implemented as proposed in `check_traits.go` and `emit_traits.go`, rebased on modules. The v1 package functions exist and return the new names only.

## Decision

Accepted 2026-10-28 with PR #3. The contract is now v2, frozen.

---

# Change request: trait-model v2 to v3

- From node: hashed-cnames
- Status: accepted

## Problem

The "C names" section names trait functions `<mangle(t)>__eq` and implementations `<Type>__0<cModule(trait module)><Trait>__<fn>`, and depends on the `__` segment rule of module-graph v2. `mangle` and `cModule` no longer exist.

## Proposed change

- Trait functions for arrays, structs, enums and instances are `eq_<type>`, `hash_<type>`, `cmp_<type>`, `show_<type>`, where `<type>` is the type's C name without a pointer star (`eq_S_<hash>_Pair_<hash>`). The operation comes first so that no declaration name can equal it. Primitives keep their runtime names.
- Implementation functions are `FuncDecl`s in the type's module named `<Type>_<Trait>_<fn>_<hash>`, where the hash is xxHash32 over `Type|trait module (relative)|Trait|fn`. `funcCName` adds `f_<module hash>_`.
- Trait names and trait function names have no segment rule any more (see module-graph v3).
- The `FuncType` mangle `F<n>__...` becomes the interned name `F_Fn<n>_<hash>`.

## Affected nodes

traits (done), closures, opaque-distinct, selfhost nodes.

## Workaround in place

Implemented in `compiler/internal/narra/cnames.go` (`implName`, `opFuncName`).

## Decision

Accepted 2026-10-30 with PR #8. The contract is now v3, frozen.
