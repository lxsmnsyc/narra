# Log: local-functions

- SPEC line 62 says only "A function declared inside another function is local to that function." It says nothing on capture, ordering or recursion. Closures section: closures capture `const` and parameters; no rule for named local functions.
- Decision: a local function captures nothing (it is a module-level function in disguise). Using an enclosing local is E0919. Declarations are hoisted within their block, so a call may come before the declaration, and sibling local functions can call each other. Recursion needs nothing.
- Parser desugaring, as for-destructure: `function name` directly in a block becomes a private module-level `FuncDecl` named `__local_<name>_<line>_<column>` (position of the name). `block()` pre-scans the block's own level for `function name` tokens, so `Ident` renaming covers uses before the declaration. file() appends the hoisted decls after the enclosing declaration.
- No AST, clone, astdump, emitter change. The checker recognises hoisted names by the prefix (`localFuncSource`).
