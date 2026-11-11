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

## Batch-7 (2026-11-11)

- Data and migrations: N/A. No stored data; the seed is regenerated at the end.
- Compatibility: the emitted C changes shape (tree shaking, macros, trailer). Covered by the bootstrap fixed point and the full suite. New syntax only adds forms.
- Auth and permissions, security, dependencies: N/A. Local compiler; no new libraries.
- Failure modes: contracts and build conditions add compile errors (E0984 to E0991) with error samples. A wrong `requires` fact is undefined behavior in C (premortem 3).
- Concurrency: N/A. Single-threaded compiler.
- Limits: place-ids must stay linear-ish on large bodies; the batch-6 exponential loop walk is the precedent (premortem 2).
- Observability: N/A beyond diagnostics.
- Performance: baseline in `bench/BASELINE.md` (end of batch-6) and the planning profile in the batch-7 brief; place-ids and the output nodes record instructions and C lines before and after; end-of-batch bench row.
- Testing: targeted suites per node, the new bench suite, full check and Linux check at the end, copy check on every program sample.
- Local files: N/A. No gitignored inputs.
- Developer experience: SPEC sections 12 and 13 and D95 updated by the syntax nodes; guide chapters if they mention contracts.
- Rollout: one story pull request.
- Cleanup: the outline briefs for contracts and build-conditions are refined in place; `extern` in build-targets waits for c-interop and is said in SPEC.
- Coverage: N/A (not imported).
