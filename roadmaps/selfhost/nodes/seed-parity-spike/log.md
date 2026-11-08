# Log: seed-parity-spike

## 2026-11-08

- Base: `origin/feat/c-seed` at b569948, branch `spike/seed-parity-spike`.
- Built Go `narra`, stage 1 and stage 2 with `build-stage2.sh` (same commands as `selfhost/bootstrap.sh`, but keeps the binaries). Stage 2 is the release build (`-O2 -DNARRA_RELEASE`). Fixed point holds: Go C, stage 1 C and stage 2 C are identical.
- `survey-tests.sh`: 34 test files, 96 tests plus `TestMain`, 4957 lines. Each test classified by reading its doc comment and which Go internals it calls (`CompileToC`, `ASTDump`, `TypesDump`, `CheckDiagnostics` and so on).
- `diagdiff.py`: `narra emit` and `narra check --json` (Go) against `narrac2 emit` on errors/ (286), diagnostics/ (18), packages/errors/ (26), stage2/errors/ (2, stage 2 only). After removing narrac's `error ` prefix, every diagnostic line (file:line:col: message) equals Go's, in order, on all 330 samples (410 Go diagnostic lines). Stage 2 has no codes and no fixes: Go gives 142 distinct codes on these samples and 65 samples carry a fix.
- `habitdiff.py`: the 20 inline `habitCases` of `diagnostic_test.go` through both compilers: 20/20 identical text.
- `codematch.py`: can the code be found from the final message alone (Go's 289 formats as regexes)? Of 410 Go diagnostics: 289 match one format with the right code, 113 match several formats (mostly E0011 parser "expected %q, found %s" against E0119 checker "expected %s, found %s", and E0409), 3 match only a wrong format (E0326 messages matched as E0115), 5 match none (E0800 messages that start with "Function"). So codes must be given at the report site, not looked up from the text.
- `rundiff.py`: programs through stage 2, debug (ASan+UBSan, leak check) and release, compared with `.out`, and C compared with Go's: testdata/*.nra 81/81, stage2/*.nra 1/1, panics 9/9 (TestPanics 3 + TestStringPanics 6), packages with run.out 8/8 (debug, release, tests mode), runner/mixed 1/1. 25 s wall on this machine.
- `golddiff.py`: fmt goldens 55/55, fmt config cases 10/10, fmt-doc 14/14, trivia goldens 3/3, `fmt --check selfhost/src` clean, query 72/72 commands in 3 cases, guide 177/177 (89 `narra` blocks run in both modes, 88 `narra error` blocks, 11 fragments skipped).
  - First run failed every fmt case with options: `narrac fmt` finds `narra.toml` through `$PWD`, not the process cwd. The Go test sets `PWD` explicitly. Fixed the harness; a runner must do the same.
- `dumpdiff.py`: narrac2 dumps against the goldens Go writes: astdump 8/8, declsdump 53/53, typesdump 56/56, pkgdump 28/28.
- Command gap: read `compiler/cmd/narra/main.go`, `driver.go` (`Build`), `diagnostic_json.go`, `test_runner.go`, and `selfhost/src/main.nra`, `emitter.nra` (`runEmit`).
- Outside files: grep for `go build|go test|cmd/narra|compiler/internal|narra <command>` outside `compiler/` and `roadmaps/`.
- Generated files: `selfhost/src/runtimetext.nra` (77 KB, from Go runtime strings and the `runtime*.h` files embedded in Go) and `selfhost/src/stdtext.nra` (from `std/src`) are written by Go tests (`TestSelfhostRuntimeText`, `TestStdSources`). Nothing else regenerates them.
- std/process: `Process.run(command, args)` blocks, closes stdin, inherits env, has no timeout. std/fs has readText, writeText, exists, isDir, createDirAll, listDir: no remove, no temp dir. No threads yet.
