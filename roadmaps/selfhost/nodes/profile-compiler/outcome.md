# Outcome: profile-compiler

## Built

- Spike only. No compiler code changed. The profile is of `narrac emit selfhost` at commit 55fb951 (`feat/batch-4`).
- **Verdict on const-params-pointer: worth doing.** Copying structs passed by value takes 5.0% to 5.8% of the run in `memcpy` alone. That is a lower bound. The detail is under "const-params-pointer" below.
- About 60% of the time goes to allocation and string building. Most of it comes from a few helpers that build strings to use as keys or names. Ten follow-up briefs are proposed below, most valuable first.
- Scripts in this folder:
  - `build.sh <dir>` builds the release compiler with `selfhost/bootstrap.sh`, as `bench/run.sh` does. It also builds `narrac-g` (the same C with `-O2 -g -fno-omit-frame-pointer`), `narrac-loop` (the same, plus `main` runs the driver `NARRA_REPEAT` times so `sample` can attach) and `narrac-noinline` (`narrac-loop` with `-fno-inline`).
  - `record.sh <narrac-loop> <out.txt> <seconds>` runs the compiler 1000 times in one process and runs `sample` on it.
  - `analyze.py` gives the top functions, category and phase shares, and the callers of a function (`--callers`).
  - `copies.py` splits `memmove` samples by the C statement that made the copy.
  - `allocators.py` splits allocation samples by runtime helper.
  - `sizes.py` gives the C size of each struct passed by value.
- Text summaries are in `profiles/`. They hold only the reports. The raw `sample` files (5 to 11 MB) are not committed.

## Result

- Timing: the release `narrac` takes 0.60 s per run (5 runs, 0.598 to 0.621 s after a cold first run). `narrac-g` and `narrac-loop` take the same 0.60 s, so the profiled code is the code that ships. `narrac-noinline` takes 0.83 s. Peak footprint is 131 MB.
- Profiles: `sample` at 1 ms. There are two 30 s runs of `narrac-loop` (`opt`: 20,493 main-thread samples, `opt2`: 20,498) and one of `narrac-noinline` (21,194). The two `opt` runs agree within 0.5 points on every share. The tables below use `opt`.
- Load: the 1-minute load average was 2.1 to 2.5 on 8 cores during all three runs, while other worktree agents were running. These shares are fractions of the samples, so the load matters less than it would for wall-clock times.

### Top 10 by self time (all frames)

| # | function | self |
| - | -------- | ---- |
| 1 | `_platform_memmove` (libc `memcpy`) | 12.1% |
| 2 | `_xzm_free_tc` (malloc) | 9.9% |
| 3 | `_xzm_malloc_tc` (malloc) | 9.3% |
| 4 | `_malloc_zone_malloc` | 6.2% |
| 5 | `_free` | 6.1% |
| 6 | `_platform_memcmp` | 3.4% |
| 7 | `__vfprintf` (from `snprintf` in `lstr_from_int`) | 2.5% |
| 8 | `lexer.runLexer` (`f_bc628247_runLexer`) | 2.4% |
| 9 | `lexer.createLexer` | 1.9% |
| 10 | `DYLD-STUB$$malloc` | 1.6% |

Top 10 compiler and runtime functions by self time, without libc:

- `lexer.runLexer` 2.4%, `lexer.createLexer` 1.9%, `effects.walkExpression` 1.5%.
- `retain_S_b2b309ba_Checker` 1.4%, `lstr_concat` 1.0%, `parser.isAt` 1.0%.
- `release_S_b2b309ba_Checker` 0.8%, `parser.parseBinary` 0.8%.
- `check.findLocal` 0.7%, `unused.reportUnusedTopLevel` 0.7%.

### Top 10 by inclusive time

`start`, `main` and `emitter.runEmit` are left out, because each is 100%. The top of the call tree is:

| # | function | inclusive |
| - | -------- | --------- |
| 1 | `emit.emitProgram` | 36.1% |
| 2 | `emitconsts.emitFunction` | 30.8% |
| 3 | `emitstmts.emitStatement` | 27.6% |
| 4 | `modules.loadPackage` | 24.1% |
| 5 | `parser.parse` (includes lexing) | 23.4% |
| 6 | `body.checkModules` | 20.6% |
| 7 | `body.checkFunctionBody` | 18.3% |
| 8 | `effects.checkEffects` | 15.9% |
| 9 | `effects.walkExpression` | 13.2% |
| 10 | `lexer.lex` | 11.5% |

