# Log: const-functions-std

## Proof first

- Base: feat/batch-15 at 37ffedff, compiler built with `sh selfhost/bootstrap.sh <out>` (release).
- Command: `<base>/narrac run testdata/const_functions_std.nra`.
- Result on the base: five E1052 errors, one per function that calls a module outside the allowed list: `Math.powInt` (std/math), `Bytes.fromString` (std/bytes), `Fixed.get` (std/fixedarray), `I8.wrap` (std/i8), `F32.truncate` (std/f32). std/char, std/option and std/result were already allowed.

## Notes
