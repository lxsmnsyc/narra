# Outcome: box-recursive

## Built

All in `compiler/internal/narra/` unless noted.

- Enums with at least one variant that has fields keep those fields on the heap with a reference count. C layout: `struct E_x { int32_t tag; Q_x* u; }`; `Q_x` is `{ int64_t rc; union { struct {...} v_A; ... }; }`. Field access is `value.u->v_A.f_name`. A variant without fields has `u == NULL` (no allocation); an enum with no fielded variant is just `{ int32_t tag; }` as before. `isBoxedEnum(t)` (types.go) is the test; `needsRC` is true for every boxed enum.
- `retain` is `if (v.u) v.u->rc++`; `release` drops the count and, at zero, releases each field of the live variant and frees `Q`.
- Enums can be recursive: `orderTypes` (check_types.go) treats a boxed enum as complete before its fields, so `enum Expr { Add(left: Expr, right: Expr) }` and a struct that holds an enum that holds the struct both work. A struct that holds itself directly is still an error, now worded "put the recursive field in an enum, Box or array".
- `Box<T>`: `BoxType` (types.go), made through `c.boxOf` (interned, `Program.Boxes`). `Box.new(x)` (std_prelude.go, namespace `Box`) and `.value` (read only). C: pointer to `struct B_... { int64_t rc; T f_value; }` with `_new`, `_retain`, `_release`. Box supports `==`, hash, compare and Show when `T` does (`Box(1)` for show); `unify`, `substitute`, `containsTypeParam`, `containsOpen`, template edges, `implements`, `containsFunc` have Box cases.
- New file `emit_box.go`: heap struct definitions, Box functions, `enumAlloc` (constructor allocation).
- C names (cnames.go): `boxCName` (`B_Box_<hash>`) and `payloadCName` (`Q_<same suffix as the enum>`), both registered in the name table; boxes appear in the C name comment table.
- Changed: `emit_types.go` (enum typeDef, retain/release for enums and boxes), `emit.go` (typedefs, heap defs, box functions), `emit_match.go`, `emit_expr.go` and `emit_traits.go` (`.u.` became `.u->`), `emit_call.go` (constructor).
- Tests: `testdata/recursive.nra` (+ `.out`): expression tree with eval, render, Show, `==`, compare; a `Box` linked list (`Option<Box<Node>>`); `Box.new`/`.value`; a 20000-long recursive `List` built and dropped. Errors: `testdata/errors/recursive_struct_self.nra`, `recursive_box_field.nra`.

## Deviations

- Error tests are in `testdata/errors/` (the harness reads that folder) with the `recursive_` prefix.
- A user type named `Box` hides the builtin type (`Box<T>` resolves to the user type); existing tests declare `struct Box<T>`. `Box.new` is still the builtin namespace.
- `compiler/README.md` not edited (it did not mention enum layout).
- Files outside owned/shared paths: check.go, check_call.go, check_expr.go, check_traits.go, cnames.go, emit.go, emit_call.go, emit_expr.go, emit_traits.go, modules_check.go, std_prelude.go, types.go (small Box and layout cases).

## Stubs left

- `.value` on a Box cannot be assigned. No Box mutation or in-place reuse.
- Releasing is recursive: dropping a 1,000,000-deep list overflows the C stack. 20000 deep is tested. The spec's batched free would fix it; not built.

## Gotchas

- Every `Some(x)`, `Ok(x)`, `Err(e)` now allocates. Perceus/reuse (node `perceus`) is the place to win that back.
- Debug builds (ASan) use much more stack: a recursive Narra function 20000 deep overflowed under ASan, so samples keep recursion at 2000.
- `while (x is P(b))` cannot bind names (only `if`); the sample uses `while (true) { if ... else return }`.
- Enum payloads are shared by copies of the value and never mutated after construction.

## Proposed learnings

- Reach enum fields through `v.u->variant.field` and allocate through `e.enumAlloc`, never build `{.tag=..., .u...}` by hand | the payload is a heap pointer | compiler/internal/narra | no
- Make heap or interned C names in cnames.go with their own kind letter (B, Q) | avoids clashes with declaration names | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (every program under ASan, UBSan, leak check, and release).
- Acceptance: expression tree, Box linked list, deep drop leak-clean: met by `testdata/recursive.nra`.

## Size

- Estimated M, actual M.
