# Outcome: selfhost-emit-options

## Built

- `selfhost/src/emitoptions.nra` replaces the stub: `emitOptionCall` (Option.map, andThen, withDefault, orElse, isSome, isNone, okOr; Result.map, mapErr, andThen, withDefault, ok, err; Test.eq, ne, panics, panicsWith, golden) and `emitTestMain` (test runner text, the case table with each test's raw name from the module's `Decl.Test`, and `narra_test_run`). Callbacks go through `callFunctionValue`; payloads through `enumAllocation`.
- Removed the probe `reportOptionsWork` and its call in `emit.nra`.
- Deleted `compiler/testdata/emit/skips/selfhost-emit-options.txt` (7 inputs).

## Deviations

- `compiler/testdata/emit/skips/selfhost-emit-traits.txt` gains `std_test.nra` and `runner/mixed.nra`: `Test.eq` now calls `useEq` and `useShow`, so these inputs reach the traits stub, which the options stub hid before. selfhost-emit-traits removes the two lines when it lands.
- `selfhost/src/emit.nra`: one line deleted (the probe call).

## Stubs left

- None in this node. Remaining blockers for the test inputs: traits (std_test, mixed) and std (std_option, std_result).

## Gotchas

- `from` and `to` are keywords in Narra, even as parameter names.
- The test name is not in `BodyInfo.name` (that is a quoted dump). `emitTestMain` reads the n-th `Decl.Test` of the module scope.
- `Test.eq` and `Test.ne` return empty text with `owned: false`; the statement emitter prints only calls of void type.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass.
- Whole-output diff against Go: tests.nra is byte-identical; std_option, std_result and std_test differ only by std and traits stub markers and the counters they shift (temps, labels, closure numbers).

## Size

- Estimated M, actual M (about 400 Narra lines).
