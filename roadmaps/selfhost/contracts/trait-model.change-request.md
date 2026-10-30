# Change request: trait-model v2 to v3

- From node: hashed-cnames
- Status: open

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

