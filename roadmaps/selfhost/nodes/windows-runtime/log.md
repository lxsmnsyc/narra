# Log: windows-runtime

- Runtime C: binary stdio and fopen mode on Windows (constructor in runtime.h), narra_mkdir, __argv in std_env.c, rand_s in std_entropy.c, path separators and drive in std_path.c. clock, thread, fs, io compiled unchanged for x86_64-windows-gnu.
- Samples: testdata/windows_runtime.nra, testdata/windows_path.nra (checks the rule of the system it runs on, so one .out).
- Cross build with zig cc succeeds for both samples (PE32+ x86-64).
- wine: debian wine64 installs in docker (6 min) but the run failed twice: first with a wrong binary name, then with a docker input/output error on the VM disk. Skipped; the windows-latest job is the proof.
