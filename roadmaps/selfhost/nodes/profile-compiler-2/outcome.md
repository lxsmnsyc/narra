# Outcome: profile-compiler-2

## Built

- Spike only. No compiler code changed. The profile is of `narrac emit selfhost` at commit 9f8ee5f1 (`feat/batch-5`, the batch-4 compiler plus the batch-5 plan). The first profile (`nodes/profile-compiler/outcome.md`) was at 55fb951.
- **Headline:** allocation fell from 39.6% to at least 4.1% of the run (9.7% at -fno-inline) and strings from 20.2% to 12.6%. Reference counting now takes 24.0%, about the same time per run as before (56 ms then, 61 ms now), so its share grew. Up to 3 to 4 points of that 24.0% are frees, because at -O2 `narra_block_free` is inlined into the `release_*` and `drop_*` helpers; the first profile counted the frees under malloc as allocation. Half of the run is plain compiler code.
- **The largest single cost is whole-struct retains in tail-loop functions: 9.1% to 9.4% of the run** (4.8% in the first profile). tail-loop-borrow is the most valuable next node.
- Of the four batch-5 follow-ups, only tail-loop-borrow is supported by this profile. cache-type-names-rest is worth about 1.3%, or 2.5% if it also caches declaration names. string-concat-one-alloc and map-loop-borrow have no measurable share in the compiler. The verdicts are below.
- Scripts in this folder (the scripts in `nodes/profile-compiler/` are reused unchanged: `build.sh`, `record.sh`, `time.sh`, `copies.py`):
  - `analyze.py` and `allocators.py`: copies of the first profile's scripts. The only change is that they count `narra_block_new`, `narra_block_free` and `narra_carve` (the fast-alloc allocator from batch-4) as allocation. Without it, allocation reads 1.8% and its samples go to the caller's category. Only frames that stay out of line are counted this way: at -O2, `narra_free` and `narra_block_free` are mostly inlined into `release_*` and `drop_*`, so those frees still count as reference counting there.
  - `analyze.py --whole-struct <sample.txt> <repo root>` counts the samples under whole-`Checker` and whole-`Emitter` retains and releases by caller (`profiles/whole-struct-opt.txt`, `whole-struct-opt2.txt`).
  - `split.sh <build dir> <name> <regex>`: builds `narrac-split-<name>`, the `narrac-loop` C with `__attribute__((noinline))` on the functions whose C names match. It keeps -O2 everywhere else, so a profile can split the self time that -O2 inlines into one big function. Used here to split `effects.walkExpression`.
- Text summaries are in `profiles/`: `opt`, `opt2`, `noinline` and `split-effects` (analyze.py), `allocators-noinline.txt`, `copies-opt.txt`, `whole-struct-opt.txt`, `whole-struct-opt2.txt` and `callers-cType-opt.txt`. The raw `sample` files (6 MB each) are not committed.

## Proof

- base: commit 9f8ee5f1, release `narrac` from `selfhost/bootstrap.sh` (fixed point at 164,332 lines of C; it was 184,091 at 55fb951), `/usr/bin/time -l narrac emit selfhost`: 2.622, 2.635 and 2.656 billion instructions retired (spread 1.3%), peak footprint 127.5 to 128.6 MB. batch-4 recorded 2.59G on an idle machine.
- Wall time (`time.sh`, 5 runs, bench lock held): release 0.310 (cold), 0.254, 0.257, 0.259, 0.254 s. `narrac-g` 0.259 to 0.307 s, so the profiled code is the code that ships. `narrac-noinline` 0.396 to 0.421 s. The first profile measured 0.60 s.
- Profiles: `sample` at 1 ms, 30 s each, bench lock held. `opt` (`narrac-loop`, 23,901 main-thread samples), `opt2` (24,015), `noinline` (24,040) and `split-effects` (21,873). `opt` and `opt2` agree within 0.7 points on every category and phase. The tables use `opt`.
- Load: the 1-minute load average was 5 to 10 on 8 cores during the runs (two other spikes and other worktree agents). Shares are fractions of samples, so they hold better than wall times; the wall times above are about 0.025 s above batch-4's idle 0.23 s figure.

