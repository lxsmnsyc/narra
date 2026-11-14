# Log: coroutine-send

- Proof on base (commit d64eee66, compiler built from `git archive` of it): `base/narrac run testdata/coroutine.nra` with `Co.send` calls fails: `module "std/coroutine" has no export "send"`, then `expected Step<int>, found void`.
- Branch: the same command prints the yielded steps; the errors suite reports `coroutine_resume_input` (two-argument `Co.resume`) as `this call passes 2 arguments, the function takes 1`.
- Async generators (D110) do not use the name: only std/coroutine, `checkModuleCall` and the effect label knew `resume`.
- A first base build from the worktree read edited sources and accepted `Co.send`; it was discarded and rebuilt from the archive.
- `emit selfhost` of the base source with the base and branch compilers: byte-identical (11,851,618 bytes).
