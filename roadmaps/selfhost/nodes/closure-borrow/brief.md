---
id: closure-borrow
title: Non-escaping closures capture let and mutable by reference
status: done
priority: should
type: feature
story:
early: false
outline: false
milestone: m11-closures
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - compiler/testdata/closure_borrow*
  - compiler/testdata/errors/closure_borrow_*
  - roadmaps/selfhost/nodes/closure-borrow/**
shared:
  - compiler/internal/narra/check_closure.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/emit_closure.go
  - compiler/internal/narra/emit.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0612, diag E0613, diag E0614, decision D77]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes after bootstrap go into both compilers together, and stage 2 must match Go on the new tests
---

# Non-escaping closures capture let and mutable by reference

## Goal

An arrow function passed straight to a non-escaping function-type parameter (a stack closure, `Closure.Stack`) may read and assign `let` variables and `mutable` parameters of the enclosing function, by reference:

```ts
function each(items: Array<int>, f: (item: int) => void) {
  for (const item of items) { f(item) }
}
let total = 0
each(numbers, (item) => { total = total + item })
```

This is sound with today's rules: a stack closure's environment lives on the caller's stack (D46) and the callee can only call it during the call, so a pointer to the caller's variable cannot dangle. Every other closure keeps today's errors.

## Context

- Capture check: `useLocal` in `compiler/internal/narra/check_closure.go` (Go) and `useLocal` in `selfhost/src/funcvalues.nra`. Both report E0602/E0603/E0604 only for the innermost arrow. `funcArg` marks `Closure.Stack` before the arrow is checked, so the flag is known at capture time.
- Emitter: `arrow` and `closureFunc` in `emit_closure.go`, and `selfhost/src/emitclosures.nra`. A stack environment is `{.h = {-1, NULL}, .c_x = x}` with borrowed captures. `localRef` in `emit.go` already emits `(*name)` for `Local.ByRef` locals (D50 and `mutable` parameters); reuse it for by-reference captures inside the closure body.
- Exclusive-use scan: E0929 (`checkMutableConflicts` in `check_call.go`, `checkMutableConflicts` in `selfhost/src/stmts.nra`) and the per-call check E0319/`checkExclusive`. See `nodes/mutable-read-conflict/outcome.md`.
- Narrowing after `is` (let-narrowing): `nodes/let-narrowing/outcome.md`.
- Source: D46, D51, D52; SPEC.md section 2 "Closures" and section 6 "Mutable parameters" and "Cycles".

## Rules

1. A stack closure may capture a `let` variable or a `mutable` parameter of an enclosing function. It reads and assigns the caller's variable through a pointer. Every `let` and `mutable` it uses is captured by reference, also when the closure only reads it (maintainer decision 2026-11-05). `const` bindings and non-mutable parameters keep today's capture.
2. Every arrow between the use and the variable's declaration must be a stack closure. A stack closure nested in an escaping one, or an escaping one nested in a stack one, keeps today's errors.
3. A variable that a stack closure assigns is borrowed for the call the closure is passed to. In that call it cannot be passed as `mutable`, read or assigned in another argument, or captured by another closure argument, at any depth (`items.length`, `xs[0]`). Reads by two closures of the same call are fine. Reason: constant parameters are passed by pointer (D50), so `each(items, (x) => { items = [] })` would free the array the callee is reading.
4. A captured `let` is not narrowed inside the closure, and a narrowing of a variable that a closure assigns ends at the call.
5. A closure that is stored, returned, captured by an escaping closure, passed to an escaping parameter, or passed through a function value (D51) keeps today's errors, with a message that says where capture is allowed.
6. Effects of the closure flow to the call as now. Assignment adds no effect.

## Acceptance

- [ ] Sample `compiler/testdata/closure_borrow.nra` and `.out`: a `let int` counter written by a closure passed to a user higher-order function; a `let string` and a `let Array<T>` reassigned and pushed inside the closure (old values released, ASan and leak check clean); a `mutable` parameter written by a closure; two nested stack closures writing the same outer `let`; two closures in one call that only read the same `let`; a closure that writes a `let` and the caller reads it after the call. Output proves the writes are visible.
- [ ] Error tests in `compiler/testdata/errors/closure_borrow_*.nra`, one per rule: capture of a `let` by an escaping closure (stored in a `const`, returned, passed to an `escaping` parameter, passed through a function value); a stack closure inside an escaping one; a written variable also passed as another argument, as `mutable`, and read in a field of another argument; narrowing used after a call whose closure assigns the variable. `// error:` headers with positions.
- [ ] Existing tests for E0603/E0604 still pass or are updated to the new messages.
- [ ] SPEC.md section 2 "Closures" and section 6 ("Mutable parameters", "Cycles": a `let` still cannot be stored, so no cycles) state the rules. DECISIONS.md gets D77 (maintainer decision 2026-11-05) with these reasons: the same semantics as TypeScript; one simple rule; cheaper, since the environment is on the caller's stack and by-reference captures need no retain or release of heap values. Prior art: Swift's non-escaping capture of `inout`. Rule 3 with its D50 reason. Rejected: by-value capture of `let` variables a closure only reads, which would be observably the same only because of the conflict rule, so it buys nothing. `const` bindings and non-mutable parameters keep today's capture.
- [ ] The Go compiler and `selfhost/` change together and match on the new tests; fast check passes including `TestSelfhostBootstrap`.

## Sizing

M. Uncertainty medium (new capture mode, known pattern from Swift), coupling low (one dependency), blast radius medium (checker capture and conflict scans, closure emitter, in both compilers), verifiability low (samples under ASan and error tests), size medium (about 6 Go files and 4 selfhost files), context medium.

## Notes for the worker

- Codes: E0612, E0613 and E0614 are reserved. A format change is a new code (D59): if the E0603/E0604 texts change, give the new texts new codes and leave the old rows unused, never reuse them. Likely uses: escaping closure captures a `let`, escaping closure captures a `mutable` parameter, a borrowed variable used in another argument.
- In the environment struct, a by-reference capture is a `T*` field. Initialize it with `&x` for a `let` local and with the pointer itself for a `mutable` parameter or a capture that is already by reference (nested closures), so no pointer to a pointer is built. Inside the closure, mark the local `ByRef` so `localRef` and assignment go through the pointer and release the old value.
- No reference count changes for by-reference captures: the environment borrows, as stack environments do now.
- selfhost/src has no higher-order functions today (no `Arrays.forEach` or `Arrays.map` exists in std, and no selfhost function takes a function-type parameter), so there are no loop workarounds to convert. Say so in the outcome; do not add higher-order code to the compiler for this node.
- The rule 3 scan can reuse the E0929 use collector: treat each variable a stack closure argument assigns as passed `mutable` for that call.

Merged as PR #92.
