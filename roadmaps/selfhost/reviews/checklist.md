# Blindspot checklist

- Data and migrations: N/A. The compiler stores no data. Generated C is rebuilt every time.
- Compatibility: covered by the rule that every node keeps existing `testdata` programs passing. Language changes follow `SPEC.md`; differences go in `compiler/README.md`.
- Auth and permissions: N/A. Local command-line tool.
- Failure modes: compile errors (every node adds `testdata/errors/` cases), runtime panics (`testdata/panics/`), and `Result` errors in io-std.
- Concurrency: threads and async nodes. The compiler itself is single-threaded.
- Limits: deep recursion in programs (box-recursive leak check on deep trees), large generic instance counts (cap of 10,000 in the checker), large files (selfhost-bootstrap compiles the compiler itself).
- Security: io-std runs processes. `std/process.run` takes an argument list, never a shell string.
- Observability: sanitizer and leak check runs in every test. bench-harness for speed.
- Performance: bench-harness records the baseline, perceus reports before and after.
- Testing: `go test ./...` runs in about 10 s with no shared resources, safe in parallel worktrees. Benchmarks are exclusive.
- Developer experience: json-diagnostics, fmt, query-cli.
- Rollout: N/A. No users yet.
- Cleanup: maps and strings-std may leave special cases for `for ... of` that generators replaces. generators' outline should remove them.
- Coverage: mapped in `GOAL.md`.
- Dependencies: none. Go standard library only, system C compiler.
