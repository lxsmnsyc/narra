# Wave 2, second half

Nodes: closures (PR #5) and patterns (PR #6), both merged. The user asked to start only these two. strings-std, io-std, maps, poly-recursion and opaque-distinct were held back.

## Changes to the plan

- The user changed the tuple syntax during patterns. Tuple literals, patterns, destructuring and types are all written `#[...]` (D53). `Tuple(...)`, `Tuple<A, B>` and `[a, b]` as a tuple are gone.
- The user decided to keep `Array<T>` and reject `T[]` (D23). `T[]` gets a TypeScript-habit fix in json-diagnostics.
- New decisions D51 and D52 cover escape rules that are stricter than SPEC.
- New node `hashed-cnames` (must, M). C names become `<kind>_<readable>_<FNV-1a 32>`, and it replaces node `prelude-cnames`, which is dropped. The user chose FNV-1a over xxhash32 because it is in the Go standard library and easy to port. Build cost is negligible.
- opaque-distinct brief: it must block patterns, destructuring and spread on opaque structs. The inline brief now warns about stack closure environments.

## Learnings promoted

`compiler/internal/narra/CLAUDE.md` gained these rules:
- `c.funcOf`;
- `closureFunc` state;
- pattern tests as `&&` chains;
- pattern bindings retain and release.

## Notes

- The patterns brief still describes the old tuple syntax. It is done, so it was left as history. outcome.md describes what was built.
- Three name-collision bugs came from hand-built C names: `__` in names, prelude names, and tuple arity. hashed-cnames removes that whole class.
- The fast check now takes 25 to 130 s here, depending on load.