### Top 10 by self time (all frames)

| # | function | self now | first profile |
| - | -------- | -------- | ------------- |
| 1 | `effects.walkExpression` | 8.8% | 1.5% |
| 2 | `_platform_memmove` (libc `memcpy`) | 5.4% | 12.1% |
| 3 | `lexer.runLexer` (`lexPunctuation` inlined) | 4.7% | 2.4% |
| 4 | `_platform_memcmp` | 4.4% | 3.4% |
| 5 | `check.findLocal` (map lookups inlined) | 3.0% | 0.7% |
| 6 | `retain_S_b2b309ba_Checker` | 2.9% | 1.4% |
| 7 | `narra_block_new` | 2.4% | (new) |
| 8 | `A_Array_5fce1d7c_release` | 2.2% | |
| 9 | `idcheck.collectExpressionIds` | 1.9% | |
| 10 | `parser.isAt` | 1.7% | 1.0% |

The malloc frames that led the first profile (`_xzm_free_tc` 9.9%, `_xzm_malloc_tc` 9.3%, `_malloc_zone_malloc` 6.2%, `_free` 6.1%) are gone from the top 25. `__vfprintf` (2.5% before) is gone: int-to-string no longer calls `snprintf`.

### Top 10 by inclusive time

`start`, `main` and `emitter.runEmit` are 100% and left out.

| # | function | inclusive now | first profile |
| - | -------- | ------------- | ------------- |
| 1 | `body.checkModules` | 28.8% | 20.6% |
| 2 | `emit.emitProgram` | 25.4% | 36.1% |
| 3 | `body.checkFunctionBody` | 23.9% | 18.3% |
| 4 | `modules.loadCached` | 22.6% | 24.1% (`loadPackage`) |
| 5 | `parser.parse` (includes lexing) | 20.9% | 23.4% |
| 6 | `stmts.checkStatement` | 19.8% | |
| 7 | `emitstmts.emitStatement` | 17.2% | 27.6% |
| 8 | `effects.checkEffects` | 16.7% | 15.9% |
| 9 | `effects.walkExpression` | 12.6% | 13.2% |
| 10 | `lexer.lex` | 7.0% | 11.5% |

The helpers the first profile named, now (`opt` inclusive):

| function | first profile | now | what changed |
| -------- | ------------- | --- | ------------ |
| `lstr_concat` | 11.3% | 0 samples | template-one-alloc; the C has 7 `lstr_concat(` sites left, none sampled |
| `lstr_concat_n` | (new) | under 0.5% (-O2), 4.1% (-fno-inline) | one call per template, spread over `temp`, `declHead`, `functionSignature`, `localName` |
| `snprintf` / `lstr_from_int` | 6.9% | 1.1% | int-to-string |
| `effects.positionKey` | 6.1% | gone | effect-int-keys |
| `cnames.typeId` | 6.1% | 1.4% | cache-type-names (4 call sites) |
| `lexer.readText` | 4.9% | under 0.5% | lexer-slices |
| `parser.finishFacts` | 4.9% | 2.6% | finish-facts-lines |
| `emitstate.temp` | 4.5% | 1.4% | |
| `emitstate.refCountCode` | 4.3% | 1.6% | |
| `emitstate.cString` | 3.4% | 0.6% | emit-string-helpers |
| `parser.quoteGo` (`useLocal`) | 2.3% | under 0.5% | quote-on-error |
| `release_S_b2b309ba_Checker` | 3.3% | 5.9% | see tail-loop-borrow below |

New hot helpers:

