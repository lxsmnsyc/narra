# Log: std-design

Machine: Apple M2, macOS 27. Compiler: `cc -O2 -DNARRA_RELEASE -std=c11 -w seed/narrac.c -lm` at `feat/batch-14` (4e2cf3c4). Programs in `/private/tmp/claude-501/std-design/` (scratch, not committed).

## D118 check

DECISIONS.md had no D118 and no std design for these modules. D12 rules out a pipe; D31 makes iterators generators; D83 and D84 give `Value` and `Shape`; D105 gives `@const` and its C writer.

## Iterator cost per stage

`iter/main.nra`: sum of `i * 2` over 20 million integers where the result is divisible by 3; `-release` build; `time`, two runs each, the second shown.

| mode | time | per item |
| --- | --- | --- |
| plain `for` over `Ranges.upTo` | 9 ms | (clang folds most of it) |
| `for` over a `function*` range | 63 ms | 3.0 ns |
| `mapStage` over that | 111 ms | +2.4 ns |
| `mapStage` then `filterStage` | 172 ms | +2.8 ns per stage |

`function* mapStage<T, U, effect E>(source: Generator<T> uses E, escaping transform: (value: T) => U uses E): Generator<U> uses E` builds and gives the same result. `escaping` is written before the parameter name, not the type.

## Arena prototype

`arena/main.nra`: generic `struct Handle<T> derives Equatable, Hashable` with an unused `T`, as a map key, and `get` over an `Arena<T>` of `enum Slot<T>`. Builds; `NARRA_CHECK_LEAKS=1` clean. `get` needs `from arena` (D94 shared result). `Arrays.get` does not exist on this branch (SPEC section 5 names it); the prototype checks the index by hand.

## Unicode table form

`table/lit.nra`: `const TABLE: Array<int>` of 6,000 literals. `narrac build -release`: 51.8 s total, of which `-emit-c` alone is 1.4 s; binary 269 KB against 34 KB for an empty program. `table/str.nra`: a 120,000-character ASCII string constant read with `Str.byteAt` at a run-time index: 0.15 s, binary 147 KB (+112 KB). A constant index let clang fold the read and drop the string, so the index must come from input.

Counts from Python 3.9 `unicodedata` (Unicode 13.0): 1,485 code points change in upper case, 1,393 in lower case; 2,061 canonical and 3,675 compatibility decompositions; 549 combining-class ranges. Sizes in D118 are estimates from these counts at about 6 to 12 bytes per record.

## narrac today

- `driver.nra` `parseOptions`: Go `flag` syntax, one or two dashes, `-name=value`, `-h`. `fmt.nra` and `query.nra` have their own loops. Scripts using single-dash long flags: `scripts/test.sh:61` and `bench/run.sh:91` (`build -release`).
- `manifest.nra`: a line-based TOML subset, E0405 to E0409.
- No `isatty`, `ioctl` or color code anywhere in the runtime or the compiler.
- Std sources are embedded in narrac through `stdtext.nra`, so a new std module reaches `selfhost/src` only after a seed update.

## Checks

- `NARRAC=narrac0 sh scripts/test.sh --suite spec`: PASS, 62 cases.
- `NARRAC=narrac0 sh scripts/test.sh --suite sources`: PASS, 1 case.
