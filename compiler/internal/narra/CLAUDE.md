# compiler/internal/narra

- Add a std function only as an entry in its module's `std_*.go` file (`registerModule`). Do not add cases to `check_call.go` or `emit_call.go`. A new module is a new `std_<name>.go`.
- Put a module's runtime C in `Module.Runtime`, not in `runtime.h`.
- Report compile errors through `c.errorf` or `c.report`, which build a `Diagnostic`. Do not build error strings by hand.
- Every `Builtin.Name` must be unique. Registration panics on a duplicate.
- Registration runs in `init()`, and Go runs `init` in file name order. Do not depend on the order between modules.
- A refactor must keep the generated C and the error text byte-identical. Compare the output for every file in `compiler/testdata/` before and after.
- Build C names only through `CName()` and `funcCName`, never from `Name`. Names carry the module.
- Compare struct and enum types by module and name, never by name alone.
- Before resolving another module's types or checking its bodies, `enter` that module's scope, as `funcInstance` and `buildGeneric` do.
- Cross-module tests go in `compiler/testdata/packages/<name>/` (with `run.out`) or `compiler/testdata/packages/errors/<name>/` (first line of `narra.toml` is `# error: <text>`).
- Emit calls to the standard trait functions through `e.useEq`, `e.useHash`, `e.useCmp` and `e.useShow`, not `eqFunc(t)` or a hand-built `eq_<type>`. The definition is only emitted when requested.
- All C names are made in `cnames.go`. Names may contain `_` anywhere. A new kind of generated name goes there too.
- Create function types only through `c.funcOf`. The emitter writes C structs only for function types it has listed.
- New per-function emitter state must be saved and restored in `closureFunc`, because closures are emitted in the middle of another body.
- Emit pattern tests as `&&` chains, with each tag test before any read of its union member.
- Names bound by a pattern retain when bound and release when the scope pops. Never borrow from the match subject.