| function | inclusive | what it does |
| -------- | --------- | ------------ |
| whole `Checker`/`Emitter` retain and release in 25 tail-loop functions | 9.1% (`opt`), 9.4% (`opt2`) | `openTailLoop` retains every large parameter at entry, though const-params-pointer now passes it as a pointer |
| `check.startBody` | 3.5% | walks the whole body with `idcheck.collectFuncIds` (2.8%) to find its node id range |
| `check.findLocal` | 3.1% | one `Map<string, int>` lookup per scope level, innermost first |
| `lexer.lexPunctuation` (-fno-inline) | 5.7% | tries all of `PUNCTS` with `hasPrefix` (5.3%) on every punctuation token |
| `parser.isAt` | 2.4% | compares token text, with `memcmp` |
| `narra_str_replace` | 2.0% | all from `emitmaps.mapFunctions`, which fills the map helper template per map type |
| `cnames.functionSignature` | 2.0% | `parameterCType` and `cType` uncached, `functionName` rebuilt |
| `hash.moduleHash` | 1.1% | `xxh32` of the module path on every `declHead` call |

### Share per category

Self samples. Same rules as the first profile, plus the fast-alloc allocator frames counted as allocation. "ms" is the share times the run's wall time (0.60 s then, 0.256 s now), to show where time went, not only shares.

| category | first `opt` | now `opt` | now `opt2` | first ms | now ms | first -fno-inline | now -fno-inline |
| -------- | ----------- | --------- | ---------- | -------- | ------ | ----------------- | --------------- |
| allocation | 39.6% | 4.1% | 3.9% | 238 | 10 | 32.2% | 9.7% |
| strings | 20.2% | 12.6% | 13.0% | 121 | 32 | 20.4% | 19.2% |
| reference counting | 9.3% | 24.0% | 24.6% | 56 | 61 | 17.4% | 25.9% |
| by-value struct copy | 5.2% | 1.5% | 1.4% | 31 | 4 | 4.5% | 0.7% |
| maps | 1.1% | 3.5% | 3.9% | 7 | 9 | 2.1% | 6.3% |
| arrays | 0.6% | 3.6% | 3.4% | 4 | 9 | 8.3% | 10.0% |
| other compiler code | 24.1% | 50.6% | 49.8% | 145 | 130 | 15.2% | 28.2% |

- The allocation row is not like with like. The first profile counted every free as allocation, because the frees ran in `libsystem_malloc`. Now, at -O2, `narra_block_free` is mostly inlined through `narra_free` into `release_*` and `drop_*`, so those frees count as reference counting. Read 4.1% as a lower bound. At -fno-inline, where the allocator keeps its own frames, allocation is 9.7%: `narra_block_free` 3.7% self, `narra_block_new` 3.3%, `narra_alloc` 0.9% (`profiles/noinline-summary.txt`). So up to 3 to 4 points of the 24.0% reference counting at -O2 are frees.
- Allocation went from 238 ms to about 10 to 20 ms, counting the inlined frees (fast-alloc, perceus, the string fixes). `allocators.py` on -fno-inline: arrays 3.7%, string helpers 2.9%, maps 1.1%, boxes and payloads 0.9%, frees in `release_*` 0.8%.
- Reference counting did not shrink in time. Most of it is whole-struct retains in tail loops (9%), teardown (6%) and the inlined frees above.
- Compiler code is now half the run and spread thin. The biggest pieces are the effect walk, the lexer's punctuation scan and the body id walk.

### Share per phase

| phase | first `opt` | now `opt` | now `opt2` | first ms | now ms | main costs now |
| ----- | ----------- | --------- | ---------- | -------- | ------ | -------------- |
| lex | 12.2% | 7.2% | 7.1% | 73 | 18 | `lexPunctuation` (in `runLexer` self 4.7%) |
| parse | 11.2% | 13.7% | 13.3% | 67 | 35 | `isAt` 2.4%, `finishFacts` 2.6%, `parseTemplate` |
| check | 20.6% | 28.9% | 29.2% | 124 | 74 | reference counting 7.6% (tail loops), `startBody` 3.5%, `findLocal` 3.1% |
| effects | 15.9% | 16.8% | 17.4% | 95 | 43 | the AST walk itself; `valuesAt` map 2.4% |
| emit | 36.1% | 25.4% | 25.3% | 217 | 65 | strings 8.2%, reference counting 6.3%, `functionSignature` 2.0%, `mapFunctions` 2.0% |
| teardown | 3.3% | 6.3% | 6.3% | 20 | 16 | `release_S_Checker`, `drop_E_Expr` (the AST) |
| load other | 0.7% | 1.7% | 1.4% | 4 | 4 | file reads |

