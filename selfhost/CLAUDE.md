# selfhost

The Narra compiler written in Narra. Each part is checked against the Go compiler through a text dump.

- Compare each ported part with the Go side through a dump with a fixed text format (`narra lex` and `LexDump` for tokens). A byte diff points at the first wrong token or node.
- Print optional and repeated children under named slot lines (`.result`, `.body`) so an absent child and an empty list differ in the dump. The `astdump.go` doc comment is the authority on the AST dump format, including positions.
- `narra build selfhost` type-checks every module, even ones nobody imports, so a broken module fails `TestSelfhostLexer`.
- Keep the Go lexer's quirks (`0b102` is one int, `3.x` is `3` `.` `x`, `7e` is a float) unless both sides change together. The dump tests compare them.
- Thread compiler state through one `mutable` context struct, and keep every array in it unique. A second owner makes each push copy the whole array.
- Give every expression, pattern and local an integer id, and keep checker results in side tables indexed by id (`Array<T>`, `Map<int, T>`). Enum payloads are immutable and shared, so nothing can be written back into the AST.
- Narra has no recover. The parser keeps a `failed` field; once set, functions return dummy nodes and loops stop. Start every loop condition with `ok(p)` and never index tokens past EOF (`tokAt` clamps).
- Enum variants are global constructors. Prefix variants that would clash across modules imported by name (token kinds are `TokIdent`, `TokInt`, ...).
- `f(mutable p, g(p))` is rejected. Read the value into a local first.
- Loop over lists of tokens, statements and declarations. User code overflows the C stack at about 20000 frames; recurse only on nesting.
- Exported functions declare their effects (`uses alloc`, `uses io, alloc`). Private ones are inferred.
- Keep `Check`/`Emit` closures in struct fields, not enum variant fields. A closure in a variant field cannot be called after narrowing.
- Use top-level `const` for fixed tables (keywords, punctuation, precedence) and `std/char`, `Int.parse`, `Int.toHex` for conversions. Do not hand-write them.
- `compiler/internal/narra/selfhost_test.go` builds this folder from `../../../selfhost`. Keep the package at the repo root.
