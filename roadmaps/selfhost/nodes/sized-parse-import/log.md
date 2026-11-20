# Log: sized-parse-import

- Base proof (feat/batch-21, compiler built with `sh selfhost/bootstrap.sh`): `narrac build a.nra` with the brief's program (`U8.parse("7", 10)`, no std/int) failed: `error: call to undeclared function 'narra_sized_parse'`.
- Cause: the emitter writes the C runtime of a module only when the module is in `checker.used` and a helper of it was noted. `narra_sized_parse` lived in `selfhost/runtime/std_int.c`, so only a program that imports std/int got it. The std/u8..std/u64 modules have no runtime of their own.
- Fix: moved the function into `selfhost/runtime/std_sized.c` (runtime of `std/internal/sized`), `emitParse` notes that helper, and `emitProgram` writes it when noted.
- Other sized functions: toString uses `lstr_from_int`/`lstr_from_uint` and the checked arithmetic and tryFrom use compiler builtins, all in runtime.h; no gap. `std/f32` has no runtime function. Sample `testdata/sized_parse_alone.nra` uses every sized module and f32 without std/int.
