# Outcome: selfhost-spike

## Built

- `selfhost/`: package `narrac` (`narra.toml`, one binary `narrac`, entry `main.run`).
  - `selfhost/src/lexer.nra` (545 lines): line-by-line port of `compiler/internal/narra/lexer.go` (372 lines). Exports `TokenKind`, `Token { kind, text, line, col }`, `lex(src): Result<Array<Token>, string> uses alloc`, `kindName`, `escapeText`, `hex`. It lexes an `Array<char>` from `Str.chars`, so columns count characters like the Go `[]rune` lexer.
  - `selfhost/src/main.nra`: `narrac <file>` prints the token dump.
- Dump format, shared by both sides: one line per token, `<line>:<col> <kind> <text>`, kinds `eof newline ident int float string char template keyword punct`. Text escapes `\` as `\\`, newline, tab and CR as `\n \t \r`, other control characters as `\u{<hex>}`. A lex error prints `error <line>:<col>: <message>` and exits 1.
- `compiler/internal/narra/lexdump.go`: `LexDump(src) (string, bool)`. `compiler/cmd/narra/main.go`: hidden command `narra lex <file>`.
- `compiler/internal/narra/selfhost_test.go`: `TestSelfhostLexer` builds `selfhost/` in debug mode (ASan, UBSan, `NARRA_CHECK_LEAKS=1`), runs it on every `.nra` under `compiler/testdata/` (programs, `errors/`, `panics/`, packages), on `selfhost/src/*.nra`, and on 15 edge cases (escapes, NUL, control characters, every lexer error), and compares stdout and exit code with `LexDump`. 161 files, all equal. Adds about 2 s to the test run.

## Gap list

Found while porting the lexer, by probe programs, and by reading the Go parser, checker, emitter and driver. Sizes: S (under a day of agent work, < 200 lines), M, L.

### Blockers (the port cannot be equivalent without them)

1. **No char/int conversion: `Char.code(c): int`, `Char.fromCode(n): Option<char>`.** Needed by the lexer (`\u{...}` escapes build a char from a number), the emitter (`CharLit` is emitted as its code, `%dU`; `testdata/strings2.nra` has `'λ'`) and `cString` (bytes as octal escapes). Workaround: `code` can be computed by decoding the UTF-8 bytes of `Str.fromChar(c)` with `Str.byteAt` (about 20 lines; the spike does the ASCII case). `fromCode` has no workaround above 0x7F: the spike indexes a 128-character string for ASCII and reports an error for larger escapes. No file in `testdata/` uses a non-ASCII `\u{}` today, so the spike test passes, but any later sample that does will break stage 1. Size S.
2. **Bootstrap runtime bug: `Str.fromChar('\0')` and `Str.appendChar(sb, '\0')` give an empty string.** `lstr_from_char` in `runtime.h` ends with `lstr_from_cstr(buf)`, which uses `strlen`. Hits the lexer (a NUL inside a string literal or source) and anything that builds strings char by char. Workaround in `lexer.nra`: `charString`/`pushChar` go through the literal `"\0"`. Fix: pass the encoded length to a length-taking constructor. Size S (one function plus a sample). Listed as a blocker only because it silently corrupts data; fold it into node 1.

### Decision needed before selfhost-front (no language change)

3. **The Go compiler is a mutable pointer graph; Narra has value semantics only.** The checker writes results into AST nodes (`Expr.T`, `CallExpr.Func`, `Pattern.Local`, `Ident.Func`), compares `*StructType` and `*Local` by pointer, and keeps caches keyed by pointers. Enum payloads in Narra are immutable and shared, and nothing has identity except future `Shared<T>`. The port cannot be line by line here. Options:
   - (a) Arena by hand: the parser numbers every expression, pattern and local (`id: int`), and the checker fills side tables (`Array<Type>` indexed by id, `Map<int, ...>`). Types become a recursive `enum Type` with structs and enums referred to by an index into a declaration table. Works today: probes show recursive enums as `Map` keys, `==` on them, and in-place `Array` updates through `mutable` all work. Cost: about 10-15% more code in check and emit, and every pass threads a context struct through `mutable`.
   - (b) Add `std/arena` (`Arena<T>`, `Handle<T>`, SPEC section 6 "Cycles") first. Same design as (a) with checked handles. Size M.
   - Recommendation: (a). The node ids must be designed in selfhost-front, so the user should confirm before it starts.

### Annoyances (workaround costs code or speed)

4. **No top-level `const`**: specified in SPEC section 10, not implemented. Hits keyword and punctuation tables (lexer), `binaryLevels` (parser), the builtin registry (about 100 entries with `Check`/`Emit` closures, checker and emitter), the prelude source, and the C runtime text. Workaround: functions that build the table, called once and kept in a context struct passed everywhere. Size M as a feature (constant expressions, C emission of static data).
5. **No embedding of files; long C text as Narra strings.** The emitter embeds `runtime.h`, `runtime_string.h`, `runtime_io.h`, `runtime_map.h` (1076 lines) with `//go:embed`. Plain strings cannot span lines, and template strings process `${` and escapes. Workaround: a small generator (Go or Narra) that writes `selfhost/src/runtime_c.nra` with one function per file, escaping `\`, `"` and newlines; or read the files at run time from a path next to the executable. Size S (generator).
6. **No `Result<void, E>`** (no unit value). Every fallible statement-like function returns `Option<string>` (Some = error) or `Result<int, E>`. Hits parser, checker passes, driver. Workaround as described; it reads backwards. Size M if a unit type is added.
7. **No early-exit error flow; the Go parser uses `panic`/`recover`.** `p.fail` panics and `file()` recovers. Narra has no recover, so every parse function must check and return. Workaround: a `failed: Option<string>` field in the parser state; functions return dummy nodes once it is set, and loops stop on it. About +10% parser lines. No feature planned (SPEC section 5 forbids `try`). Size: none, cost carried by selfhost-front.
8. **No integer parsing or formatting beyond `Int.toString`.** The checker validates int literals with `strconv.ParseInt(text, 0, 64)` (hex, octal, binary, overflow); the emitter re-prints them in decimal; C names need 8-digit hex (`hex8`). Workaround: hand-written `parseInt` with overflow check by comparison before multiply, and `hex` (in `lexer.nra`). Size S as std (`Int.parse`, `Int.toHex`).
9. **No 32-bit unsigned type.** xxHash32 for C names (`cnames.go`) works on `int` with `*%`, `+%` and a `& 0xFFFFFFFF` after each step; a probe matches the reference vectors in debug and release. Cost: helper functions `mul32`, `add32`, `rotl32`. Covered by the fixed-width integer types in SPEC; no node needed for the port.
10. **No sort.** `sort.Strings` is used in `check.go`, `check_call.go`, `check_closure.go`, `cnames.go`, `modules.go`; the C output order depends on it. Workaround: a generic merge sort `<T: Ordered>` in the port (about 40 lines). Size S as std (`Array.sort`).
11. **No stderr output and no `exit`.** Diagnostics go to stderr. Workaround: `Fs.writeText("/dev/stderr", text)` works on macOS and Linux (probe); exit codes come from the entry's return value. Size S as std (`eprint`).
12. **Value `match` arms must be single expressions.** Checker and emitter switches often need several statements per case. Workaround: statement `match` with `return` in each arm (exhaustive matches need no trailing return). Size S-M.
13. **`Map` has no `Equatable`, `Hashable` or `Show`.** A struct or enum holding a `Map` cannot use `==`, `derives`, or be a key. Hits types that hold field tables (struct types, module scopes). Workaround: `Array<#[string, T]>` for small tables, or keep maps only in the context struct. Size S (maps outcome lists it as a stub).
14. **Closures in an enum variant field cannot be called after narrowing** (closures outcome). The builtin registry stores `Check`/`Emit` functions; keep them in structs, not enum variants. No cost if designed that way.
15. **No directory test or recursive walk.** `LoadPackage` walks `src/**`. Workaround: `Fs.listDir` returns Err on a file, so recurse on Ok. No temp directories: `Process.run("mktemp", ["-d"])`, `Process.run("rm", ["-rf", dir])`. Size S.
16. **Go `%q` quoting in about 80 error messages.** Non-ASCII printable characters need `unicode.IsPrint` tables. Workaround: ASCII rules only (as `quoteChar` in `lexer.nra`), print other characters as they are. Only error texts differ, not C. Size S.
17. **Enum constructors and struct literals through a namespace are not supported** (`X.Token.Word`, `B.Pub { }`). The port imports AST and type enums by name; collisions (`Ident`, `Call` in both AST and types) need `Enum.Name`. Small code cost.

### Nice-to-have

18. **Speed.** Every fielded enum value allocates (box-recursive) and every copy counts references. The spike lexer in release mode takes 0.11 s for 18k lines, about 10x the Go lexer. The whole compiler on its own source (about 17k lines) should still take seconds, not minutes. perceus and `Array.values` by borrow would win most of it back.
19. **Contextual keywords.** `from`, `as`, `of`, `test`, `is` cannot be names (the spike had to rename a `from` parameter). Size S.
20. **No tuple index `t.0`, no `for (const #[i, x] of ...)`.** Destructure instead. Size S each.
21. **Generators** (`Map.keys` is an array today). Not needed for the port.

### Recursion depth

The Go parser recurses once per precedence level (10 levels in `binary`, then unary, postfix, primary), so one nesting level of an expression costs about 13 frames; `else if` chains recurse once per branch; statement and declaration lists loop. Depth grows with nesting, not with file size. The compiler's own source nests under 15 levels, so a Narra parser stays far below the 20000-frame limit (expected under 300 frames). The checker and emitter walk the same tree and have the same bound. The lexer loops.

## Port estimate

Ratio from the spike: 372 Go lines became 545 Narra lines (1.45x, including comments and gap workarounds). Go line counts are on main after effects (#14).

| Node | Go source | Go lines | Narra lines (est.) | Size |
| --- | --- | --- | --- | --- |
| selfhost-front | parser, ast, clone, modules, modules_toml, diagnostic (lexer done) | 2250 | 3500-4000 (+10% error flow) | L |
| selfhost-check | check*.go incl. effects, types, modules_check, builtins, Check half of std_*.go | 5600 | 8000-9000 | XL; split it (below) |
| selfhost-emit | emit*.go, cnames, Emit half of std_*.go; plus 1076 lines of runtime C as data | 2650 | 3800-4200 + C data | L |
| selfhost-bootstrap | driver, cmd/narra; stage 1 vs stage 2 vs Go diff harness | 250 | 400 + Go test | M |

Total about 16000-18000 Narra lines. The Go compiler keeps growing: effects alone added 500 checker lines. Every m3-m7 node that lands before the port adds to these numbers.

Suggested split of selfhost-check (XL): `selfhost-check-decls` (types, declarations, modules, generics, traits; L) and `selfhost-check-bodies` (expressions, statements, patterns, closures, effects; L).

## Proposed nodes

- id: `std-convert`, title: Char and integer conversions, size: S, priority: must, depends_on: [strings-std (hard)]. Scope: `Char.code`, `Char.fromCode` (Option), `Int.parse(s, base)` (Result), `Int.toHex`, and the NUL fix in `lstr_from_char`. Unblocks gap 1 and 2 and removes gap 8.
- id: `top-level-const`, title: Top-level constants, size: M, priority: should, depends_on: [modules (hard)]. Removes gap 4. Not required for the port; the context-struct workaround costs code in every pass.
- (Optional, if the user picks option (b) of gap 3) id: `arena-std`, title: std/arena with Arena and Handle, size: M, priority: should, depends_on: [box-recursive (hard)].

## Deviations

- None from the brief. Files outside owned paths: none (`compiler/cmd/narra/main.go` is shared; one `case`).
- `TestSelfhostLexer` also lexes `selfhost/src/*.nra` and 15 inline edge cases, beyond the brief's testdata files.
- Non-ASCII `\u{...}` escapes are a known difference (gap 1); the edge cases leave them out.

## Stubs left

- `lexer.nra` reports `\u escapes above 7f are not supported by the self-hosted lexer yet` where the Go lexer builds the character. Replace with `Char.fromCode` (node std-convert).
- `charCode` handles ASCII only; `charString`/`pushChar` work around the NUL bug. Replace in std-convert.
- `main.nra` prints errors to stdout.

## Gotchas

- The dump compares the Go lexer's quirks too: `0b102` is one int token (the Go lexer reads hex digits after `0b` and `0o`), `3.x` is `3` `.` `x`, and `7e` is a float token `7e`. Keep them when porting the parser, or fix both sides together.
- `quoteChar` copies Go's `%q` for ASCII only. A non-ASCII unexpected character that Go would escape (`unicode.IsPrint` false) prints differently.
- `selfhost_test.go` builds from `../../../selfhost`; it must stay at the repo root.
- `uses` is checked now: exported functions in `selfhost/src` need `uses alloc` or `uses io, alloc`.

## Proposed learnings

- Thread compiler state through one `mutable` context struct and keep every array in it unique | a second owner makes each push copy the whole array | selfhost/ | no
- Compare the self-hosted and Go outputs through a text dump with a fixed format | byte diffs point at the first wrong token or line | selfhost/, compiler/internal/narra/selfhost_test.go | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (on origin/main 2a0895f, after effects merged).
- `go test ./internal/narra -run TestSelfhostLexer -v`: 161 subtests pass.
- `narra build selfhost` and `narra build -release selfhost`: build; release lexes 18130 lines in 0.11 s.
- Probes in the scratchpad (not committed): AST-shaped mutual recursion, recursive enum as map key, xxHash32 vectors, `/dev/stderr` write.

## Size

- Estimated M, actual M.
