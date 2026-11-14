# Calibration

Notes:

- selfhost M nodes took 178k-607k tokens in wave 4 (median about 390k).
- Nodes that touch the checker edit `check.go`, `check_call.go`, `modules_check.go`, `cnames.go`, `diagnostic_codes.go` and `compiler/README.md`, whatever the brief says. List them as shared up front.

| Roadmap | Node | Estimate | Actual | Review rounds | Note |
| --- | --- | --- | --- | --- | --- |
| selfhost | split-compiler | M | M | 2 | Round 1 found two changed error messages that no test covered. |
| selfhost | bench-harness | M | M | 2 | Round 1 found an all-skipped row that reported ok. |
| selfhost | modules | L | L | 2 | Round 1 found types compared by name only across modules, and a C name collision through `__`. |
| selfhost | traits | L | L | 2 | Built in parallel with modules, then rebased onto it. Rounds found a constructor lookup that depended on map order, an implementation name collision, and function types that never unified. |
| selfhost | closures | L | L | 2 | Round 1 found two use-after-free bugs, both proven by ASan: a stack environment escaping through a trait, and a callee freed during its own call. |
| selfhost | patterns | L | L+ | 3 | Grew mid-node when the user changed the tuple syntax to `#[...]`. Rounds found tuple C name collisions twice. |
| selfhost | poly-recursion | S | S | 1 | Found by a hang: on main the new error tests time out at 15 s. |
| selfhost | hashed-cnames | M | M | 2 | Touched about 20 files, and the generated C differed in definition order for 4 of 31 samples. Needed a 31-sample before/after comparison. |
| selfhost | box-recursive | M | M | 2 | Heap payloads changed every enum read. A bench claim in review did not reproduce, and `interp` is a real 2x slowdown, accepted until perceus. |
| selfhost | io-std | M | M | 1 | Needed a hotfix (PR #11): merged before rebasing onto box-recursive, and main broke. |
| selfhost | strings-std | M | M | 2 | 331k tokens. Review found `StringBuilder` nameable without its runtime emitted. Fixed a bug that made every `Mutable` builtin unusable. |
| selfhost | maps | M | M | 2 | 468k tokens. Review found the key Equatable/Hashable check skipped in signatures. Touched 17 files. |
| selfhost | effects | M | M | 2 | 490k tokens. Review found private higher-order functions wrongly strict. Six existing samples needed `uses`. |
| selfhost | json-diagnostics | M | M | 2 | 385k tokens. Review found `clone` dropping `StringLit.End`, which broke a fix in generics. Codes by format string kept hot files nearly untouched. |
| selfhost | selfhost-spike | M | M | 2 | 399k tokens. Gate. Review asked for doc fixes. 372 Go lexer lines became 545 Narra lines (1.45x). |
| selfhost | std-convert | M | M | - | 178k tokens. Spike guessed S; nine small functions and a runtime fix. |
| selfhost | opaque-distinct | M | M | - | 340k tokens. Edited 13 files outside its owned paths, a few lines each. |
| selfhost | top-level-const | M | M | 2 | 607k tokens, the most of the wave. Review found a constant/type name clash undetected. Lazy getters replaced static data. |
| selfhost | selfhost-ast | M | M | 2 | 446k tokens. Round 1 asked for escape goldens and position rules in the format comment. Found a Go parser bug (`mutable` after `uses` in parameters), split into fix PR #21. |
| selfhost | selfhost-front | L | L | 1 | 292k tokens. 1501 Narra lines for about 1180 Go (1.27x), under the spike's 1800 because constants and std-convert removed workarounds. Review had nits only. |
| selfhost | selfhost-modules | M | M | 2 | 512k tokens. 745 Narra lines for 569 Go (1.3x). Round 1: no test of the id chain across files; added `narrac pkg --ids`. Reviewer ran about 60 handwritten packages against both loaders. |
| selfhost | std-test | L | L | 2 | 385k tokens. Gate (spec text). Fork-based panic capture clean under ASan and leaks. Review found a hang on a closure that never returns; fixed with alarm. Brief named the module file `std_test.go`, which Go builds only under go test. |
| selfhost | selfhost-check-decls | L | L | 1 | 401k tokens. 1788 Narra lines (plus a 251-line Go dump) for about 1300 Go lines of the decl half. Split from a 3900-line estimate into decls, traits, imports. Review had nits only; reviewer diffed 470 emit/check runs to confirm the Go refactor changed nothing. |
| selfhost | test-runner | M | M+ | 3 | 510k tokens over three fix rounds. Reviews found a hang with no per-test timeout, a zero-match filter passing, an absolute golden path baked into the C, untested leak detection, and an unreachable inner timeout. |
| selfhost | prelude-imports | M | M | 3 | 561k tokens. Gate. Scripted migration of 59 files (90 import lines). Reviews found a rebase conflict on decision numbers, a missing E0908 for functions used as values, and a false E0908 when the alias was imported. |
| selfhost | literal-type-args | M | M | 2 | 311k tokens. Gate. One token-scan lookahead rule shared by both parsers. Review found an error cascade (void after E0237); reviewer showed no valid program changes meaning. |
| selfhost | selfhost-check-traits | M | M+ | 2 | 340k tokens. About 1050 Narra lines against 850 estimated. Review found a quadratic impl lookup in the dump (4.6 s to 0.85 s at 2000 types). Merged automatically. |
| selfhost | std-option-result | M | M | 2 | 375k tokens. Gate. Capturing arrows passed to helpers use a stack environment: 0 heap allocations per call besides the result payload. Review found Err(...) callbacks needing annotations; fixed by taking the result type from the call's expected type. |
| selfhost | selfhost-check-imports | L | M+ | 1 | 369k tokens. About 930 Narra lines against 1000 estimated. Clean review; whole selfhost/ package now compared. Reviewer found a quadratic cost per module on main (inbox). |
| selfhost | selfhost-emit-core | M | M | 1 | 262k tokens. Gate. Contract change request accepted (v2). Clean review. |
| selfhost | selfhost-emit-program | M | M | 1 | 268k tokens. Clean review. Edited emit.nra, body.nra, main.nra and the test helper outside its paths. |
| selfhost | selfhost-emit-decls | M | M | 1 | 196k tokens. Clean review with nits. Output proven only after program merged; no diffs then. |
| selfhost | selfhost-emit-bodies | M | M+ | 2 | 441k tokens. Round 1: functionSignature read an unset body, test lookup bypassed the findBody key. Round 2: a built binary committed at the repo root. |
| selfhost | selfhost-emit-closures | M | M | 1 | 180k tokens. Clean review with nits. |
| selfhost | selfhost-emit-match | M | M | 2 | 210k tokens. Round 1: outcome.md uncommitted after a blocked commit. Edited emitstate.nra and emit.nra outside its paths. |
| selfhost | selfhost-emit-std | M | M | 1 | 183k tokens. Clean review. |
| selfhost | selfhost-emit-options | M | M | 1 | 212k tokens. Clean review with nits. |
| selfhost | selfhost-emit-traits | M | M | 1 | 169k tokens. Clean review; filled traitSelf in the checker. |
| selfhost | selfhost-emit-maps | S | S | 1 | 134k tokens. Self-review; `can-merge` sent it to the maintainer as a PR. |
| selfhost | selfhost-emit | S | S | 1 | 67k tokens. Self-review. The selfhost package compared byte-identical on the first run; no divergence fixes. |
| selfhost | selfhost-bootstrap | M | S | 2 | 124k tokens (worker 67k, fresh review 57k). Fixed point on the first run. Round 2: `time` in a `/bin/sh` script, error wrapping. |
| selfhost | spec-gap-audit | M | M | 1 | 164k tokens. 241 programs, 21 confirmed gaps, 18 proposed fix nodes. |
| selfhost | spec-fixes | S | S | 1 | 60k tokens. Invented a placeholder name (`Int.toI32`) in SPEC.md; left to sized-numbers. |
| selfhost | box-assign | S | S | 1 | 73k tokens. Self-review, no findings. Only the emitter was broken. |
| selfhost | block-arrow-args | S | S | 1 | 60k tokens. Cause was in the lexer (paren depth), not the parser. |
| selfhost | mutable-escape | S | S | 2 | 114k tokens. Round 2: maintainer decided constructors count as stores. |
| selfhost | tail-loops | M | M+ | 2 | 328k tokens (worker 231k, review 96k). Round 1 missed returned match arms and void tail calls; round 2 redid tail detection as a pre-pass. |
| selfhost | for-destructure | M | S | 2 | 204k tokens (worker 126k, review 79k). Parser desugaring; round 2: E0232 list, error positions, cascading error. |
| selfhost | labels | M | M | 1 | 230k tokens (worker 134k, review 96k). Clean review; the new `label` field on four Stmt variants made every open branch conflict. |
| selfhost | let-narrowing | M | M+ | 3 | 237k tokens. Review found a blocking soundness bug (`.value` of None read at run time through hoisted `mutable` calls); orchestrator found a third hole (for-of source). |
| selfhost | for-of-sources | M | M | 2 | 256k tokens. Round 2: Range map keys, nested for-of state, Range typedef without import. |
| selfhost | local-functions | M | dropped | 1 | 221k tokens (worker 146k, review 75k) plus a 149k goal debate. Six defects from parse-time renaming; debate dropped the feature (D74). |
| selfhost | ts-habit-fixes | M | M | 2 | 227k tokens. Round 2: four fixes produced code that still failed to check; TestHabitCases now applies every fix and re-checks. |
| selfhost | nested-function-error | S | S | 1 | 73k tokens. Self-review clean; orchestrator confirmed selfhost parity. |
| selfhost | mutable-read-conflict | S | M+ | 3 | 225k tokens. Found by the let-narrowing review; rounds added match arms and block expressions to the scan. |
| selfhost | block-expressions | M | L | 2 | 413k tokens (worker 298k, review 115k). Round 2: exclusivity walkers skipped BlockExpr, nested break codes, duplicate labels. |
| selfhost | match-block-values | S | S | 1 | 137k tokens. Self-review clean; reused block-expression checking. |
| selfhost | ternary | S | M | 2 | 264k tokens (worker 162k, review 102k). Round 2: `c ? (a + b) : 0` parsed as an arrow (blocking), messages naming the hidden block expression. |
| selfhost | generators | L | L | 2 | 482k tokens (worker 310k, review 173k). Round 2: three emitter sites kept raw C locals across a yield; Go and selfhost emitted the same wrong C, so TestSelfhostEmit could not catch it. |
| selfhost | query-cli | L | L | 3 | 528k tokens (worker 353k, review 175k). Rounds 2 and 3: template positions, function-type effects, one rule for `effects`, stream checks, then contract wording. |
| selfhost | guide-basics | M | L | 4 | 529k tokens. Sonnet draft, two maintainer-direction rewrites (standalone, tutorial voice), then an Opus rewrite and review; the briefs set the wrong audience. |
| selfhost | guide-advanced | M | L | 3 | 514k tokens. Sonnet draft replaced by an Opus rewrite; two Opus review rounds, mostly reteaching chapters 1 to 5. |
| selfhost | place-ids | M | M+ | 2 | 620k tokens. -0.37G. Worker stopped after its report; the orchestrator fixed the review's alias message. 32,538 differential inputs. |
| selfhost | shared-results-walk | S | S | 1 | 185k tokens. 3.8% to 1.3%. 4,391 differential inputs, identical. |
| selfhost | return-trailer | S | M | 3 | 372k tokens. Review found a silent fall-off through a shadowed `panic`, and a narrowing bug on main from the same name-only test. |
| selfhost | copy-check-macros | M | S+ | 2 | 191k tokens. Clean review, two macro-hygiene nits. |
| selfhost | tree-shake | M | M | 2 | 328k tokens. Clean review; rebase round cut peak memory below base. |
| selfhost | tree-shake-types | M | M | 1 | 399k tokens. Clean review, 228 probes. First version +2.5% instructions, incremental trait pass brought it to +1.6%. |
| selfhost | contracts | M | L | 4 | 493k tokens. Three review rounds on soundness (sentinel bounds, shadowing, wrapped bounds). |
| selfhost | contracts-ensures | M | L | 4 | 661k tokens. `!=` added mid-node by the maintainer. No false fact found. |
| selfhost | build-conditions | M | M+ | 2 | 319k tokens. Review: a hanging `uname` hung the compiler and LSP. |
| selfhost | build-targets | M | M+ | 3 | 415k tokens. Review: fmt lost comments around `@target`; `--all-targets` checked debug only. |
| selfhost | bench-check-suite | S | S | 1 | 80k tokens. Reviewed inline by the orchestrator (41 lines). |
| selfhost | attributes | S | S | 1 | 84k tokens. Reviewed inline. |
| selfhost | const-generics | M | M+ | 3 | 459k tokens. Parse regression; maintainer chose TypeScript's syntactic rule mid-node. |
| selfhost | fixed-arrays | M | L | 2 | 576k tokens. Review found a narrowing hole on main, shipped as its own fix commit; 1 MiB cap added. |
| selfhost | unsafe-ptr | M | L | 3 | 469k tokens. Effect escapes through generators and trait implementations closed for unsafe. |
| selfhost | c-interop | M | L | 2 | 462k tokens. Assembler-label prototypes; prefix-helper name clash closed. |
| selfhost | c-audit-target | S | S | 1 | 142k tokens. Reviewed inline; sample renamed. |
| selfhost | test-runner-fixes | M | M | 1 | 253k tokens. The 0-case PASS was a misread shard line; the fix is the case total on the final line. |
| selfhost | process-exit | M | M | 1 | 246k tokens. Clean review. |
| selfhost | host-target-builtin | M | M | 1 | 239k tokens. Clean review; Linux check passed. |
| selfhost | ptr-cast | M | M+ | 2 | 482k tokens. Review: fmt comment fix needed a trivia move in fmtprinter, outside owned paths. |
| selfhost | hidden-copy-fix | M | S | 1 | 159k tokens. Narrowed to loadFile and loadPackage; checker and emitter copies moved to the inbox. |
| selfhost | process-signals | M | L | 2 | 466k tokens. Review: `narrac test` has its own forks (runner, Test.panics) that needed the same forwarding. |
| selfhost | diagnostics-polish | M | M+ | 2 | 395k tokens. Review: per-file cascade check, constant cycles. |
| selfhost | guide-multifile-samples | M | M | 2 | 243k tokens. Review: extractor case for `narra file`. |
| selfhost | fmt-block-hug | M | M+ | 2 | 482k tokens. Operand hug dropped after comparing with oxfmt 0.66.0. |
| selfhost | seed-switch | S | S | 1 | 177k tokens. Two seed updates; -1.3% instructions, close to noise. |
| selfhost | destructuring-design | M | M | 2 | 423k tokens. Gate answered in session; parameters and assignment dropped, named rest built as a copy. Review: tail-loop count was 0, really 15. |
| selfhost | seed-switch-cleanup | S | S | 1 | 118k tokens. Clean review. |
| selfhost | fmt-std-sources | S | S | 2 | 129k tokens. Review: 5 typesdump goldens moved with the std reformat; worker's first PASS line was not from its tree. |
| selfhost | copy-check-holders | S | M | 2 | 232k tokens. Review: one-copy case still exited 103; needed an emitter field outside owned paths. |
| selfhost | checker-emitter-copies | L | L+ | 4 | 613k tokens. 1927 hidden copies to 0; D104 rule took 171 site fixes. Reviews: enum and `is`/`match` holders, then the `copy(...)` initializer exception. |
| selfhost | async-design | S | M | 2 | 221k tokens. D109 answered eleven questions; maintainer changed await to any position. |
| selfhost | arrow-arg-order | S | S | 1 | 72k tokens. Needed a checker flag (writesCaptured) outside owned paths. |
| selfhost | poly-recursion-hang | S | S | 1 | 83k tokens. Depth limit 64 on type arguments, reuses E0310. |
| selfhost | shape-names-private | S | S | 1 | 100k tokens. Lexes Shape-deriving files for hidden names. |
| selfhost | bound-trait-shared-results | S | M | 1 | 128k tokens. Fix exposed 17 unmarked shared results in std and the compiler. |
| selfhost | review-nits | S | S | 1 | 91k tokens. Four of five fixed; arrow hint not reproduced. |
| selfhost | async-core | L | L+ | 2 | 662k tokens. Review: unused-handle holes, map key across await. Built Async.idle early. |
| selfhost | event-loop | M | L | 2 | 448k tokens. Review: deadline overflow, kqueue spin, duplicate fd waits; maintainer accepted a quiescent idle rule. |
| selfhost | async-api | M | L | 3 | 280k tokens. Review: race watcher growth; Async<void> inference; then queued. |
| selfhost | std-async-io | L | L | 2 | 247k tokens. Review: accept errors, read buffer; maintainer made cancel pass down the await chain. |
| selfhost | batch-13 design spikes (8) | S | S-M | 1 | 1.25M tokens in all. Regions dropped by re-measuring; await-anywhere, idle and cancel-chain amended by the maintainer earlier. |
| selfhost | threads-process | S | S | 1 | 92k tokens. |
| selfhost | constructor-field-defaults | S | S | 1 | 81k tokens. One match arm. |
| selfhost | effect-var-nested-result | S | M | 2 | 207k tokens. Added the private-uses rule. |
| selfhost | async-timing-flake | S | S | 1 | 58k tokens. Wider timer gaps; 0 of 400 under load. |
| selfhost | field-default-alloc | S | S | 1 | 92k tokens. |
| selfhost | async-generators | L | L | 1 | 340k tokens. Clean review. |
| selfhost | refinement-types | L | L | 2 | 310k tokens. Review: field defaults, escaping, argument names. |
| selfhost | tagged-templates | M | L | 2 | 373k tokens. Review: void holes gave invalid C. |
| selfhost | parameter-defaults | M | L | 2 | 326k tokens. Review: tail self-calls with left-out defaults. |
| selfhost | for-await | M | L | 1 | 230k tokens. Clean review. |
| selfhost | threads | L | XL | 2 | 460k tokens. Review: share race, recursive share walk, copy check. |
| selfhost | simd | M | L | 2 | 344k tokens. Review: `!v`, unused import. |
| selfhost | uses-list-syntax | M | M | 2 | 242k tokens. Review: old-form hint, duplicates. |
| selfhost | threads-check | M | L | 2 | 263k tokens. Review: Scope escape through generator frames. |
| selfhost | simd-library | M | M | 1 | 222k tokens. Clean review. |
| selfhost | threads-channels | M | L | 2 | 316k tokens. Sync instead of SyncSlot; review: receive with from, void thread functions. |
| selfhost | threads-shared | M | L | 2 | 288k tokens. Review: relaxed load race in Threads.take (190/200 under TSan). |
| selfhost | feature-review | M | M | 1 | 270k tokens. 25 proposals, 7 kept, D119. |
| selfhost | std-design | S | M | 1 | 176k tokens. One gate for seven module APIs; split unicode and regex into two nodes each. |
| selfhost | std-iterator | M | M | 1 | 159k tokens. Clean review; chunk/window narrowed by an effect-variable gap found on the base. |
| selfhost | std-args | M | M | 2 | 192k tokens. Review: flags of nested commands clashed; chose a BadSpec error. |
| selfhost | std-arena | S | S | 1 | 142k tokens. No review beyond the batch review. |
| selfhost | std-toml | M | L | 2 | 207k tokens. 1,460 lines against 900 estimated; review: -0.0 and stringify depth. |
| selfhost | std-terminal | S | M | 2 | 196k tokens. Two rebases on shared lists; review: colon SGR form. |
| selfhost | std-regex | L | L | 2 | 248k tokens. Review: per-pass capture reset, empty passes, slot cap, all checked against node. |
| selfhost | std-regex-const | M | M | 2 | 209k tokens. Review: hints, renamed imports. |
| selfhost | std-unicode | L | L | 2 | 251k tokens. Review: final sigma order, three lookups per code point. |
| selfhost | std-unicode-normalize | M | L | 2 | 215k tokens. Review: quadratic mark ordering; the linear-time test needed three designs to be load proof. |
