---
goal: Find the SPEC features to remove or simplify, and the features to adopt, before Narra is released
date: 2026-11-14
rounds: 1 of 1
status: proposed        # approved once the maintainer signs off
roadmap: roadmaps/selfhost
supersedes: none
decision: none yet (D119 is reserved; the maintainer decides first)
---

# Debate: what to remove, simplify or adopt before release

Node feature-review (phase p3), on `feat/batch-13` at 6c9c740e. Short debate: one round. The judge wrote the facts and the 25 proposals. A fresh opponent agent, which saw only the ledger and the repo, attacked each one with 40 claims. The judge ruled on every claim. The ledger is `ledger.json`. The counts are in `evidence/`.

How the counts were taken. `evidence/count-code.sh` drops comment lines and replaces string and template literals before matching, so diagnostic texts that name a feature do not count. `evidence/usage-counts.txt` counts raw matches, including those texts and the generated `runtime*.nra` and `stdtext.nra`. Four opponent claims rested on it and were rejected for that reason (J16). Counts are one-line greps, so they are lower bounds for forms that span lines. No build or test ran: the load average was 30 to 41 (J15).

The corpus (J1): `selfhost/src` 171 files, 81,580 lines (73,311 code lines); `std/src` 21 files, 1,772 lines; `testdata` 1,747 `.nra` files.

Rules every proposal was held to (J13): one call form and no overloading (D12), no hidden or implicit calls, no early-return operator (SPEC 5), no deprecation paths (GOAL.md), and TypeScript's meaning where the syntax is the same.

## Settled structure

Eight proposals stand, seven of the 25 and one opponent alternative: three removals, one removal after a measurement, one SPEC simplification and three adoptions. Each one would become its own node, or a part of D119.

### 1. Remove `Coroutine<In, Out>` (P1)

- **Kind:** remove.
- **Evidence.** No code in `selfhost/src` or `std/src` uses a coroutine; `std/coroutine` only declares `resumeCoroutine` (J2). `Co.resume` is two functions chosen by argument count (`calls.nra:715`, `scopes.nra:428`), the one overload in a language whose SPEC says it has none (J3). Every generator frame carries an `in` field for coroutine input (`runtime.h:416`). 16 testdata files mention `Coroutine`; 4 samples are about it (`coroutine*.nra`).
- **What stays.** `Generator<T>`, `Co.resume(mutable g)`, `Step` with `Panicked`. Async generators (D110) now cover streams that wait.
- **Cost.** About 100 compiler lines out (estimate: `Coroutine` branches in 12 files, the input slot in `emitgenerators.nra`), E0197 and the three yield-input errors in `stmts.nra:1078-1364`. The runtime field goes, so the batch needs two seed updates (runtime text rule). Edits in `types.nra:74`, `check.nra:1136`, `declsdump.nra:448` and the generated `stdtext.nra` (O1). The compiler uses no coroutine, so no source sweep.
- **Value.** D12 has no exception left. One generator type to learn. A smaller generator frame.
- **Prior art.** TypeScript sends values with `next(v)` on its one `Generator<T, TReturn, TNext>`; Python has `send`. Go, Kotlin sequences, Swift and Gleam have no two-way generator. D79 records that the maintainer asked for coroutines on 2026-11-07, so this reverses a maintainer request.
- **Recommendation:** remove, in one node with the runtime change.

### 2. Remove `lazy` top-level values (P2)

- **Kind:** remove.
- **Evidence.** No use in `selfhost/src` or `std/src` (J2). The one raw hit is a string (J16). 7 `lazy_values*` samples and 32 testdata lines. `@const` (D105) computes constant tables at compile time. It is used once in std (`Template.raw`).
- **Cost.** About 260 compiler lines out: `lazyvalues.nra` 106, `stdlazy.nra` 82, 37 lines in `effects.nra`, about 30 in the parser, emitter and fmt (J4). Codes E1032 to E1035 retire. No seed step beyond the batch's usual one. User-visible: `const x = lazy f()` and `std/lazy` go. A table that needs a call to a function `@const` cannot run becomes a function the program calls and passes on (O3).
- **Value.** One fewer top-level form, with its four rules and `Lazies.force`. "Modules have no initialization code" holds without an exception.
- **Prior art.** TypeScript has none. Rust (`LazyLock`) and Go (`sync.OnceValue`) put it in the library, which Narra cannot do without globals. Kotlin `by lazy` and Swift lazy globals are language forms.
- **Recommendation:** remove.

