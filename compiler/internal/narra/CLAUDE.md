# compiler/internal/narra

- Add a std function only as an entry in its module's `std_*.go` file (`registerModule`). Do not add cases to `check_call.go` or `emit_call.go`. A new module is a new `std_<name>.go`.
- Put a module's runtime C in `Module.Runtime`, not in `runtime.h`.
- Report compile errors through `c.errorf` or `c.report`, which build a `Diagnostic`. Do not build error strings by hand.
- Every `Builtin.Name` must be unique. Registration panics on a duplicate.
- Registration runs in `init()`, and Go runs `init` in file name order. Do not depend on the order between modules.
- A refactor must keep the generated C and the error text byte-identical. Compare the output for every file in `compiler/testdata/` before and after.
