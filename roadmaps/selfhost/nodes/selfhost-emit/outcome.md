# Outcome: selfhost-emit

## Built

- `compiler/internal/narra/selfhost_emit_test.go`: the skip machinery (`emitNodes`, skip files, markers, `compareEmitted`) is gone. `TestSelfhostEmit` fails on any `unsupported:` line, on exit code other than 0, and on any byte difference with Go's `CompileBinary`. Inputs that Go cannot compile must fail with exit 1.
- The selfhost package (`selfhost/`) is an input named `selfhost`, and `narrac emit selfhost` equals `narra emit selfhost` byte for byte (95,033 lines of C today; the brief said 58,893, the package has grown). No divergence fixes were needed.
- `compiler/testdata/emit/skips/` does not exist (earlier nodes deleted the files).
- Stub probes deleted: `Std.reportStdWork` and its call and import in `selfhost/src/emit.nra`; `hasWork`, `bodyNeeds`, `typeNeeds`, `refNeeds` in `selfhost/src/emitstate.nra`. `emit.nra` calls no `report*Work` function.
- `compiler/README.md` says the self-hosted compiler emits C.
- Time of `emit selfhost` (release build of narrac, `narra build -release`, three runs): Go `narra emit` 0.09 s; `narrac emit` 0.19 s (first run 0.57 s, cold). About 2x slower than Go. The harness narrac (ASan/UBSan) takes about 4 s for the same input.

## Deviations

- None. `reportUnsupported` and the `unsupported` field of `EmitResult` stay, since the emitter still uses them for constructs that should not occur; the harness treats any report as failure.

## Stubs left

- None in the emitter.

## Gotchas

- selfhost-bootstrap input: build `narrac` with `narra build -release -o narrac selfhost`, then `narrac emit selfhost` gives the stage 1 C. The test path is `../../../selfhost` from `compiler/internal/narra`.
- The selfhost emit comparison dominates `TestSelfhostEmit` runtime (about 4 s under sanitizers).

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass (all inputs byte-identical, including `selfhost`).
- `cd compiler && go vet ./... && go test ./...`: pass.

## Size

- Estimated S, actual S.
