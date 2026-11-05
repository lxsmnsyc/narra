---
id: generators
title: "Generators: state machines in C"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m5-concurrency
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - compiler/testdata/generator*
  - compiler/internal/narra/emit_generator.go
  - roadmaps/selfhost/nodes/generators/**
shared:
  - compiler/internal/narra/emit.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/emit_closure.go
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/cnames.go
  - compiler/internal/narra/runtime.h
  - compiler/internal/narra/check*.go
  - selfhost/src/**
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: generators-check
    kind: hard
    why: the emitter reads the checker facts generators-check adds (generator flag, yield statements, for-of source kind), in the same files of both compilers
---

# Generators: state machines in C

## Goal

Compile every `function*` that generators-check accepts to a state machine in plain C11, in the Go compiler and `selfhost/` together, so generator programs run, clean under ASan and the leak check, and stage 2 matches Go. `for ... of` over a `Generator<T>` resumes it until it finishes. No separate stack per generator (D19).

## Context

- Read `nodes/generators-check/outcome.md` first: it lists the checker facts this node reads.
- Function emission: `function` and `signature` in `compiler/internal/narra/emit.go`; scopes, frames and `unwindAll` in `emit_stmt.go`; `retain`/`release` in `emit_types.go`; tail-loop jumps (`narra_tail`) in `emit.go`. Selfhost: `emitter.nra`, `emitstmts.nra`, `emittail.nra`, `emitstate.nra`.
- Heap environments and their drop functions (`envRuntime`, `envStruct` in `emit_closure.go`) are the closest existing pattern: a counted header with a drop function pointer.
- Hashed C names: `cnames.go`, D54. New name kinds for the frame struct, resume, clone and drop functions follow the same scheme.

## Lowering

1. **Value.** `Generator<T>` is a C struct holding one pointer to a counted header `{ rc, state, resume, clone, drop }`. `resume(header*, T* out)` returns true with a value or false when finished. One C type per element type `T`, shared by all generators of that `T`.
2. **Frame.** Each generator (each instance of a generic one) gets a heap frame struct: the header, the parameters, every Narra local of the body, and every loop source and statement temp that is alive at a `yield` (for-of arrays and ranges, `match` statement subjects). Zero-initialized by `narra_alloc` or an explicit zeroing.
3. **Calling** the generator function allocates the frame with `rc = 1` and `state = 0`, stores the parameters with a retain, and runs nothing else.
4. **Resume.** A `switch (state)` at the top jumps to a `goto` label per resume point. `yield v` stores `v` into `*out`, sets `state` to the next point and returns true. Ending the body (or `return`) releases what is live, sets the finished state and returns false. Resuming a finished generator returns false again. No C local may be alive across a resume point; values alive there are frame fields.
5. **Release discipline.** A frame field that holds a heap value is released and set to zero when its Narra scope ends, so at any state the non-zero fields are exactly the live values.
6. **Drop before finish.** `drop` releases every non-zero heap field and frees the frame. A `break` or `return` out of a for-of over a generator, and dropping a `Generator<T>` variable, reach it through the normal release of the value.
7. **Copies.** A `Generator<T>` is a value (SPEC section 8). Resuming one whose count is above 1 first clones it: `clone` copies the frame and retains every non-zero heap field, as in-place updates copy shared values (SPEC section 6). `for (const x of g)` iterates its own copy, so `g` is unchanged after the loop.
8. **yield\*.** `yield* other` lowers to a frame field holding `other` and a loop that resumes it and yields each value.
9. **Closures** in a generator body capture frame fields by value, as they capture locals now.

## Acceptance

- [ ] Samples `compiler/testdata/generator*.nra` and `.out`: the SPEC section 8 example (`counter`, `evens`, `both` with `yield*`); a generator holding strings and arrays across yields; `yield` inside nested `while`, for-of over an array and over a range, and a `match` statement arm; a generic generator; a generator that is dropped unfinished (loop `break`, and a variable never iterated); iterating the same generator variable twice shows two full runs (copy rule); a generator passed to a function that iterates it; a generator with an arrow function capturing a frame local. All clean under ASan and the leak check.
- [ ] The "generators are not compiled yet" stub and its error test from generators-check are removed.
- [ ] SPEC.md section 11 "Code generation" describes the lowering in a few lines; `compiler/README.md` "Supported" lists generators.
- [ ] Go and `selfhost/` emit the same C for every new sample; fast check passes including `TestSelfhostBootstrap`.

## Sizing

L. Uncertainty high (first state-machine lowering in this repo, and it must be byte-identical in two compilers), coupling low (one dependency), blast radius medium (function emission and scope release, appended to rather than rewritten), verifiability low (samples under ASan and the leak check), size medium (one new file per compiler plus hooks), context medium.

## Notes for the worker

- Keep the lowering in `emit_generator.go` and `selfhost/src/emitgenerators.nra`; touch the shared emitter files only for hooks (local references that become frame fields, scope-end release, for-of over a generator).
- A local that is a frame field is referenced as `f->name`. The existing `localRef` switch is the hook; check where scopes release locals and make them release and zero the field.
- `case` labels inside nested blocks of a `switch` and `goto` into a block are valid C11, but they skip C initializers. Declare no C variable whose value must survive a `yield`.
- `async` (later) will reuse this lowering. Keep the frame, header and resume-point code free of generator-only assumptions where that costs nothing.
- If the element type `T` holds heap values, the value written to `*out` is owned by the caller of `resume`.

## Open questions

- None. The semantic choices are in generators-check (D78).
