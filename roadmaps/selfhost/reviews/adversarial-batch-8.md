# Adversarial review of batch-8 (fresh-context agent), with dispositions

All eight findings were checked against the code and applied.

1. The `@target`-on-extern sample in SPEC uses the single-function form D95 replaces, and `@target` placement was open. c-interop rewrites SPEC sections 9 and 12 to the block form; `@target` goes on the block; c-audit-target's acceptance names the new sample.
2. The link path: `[link] libraries` is parsed and carried in `Package` but never reaches `cc`; `cCompilerArguments` always adds `-lm`; `compileToC` returns only text. c-interop now states the `compileToC` change, keeps `[link] search` (passed as `-L`), and lists `modules.nra` and `main.nra`.
3. New `Type` variants touch every exhaustive `Type` match. The type, key, C-name, trait, resolution and dump files are added to the three type-adding nodes.
4. const-generics and unsafe-ptr both change the `Type` key and C-name plumbing in wave 1. unsafe-ptr now waits for const-generics.
5. fixed-arrays and c-interop share emitter files. Each feature's emitter code goes in its own module, listed in `owns:`.
6. `@target` parsing cannot be reused for `@trusted` and `@extern`. New node attributes (S) makes attributes a general list first and adds the E1000 range row.
7. `narrac audit` had no home or output format. c-audit-target names `main.nra` dispatch, a new `audit.nra`, and the line format; `buildtarget.nra` dropped from its paths.
8. Redundant edge c-audit-target to unsafe-ptr removed; empty `added:` fields set.
