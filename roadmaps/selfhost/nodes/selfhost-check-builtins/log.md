# Log: selfhost-check-builtins

- Read the Go std_*.go checks; wrote selfhost/src/stdcalls.nra above closures.nra. exprs reaches it through a new Walker function, checkStdCall; checkBuiltinCall gained a `want` parameter.
- Fixed a latent mismatch exposed once tuple-typed inputs stopped skipping: Go's targs header drops every space after a comma, including in tuple types; typesdump.nra now does the same.
- Cases in testdata/typesdump/stdcalls, test in selfhost_stdcalls_test.go. Fast check passes.
