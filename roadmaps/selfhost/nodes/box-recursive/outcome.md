# Outcome: box-recursive

## Built

All in `compiler/internal/narra/` unless noted.

- Enums with at least one variant that has fields keep those fields on the heap with a reference count. C layout: `struct E_x { int32_t tag; Q_x* u; }`; `Q_x` is `{ int64_t rc; union { struct {...} v_A; ... }; }`. Field access is `value.u->v_A.f_name`. A variant without fields has `u == NULL` (no allocation); an enum with no fielded variant is just `{ int32_t tag; }` as before. `isBoxedEnum(t)` (types.go) is the test; `needsRC` is true for every boxed enum.
- `retain` is `if (v.u) v.u->rc++`. `release` drops the count; at zero it calls `narra_drop(q, drop_E)` (runtime.h). `Q` carries the variant tag so `drop_E` can release the live variant's fields and free `Q`. Inside a drop, releasing a child that reaches zero queues it on a thread-local worklist instead of recursing, so freeing a 1,000,000-long list or a very deep tree uses heap, not C stack. Box releases the same way (`<box>_drop`). The worklist is freed when the outermost drop ends.
- `enumAlloc` (emit_box.go) is the way to build a fielded variant in the emitter (allocates `Q`, sets rc and tag, fills fields from already consumed values, returns the C value). Other nodes that build enum values by hand (`Result` in io-std) must call it instead of writing `.u.v_X...`; field reads are `.u->v_X.f_name`.
- Enums can be recursive: `orderTypes` (check_types.go) treats a boxed enum as complete before its fields, so `enum Expr { Add(left: Expr, right: Expr) }` and a struct that holds an enum that holds the struct both work. A struct that holds itself directly is still an error, now worded "put the recursive field in an enum, Box or array".
- `Box<T>`: `BoxType` (types.go), made through `c.boxOf` (interned, `Program.Boxes`). `Box.new(x)` (std_prelude.go, namespace `Box`) and `.value` (read only). C: pointer to `struct B_... { int64_t rc; T f_value; }` with `_new`, `_retain`, `_release`. Box supports `==`, hash, compare and Show when `T` does (`Box(1)` for show); `unify`, `substitute`, `containsTypeParam`, `containsOpen`, template edges, `implements`, `containsFunc` have Box cases.
- New file `emit_box.go`: heap struct definitions, Box functions, `enumAlloc` (constructor allocation).
- C names (cnames.go): `boxCName` (`B_Box_<hash>`) and `payloadCName` (`Q_<same suffix as the enum>`), both registered in the name table; boxes appear in the C name comment table.
- Changed: `emit_types.go` (enum typeDef, retain/release for enums and boxes), `emit.go` (typedefs, heap defs, box functions), `emit_match.go`, `emit_expr.go` and `emit_traits.go` (`.u.` became `.u->`), `emit_call.go` (constructor).
- Tests: `testdata/recursive.nra` (+ `.out`): expression tree with eval, render, Show, `==`, compare; a `Box` linked list (`Option<Box<Node>>`); `Box.new`/`.value`; a 1,000,000-long recursive `List` built, walked and dropped (clean under ASan and the leak check, debug and release). Errors: `testdata/errors/recursive_struct_self.nra`, `recursive_box_field.nra`, `recursive_user_box_new.nra`.

## Deviations

- Error tests are in `testdata/errors/` (the harness reads that folder) with the `recursive_` prefix.
- A user type named `Box` hides the builtin type and `Box.new` (`c.userBox()`): `Box<T>` resolves to the user type, and `Box.new` is no longer the builtin (the call then fails like any unknown module; tested). Existing tests declare `struct Box<T>`.
- `compiler/README.md` not edited (it did not mention enum layout).
- Files outside owned/shared paths: check.go, check_call.go, check_expr.go, check_traits.go, cnames.go, emit.go, emit_call.go, emit_expr.go, emit_traits.go, modules_check.go, std_prelude.go, types.go (small Box and layout cases).

## Stubs left

- `.value` on a Box cannot be assigned. No Box mutation or in-place reuse.
- None for freeing: iterative release is built (see Built).

## Gotchas

- Every `Some(x)`, `Ok(x)`, `Err(e)` now allocates. Perceus/reuse (node `perceus`) is the place to win that back.
- Freeing is iterative but user functions are not: a recursive Narra function 20000 deep overflowed the stack under ASan, so the sample keeps `eval` recursion at 2000.
- `while (x is P(b))` cannot bind names (only `if`); the sample uses `while (true) { if ... else return }`.
- Enum payloads are shared by copies of the value and never mutated after construction.

## Proposed learnings

- Reach enum fields through `v.u->variant.field` and allocate through `e.enumAlloc`, never build `{.tag=..., .u...}` by hand | the payload is a heap pointer | compiler/internal/narra | no
- Make heap or interned C names in cnames.go with their own kind letter (B, Q) | avoids clashes with declaration names | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (every program under ASan, UBSan, leak check, and release).
- Acceptance: expression tree, Box linked list, deep drop leak-clean: met by `testdata/recursive.nra`.

## Bench

`bench/run.sh -n 9 fib strings`, three alternating rounds, main (archive of origin/main) against the branch, same machine state, narra medians in seconds:

- fib: main 0.113, 0.114, 0.120; branch 0.114, 0.124, 0.124.
- strings: main 0.240, 0.259, 0.256; branch 0.241, 0.294, 0.274.
- The Go column moved by 5-10% between rounds too. Verdict: noise. The review's 0.124 to 0.177 s did not reproduce. Neither program uses enums; the only change to their C is the added runtime helpers, which they do not call.

## Size

- Estimated M, actual M.