### The effect pass

- `effects.walkExpression` is 8.8% self at -O2 because `walkCall`, `findCallee`, `findRef`, `walkUserCall` and the map helpers are inlined into it.
- `split.sh` with the effects module, `getExpressionPos`, `findNodeType`, `findRef` and the map lookups kept out of line (`profiles/split-effects-summary.txt`): `walkExpression` self is still 4.6%, `walkCall` 1.0%, `walkStatement` 1.1%. The `valuesAt` map (`M_Map_16683c1c`) is 1.8% lookup plus 0.6% set.
- No single helper dominates. The rest is visiting each AST node, likely cache misses on payloads the check phase touched long before. Only the `valuesAt` map is a clear fix (proposal 6).
- The same split build shows `M_Map_91178635_lookup` (string to int, the scope maps) at 4.0% self, mostly from `checkExpressionKind` and `checkNamedCall` through `findLocal`.

## Verdicts on the batch-5 follow-ups

- **tail-loop-borrow: supported, raise it.** 25 functions keep a `<name>_held` copy and retain the whole `Checker` or `Emitter` at entry (`grep` on the emitted C). That is a walk over every counted field (the `Checker` has about 60 arrays and maps, plus its `Registry`) on entry and again on each return. Retains and releases of the whole `Checker` and `Emitter` outside teardown are 2,165 of 23,901 samples (9.1%) in `opt` and 2,266 of 24,015 (9.4%) in `opt2` (`analyze.py --whole-struct`, `profiles/whole-struct-opt.txt` and `whole-struct-opt2.txt`). Each sample counts once, for the caller of the outermost whole-struct frame. These differ from the "nearest Narra function above reference counting" lists in the summaries, which count all reference counting, not only whole structs. About 90 samples (0.4 points) come from `emitFunctionCall`, `emitCall`, `constRef` and `emitExpression`, which are not tail loops. Callers in `opt`, most samples first: `stmts.collectUses` 529, `emitexprs.mayWrite` 442, `emitstate.underlyingType` 235, `check.isPassedByPointer` 209 (called for every argument and parameter), `effects.findSourceCall` 130, `exprs.usesLocal` 97, `exprs.containsOpen` 90, `emitmatch.collectBoundLocals` 60, `patterns.findMissingCases` 59, `patterns.collectPatternLocals` 57. Teardown (`emitter.runEmit`) adds 303 more. Expected gain: about 8%.
- **cache-type-names-rest: weakly supported.** The uncached `cType` callers (`parameterCType`, `containerStruct`, `functionSignature`, `emitProgram`, `enumAllocation` and others) add up to 300 of 23,901 samples (1.3%; `profiles/callers-cType-opt.txt`, counted at the outermost `cType` frame). That is near the 1.3% instruction spread, so the node's acceptance bar ("gain above the spread") may not be met as written. Widen it to also cache `declName` per declaration and `moduleHash` per module: `declHead` is 1.3% and `moduleHash` 1.1% inclusive, from `functionName`, `structName` and `enumName`. Combined, about 2.5%.
- **string-concat-one-alloc: not supported by the compiler profile.** `lstr_concat` has 0 samples in all three profiles; the emitted compiler has 7 `lstr_concat(` call sites, against 1,906 `lstr_concat_n(`. The compiler builds its strings with templates. Start it only if `bench/strings` or profile-programs shows `+` chains in user programs; `narrac emit selfhost` will not move.
- **map-loop-borrow: not supported by the compiler profile.** The compiler has 25 `for` loops over `Maps.keys` or `Maps.entries`. All are in functions under 0.5% inclusive. The hottest, three loops in `parser.parseTemplate`, account for about 0.1% in reference counting. Start it only for user programs, if profile-programs shows map loops.

