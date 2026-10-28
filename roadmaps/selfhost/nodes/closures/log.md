# Log: closures

- Branch `feat/closures` on 3d87ad0 (current origin/main). No closure code existed; only the traits slice of function types (parser, `FuncType`, `mangle`).
- Design: a function value is a C struct `{fn, env}` named by `mangle` (`F1__int__int`). The environment header `narra_env {rc, drop}` is shared by stack and heap environments; a stack one has rc -1, so retain and release skip it, like string literals.
- Escape rule kept syntactic: only a parameter written with a function type and without `escaping` is non-escaping (`Local.NoEscape`). Its only allowed uses are as a callee and as a direct argument to another such parameter. Checked at each `Ident` use (`useLocal`), with the one allowed `Ident` recorded before it is checked.
- Captures: an arrow body gets a new scope whose parent is the enclosing scope. A local found outside the arrow's scopes is a capture of every enclosing arrow it is outside of.
- Generic calls: `hintArrow` unifies the earlier arguments so `map(xs, (x) => ...)` knows `x`; the arrow's result then fixes `U`.
- `void` was not a type name; function types now accept `() => void` (`funcResult`).
- `traits_functypes.nra` stored and returned non-escaping parameters; marked them `escaping` and made the program call them.
- C names: `k<funcCName(owner)>__<n>` (function), `K...` (environment), `D...` (drop), `w<funcCName>` (thunk). Test blocks name closures after `f_<module>__test`; `test` is a keyword, so no user function has that name.
- All samples pass under ASan, UBSan and the leak check on the first run.