The helpers with the highest inclusive time point at the fixes:

| function | inclusive | what it does |
| -------- | --------- | ------------ |
| `lstr_concat` | 11.3% | one allocation and two copies per template part |
| `snprintf` | 6.9% | `lstr_from_int`, called through `Int.toString` |
| `effects.positionKey` | 6.1% | builds a string key for every expression |
| `cnames.typeId` | 6.1% | rebuilds a C type name and its hash on every use |
| `lexer.readText` | 4.9% | builds each token's text one char at a time |
| `parser.finishFacts` | 4.9% | splits the whole source into lines and lexes template parts again |
| `emitstate.temp` | 4.5% | C type name, temp name and line for each temp |
| `emitstate.refCountCode` | 4.3% | `typeId` and `needsRc` on every retain or release |
| `emitstate.cString` | 3.4% | escapes C literals with `Str.slice` per byte |
| `release_S_b2b309ba_Checker` | 3.3% | whole-`Checker` release, mostly in tail-loop functions |

### Share per category

Self samples, `opt` and `opt2`. A libc leaf (`memmove`, `memcmp`, `printf`) is counted in the category of the runtime helper that called it. When a compiler function calls one directly, it is counted by the C statement that made the call (see `analyze.py`).

| category | opt | opt2 | -fno-inline |
| -------- | --- | ---- | ----------- |
| allocation (`malloc`, `free`, `narra_alloc`, `narra_free`, `narra_realloc`) | 39.6% | 39.4% | 32.2% |
| strings (`lstr_*`, `narra_sb_*`, `narra_str_*`, int to string, `memcmp`) | 20.2% | 20.2% | 20.4% |
| reference counting (`retain_*`, `release_*`, `drop_*`, `*_retain`, `*_release`) | 9.3% | 9.4% | 17.4% |
| by-value struct copy (`memcpy` for a struct argument) | 5.2% | 5.0% | 4.5% |
| maps (`M_Map_*`) | 1.1% | 1.1% | 2.1% |
| arrays (`A_Array_*`) | 0.6% | 0.6% | 8.3% |
| other compiler code | 24.1% | 24.4% | 15.2% |

- At -O2, inlined `rc++` and array `get` are counted in their callers' "other compiler code". The -fno-inline column gives an upper bound for reference counting (17.4%) and arrays (8.3%), because calls cost more there.
- `allocators.py` on the -fno-inline profile splits the allocation time by helper. String helpers (`lstr_new`, `lstr_release` and the rest) are 17.6% of samples, more than half the allocation. Boxes, enum payloads and closures allocated in compiler code are 6.0%, arrays 5.7%, frees inside `release_*` 1.5% and maps 1.3%.

### Share per phase

The phase is the call site in `runEmit` (`selfhost/src/emitter.nra`) on the stack. Lex covers frames under `narrac/lexer`, and parse covers the rest of `Modules.load`.

| phase | opt | opt2 | main costs inside the phase |
| ----- | --- | ---- | --------------------------- |
| lex | 12.2% | 12.1% | `readText` 4.9%, `createLexer` 2.4%, `runLexer` self 2.4% |
| parse | 11.2% | 11.3% | `finishFacts` 4.9% |
| check (`checkPackage`, `findEntry`) | 20.6% | 20.5% | allocation 6.7%, `funcvalues.useLocal` 2.7% |
| effects (`checkEffects`) | 15.9% | 15.9% | `positionKey` 6.1%, `functionKey` 2.5% |
| emit (`emitProgram`) | 36.1% | 35.9% | `typeId` 6.1%, `temp` 4.5%, `cString` 3.4%, `writeLine` 2.6% |
| write (`narra_fs_write` to `/dev/null`) | 0.0% | 0.0% | |
| teardown (releases at the end of `runEmit`) | 3.3% | 3.4% | `release_S_b2b309ba_Checker` |
| load other (file reads, module graph) | 0.7% | 0.8% | |

Check and effects together are 36.5%.

### const-params-pointer

