# const-eval-design log

## 2026-11-12

- Branch `spike/const-eval-design`, reset to `feat/batch-11` (20f9a487). No earlier log: fresh start.
- D105 does not exist yet. Built narrac with `sh selfhost/bootstrap.sh /private/tmp/claude-501/const-eval-design/out` (fixed point, seed current).
- `@const function square(...)` with `const area = square(4)`: `attr.nra:1:2: expected a name, found "const"`. Without the attribute: E0192 "constant "area" cannot call function "square"". Both as the brief says.
- Read `checkConstValue` (exprs.nra:2810), `isStaticConstant` and `constRef` (emitexprs.nra:1263, 1319), `lazyConstants` and the getter (emitconsts.nra). A static constant re-emits its AST in the const body; the getter emits the AST too. Turning an evaluated value back into AST would need fresh node ids and side-table rows in a body that has none, so D105 emits the value tree with its own small writer instead.
- Attributes: `parseAttributeList` accepts any name, but `const` is a keyword, so `@const` fails in `parseAttribute`. `addAttributes` already keeps attributes on `FuncDecl`.
- Std surface an interpreter must cover: registry rows in the allowed modules, counted by `name:` lines: string 37, bytes 11, int 7, float 6, array 6, result 6, f32 5, map 5, prelude 4, range 3, char 2, option, math, box, hash 1 each. About 95 rows. Std source modules (`std/src/option.nra`, `math.nra`, the `ext__*` files) are Narra bodies and run as user code.
- Floats: the runtime prints with `lstr_from_f64` (shortest `%.*g` that reads back). The compiler binary links the same runtime, so the interpreter calls the same function. Arithmetic is IEEE double on both targets; clang contracts `a * b + c` into an FMA on arm64 by default (`-ffp-contract=on`), so a constant evaluated at compile time could differ in the last bit from the same call at run time. D105 writes `#pragma STDC FP_CONTRACT OFF` at the top of the emitted C (see Determinism below).
- Maps keep insertion order (SPEC line 410, runtime_map.h), so the interpreter only needs an ordered map. Integer arithmetic wraps (D92); division and shifts are proven by the checker, so the interpreter cannot meet them unproven.

## Prototype (scratch, thrown away)

`/private/tmp/claude-501/const-eval-design/probe/walker.nra`: a tree walker in Narra over a small `Expr`/`Stmt` enum, values as an inline `Value` enum, one step per expression node, statement and loop turn, a budget check per step. It runs the CRC-32 table body (256 entries, 8 bit turns each) and gives the same table as the native function (`t1 1996959894`, `t255 755167117`), macOS arm64, `-release`.

| run | steps | instructions | user s | peak memory |
| --- | --- | --- | --- | --- |
| native `crcTable()` x 1001 | - | 25.2 M (about 14 K per table over a 11 M process floor) | 0.00 | 1.3 MB |
| walker, CRC-32 x 1 | 26,370 | 12.0 M (floor about 11 M) | 0.00 | 1.2 MB |
| walker, CRC-32 x 1000 | 26,370 each | 1164 M (44 per step) | 0.07 | 1.3 MB |
| walker, 65,536 entries x 16 bits | 13,041,666 | 581 M | 0.03 user, 0.31 real | 2.3 MB |

- The walker is about 80 times slower than native code on this body. A real evaluator (scopes, calls, every expression kind, std dispatch) will cost more per step; taking 10 times as a margin, 10 M steps is about 0.3 s, under 1 s.
- The CRC-32 table uses 0.26% of a 10 M step budget. A 16-bit table (65,536 entries) needs 13 M steps and fails a 10 M budget: that size is what `lazy` or a generated source file is for.

## Writing results into C

`gen_table.py N` writes `const TABLE = [N literals]` (today's D61 path, no `@const` needed) and builds it with narrac:

| entries | build | C lines | program start |
| --- | --- | --- | --- |
| 256 | 0.42 s `-release` | 1280 | 12.1 M instructions (floor about 11 M) |
| 4096 | 25 s `-release` | | |
| 16,384 | 126 s `-release` | | |
| 65,536 | stopped after 4 min `-release`; 153 s debug | | 19.9 M instructions debug |

The getter writes one `t->items[t->len++] = ...` per element, and clang's time grows faster than linear in that function. `gen_c.py N static` writes the same table as a `static const int64_t` array and one `memcpy` into the counted array: 0.11 s for 4096 and 0.17 s for 65,536 at `-O2` (the stores version of 4096 took 53 s in the same standalone test, under load). So D105 writes arrays of scalars that way. The same slowness exists today for literal table constants; an inbox note for the orchestrator is in the outcome.

## Determinism

`double f(double a, double b, double c) { return a * b + c; }` compiles to `fmadd` on arm64 at `-O2` and at `-O0` (clang default `-ffp-contract=on`). With `#pragma STDC FP_CONTRACT OFF` first it is `fmul` and `fadd`, and `cc -std=c11 -Wall` gives no warning. The emitted C has no such pragma today (`grep -c FP_CONTRACT stage2.c` is 0).

## Compiled C cost

`cc -O2` on an empty `main`: 0.04 s. A whole narrac build of the 256-entry program: 0.42 s.

## Fresh review and approval (2026-11-12)

- The maintainer approved every recommended answer. D105 records it.
- Review fixes: functions that take a function are out (E1052); only built-in and derived trait implementations run; std preconditions and the memory charge are checked before each std call; non-finite floats are written as `NAN` and `INFINITY`; `testdata/declsdump/**` shared; the D104 note. Nits: the first build covers six std modules (45 rows) and the rest is a follow-up node; field defaults stay out; this log follows the pragma choice.
