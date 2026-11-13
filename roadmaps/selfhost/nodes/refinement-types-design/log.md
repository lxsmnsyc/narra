# Log: refinement-types-design

- Base: `feat/batch-13` at b23bbd4f. Compiler for probes: `sh selfhost/bootstrap.sh /private/tmp/claude-501/refinement-types-design/out` (fixed point held).
- DECISIONS.md had no refinement decision (D95 mentions "contracts and refinements" only in passing). D116 is new.
- `selfhost/src` and `std/src` declare no distinct type (`grep "distinct type"` finds only messages and comments). `ByteIndex` is a builtin. No in-repo user.
- Parser: `parsePatternBody` parses a comparison bound with `parseUnary`, so `>= 1 && <= 65535` stops at `&&`; a list of patterns joined by `&&` needs no new `Pattern` variant. `FuncDecl.ensures` is already `Array<Pattern>`.
- `isComparisonValue` (`contracts.nra:1376`) and `recordContractCall` both read `findOperandFact`, so one change there (unwrap of a refined integer type) reaches divisors, shift counts, `requires` arguments and `ensures` returns.
- Probe `b.nra`: `slot(Port.toInt(port))` against `requires index >= 1` is E0985 today. The unwrap gives no fact.
- Probes `a.nra`, `c.nra`: no distinct function is a value. `const make = UserId.from` is a parse error (`from` is a keyword), `const unwrap = UserId.toInt` is `unknown name "UserId"`. So the planned E1224 (refined `from` as a value) was dropped; codes are E1220-E1223.
- `emitexprs.nra:624` and `consteval.nra:2847` treat every `Ref.Distinct` call as its argument; `check` needs its own case in both.
- `deriveshape.nra:392` writes `Ok(T.from(__base.value))`; with a proven `from` that line would fail on a refined type, so it calls `check`.
- `std/shape` errors are one string with a path (`.users[2].age: ...`); the refinement error follows it.
- SPEC: the new "Refinements" subsection uses inline code only, since an unbuilt `ts` block would fail the spec suite and `spec-skip.txt` is outside this spike's paths. The build brief turns them into checked blocks.
- `NARRAC=<out>/narrac sh scripts/test.sh --suite spec`: PASS, 62 cases. `--suite sources`: PASS, 1 case.