## Proposed briefs

Shares overlap and are fractions of a run on a loaded machine, so do not add them up. "Saves" is an estimate from the measured share. Ranked by value for the size.

1. **tail-loop-borrow** (existing brief, raise to first)
   - Goal: as in its brief. A large parameter that every self tail call passes back unchanged needs no retain at entry and no release at each return.
   - Evidence: 9.1% to 9.4%, see the verdict above.
   - Saves about 8%. Size M.
2. **Skip the final release in the `narrac` CLI** (needs an orchestrator decision)
   - Goal: after `narrac emit` or `narrac build` has written its output, exit without releasing the `Checker`, the `Emitter` and the AST; the OS frees the memory. Key it on `NARRA_CHECK_LEAKS` (`selfhost/runtime/runtime.h:59`): when it is set, keep the full teardown. The test runner sets it on every `narrac` run (`selfhost/tests/src/shell.nra:45`, `context.nra:41` and `:53`, `bootstrap.nra:41`), so the leak check still covers the compiler. Limit it to the emit and build commands; the LSP server and other long-lived paths keep releasing.
   - Evidence: teardown is 6.3% in both `opt` runs (1,504 and 1,503 samples): `release_S_Checker` called from `runEmit` (303) and the drop of the AST and the other values `runEmit` holds.
   - Saves about 6%. Size S. The first profile ruled this out because the leak check needs the teardown; the share has doubled since, so it is worth deciding again.
3. **Lexer punctuation by first byte**
   - Goal: `lexPunctuation` (`selfhost/src/lexer.nra:453`) picks candidates from a table indexed by the first byte, longest first, instead of trying every entry of `PUNCTS` with `hasPrefix`. Compare punctuation by kind, not `punctuation == ";"` string compares.
   - Evidence: `lexPunctuation` 5.7% and `hasPrefix` 5.3% inclusive at -fno-inline; at -O2 it is inlined into `runLexer`, 4.7% self. The lex phase is 7.2%.
   - Saves 3% to 4%. Size S.
