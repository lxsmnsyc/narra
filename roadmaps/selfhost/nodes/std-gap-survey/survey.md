# Std gap survey

Which std pieces a first Narra user will miss, ranked by need. Base: `feat/batch-22` at b373c7b4, 55 modules in `docs/reference/std/README.md`. After std-async-fold, `std/async/*` becomes `std/net`, async `Time.sleep` and sync `std/fs`; this survey assumes that layout.

Evidence counts come from `selfhost/src`, `selfhost/tests`, `scripts/`, `testdata/`, `bench/` and `docs/guide/` (the generated `selfhost/src/std/stdtext.nra` is left out). "Promised" names the place in the repo that already says the piece will exist. "Source" names the std libraries that have it: Go, Rust, Zig, Node, Deno, Bun, Web APIs (Web), Crystal.

The cut: a piece goes before `v0.1.0` only when a first real program hits it and the workaround is longer than a line. A first real program here is a CLI tool, a script or a small parser. Every other piece waits.

## Ranked list

| # | Missing piece | Source | Evidence in this repo | Promised | Module | Where | Size | Release |
|---|---|---|---|---|---|---|---|---|
| 1 | Sorting: `Arrays.sort(mutable items)` for `Ordered`, `Arrays.sortBy(mutable items, compare)` | all | 19 hand-written sorts in 17 files (`emit.nra`, `emitdecls.nra`, `emitmaps.nra` each have a `mergeSort`; `sortStrings` 3 times); guide chapter 11 prints a count instead of sorted names | SPEC trait table: `Ordered` is for "`compare`, sorting"; inbox 2026-11-20 (guide-system-io) | `std/array` | `std/src` | S | before |
| 2 | Set type: `Set<T>` with `create`, `add`, `has`, `remove`, `size`, `values` | Go (map idiom), Rust, Zig, Node, Web, Crystal | `Map<_, bool>` used as a set in 42 files of the compiler, tests and scripts | brief (known gap) | new `std/set` | `std/src` | S | before |
| 3 | Array stack and search: `pop`, `contains`, `indexOf`, `reverse`, `truncate` | all | `popLast` (`driver/modules.nra`) and `dropLast` (`driver/textops.nra`) copy the array to drop one item; the checker keeps `scopeUndoCount` beside an array because nothing shrinks one; `testdata/traits.nra` writes `contains` | none | `std/array` | runtime (`pop`, `truncate`), `std/src` (rest) | M | before |
| 4 | Float from text: `Float.parse(text): Option<f64>` | all | `check/consteval.nra` `parseFloatText`; `std/json` and `std/toml` reach the hidden `FloatText.parse` of `std/internal/float` | none | `std/float` | `std/src` (wraps the existing `strtod` primitive) | S | before |
| 5 | Fixed decimals: `Float.toFixed(value, digits)` | Node, Web, Deno, Bun, Go (`strconv.FormatFloat`), Rust (`{:.2}`), Crystal (`round`, `format`) | `Float.toString` gives only the shortest text (`0.3333333333333333`); the guide keeps money in `int` cents so it never prints a float | none | `std/float` | runtime (`snprintf "%.*f"`) | S | before |
| 6 | Character classes: `isDigit`, `isLetter`, `isWhitespace`, `isAlphanumeric`, `isUpper`, `isLower` | Go (`unicode`), Rust (`char::is_*`), Zig (`std.ascii`), Crystal (`Char#digit?`) | 12 hand-written copies: `lexer.nra` (`isLetter`, `isDigit`, `isHexDigit`), `trivia.nra` `isSpace`, `manifest.nra` `isLetter`, `json.nra` and `toml.nra` `isDigit`, `scripts/errors/generate.nra`, `testdata/strings2.nra` | none | `std/char` | `std/src` (letter and space from the `std/internal/unicode` tables) | S | before |
| 7 | String edges: `lastIndexOf`, `trimStart`, `trimEnd`, `padStart`, `padEnd` | Node, Web, Deno, Bun, Go, Rust, Crystal | `fmtprinter.nra` `trimEnd`, `triviadump.nra` `padNumbers`, `scripts/unicode/generate.nra` `paddedByte`, `selfhost/tests` `trimSpaces` | none | `std/string` | `std/src` | S | before |
| 8 | File system: `Fs.appendText`, `Fs.rename`, `Fs.copyFile`, `Fs.removeDirAll`, `Process.cwd` | Go, Rust, Node, Deno, Bun, Zig, Crystal | guide chapter 11: "There is no function that appends"; `driver/depcache.nra` runs `mv` and `rm -rf` through `sh -c`; `driver/workspace.nra` runs `sh -c "pwd -P"` to learn the working folder | none | `std/fs`, `std/process` | runtime | M | before |
| 9 | `Process.run` options: input text, environment, working folder, output passed through | Go (`exec.Cmd`), Rust (`Command`), Node, Deno, Bun (`$`), Crystal (`Process.run`) | `driver/driver.nra` `startProgram` feeds stdin through `sh -c 'exec "$@" < file'`; `depcache.nra` sets `GIT_TERMINAL_PROMPT` through `sh -c` | inbox 2026-11-15 (deps-fetch) | `std/process` | runtime | M | after |
| 10 | Path: `relative`, `normalize`, `resolve`, `stem` | Node, Deno, Go (`filepath.Rel`, `Clean`), Rust | `selfhost/tests/src/files.nra` `relativePath` and `replaceSuffix`; `driver/workspace.nra` `normalizeDirectory` and `makeAbsolute` | inbox 2026-11-05 (basename suffix or stem) | `std/path` | `std/src` | S | after |
| 11 | Folder walk and glob | Go (`filepath.WalkDir`, `Glob`), Deno (`walk`, `expandGlob`), Bun (`Glob`), Node (`fs.glob`) | `tools/fmt.nra` `collectFiles`, `selfhost/tests/src/files.nra` `listPaths`, `findSources`, `listCorpus` | none | `std/fs` | `std/src` | M | after |
| 12 | Calendar time: UTC date parts and ISO 8601 text | Go (`time.Format`), Rust (none; chrono), Web (`Date`), Crystal (`Time::Format`) | none (only `unixMillis` exists) | none | `std/time` | `std/src` | M | after |
| 13 | Hex and base64 text of bytes | Web (`atob`, `btoa`), Node (`Buffer`), Go (`encoding/hex`, `base64`), Crystal (`Base64`) | 16 hand-written hex functions (`driver/hash.nra`, `lexer.nra`, `json.nra`, `toml.nra`, `scripts/unicode/generate.nra`) | none | `std/bytes` | `std/src` | S | after |
| 14 | Digests: SHA-256, SHA-1 | Web (`crypto.subtle.digest`), Go (`crypto/sha256`), Crystal (`Digest`), Bun (`CryptoHasher`) | `scripts/unicode/generate.nra` has its own `sha256` | none | new `std/digest` | `std/src` | M | after |
| 15 | Map extras: `values`, `has`, `clear` | Node, Web, Go, Rust | each is one line today (`table[key] is Some(_)`, a loop over `entries`) | none | `std/map` | runtime | S | after |
| 16 | Secure random bytes, UUID | Web (`getRandomValues`, `randomUUID`), Go (`crypto/rand`), Crystal (`UUID`) | `Random.fromSystem` gives only a seed | none | `std/random` | runtime | S | after |
| 17 | HTTP client and server, URL and query strings | Web (`fetch`, `URL`, `URLSearchParams`), Bun (`serve`), Go (`net/http`, `net/url`), Crystal (`HTTP`, `URI`) | none; `std/net` has TCP only | none | new `std/http`, `std/url` | `std/src` over `std/net`; runtime for TLS | L | after |
| 18 | Deque, priority queue, ordered map | Rust (`VecDeque`, `BinaryHeap`, `BTreeMap`), Go (`container/heap`), Zig | none | none | new `std/deque`, `std/heap` | `std/src` | M | after |
| 19 | Open files: read and write in parts, seek | Go (`os.File`), Rust (`File`), Web (streams), Deno, Zig | none (every read takes the whole file) | none | `std/fs` | runtime | L | after |
| 20 | File metadata: size, change time, permissions | Go (`os.Stat`), Rust, Node, Deno | none | none | `std/fs` | runtime | M | after |
| 21 | Temporary folders and files | Go (`os.MkdirTemp`), Deno (`makeTempDir`), Rust (crate) | `driver/depcache.nra` builds its own `tmp` folder under the cache | none | `std/fs` | runtime | S | after |
| 22 | Logging with levels | Go (`log/slog`), Crystal (`Log`), Deno (`std/log`) | none | none | new `std/log` | `std/src` | S | after |
| 23 | `Str.sliceFrom`, `Str.charAt`, `Str.nextChar` return `Option` instead of panicking | Rust (`get`) | `driver/textops.nra` wraps `Str.slice` and `Str.split` | SPEC section 5, "Std panics that remain" | `std/string` | compiler builtins | M | after |
| 24 | Async time operators: `debounce`, `throttle`, `merge` | RxJS, IxJS | none | inbox 2026-11-01 | `std/async` | `std/src` | M | after |
| 25 | `Opaque<T>` handles for C libraries | Go (`cgo.Handle`) | none | D122 item on opaque handles ("designed, not built"), node c-handles | new `std/opaque` | compiler and runtime | M | after |
| 26 | Signal handling (Ctrl-C) | Go (`os/signal`), Node, Deno, Crystal (`Signal`) | none | none | `std/process` | runtime | M | after |
| 27 | Semver compare | Bun (`semver`), Node (package) | `driver/depcache.nra` `isVersionBefore`, `driver/manifest.nra` `isVersion` | none | new `std/semver` | `std/src` | S | after |
| 28 | Thread cancellation | Go (`context`), Rust (none) | none | SPEC section 16 (deferred, no design) | `std/thread` | runtime | L | after |
| 29 | CSV, YAML | Go (`encoding/csv`), Crystal (`CSV`, `YAML`), Deno | none | none | new `std/csv`, `std/yaml` | `std/src` | M, L | after |
| 30 | Compression (gzip, deflate) | Go, Crystal (`Compress`), Web (`CompressionStream`), Bun | none | none | new `std/compress` | `std/src` or C library | L | after |
| 31 | SQLite, password hashing | Bun (`sqlite`, `password`) | none | none | packages | C library | L | after |
| 32 | Big integers, locale formatting | Web (`BigInt`, `Intl`), Go (`math/big`), Crystal (`BigInt`) | none | none | new `std/bigint`; packages for locales | `std/src` | L | after |
| 33 | UDP and DNS lookup | Go (`net`), Node (`dgram`, `dns`), Deno | none | none | `std/net` | runtime | M | after |

