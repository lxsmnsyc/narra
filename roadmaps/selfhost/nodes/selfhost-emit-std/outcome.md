# Outcome: selfhost-emit-std

## Built

- `selfhost/src/emitstd.nra`: `emitStdCall`, dispatching by `StdFunc.name` for every builtin of std/string, std/char, std/int, std/float, std/env, std/fs and std/process. 23 builtins that are one runtime call sit in the `FIXED_CALLS` table (`FixedCall`: call name, panic position, owned or plain temporary). The others have their own function: `Str.join/indexOf/split/chars/append/appendChar/length`, `Char.code/fromCode`, `Int.parse`, `Float.fromInt`, `Env.get/args`, `Fs.readText/writeText/listDir`, `Process.run`. `ioEither` and `ioResult` build Option and Result values through `Decls.enumAllocation`. An unknown name reports `internal error: no emit case for <name>` under this node, not a stub marker.
- The skip file `compiler/testdata/emit/skips/selfhost-emit-std.txt` is deleted (53 inputs). No input reports this node now.
- `reportStdWork` stays exported but returns `""`.

## Deviations

- `emit.nra` still calls `Std.reportStdWork`. I did not delete the line, because the neighbouring `Options.reportOptionsWork` line is removed by selfhost-emit-options and adjacent edits would conflict on merge. Whoever merges second (or a cleanup) deletes the `Std.reportStdWork` line and the function. The `std` arm of `refNeeds` in `emitstate.nra` is now unused for the same reason.

## Stubs left

- None in this node. The remaining diffs against Go on other inputs come from the stubs of match, closures, traits, maps and options.

## Gotchas

- `ioEither` writes the payload allocation lines inside the `if` and `else` branches with `emitter.indent` raised, as Go does. Order of `argumentText` calls matters for temp numbers; keep source order (`Str.append` emits the value before the builder pointer).
- `Str.split` needs a `lstr***` temp, which has no `Type`, so `rawTemp` takes C type text.
- Whole-output diff against Go (narrac vs `narra emit`) is byte-identical for strings.nra, strings2.nra, strings2_builder.nra, convert.nra, io_files.nra and structs.nra.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass.

## Size

- Estimated M, actual S (about 360 Narra lines).
