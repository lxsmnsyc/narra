# Log: selfhost-spike

- 2026-10-30: started from origin/main 4bf0e7f. No `selfhost/` yet, so not fulfilled. #14 (effects) still open.
- Dump format chosen: `<line>:<col> <kind> <text>`, kinds lowercase (`eof`, `newline`, `ident`, ...). Text escapes `\\ \n \t \r`, other control chars `\u{hex}`. A lex error prints `error <line>:<col>: <message>` and exits 1. Go side in `lexdump.go`, `narra lex` in `cmd/narra/main.go`.
- Gap: no char <-> int conversion (`Char.code`, `Char.fromCode`). The lexer needs `fromCode` for `\u{...}`. Workaround: ASCII via `Str.byteAt(Str.fromChar(c), 0)` and a 128-char lookup string; non-ASCII `\u{}` unsupported. No testdata file uses `\u{`.
- Gap: no way to write to stderr (`eprint`) and no exit without returning from `run`. Workaround: errors go to stdout, `run` returns the code. Later found `Fs.writeText("/dev/stderr", ...)` works (probe).
- Gap: `Result<void, E>` impossible (no unit). Lexer helpers return `Option<string>` (Some = error) instead.
- Gap: `from` is a keyword, so it cannot name a parameter (`text(lx, from, to)` failed to parse). Renamed.
- Port compiled on the second try. 143/143 testdata files matched on the first run.
- Bug found with edge cases: `Str.fromChar('\0')` and `Str.appendChar(sb, '\0')` give "" (runtime.h `lstr_from_char` ends with `lstr_from_cstr`, which uses strlen). Workaround in lexer.nra (`charString`, `pushChar` through the literal "\0"). Not fixed in runtime.h (outside owned paths).
- Gap: a `match` used as a value cannot have a block arm (probe). Use a statement match with `return` in each arm.
- Probes (scratch, not committed): mutually recursive struct/enum AST, a recursive enum as a `Map` key and `==` on it, read-modify-write of a struct in a `Map` via spread, `Fs.writeText("/dev/stderr")`: all work. xxHash32 in Narra with `*%`, `+%`, `<<`, `>>`, `&`, `|` and a `& 0xFFFFFFFF` mask after each step matches the reference vectors for "" and a 39-byte string, in debug and release.
- Speed: release narrac lexes 18k lines in 0.11 s (Go: about 0.01 s for 1.8k lines); linear, so `mutable` threading keeps arrays unique.
- Parser depth: `binary` recurses through 10 levels per nesting level, else-if chains recurse once per `else if`; statement and declaration lists loop. Depth follows nesting, not file length; far below 20000.
- #14 merged (d7e9c93) while working; rebased onto 2a0895f, test still passes with effect checking on.
- Review and gate passed. Rebased onto origin/main after json-diagnostics (#15): conflict in `cmd/narra/main.go` (kept `check`, `explain` and `lex`). `Token.End` and Diagnostic lex errors leave the dump unchanged; test passes. Review doc fixes applied.
