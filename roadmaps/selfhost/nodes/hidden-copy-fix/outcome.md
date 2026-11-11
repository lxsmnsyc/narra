# Outcome: hidden-copy-fix

## Built

- `loadFile` and `loadPackage` in `selfhost/src/modules.nra` keep the std modules in the `StdLoad` value and push onto `load.modules`, so the push no longer copies an array the old `std` name still held.
- `testdata/copy_dead_name.nra` (+ `.copies`, `.out`): the reduced case. E0978 does not catch it (write after the other name's last use), so it is a known-gap example added to SPEC section 6 (shared writes, "Known gap").
- `takeOwnCopy` unguarded `NARRA_COPY_GRANT`: not reproduced. The macro is `((void)0)` under `NARRA_RELEASE` (`selfhost/runtime/runtime.h`), like `NARRA_COPY_GRANTED`; no code change.

## Proof

- base: 7a1a817f. Debug compiler = seed built without `-DNARRA_RELEASE`, emitting selfhost, built again. `NARRA_CHECK_COPIES=1 narrac check testdata/hello.nra` reports `array 1`; `narrac emit selfhost` exits 103 (`array 953, map 964`).
- branch: same command: `check testdata/hello.nra` reports nothing, exit 0; `emit selfhost` reports `array 952, map 964` (one fewer).

## Deviations

- SPEC.md edited (listed shared path): one example sentence in the known gap.
- Acceptance "a debug-built compiler runs without exit 103" holds for loadFile and a trivial check, but not for larger inputs (see Gotchas).

## Gotchas

- Other hidden copies remain in the compiler: `check testdata/arrays.nra` reports `array 3, map 1` (was `array 4`), and `emit selfhost` reports about 950 array and 964 map copies. These come from the checker and emitter, not from loadFile. They need their own node (find with the debug compiler; bisect by command).
- Pattern to avoid: `let x = holder.field` then a write to `x` when `holder` is dead. Push through `holder.field` instead.

## Review notes

- Acceptance narrowed by the orchestrator (2026-11-11): this node removes the copies in `loadFile` and `loadPackage` only. A debug-built compiler still exits 103 on most runs (`array 2, map 1` on `run testdata/copy_dead_name.nra`, about 950 array and 964 map copies on `emit selfhost`). The rest is an inbox item ("hidden copies left in the checker and emitter").
- `NARRA_COPY_GRANT` in `takeOwnCopy` needs no guard: it is `((void)0)` under `NARRA_RELEASE`, and copy-check-macros kept these two lines on purpose. The inbox text was stale.
- `testdata/copy_dead_name.*` documents the language gap; it reports `array 1` with either compiler, so it is not a regression test for this fix. `sh scripts/test.sh` runs a release compiler, so nothing automatic catches the `loadFile` copy coming back. The proof is the manual `check testdata/hello.nra` command in `log.md`.

## Verification

- `sh scripts/test.sh`: PASS, 7030 cases in 219s (includes the new copy_dead_name sample in both builds).
- `narrac fmt` run on modules.nra and the new sample.

## Size

- Estimated M, actual S.
