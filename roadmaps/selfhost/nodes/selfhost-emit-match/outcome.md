# Outcome: selfhost-emit-match

## Built

- `selfhost/src/emitmatch.nra`: `emitIsCondition`, `emitIfIs`, `emitMatchStatement`, `emitMatchValue`, `emitDestructure`, `emitSpreadLiteral`, `emitTupleLiteral`, with the contract signatures. Private helpers: pattern conditions (tag tests chained with `&&` before payload reads), pattern binding (copy and retain, local released with its scope), match arms with guards.
- `reportMatchWork` and its call in `emit.nra` are deleted, as is the `selfhost-emit-match` arm of `refNeeds` in `emitstate.nra`.
- The skip file `compiler/testdata/emit/skips/selfhost-emit-match.txt` is deleted. It listed 37 inputs; none reports this node now: closures_nested, const, convert, ctor_qualified, enums, generic_self_instance_box, generic_self_instance_enum, generic_self_instance_two, generics, import_rename, io_files, maps, options, patterns_closures, patterns_generic, patterns_nested, patterns_structs, patterns_tuples, recursive, std_option, std_result, std_test, strings2, trait_method_generic_return, traits, traits_ctor_shadow, traits_derive, Tuple, patterns_tuple_shapes, literal_type_args (all `.nra` in testdata), and the packages app, cnames_modules, const, ctor_qualified, opaque_range, patterns, traits.

## Deviations

- `emitstate.nra` (shared): `ScopeFrame` gains `held: Array<TempVar> = []`, released by `releaseLocals` after the locals. Needed for the subject of `if (x is P)`: Go registers a fake local, and the emitter cannot add a LocalInfo to the read-only Checker.
- `compiler/testdata/emit/skips/selfhost-emit-traits.txt` (traits owner's file) gains `packages/opaque_range`: its `==` was hidden behind the old match stub and now reports `equality (selfhost-emit-traits)`. TestSelfhostEmit needs reported == listed.
- `emit.nra`: one line (the `reportMatchWork` call) removed, as the contract asks.
- `emitmatch.nra` has a private `declaredLocal` copy, because importing emitexprs would be a cycle.
- The commit subject is shorter than the prescribed one (the plain-writing hook limits subjects to 72 characters).

## Stubs left

- None in this node. Whole-file byte equality still waits for std (51 inputs), closures, traits, maps and options.

## Gotchas

- `test` is a keyword, so it cannot name a local.
- Calls with `mutable emitter` cannot also read `emitter.checker` in the same call or template argument; read into a local first.
- `emitMatchStatement` relies on the caller's `beginStatement`/`endStatement`; `emitIfIs` and `emitDestructure` open their own frames, as in Go.
- TestSelfhostEmit compares only the first five parts, so function bodies are not byte-checked until stubs are gone. I diffed whole outputs against Go for enums, patterns_structs, patterns_nested, patterns_closures, options, generics and Tuple: each first difference is a std, traits or closures marker; text before it, including match, is-conditions, bindings, tuple and destructuring code, is identical.
- Literal pool differences in `patterns_tuples` and `patterns_tuple_shapes` (the literals section, not compared by the test) come from the traits stub, not from this node: the `==` and `!=` on tuples (`#[1, "x"] == #[1, "x"]`, `left() == #[#["a", "b"], "c"]`) report `equality (selfhost-emit-traits)` and skip their operands, so the string literals inside them never reach the pool and every later literal number shifts. Go does not deduplicate the pool. selfhost-emit-traits owns this; it disappears when the equality emitter lands.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass (narra/internal/narra ok in 146 s).

## Size

- Estimated M, actual M (about 330 Narra lines).
