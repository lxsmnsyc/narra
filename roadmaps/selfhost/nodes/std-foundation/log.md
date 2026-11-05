# Log: std-foundation

- Started from origin/main f7fa23e. Nothing of the three steps existed.
- Generic mismatch: `shapeMismatch` in `check_call.go` and `calls.nra`; E0325 has its own format ("...; the argument does not fit the generic parameter") because E0119 owns "expected %s, found %s" and a code is found from the format.
- `std/internal/probe` (`std_internal.go`, row in `stdlib.nra`) proves the import rule; E0426 in `check.go` and `scopes.nra`.
- Stage 2 tests in `stage2_test.go`; samples under `testdata/stage2/`.
