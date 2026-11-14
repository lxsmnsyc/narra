# float-shortest log

## Proof on the base

Base: feat/batch-15 at 6c616a46, compiler built with `sh selfhost/bootstrap.sh`.
Command: `narrac run -release testdata/float_shortest.nra`.
Result: `20.0` prints `2e+01`, `100.0` prints `1e+02`, `1e-7` prints `1e-07`, `F32.toString(20.0)` prints `2e+01`. Every other line reads back.

## Findings

- The formatter was `lstr_from_f64` / `lstr_from_f32` in `selfhost/runtime/runtime.h`: `%.*g` with the first precision that reads back. `%g` switches to exponent form when the exponent is at least the precision, so 20 at precision 1 became `2e+01`.
- Fix: find the shortest digits with `%.*e`, then lay them out like JavaScript (`lstr_from_scientific`): plain digits up to 21 integer digits, `0.000001` down to six leading zeros, otherwise `d.ddde+X` with no zero padding in the exponent.
- A whole number keeps `.0` (`20.0`, as in the existing goldens `3.0`, `100.0`); JavaScript would print `20`. SPEC does not say, and a float that reads back as an int would break `Json` and const output. Decision for the maintainer.
- NaN prints `nan`, infinities `inf` and `-inf`, `-0.0` prints `-0.0`, as before.
- The template holes, `print` of floats, `Show` and `Float.toString` all call the same two functions.