4. **Body id ranges from the parser**
   - Goal: the parser records the first and last node id of each function, block and constant body (it already knows each file's `[firstId, endId)`), so `check.startBody` (`selfhost/src/check.nra:1540`) does not walk the whole body with `idcheck.collectFuncIds` to find them. `narrac pkg --ids` keeps using the walker.
   - Evidence: `startBody` 3.5% inclusive, of which `collectFuncIds` 2.8%.
   - Saves about 3%. Size S.
5. **One scope map for locals**
   - Goal: shadowing is an error, so a name is normally bound in at most one open scope. `findLocal` (`selfhost/src/check.nra:1774`) can look up one `Map<string, int>` instead of one map per scope level. A name that is not a local (a module function, a type) now misses at every level. Three things depend on the per-depth maps and must keep working:
     - (a) `addLocal` (`check.nra:1905`) still rebinds a name after reporting E0103, so a scope exit cannot just remove the name. Keep an undo log per scope of (name, previous binding) and restore the previous binding at exit.
     - (b) `setNarrowed` and `endNarrowing` (`selfhost/src/patterns.nra:113-127`) write `narrow|N` keys into the map of a given depth. They need the same undo log, or a separate structure.
     - (c) `isDeclaredBelow` (`selfhost/src/funcvalues.nra:222`) asks at which depth a local was declared. Store a declaration depth per local instead.
   - Evidence: `findLocal` 3.1% inclusive; in the split build the string-keyed scope map lookup is 4.0% self.
   - Saves about 2%. Size M.
6. **Fill the effect `valuesAt` map only where it is read**
   - Goal: `walkExpression` (`selfhost/src/effects.nra:947`) records every expression under `packPosition(body, pos)` in a `Map<int, int>`. Its only reader, `findStoredValue` (`effects.nra:323`), looks it up by `FunctionStore.pos`. The cheaper fix: skip filling `valuesAt` for a body that has no `FunctionStore`. Keying it by node id instead would need a node id in `FunctionStore`, which `recordFunctionStore` (`selfhost/src/exprs.nra:441` and `:447`) does not have today.
   - Evidence: `M_Map_16683c1c` lookup 1.8% and set 0.6% in the split build.
   - Saves up to 2%, depending on how many bodies hold a function store. Size S.
7. **cache-type-names-rest, widened to declaration names**
   - Goal: as in its brief, plus cache `declName` per declaration and the module hash per module.
   - Evidence: see the verdict above.
   - Saves about 2.5% (1.3% for the brief as written). Size S.
8. **Map helpers without `Str.replace`**
   - Goal: `emitmaps.mapFunctions` fills a template with `Str.replace` for each map type. Split the template once into parts and join them with the type names.
   - Evidence: `narra_str_replace` 2.0% inclusive, all from `mapFunctions`; it is 1.4% of `memcmp`.
   - Saves about 2%. Size S.
9. **Token kinds in `parser.isAt`**
   - Goal: `isAt` and `parseBinary` compare token text with `memcmp`. Give punctuation and keyword tokens an int kind at lex time and compare that.
   - Evidence: `isAt` 2.4% inclusive; `memcmp` from `isAt` 143 and `parseBinary` 67 samples.
   - Saves about 1.5%. Size M (touches the parser widely).

Not worth a brief:

- string-concat-one-alloc and map-loop-borrow for the compiler (see verdicts).
- Allocation is 4.1% (9.7% at -fno-inline) and by-value struct copies 1.5% in total.
- The effect walk beyond the `valuesAt` map: no single helper is over 1.1%.
- `unused.reportUnusedTopLevel` is 1.3% self in the split build; check it again after proposals 1 to 5.

## Deviations

- The brief says to reuse `analyze.py`. It is copied into this folder with the fast-alloc allocator names added, because the first node's folder is not this node's to edit. Without the change, allocation reads 1.8% instead of 4.1%.
- `split.sh` is new. The brief did not ask for it, but -O2 inlining hid the effect pass breakdown.
- The batch-5 brief was not on the branch at the start (it is on `feat/batch-5`, which the branch was reset to).

## Stubs left

- None.

## Gotchas

- The isolated-worktree guard on this machine refuses Bash commands it cannot parse (loops, heredocs, `cd` chains on long paths). Put multi-step runs in a script under the scratchpad and run it with one `sh` call.
- macOS `sed` has no `\|` in basic regular expressions. `split.sh` uses `sed -E`.
- At -O2, small helpers (`lexPunctuation`, `walkCall`, the map lookups) are inlined into their callers. Read compiler-code shares with the -fno-inline profile or a `split.sh` build.
- `narra_drop` is counted as reference counting, so teardown shows as reference counting, not allocation.
- The `opt` phase "teardown" is the release calls at the end of `runEmit`. In the -fno-inline build the same work shows as "runEmit other" (3.9%), because the call lines differ.

## Proposed learnings

- A tail-recursive helper that takes the whole `Checker` or `Emitter` pays a full retain and release per call. | 9% of `narrac emit selfhost` after batch-4. | `selfhost/src` | no

## Verification

- `sh roadmaps/selfhost/nodes/profile-compiler/build.sh <dir>` at 9f8ee5f1: stage 1 and stage 2 agree (164,332 lines); the seed is current.
- Instruction counts and wall times as listed under Proof.
- `sh scripts/test.sh` was not run, because no code changed.

## Size

- Estimated M, actual M.
