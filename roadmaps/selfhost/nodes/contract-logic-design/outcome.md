# Outcome: contract-logic-design

## Built

- D102 in `DECISIONS.md`, answered by the maintainer on 2026-11-12.
  - (1) `||` between comparisons in `requires`, as clauses joined by `&&`. A caller proves one side per clause. The body and D92 checks get the hull of a one-name `||`, with `u64` sides built as `refineFact` builds them.
  - (2) Relations between two integer parameters: a set of possible orders per pair of locals, from `requires` and from D92 checks, with bounds flowing one step.
  - (3) `!` in `requires`: not built. E0984 prints the rewritten condition.
  - (4) `&&` inside an `ensures` pattern: not built. Several lines stay.
- `roadmaps/selfhost/nodes/contract-logic-design/count.py`: counts the checks each form would serve (`-v` lists sites).
- `roadmaps/selfhost/nodes/contract-logic/brief.md` rewritten as a full brief: size L in three parts, depends on array-patterns and this node, no new error codes (E0984 and E0985 change their details), no `Pattern` change.

## Proof

- Not a fix node.

## Result

- Not an experiment node.

## Deviations

- The brief's example for question 2 ("`high - low` cannot be negative") does not hold under D92 wrapping. D102 lists what a relation proves instead: calls, and bounds through one step.
- No codes from E1040 to E1049 are reserved: every new error is a form E0984 or E0985 already covers.

## Stubs left

- None.

## Gotchas

- `testdata/errors/contracts_unsupported.nra:21` expects E0984 on a `||` condition, and the file also has a two-parameter case. contract-logic must move or replace both.
- A `u64` comparison with a bound from 2^63 up must give the unknown fact on every new path: hull sides, the bound step, the interval proof of a relation.
- `negateComparison` in `operands.nra` is private. contract-logic exports it for the `!` hint and the false `&&` hull.
- The counts are text heuristics: a check counts only when it compares a plain integer parameter with a literal or an upper-case constant.

## Proposed learnings

- A proof rule about a difference (`high - low`) needs an overflow argument, since every `-` wraps (D92) | relation and offset facts look sound until wrapping is considered | roadmaps and DECISIONS.md | no
- A new fact path on `u64` must go through the `refineFact` rule for bounds from 2^63 up | a join or a passed bound ordered as int is unsound there | selfhost/src/operands.nra | yes, with a `u64` error test per path

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/contract-logic-design/out`: fixed point, seed current.
- Samples in `/private/tmp/claude-501/contract-logic-design/samples/` run with that compiler: the E0984 and parse errors of today, two `ensures` lines proving a shift count, and the weaker return proof of lines.
- `python3 roadmaps/selfhost/nodes/contract-logic-design/count.py`.
- `narrac emit selfhost`: 2.434G and 2.411G instructions retired.
- No compiler code changed, so `sh scripts/test.sh` was not run.

## Size

- Estimated M, actual M. The node it refines grew from M to L when the maintainer chose to build relations.