Already covered, so not in the table. From the Web APIs: `TextEncoder` and `TextDecoder` (`std/bytes`), `structuredClone` (`copy`), `AbortController` (`Async.cancel`), `performance.now` (`Time.now`). From Bun: `Bun.hash` (`std/hash`), `Bun.file` (`std/fs`). From Crystal: `OptionParser` (`std/args`), `Spec` (`test` blocks and `std/test`). Others: JSON and TOML, regular expressions, terminal styles, seeded random numbers, TCP sockets (`std/net`, async only, as decided).

## Before the release

**1. Sorting.** Every program that prints a list in order needs it, and `Ordered` already promises it. Add `Arrays.sort(mutable items)` for `T: Ordered` and `Arrays.sortBy(mutable items, compare)` with an effect variable on `compare`, both a stable merge sort in `std/src/ext__array.nra`. The compiler's 19 sorts then become calls, a cleanup for later.

**2. Set.** A TypeScript user writes `new Set()` within the first hour, and the compiler itself keeps 42 files of `Map<_, bool>`. Add `std/set` as an `opaque struct Set<T>` over `Map<T, bool>` with `create`, `add`, `has`, `remove`, `size` and `values` (an array, so `for ... of` works). Insertion order follows the map.

**3. Array stack and search.** `push` exists with no `pop`, so a stack needs a counter beside the array or a full copy (`popLast`, `dropLast`). Add `pop(mutable items): Option<T>` and `truncate(mutable items, length)` as builtins that shrink in place, and `contains`, `indexOf` and `reverse` in `std/src` over `Equatable`.