- **Verdict: worth doing.** It saves an estimated 5% of `narrac emit selfhost` (about 0.03 s), and it speeds up any program that passes large structs.
- **Measured:** `copies.py` attributes 5.8% of `opt` samples (1,179) to `memcpy` calls whose calling C statement passes a struct by value, and 5.6% of `opt2`. The callers are spread over the checker and emitter. The largest are `stmts.collectUses` 0.5%, `emitstate.needsRc` 0.5%, `exprs.checkExpressionKind` 0.2%, `retain_S_4d4d5167_Emitter` 0.2% and `effects.walkCall` 0.2%.
- **Lower bound:** clang copies small structs with inline loads and stores, and the profile cannot see those. Only the larger copies call `memcpy`.
- **Struct sizes** (`sizes.py`): `Checker` is 624 bytes and is passed by value to 255 parameters. `Emitter` is 928 bytes in 72 parameters, and `Parser` is 200 bytes in 23. Of 2,644 by-value struct and enum parameters, 864 are over 16 bytes and 500 are over 128 bytes.
- **Size limit:** use 16 bytes. Above 16 bytes, the arm64 C ABI already passes a struct as a pointer to a copy the caller makes. The copy is all that a by-value parameter adds.
- **Helpers:** the generated `retain_S_*` and `release_S_*` helpers also take the struct by value. `retain_S_Emitter` copies the 624-byte `Checker` to call `retain_S_Checker`. They should take a pointer as well.
- **Separate cost:** whole-struct retains in tail-loop functions are a different 4.8% (proposal 3). const-params-pointer does not remove them.

## Proposed briefs

Shares overlap, for example `positionKey` time includes `snprintf` and `lstr_concat` time, so do not add them up. "Saves" is an estimate from the measured share.

1. **Int to string without `snprintf`**
   - Goal: `lstr_from_int` and `lstr_from_uint` in `selfhost/runtime/runtime.h` write the digits into a stack buffer themselves. Regenerate `selfhost/src/runtimecore.nra` from it.
   - Evidence: `snprintf` is 6.9% inclusive and `__vfprintf` is 2.5% self. Callers are `effects.positionKey`, `effects.functionKey`, `emitstate.localName`, `emitstate.temp` and `cnames.internedName`.
   - Saves about 5%, and the change helps every Narra program that formats ints.
   - Size: S.
2. **Integer keys in the effect pass**
   - Goal: `effects.nra` keys `valuesAt`, `bodyIndex`, `entries`, `sources` and `locals` by ints (body index and node id, or a packed position) instead of `positionKey`, `functionKey` and `localKey` strings. Stop copying the `valuesAt` array on every push, too. `walkExpression` reads the array from the map, pushes, and stores it back while the map still holds it, so the array is copied each time.
   - Evidence: `positionKey` 6.1%, `functionKey` 2.5% and `localKey` 0.6% inclusive, out of 15.9% for the whole effects phase.
   - Saves up to 8%.
   - Size: S.
3. **Tail-loop functions borrow the parameters they pass through**
   - Goal: `openTailLoop` (`selfhost/src/emittail.nra:235`) retains every non-`mutable` parameter at entry, so that a tail jump can replace it. A parameter that every tail call passes back unchanged, such as `emitter` in `needsRc` or `checker` in `collectUses`, needs no retain and no release.
   - Evidence: 25 functions retain a whole `Checker` or `Emitter` at entry. Retains and releases of those two structs inside compiler functions take 4.8% (974 samples). The top functions are `needsRc`, `collectUses`, `mayWrite`, `underlyingType` and `findSourceCall`.
   - Saves about 4%.
   - Size: S to M.
4. **Cache C type names**
   - Goal: `cnames.typeId` and `cnames.cType` rebuild the name, `xxh32` hash and `hex8` text on each call. Cache them per type key. `Emitter.typeNames` (`selfhost/src/emitstate.nra:53`) is declared but never read or written.
   - Evidence: `typeId` 6.1% and `cType` 4.9% inclusive. The callers are `temp`, `refCountCode`, `functionSignature` and `declareLocalVariable`.
   - Saves about 4%.
   - Size: S.
5. **Template literals in one allocation**
   - Goal: emit a template literal with n parts as one call that sums the lengths and copies each part once, such as `lstr_concat_n(count, parts)`. The emitter now makes n-1 `lstr_concat` calls, so a template does n-1 allocations and copies its prefix again at each step.
   - Evidence: `lstr_concat` is 11.3% inclusive (`memmove` 3.6%, `malloc` 6.0%), and the freeing of the intermediate strings comes on top. The callers are spread over `positionKey`, `writeLine`, `internedName`, `localName` and `temp`.
   - Saves about 5% after proposals 1, 2 and 4.
   - Size: M (emitter, runtime, seed regeneration).
6. **const-params-pointer** (existing brief)
   - Goal: as in its brief. Use a limit of 16 bytes and include the `retain_S_*` and `release_S_*` helpers.
   - Evidence: see "const-params-pointer" above, 5.0% to 5.8% lower bound.
   - Saves about 5%.
   - Size: M.