### 3. Remove bare block expressions and `if` as a value (P4)

- **Kind:** remove.
- **Evidence.** 0 uses of either in `selfhost/src` and `std/src`; 8 testdata lines use `if` as a value and 17 a bare block (J2). `parseIfExpr` is `parser.nra:3904-3935`. The bare block is one branch at `parser.nra:4107-4111` (J12). Match arms parse their own block (`parser.nra:4459`), and `?:` and `async { }` build `BlockExpr` themselves, so the AST kind and its checker and emitter paths stay (J16; O7 rejected).
- **Cost.** About 40 parser lines and the matching fmt cases out (O8). E0933 changes: `{ a }` in value position is no longer a block. Two samples (`block_expressions`, the if-value lines) and their goldens. No seed step.
- **Value.** TypeScript has neither form. In TypeScript `{` in value position starts an object literal, so after this change `{` there always starts a struct literal. That removes the trap D75 records (`const o = { a }` is an error with a pointer to struct literals) and makes room for proposal 4.
- **Recommendation:** remove. Keep `?:`, value `match` with block arms, and `async { }`.

### 4. Adopt field shorthand in struct literals (P5)

- **Kind:** adopt.
- **Evidence.** 512 fields or named arguments written `name: name` in `selfhost/src`, 6 in `std/src`, 39 in testdata (J6). Destructuring patterns already write `{ name }`.
- **Rule.** `Point { x, y }` and `{ x, y: 0 }` mean `x: x`, as in TypeScript. Context-typed `{ x }` alone needs proposal 3, since today it is a block (O10). Named arguments keep `name: value`.
- **Cost.** About 50 compiler lines (parser field without `:`, fmt, a fix text). It reverses the "no literal shorthand" item of D26, whose reason was "a second way to write a field" (O9). The compiler can use it only after the seed update that ships it. The 512 sites need no sweep.
- **Value.** TypeScript code reads as written. Literals match patterns. No call or conversion is involved, so D12 and the no-hidden-calls rule hold.
- **Prior art.** TypeScript and ES2015 shorthand properties; Rust field init shorthand (RFC 1682). Go, Swift and Kotlin have none.
- **Question for the maintainer.** Should `narrac fmt` rewrite `x: x` to `x`? That keeps one form in formatted code, which answers D26's reason, but oxfmt does not rewrite it. Recommended: fmt keeps what is written, as oxfmt does.
- **Recommendation:** adopt.

### 5. Adopt integer values in template holes (P6)

- **Kind:** adopt.
- **Evidence.** 359 of 4,009 holes in `selfhost/src` are `${Int.toString(...)}`; 536 of 1,487 in testdata; 7 of 20 in `std/src` (J7).
- **Rule.** A hole takes a `string` or a value of any integer type. The runtime writes an integer in decimal, the text `Int.toString` gives. No user code runs, and nothing else changes: floats, `bool`, `char`, distinct types and enums still need a call. The judge narrowed P6 to integers: the other types had no count, and floats need a formatting rule.
- **Cost.** About 40 compiler lines (the hole type check, one emitter case, the error text). The seed must ship it before the compiler uses it.
- **Value.** TypeScript's `${n}` means the same thing. The most common conversion in the compiler goes away.
- **Risk, open (O11).** SPEC 2 says there are no implicit conversions, and D114 rejected `Shape.toValue` per hole as an implicit call. Here no user function runs, and Lua, the maintainer's reference for explicit calls, converts numbers in `..`. Gleam and Zig keep the call explicit. Whether a built-in integer format in a hole is an implicit conversion is the maintainer's call.
- **Prior art.** TypeScript, Kotlin (`"$n"`), Swift (`"\(n)"`), Lua (`..`); Rust's `format!` goes through `Display`, a trait call; Gleam and Zig require the call.
- **Recommendation:** adopt for integer types, if the maintainer rules that it is not a hidden conversion.