**4. `Float.parse`.** Reading a number from input or arguments is a first-program task, and `Int.parse` exists while the float form does not. `std/internal/float` already has `FloatText.parse` over `strtod`, so the public function is a wrapper that returns `None` for text `strtod` does not consume whole.

**5. `Float.toFixed`.** Printing an average or a time in seconds gives 17 digits today. Add `toFixed(value, digits)` with the rounding of C `snprintf("%.*f")`; it returns `None` for a negative digit count or one above 20.

**6. Character classes.** Every tokenizer starts with `isDigit` and `isLetter`, and the repo has 12 copies. Add them to `std/char` for `char`, Unicode-aware for letters and white space through the existing tables, ASCII for digits.

**7. String edges.** `trimStart`, `trimEnd`, `padStart`, `padEnd` and `lastIndexOf` are the TypeScript names a user types first. Aligned table output needs the pad functions. All five are short Narra functions in `std/src/ext__string.nra`.

**8. File system gaps.** A log file needs append, and the guide tells the reader to read, join and write the whole file back. A tool that writes output safely needs `rename`, and one that cleans up needs `removeDirAll`; the compiler shells out to `mv` and `rm -rf` for both. `Process.cwd` replaces `sh -c "pwd -P"`. `copyFile` comes with them. All are thin C over POSIX and Windows calls, with a row in `selfhost/src/buildconditions.nra` for WASI where needed.

## After the release

Rank 9 (`Process.run` options) is the first item after the cut. The compiler needs it, but a first program can run a command without input or environment. Parameter defaults let the options arrive later without breaking a call. The rest follow the rank. The larger network and data-format modules (17, 29 to 32) can live as packages before they enter std.
