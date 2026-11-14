# Log: remove-inline

- Base: feat/batch-15 at 76f911ff. Base compiler built with `sh selfhost/bootstrap.sh` (249,695 lines of C).
- Proof, base: `bench/inline` built with `-release` retires 1926.2M instructions; `testdata/std_option.nra` output equals `std_option.out`; `narrac check` of a file with `@inline function double` reports no attribute error.
- Proof, branch: `bench/inline` retires 1929.9M instructions (+0.2%, inside the run noise; the benchmark's own figure was 1911.5M on a quieter machine); `std_option` output unchanged; the same file reports `unknown attribute @inline; the known attributes are @target, @trusted, @extern, @const`.
- Emit: `narrac emit selfhost` of the branch source, base compiler vs branch compiler: byte-identical (246,442 lines).
- Scratch: /private/tmp/claude-501/remove-inline/ (bench.sh).
