# selfhost

The Narra compiler written in Narra. A checked-in C seed (`seed/narrac.c`) builds it, and the test runner in `selfhost/tests` checks it. There is no Go compiler.

- Check each part through a dump with a fixed text format (`narrac lex` for tokens, `testdata/astdump`, `declsdump`, `typesdump` and `pkgdump` for the rest). A byte diff against the golden points at the first wrong token or node.
- Print optional and repeated children under named slot lines (`.result`, `.body`) so an absent child and an empty list differ in the dump. The doc comment of `astdump.nra` is the authority on the AST dump format, including positions.
- `narrac build selfhost` type-checks every module, even ones nobody imports, so a broken module fails the `bootstrap` suite.
- Keep the lexer's quirks (`0b102` is one int, `3.x` is `3` `.` `x`, `7e` is a float) unless the goldens in `testdata/lex` change with them.
- Thread compiler state through one `mutable` context struct, and keep every array in it unique. A second owner makes each push copy the whole array.
- Give every expression, pattern and local an integer id, and keep checker results in side tables indexed by id (`Array<T>`, `Map<int, T>`). Enum payloads are immutable and shared, so nothing can be written back into the AST.
- Narra has no recover. The parser keeps a `failed` field; once set, functions return dummy nodes and loops stop. Start every loop condition with `isOk(parser)` and never index tokens past EOF (`peekToken` clamps).
- Enum constructors are always qualified (`Token.TokIdent`). The `Tok` prefixes in token kinds are kept; renaming them is a separate change.
- `f(mutable p, g(p))` is rejected. Read the value into a local first.
- Loop over lists of tokens, statements and declarations. User code overflows the C stack at about 20000 frames; recurse only on nesting.
- Exported functions declare their effects (`uses alloc`, `uses io, alloc`). Private ones are inferred.
- Keep `Check`/`Emit` closures in struct fields, not enum variant fields. A closure in a variant field cannot be called after narrowing.
- Use top-level `const` for fixed tables (keywords, punctuation, precedence) and `std/char`, `Int.parse`, `Int.toHex` for conversions. Do not hand-write them.
- `scripts/test.sh` and `selfhost/bootstrap.sh` build this folder as the package at the repo root. Keep it there.
- The compiler source may only use features that the current seed compiles. A new feature lands in the compiler first, written without itself. Then run `sh scripts/seed-update.sh` to regenerate `seed/narrac.c`, and only then use the feature in this folder. Commit the seed after the last change to `selfhost/`.
- `src/runtimetext.nra` and `src/stdtext.nra` are generated, from `selfhost/runtime/` and `std/src/`. Edit those, never the generated files. `sh scripts/test.sh --suite generated --update` rewrites them; then regenerate the seed.
- Every report site passes a code from the table in `src/diagnostics.nra`. The `codes` suite checks it.
- Format every `.nra` file you edit, here and under `testdata/`: build a compiler with `sh selfhost/bootstrap.sh out`, then `out/narrac fmt <paths>`. The `sources` suite fails on unformatted files. Formatting can move lines, so update the positions in `// error:` lines and rewrite the goldens with `sh scripts/test.sh --update`. A sample that tests a layout the formatter would change goes in `LAYOUT_SAMPLES` in `selfhost/tests/src/sources.nra`.