### 6. Make SPEC 11 and 12 describe what is built (P7)

- **Kind:** simplify.
- **Evidence.** SPEC promises layout tables in debug builds, `@specialize`, `generics = "compact"`, `.nrai` interface files, TCC debug builds and spread-out large frees. `compiler/README.md:250` says every build monomorphizes and layout tables are not built; `README.md:224` and `:253` say there are no `.nrai` files. No code in `selfhost/src` handles `@specialize` or `compact`. Node layout-tables is todo, could (J5).
- **Cost.** SPEC text only (sections 6, 11, 12 and the `narra.toml` sample in 15), D7 and D13 amended, node layout-tables dropped after its brief is read (O40). No code and no seed impact.
- **Value.** The SPEC stops promising a second code generation path that no node builds. Keep the removed text in a parked-ideas note, since incremental builds may want interface files later (O13).
- **Prior art.** Rust, C++ and Zig always monomorphize; Swift specializes with a witness-table fallback; Go uses GC-shape stenciling.
- **Recommendation:** adopt.

### 7. Measure `@inline`, then drop it if it buys nothing (P8)

- **Kind:** remove, after a measurement.
- **Evidence.** No `@inline` in `selfhost/src`; 3 in `std/src`, all in `std/option`; 36 testdata lines (J2, J16). The feature is `inlining.nra` 371 and `emitinline.nra` 458 lines, codes E1030 and E1031 (J4): the largest saving on this list. The program is one C file (D54), and a non-escaping closure already keeps its environment on the stack, so clang `-O2` may inline the same calls.
- **Cost of the measurement.** One small spike: a benchmark of `Options.map`, `Options.andThen` and an `@inline` loop helper, with and without the attribute, at `-O2`, on a quiet machine (bench runs are exclusive). A result inside the noise drops the feature: about 830 lines and 2 codes out, 3 attributes off `std/option`, no seed step.
- **Value.** `@inline` is a guarantee (SPEC 2), not a hint. The spike tells whether the guarantee buys speed anyone sees.
- **Prior art.** Zig `inline fn` and Kotlin `inline fun` guarantee it; Rust `#[inline]` and Swift `@inline(__always)` are hints; C compilers inline static functions at `-O2`.
- **Recommendation:** run the spike, then decide.

### 8. A descending range that needs no `Option` (O23)

- **Kind:** adopt (std).
- **Evidence.** Counting down is the one thing a C-style `for` does that `Ranges` does not do short: `Ranges.stepped` returns an `Option` because a step can be 0 (SPEC 2). The compiler writes it as `for (let index = n - 1; index >= 0; index--)` (`lspnavigation.nra:503`, `inlining.nra:189`).
- **Cost.** One Narra function in `std/range`, about 20 lines. No syntax.
- **Value.** Counting loops in both directions have a Range form. C-style `for` stays (P12 rejected).
- **Recommendation:** adopt as a std node of batch-14 or later.

## Budget

- device: no build or test during the debate. The `@inline` spike needs a quiet machine (J15).
- time: proposals 1 to 3 and 6 are removals with no compiler use, so they need no source sweep. Proposals 4 and 5 can be used in `selfhost/src` only after the seed update that ships them. Proposal 1 changes runtime text, so its batch needs two seed updates (O1, O2).
- tokens: one opponent agent (about 54,000 tokens) and the judge.
- maintenance: about 400 compiler lines out for proposals 1 to 3, about 830 more if proposal 7 drops `@inline`, about 90 in for proposals 4 and 5 (estimates from J4 and J12).
- user: three forms and one type go (`lazy`, bare blocks with `if` values, `Coroutine`). Two TypeScript forms arrive (field shorthand, integer holes).

