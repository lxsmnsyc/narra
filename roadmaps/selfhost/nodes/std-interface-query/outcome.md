# Outcome: std-interface-query

## Built

- `narrac query [--json] interface std/<module>` answers for every public std module, builtin and source, from any folder (no package needed). Code: `selfhost/src/tools/querystd.nra`, called from `runQuery` in `query.nra` before the package loads.
- Builtin module: the functions of the registry (`stdRegistry`), sorted by name. Source module: the exported declarations of the embedded std text. A builtin module with a Narra part (`std/int`, `std/array`, `std/string`, `std/float`) also lists the `std/ext/<name>` declarations. `std/internal/*`, `std/ext/*` and unknown names give `unknown module`.
- Checked for every module of the registry plus every source module (std/async/*, arena, args, atomic, build, cell, channel, coroutine, cstring, fixedarray, hash, iterator, json, math, option, parallel, random, regex, shape, shared, template, terminal, thread, time, toml, unicode): all answer.
- Case: `testdata/query/std_modules` (std/path builtin, std/iterator and std/unicode source, std/int builtin plus Narra part, and the two error answers).

JSON shape (the shape of package modules, one extension):

```json
{"query": "interface", "module": "std/path", "declarations": [
  {"name": "join", "kind": "function", "signature": "export function join(string, string): string uses alloc", "location": null},
  {"name": "map", "kind": "function", "signature": "export function* map<T, U, effect E>(source: Generator<T> uses E, escaping transform: (T) => U uses E): Generator<U> uses E", "location": {"file": "std/src/iterator.nra", "range": {"start": [11, 18], "end": [11, 21]}}}
]}
```

- Effects are in `signature` after `uses`, as for package modules; there is no separate effects field.
- Extension: `"location"` is `null` for a builtin (no source file). Source declarations carry `std/src/<name>.nra` and the name's range.
- Builtin signatures have no parameter names, since the registry has none: `export function join(string, string): string uses alloc`; a `mutable` parameter prints `mutable T`.
- 135 builtins have a check function instead of a type list (`Array.push`, `Map.remove`, `Range.upTo`, the std/ptr and std/simd functions and the sized `from`/`tryFrom`/`wrap`). They print `(...)`, for example `export function push(...) uses alloc`. Their effects are kept. node std-reference must write these signatures by hand (SPEC.md section 3 has them).

## Deviations

- Signatures of builtins are type-only and 135 are `(...)`; the registry holds nothing better. Not a change to the registry, to keep the node inside its owned paths.

## Stubs left

- Names for builtin parameters and signatures of the check-computed builtins: a registry change, not in this node.

## Gotchas

- `interface std/...` ignores `--package`; a package cannot define a module named `std/...`.
- `narrac query --json interface` at a workspace root still merges members; std queries skip that.

## Verification

- `sh scripts/test.sh --suite query`: PASS, 17 cases.
- Full `sh scripts/test.sh`: `PASS, 10503 cases`.

## Size

- Estimated M, actual S.