7. **Lexer reads bytes**
   - Goal: the lexer works on the source string's bytes (`Str.byteAt`) and slices token text out of the source. Today it decodes the whole file with `Str.chars` in `createLexer`, builds each token's text one char at a time in `readText`, and rebuilds `punctuationChars` for every lexer, template parts included.
   - Evidence: the lex phase is 12.2%, with `readText` 4.9%, `createLexer` 2.4% and `runLexer` self 2.4%.
   - Saves 6% to 8%.
   - Size: M (non-ASCII positions and columns need care).
8. **`finishFacts` without a full split and a second lex**
   - Goal: `findHeaderLine` (`selfhost/src/parser.nra:261`) scans for the header instead of splitting the whole source into lines. `collectReferencedNames` reuses tokens it already has instead of calling `Lexer.lex` on template parts again.
   - Evidence: `finishFacts` 4.9% inclusive, with `narra_str_split` 2.1% and `collectReferencedNames` 1.8%, of which 0.9% is `lexer.lex`.
   - Saves about 3.5%.
   - Size: S.
9. **Emitter string helpers**
   - Goal: `cString` (`selfhost/src/emitstate.nra:198`) appends bytes directly. `findByteChar` now makes a one-char `Str.slice` per byte. `writeLine` appends the indent and the text to `emitter.out` without building `padding + text + "\n"` first.
   - Evidence: `cString` 3.4% inclusive (`findByteChar` 1.6%) and `writeLine` 2.6%.
   - Saves about 3%.
   - Size: S.
10. **Quote names only on the error path in `funcvalues.useLocal`**
    - Goal: `useLocal` (`selfhost/src/funcvalues.nra:73`) calls `quoteGo(local.name)` on every use of a local, but needs the quoted name only when it reports E0601. Quote it inside the error branch.
    - Evidence: `parser.quoteGo` 2.3% inclusive, of which 99% is from `useLocal`.
    - Saves about 2%.
    - Size: S.

Not worth a brief:

- Maps are 1.1% in total.
- Teardown, which releases the `Checker` at the end of `runEmit`, is 3.3%. The leak check needs it.
- `check.findLocal` is 0.7% self. It does a map lookup per scope level, not a linear scan.
- The profile shows no new quadratic step. Template concatenation (proposal 5) is quadratic in the number of parts, but templates are short.

## Deviations

- None. Only files under `roadmaps/selfhost/nodes/profile-compiler/` were added.

## Stubs left

- None.

## Gotchas

- `xctrace` needs full Xcode, and this machine has only the command line tools. Use `/usr/bin/sample`.
- `sample` cannot attach to a 0.6 s process in time. `narrac-loop` (made by `build.sh` with `sed` on `main`) runs the driver `NARRA_REPEAT` times in one process. `NARRA_REPEAT=3` prints the stage 2 C three times, byte for byte.
- Link with `-Wl,-no_deduplicate`, or the linker folds identical functions into `<deduplicated_symbol>` frames.
- At -O2, `lstr_new`, `lstr_retain`, `A_Array_*_get` and many `retain_*` helpers are inlined into the compiler functions that call them. Read reference counting and arrays from the -fno-inline profile as upper bounds.
- Map hashed names with `analyze.py`. The module hash is `xxh32` of the module path after its first `/` (`selfhost/src/hash.nra`), so `f_bc628247_runLexer` is `lexer.runLexer`. A suffix such as `_01c2d782` on `emitStatement_01c2d782` is a generic instance.
- The emitted C takes its runtime from `selfhost/src/runtimecore.nra`, which is generated from `selfhost/runtime/runtime.h`. After editing `runtime.h`, run `sh scripts/test.sh --suite generated --update` (see `selfhost/CLAUDE.md`).

## Proposed learnings

- Do not build a string to use as a map key in a per-node pass. | `positionKey` and `functionKey` cost 8.6% of the compiler run. | `selfhost/src` | no

## Verification

- `sh roadmaps/selfhost/nodes/profile-compiler/build.sh <dir>`: the fixed point holds (184,091 lines). The note "seed/narrac.c is older than the source" is expected on `feat/batch-4`.
- `NARRA_REPEAT=3 narrac-loop emit selfhost`: its last 184,091 lines match `stage2.c`.
- The fast check (`sh scripts/test.sh`) was not run, because no code changed.

## Size

- Estimated M, actual M.
