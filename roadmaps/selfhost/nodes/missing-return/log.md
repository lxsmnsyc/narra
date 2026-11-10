# Log

## Proof on the base (dca242d)

Sample (`missing.nra`): `function f(x: int): int { if (x > 0) { return 1 } }`, called as `f(0)` from `run`.

- `narrac check missing.nra` with the compiler built from the committed seed (equal to the base): exit 0, no output.
- `narrac run missing.nra`: `panic at missing.nra:1:1: function ended without returning a value`, exit 101.

## Branch

- `narrac check missing.nra` with the compiler built from this branch: exit 1, ``missing.nra:5:1: this function must return int on every path, but the `if` has no `else`, so the path where its condition is false does not return``.
- `narrac check selfhost` and `narrac check selfhost/tests`: no site rejected.
- std and the testdata samples: no site rejected (full suite).
- A generic function reports once: its instances are skipped, the declaration is checked.
