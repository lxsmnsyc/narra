# Change request: module-graph v2 to v3

- From node: hashed-cnames
- Status: open

## Problem

The v2 C name rule joins module paths, names and type arguments with `__`. It needed rules against `__` and leading or trailing `_` in names, a `0` marker for implementations, and a separate tuple prefix. Three collision bugs came from it.

## Proposed change

Replace the "C names" and "injective mapping" bullets of Layout and names with:

- A declaration's C name is `<kind>_<module hash>_<name>`, with kind `f`, `S` or `E`. The module hash is xxHash32, seed 0, as 8 lowercase hex digits, over the module path relative to the package (`app/text` hashes `text`; a single-file module hashes its whole path). Example: `app/text.show` is `f_<hash of "text">_show`. Prelude names have no hash segment (`E_Ordering`).
- Composite types (generic instances, tuples, function types, arrays) are interned and named `<kind>_<head>_<hash>`. The head is `<module hash>_<name>` for a generic instance (just `<name>` in the prelude), `Tuple<n>`, `Fn<n>`, or `Array`. The hash is xxHash32 over `head|arity|arg1|arg2...`, where each arg is the C name of the argument without a pointer star (a primitive is its Narra name). So `#[int, #[int, int]]` is `S_Tuple2_<h>` built from `int` and the name of `#[int, int]`. Function types list the parameters, then the result, as arguments.
- Generic function instances are named `<name>_<hash>` (FuncDecl.Name), with the hash over `<module hash>_<name>|arity|arg names`. Implementation functions are named `<Type>_<Trait>_<fn>_<hash>`, the hash over type, trait module, trait and function.
- Generated helper functions come first: `retain_<type>`, `release_<type>`, `eq_<type>`, `hash_<type>`, `cmp_<type>`, `show_<type>` (arrays keep `<array>_new`, `_push` and so on). Arrow functions are `k_<hash>_<owner name>_<n>`, with `K` for the environment and `D` for the drop function, and a thunk for a named function is `w_<hash>_<name>`.
- The compiler keeps a table from C name to what it names, and stops with an internal error if two different things get one name.
- The generated C has a comment table before the struct definitions, sorted by name: `/* S_Tuple2_<h> = #[int, #[int, int]] */`.
- Removed rules: no `__` or edge `_` restriction on top-level names, module path segments, package names, trait names or trait function names. Modules `a-b` and `a_b` no longer collide; two modules whose relative paths hash alike are an error.

## Affected nodes

modules (done), traits (done), closures, patterns, opaque-distinct, c-interop, every selfhost node that ports the naming.

## Workaround in place

None needed; the node implements this in `compiler/internal/narra/cnames.go`.

## Decision

