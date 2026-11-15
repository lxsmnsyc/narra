# Log: cross-targets

- Branch renamed to feat/cross-targets, reset to feat/batch-16.
- Script written at scripts/cross-check.sh; workflow windows.yml linted (actionlint, clean).
- Rebased onto feat/batch-16 (exit 0). Built the host compiler with selfhost/bootstrap.sh (stage 3 ok).
- First run: all rows built; docker was stopped, Rosetta absent. Started colima (2 cpu, 3 GB) and reran the Linux rows: all pass, seed C identical.
- Wine rows failed with a wine prefix error; the image now sets WINEARCH=win64 and initialises the prefix. Programs run, but stdout has CRLF: a windows-runtime matter, row stays an expected failure.
- Wasm rows fail on setjmp flags: wasi-runtime matter.
- Rebased onto windows-runtime: added windows_runtime and windows_path, Windows row passes under wine. Review fixes applied.
