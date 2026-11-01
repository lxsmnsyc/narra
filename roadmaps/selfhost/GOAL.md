---
milestones: [m3-core, m4-systems, m5-concurrency, m6-performance, m7-tooling, m8-selfhost, m9-adoption, m10-std]
merge:
  mode: auto
  method: squash
  max_size: M
  exclude: [release, hotfix, contract, gate]
branches:
---

# Bootstrap compiler to a self-hosted Narra

## Outcome

The Narra compiler is written in Narra and compiles itself. The Go bootstrap compiler `narra` first grows to cover the language features the compiler needs. The compiled compiler (stage 1) builds itself (stage 2), and stage 1 and stage 2 produce byte-identical C for the test programs. The remaining spec features (`SPEC.md`) are implemented or explicitly deferred, and a benchmark compares Narra with Go and Rust. After self-hosting, milestone m9-adoption makes Narra usable by others: a language guide, an installation guide, `narra fmt`, `narra query`, a language server, agent skills and maybe a linter (should, except the linter: could). The language was named Narra on 2026-10-28, before self-hosting.

## Acceptance

- `cd compiler && go test ./...` passes, with every sample program clean under AddressSanitizer and the leak check.
- The self-hosted compiler builds itself, and its output matches the Go compiler's output on every program in `compiler/testdata/`.
- `bench/` reports Narra, Go and Rust times for the same programs.

## Priorities

- must: needed for the self-hosted compiler (milestones m3-core and m8-selfhost).
- should: spec features that real programs need soon, and the measurement and tooling that make the language usable.
- could: spec features with no current user.
- later: parked.

## Constraints

- `SPEC.md` is the source of truth for language behavior. A node that must differ records the difference in `compiler/README.md` under "Differences from the spec".
- Every new feature adds sample programs to `compiler/testdata/` (run under sanitizers and the leak check) and compile-error tests to `compiler/testdata/errors/`.
- The generated C must stay valid C11 for clang and gcc.
- Go standard library only in the bootstrap compiler. No third-party Go modules.

## Non-goals

- A package registry, an editor plugin, or a debugger.
- Keeping the Go compiler after self-hosting beyond what bootstrapping needs.

## Verification

- Fast check (safe to run in parallel worktrees): `cd compiler && go vet ./... && go test ./...` (about 10 s, no shared resources)
- Full check: the same, plus `bench/run.sh` once it exists (exclusive: benchmarks need a quiet machine).

## Numbered resources

- Diagnostic codes (`E0001` and up) are assigned by node `json-diagnostics`. Earlier nodes use the placeholder code `E0000`. Nodes after it reserve their codes with `reserves: [diag E0NNN]`.

## Hot files

`parser.go`, `ast.go`, `clone.go`, `check_expr.go` and `emit_expr.go` are edited by several nodes. split-compiler splits the checker and emitter by feature, but the parser stays one recursive-descent file and each feature adds a few `case` lines to the shared switches. Nodes in the same wave merge in the order `plan.py status` prints, and each later branch rebases before review. In wave 2, modules merges first, because it changes how every C name is generated.

## Sources

- `SPEC.md` and `DECISIONS.md` in the repo root.
- `compiler/README.md`: what the bootstrap compiler supports and its gaps.

## Coverage

- closures with non-escaping parameters and `@inline` -> closures, inline
- multiple modules and `narra.toml` -> modules
- traits and constraints, struct and enum equality -> traits
- nested patterns, spread updates, destructuring -> patterns
- effects checking -> effects
- generators -> generators; async -> async; threads, channels, shared state -> threads
- unsafe, extern, C interop -> c-interop
- opaque and distinct types -> opaque-distinct
- regions -> regions; fixed arrays -> fixed-arrays; SIMD -> simd; build conditions -> build-conditions
- JSON diagnostics -> json-diagnostics; query CLI -> query-cli; fmt -> fmt
- no global functions except print, assert, panic (user decision 2026-10-30) -> prelude-imports
- literal type arguments, `Map<K, V> {}` (user decision 2026-10-30) -> literal-type-args
- `std/option` and `std/result` helpers (user decision 2026-10-30) -> std-option-result
- testing library `std/test` (added 2026-10-30, not in SPEC.md) -> std-test, test-runner
- Perceus reference count optimization -> perceus; layout tables -> layout-tables
- benchmarks against Go and Rust -> bench-harness
- self-hosting -> selfhost-spike, selfhost-ast, selfhost-front, selfhost-modules, selfhost-check-decls, selfhost-check-bodies, selfhost-emit, selfhost-bootstrap
- renaming the language -> rename (done early: Narra)
- language guide -> language-guide; installation guide -> install-guide; agent skills -> narra-skills; language service -> language-service; linter -> lint (user request 2026-11-01)
- std library growth -> std-source, std-json, std-io-more, std-basics, std-unicode, std-arena (2026-11-01); `Range.stepped` -> std-basics; std modules of parked features stay with those nodes
- not in the original list, needed for self-hosting: split-compiler, box-recursive, maps, strings-std, io-std

## Decisions

- 2026-10-28: the user confirmed the plan. The must line covers only what self-hosting needs. split-compiler runs first. Every pull request waits for the user.
- 2026-10-28: the repo was published at https://github.com/lxsmnsyc/narra (public). Node pull requests go there. The repo has no CI workflows yet; checks run locally.
- 2026-10-28: wave 1 (split-compiler, bench-harness) started.
- 2026-10-28: the user switched merging to auto. The orchestrator merges a pull request without the user when `plan.py can-merge` allows it: clean fresh-context review, local checks pass on the rebased branch, size M or smaller. Gated nodes, contract-freezing nodes, releases and hotfixes still go to the user. This replaces "every pull request waits for the user" from the first decision.
- 2026-10-30: wave 3 showed that two reviewed pull requests can break main together (io-std built enum values the old way after box-recursive merged). Rule: before merging a second pull request of the same wave, rebase it on main and rerun the fast check. `can-merge` answers hold only for the base they were run on.
- 2026-10-30: selfhost-spike gate, user decisions. Add `top-level-const` (should). The self-hosted compiler uses integer node ids and side tables, not an arena library. selfhost-check is split into selfhost-check-decls and selfhost-check-bodies. Port soon: every m4-m7 node except opaque-distinct is parked (`later`) until after self-hosting, because each feature built in Go first must also be ported. std-convert names approved (Char.code, Char.fromCode, Int.parse, Int.toHex, Str.endsWith, contains, replace, join).

- 2026-11-01: the user added milestone m9-adoption. fmt, query-cli, the language service and lint are written in Narra after selfhost-bootstrap, not in Go. Docs and skills are should; lint is could and may fold into checker warnings. fmt and query-cli move from m7-tooling (later) to m9-adoption (should). Editor plugins stay a non-goal; the LSP server is in scope.
- 2026-11-01: the user added milestone m10-std. A gated spike (std-source) first decides whether std modules can be Narra source with only primitives as builtins. The language service waits for std-json and std-io-more.

## Gates

- Contract freeze after split-compiler (the diagnostic and builtin registry contracts).
- selfhost-spike: the user reviews the list of missing features before the port starts.
