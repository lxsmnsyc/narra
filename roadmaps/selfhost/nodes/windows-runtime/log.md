# Log: windows-runtime

- Runtime C: binary stdio and fopen mode on Windows (constructor in runtime.h), narra_mkdir, __argv in std_env.c, rand_s in std_entropy.c, path separators and drive in std_path.c. clock, thread, fs, io compiled unchanged for x86_64-windows-gnu.
- Samples: testdata/windows_runtime.nra, testdata/windows_path.nra (checks the rule of the system it runs on, so one .out).
- Cross build with zig cc succeeds for both samples (PE32+ x86-64).
- wine: debian wine64 installs in docker (6 min) but the run failed twice: first with a wrong binary name, then with a docker input/output error on the VM disk. Skipped; the windows-latest job is the proof.
- 2026-11-15: rebased onto feat/batch-16. `sh scripts/test.sh` found one failure, `sources/formatted` on `testdata/windows_path.nra`; fixed with `narrac fmt`. Generated and guide suites pass.
- Cross build with zig cc repeated after the rebase: both samples give PE32+ x86-64. Native runs of both samples match the goldens. wine not retried: the machine load was above 100 from other workers.
- No workflow file: `.github/workflows/windows.yml` and `scripts/cross-check.sh` belong to node cross-targets.
- Review (fresh agent): no must-fix. Fixed: `Path.dirname("C:a")` and `("C:")` now give `C:` (was `C:.`), with a sample line; SPEC says UNC paths are unsupported and that stdio is binary; comments on `__argv` encoding and `_CRT_RAND_S`. Full `sh scripts/test.sh` after the fixes: one failure, a 1 s timing check in the regex sample under load 50; `--filter regex` passes alone.
