# Outcome: perceus-reuse

## Built

- No compiler change. In-place reuse and drop specialization were measured and dropped: the compiler almost never builds a heap payload in an arm whose subject dies there, and the bench programs have no heap payloads at all. The brief allows stopping when the gain is within noise.
- Samples `testdata/perceus_reuse_lists` and `testdata/perceus_reuse_tree` pin the behavior a later reuse must keep: a list mapped while nothing else holds it, the same list mapped while two other names hold it (originals intact), an owned temporary subject rewritten in its arm, a rewrite while a closure and a suspended generator hold the list, an arm that builds the new cell and then reads the old fields, a tree insert and mirror over a fresh and a shared tree (also held in a map), and a 500-node chain.
- Measurement scripts in `roadmaps/selfhost/nodes/perceus-reuse/measure/`: `count-arms-mutant.py` and `count-arms-runtime.py` (the arm and allocation counters), `leak-bound-mutant.py` (the release cost bound), `unsound-lists.py` (a reuse with no uniqueness check, for the samples).

## Proof

Instructions retired, base feat/batch-4 9501a3c, compiler built by `sh selfhost/bootstrap.sh`, macOS arm64, load 7 to 10.

| program | base | branch |
| --- | --- | --- |
| compiler (`narrac emit selfhost`) | 3.085 to 3.122 G, 127 MB | same binary (no code change) |
| fib | 1.216 G | same |
| sieve | 2.285 G | same |
| matmul | 2.447 G | same |
| strings | 1.097 G | same |
| interp | 1.965 G | same |
| particles | 2.612 G | same |

Bench values are the median of 3. "Same" means the emitted C is byte-identical because no compiler code changed.

- Where reuse could apply. On base, parameters and `match`/`is` bindings borrow and locals are released at scope end, so a payload dies inside an arm only when the subject is an owned temporary (`match (f(x))`). Counted on `narrac emit selfhost` with an instrumented compiler: 3,509,453 arms entered on heap-enum subjects; 59,467 with an owned temporary subject; 56,038 of those with count 1. Of 253,278 heap payload allocations (24.5 MB), 78 are a same-size allocation in an arm whose owned subject has count 1. With owned parameters and last-use moves (not on base), 16 more. Reuse could save at most 94 allocations of 253,278 (0.04%).
- Why so few: the compiler's heap enums (AST, `Type`) are built once and read many times. Results go into side tables by id, so no pass rebuilds a tree of the same shape. The 46 owned subject sites are lookups such as `match (underlyingType(...))` and `Option` results, and their arms read fields.
- Bench programs: every enum is inline since option-inline (`interp`'s `Op` included). The emitted C of all six has no payload type, so neither step can change them.
- Drop specialization bound: a compiler that never releases an owned heap-enum subject of `match` or `is` (46 sites, leaking) took 3.045 to 3.062 G against 3.080 to 3.096 G for base (5 interleaved runs each). So these releases cost 34 M instructions (1.1%), subtree frees included. Specialization keeps the subtree frees and removes only the dispatch (`narra_drop`, the indirect call and the tag switch) for 56,038 freed cells, under 2 M (0.05%). No field retains to cancel: the arms borrow their bindings and only read them. The base spread is 0.5% to 1.2%.
- Samples against an unsound reuse (`unsound-lists.py`, hand-patched C of `perceus_reuse_lists`, each arm writes into its subject's cell with no uniqueness check): under ASan, heap-use-after-free in `swapFirst` (old head read after the rewrite, closure case). Without that site, the shared-list case prints `11!,22!,33!` for both other names instead of `11,22,33`. The owned unique case (`bumpFirst`) still prints the right output, as a sound reuse would.

## Deviations

- The brief's code (drop specialization and reuse) is not built; reason above. DECISIONS.md and SPEC.md are unchanged: D55's "never mutated" holds as written.
- No mutant compiler with the uniqueness check removed, because there is no reuse code to remove it from. The samples were checked against a hand-made unsound C mutant instead.
- `bench/run.sh` wall times not recorded: no code changed, and the load stayed at 7 to 10.
- The seed is not regenerated and `sh scripts/linux-check.sh` is not run (batch rules).

## Stubs left

- None.

## Gotchas

- Reuse needs the payload to die in the arm. Under the current rules only an owned temporary subject does. A later attempt needs owned parameters or last-use moves first (perceus dropped moves), and the count above says even that adds 16 reuses in the compiler.
- A sound reuse must also wait until the arm no longer reads its borrowed bindings: `swapFirst` in the lists sample builds the new cell and then reads the old head.

## Proposed learnings

- Count the dynamic opportunities of an optimization with an instrumented compiler before building it | 78 reusable cells of 253,278 allocations killed a size L node in a few hours | selfhost/src | no

## Verification

- `NARRAC=<base narrac> sh scripts/test.sh --filter perceus_reuse`: PASS.
- `sh scripts/test.sh` (full fast check) after rebasing on feat/batch-4 49cc207: PASS in 137 s, 8 shards with 5838 passed and 0 failed.

## Size

- Estimated L, actual S. The measurement settled it before any emitter code.