## Rejected

- P3, remove `typeof` in types: TypeScript form with the same meaning, already built. Saves about 250 lines and no user rule (O5, O6).
- P9, remove `Dynamic<Trait>`: built on 2026-11-12, the only dispatch another module can extend. Removing it is a product call with no measured cost to keep (O16, O17).
- P10, remove `ensures`: only about 121 lines are `ensures`-only. The prover stays for refinement types (O19).
- P11, remove `x++` and `x--`: 405 uses in the compiler against 3 `+= 1`, and the TypeScript form (J8, O20).
- P12, remove C-style `for`: TypeScript form, 45 compiler uses, FixedArray proofs read its condition, and counting down has no Range form yet (O22, O23).
- P13, remove labeled `break` and `continue`: TypeScript form, no measured saving (O24, O25).
- P14, remove match guards: four parser lines (`parser.nra:4453-4457`), Rust and Swift form, no measured saving.
- P15, remove type aliases: the re-export path of SPEC 3 and the only way to name a function type (O27).
- P16, infer `escaping`: a signature would depend on its body (O28).
- P17, adopt let-else (Rust) or guard (Swift): a second spelling of `if (x is None) { return }` (O30).
- P18, adopt `?` (Rust) or `try` (Zig): the early-return operator SPEC 5 forbids (O31).
- P19, adopt a pipe `|>`: a second call form against D12, and TypeScript has not shipped it (O33).
- P20, adopt `items.length`: D12, and on a string it would mean bytes where TypeScript means UTF-16 units (O34).
- P21, adopt range literals `a..b`: D23 and D45 chose names; 416 calls to rewrite for shorter text (O35).
- P22, adopt `defer`: a hidden call at the end of a scope. Reference counting closes resources (O36).
- P23, adopt optional chaining `a?.b`: a hidden branch, and `Option` is an enum, not a nullable (O37).
- P24, adopt trait default implementations: no demand (6 traits and 4 implement blocks in the compiler), and TypeScript interfaces have none (O38).
- P25, adopt `x !is P` (Kotlin): a second spelling for 19 sites (O39).
- O12, leave holes explicit and rely on editor snippets: a snippet shortens typing, not the 359 holes a reader reads.
- O32, cover early exits with std helpers: the helpers exist and are used twice.

## Scope changes

- drop layout-tables: SPEC will no longer promise layout tables (proposal 6, J5).

## Proofs

- None. The `@inline` question needs a quiet machine. It is a spike (proposal 7), not a proof inside the debate.

## Prior art

- TypeScript: shorthand properties, `${n}` on numbers, labels, C-style `for`, `x++`, `typeof` types. No block expressions, no if-expressions, no pipe.
- Rust: field init shorthand (RFC 1682), `LazyLock`, `#[inline]` hints, always monomorphizes, `?` (rejected here by SPEC 5).
- Go: `sync.OnceValue`, GC-shape stenciling, `defer`.
- Swift, Kotlin, Zig, Gleam, Lua: cited per proposal above.
- This repo: D26 rejected literal shorthand; D75 chose block expressions; D79 added coroutines at the maintainer's request; D96 added `typeof`; D114 rejected an implicit `Shape.toValue` per hole.

## Open risks

- O11, whether integer holes are an implicit conversion: the maintainer decides before proposal 5 becomes a node.
- Proposal 7: the spike decides. Until then `@inline` stays.

## Inferred choices

- TypeScript's meaning wins where the syntax is the same (GOAL.md, SPEC 2, D2).
- No hidden or implicit calls, no operator overloading (maintainer notes, D12, D26).
- No migration errors for removed forms (GOAL.md constraints).

## Answers from the user

- None asked during the debate. Two questions go to the maintainer with this file: integer holes (O11), and whether fmt rewrites `x: x` (proposal 4).

## Overruled

- none

## Cost of this debate

- 1 of 1 rounds. The opponent took about 54,000 tokens.
