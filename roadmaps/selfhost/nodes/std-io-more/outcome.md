# Outcome: std-io-more

## Built

- New modules `std/io` and `std/path`, as builtins in both compilers.
  - `compiler/internal/narra/std_io.go`: `Io.readLine`, `readBytes`, `readAll`, `write`, `writeError`, `flush`.
  - `compiler/internal/narra/std_path.go`: `Path.join`, `dirname`, `basename`, `extname` (Node.js `path.posix` results; checked against Node 26 in `testdata/path.out`). `std/path` is builtin C, not Narra source, because the Go compiler loads source modules only for `selfhost/src` (std-source gate, option b) and `testdata/path.nra` has to run under Go.
  - `std_fs.go`: `Fs.isDir`, `createDir`, `createDirAll`, `removeFile`, `removeDir` (`Result<bool, string>`).
  - `runtime_io.h`: sections `io` and `path`, and the new `fs` functions.
- Selfhost port: `STD_FUNCS` rows in `selfhost/src/stdlib.nra`, `FIXED_CALLS` rows and emitters in `selfhost/src/emitstd.nra`, `runtimetext.nra` regenerated. `TestSelfhostBuiltins` and `TestSelfhostEmit` pass, so both compilers register and emit the same C.
- `run_test.go`: `<name>.in` is fed to the program's stdin; other samples keep `/dev/null`.
- Samples (debug mode runs under ASan/UBSan with the leak check): `io_stdin` (+ `.in`), `io_stdin_all` (+ `.in`), `io_write`, `io_dirs`, `path`. Error tests: `errors/io_no_import.nra`, `errors/io_remove_effect.nra`.
- Docs: SPEC.md section 7 module list, `compiler/README.md` "Supported".

## Deviations

- Brief said `io_stdin.nra` covers `readAll` of the rest; after the last-line-without-newline case stdin is at the end, so `readAll` of non-empty input is in a second sample, `io_stdin_all.nra`.
- Samples print through a local `yesNo` helper because bools cannot go into template strings.
- `Fs.removeFile` on a directory prints only `Err` in `io_dirs` (the reason is EPERM on macOS and EISDIR on Linux).

## Stubs left

- None.

## Gotchas

- `Io.readLine` keeps the `\r` of a `\r\n` line.
- `Result<Option<string>, string>` is built by hand in both compilers (`std_io.go`, `emitReadLine` in `emitstd.nra`): the Option is a plain temporary moved into Ok. Using `ioEither` for it would double-release.
- `Path.dirname("//a")` is `"//"`, copied from Node.
- `stdAliasPaths` in `diagnostic_fixes.go` (add-import fix) does not know `Io` and `Path`; not touched (outside the node's paths).
- No new diagnostic codes.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhostBuiltins|TestSelfhostEmit|TestSelfhostRuntimeText'`: pass.
- Fast check `cd compiler && go vet ./... && go test -timeout 45m ./...`: see the pull request.

## Size

- Estimated M, actual M.
