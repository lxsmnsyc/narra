# Log: std-json

- Go is retired; built only in `selfhost/src`, `std/src`, `testdata/`.
- `Float.toString` (`lstr_from_f64`) already gives the shortest text that reads back, so only a parse primitive was needed: `FloatText.parse` (`std/internal/float`, emits `strtod`). `std/internal/probe` stays (test-only, used by existing samples).
- Map assignment to an existing key keeps its place and replaces the value (checked), so "last wins, first position" needs no special code.
- `Float.toString(100.0)` is `1e+02`; `std/json` writes whole floats below 1e15 as digits plus `.0` itself (`floatText`).
- Single-file packages did not order std modules; `addStdImports` now pushes a module after the modules it imports, so `std/json` comes after `std/shape` without `orderModules`.
- `narrac run -sanitize` is how a sample runs; `--update` of the programs suite did not write the `.out` files, so they were written from the run output and checked by eye.
