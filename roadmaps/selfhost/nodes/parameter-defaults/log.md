# Log: parameter-defaults

## 2026-11-13

- Branch feat/parameter-defaults, reset to feat/batch-13 (82e54d0b).
- Design: the default body is keyed per call ("paramdefault|<caller body>|<call id>|<parameter>"), checked with the call's parameter types, so a generic default gets the instance types and two instances in one caller do not share node types. The declaration check ("paramdefault|func|<funcIndex>|<parameter>") reports the errors; a call fills only a default that held there, so a broken default is reported once.
- argumentOrders keeps its meaning (written index per passed parameter); a left-out parameter has no entry, so orderArguments gives the written arguments only and every existing reader stays correct. `ParamDefaults.fillArguments` gives the full list for the emitter, the effect pass, the constant evaluator and contracts. A tail self-call that fills a default has fewer arguments than parameters, so emittail's length check keeps it a plain call.
- The emitter emits each filled default from its call body into `heldArguments`, the slot the async spill already uses, so the plain call, the `@inline` copy and every other reader of the argument list pick it up without a change. A nested call that fills the same default saves and puts back the outer entry.
- A first bootstrap of the base ran while I edited sources, and its stage 3 differed; rebuilt the base from `git archive feat/batch-13` in a separate folder.
- Proof: base 82e54d0b, `narrac run testdata/param_defaults.nra` reports `param_defaults.nra:9:42: expected ")", found "="`, exit 1. Branch: prints the expected output, exit 0.
- Confirmed the struct field default `alloc` gap (see outcome, Gotchas); not fixed here.
- Rebased on feat/batch-13 8ea130c1 (clean). Full check: PASS, 8815 cases. Emit of selfhost from base and branch compilers identical.

## 2026-11-14

- Review fixes: a tail self-call that leaves out a default is a jump again (`emittail.nra` matches on the filled list; `testdata/param_defaults_tail.nra` overflows the stack under ASan with the c825d80e compiler and passes now); `name?: T = value` gets E1216 without a fix; an arrow's `name?: T` gets a second E1216 row.
- Rebased on feat/batch-13 11d255b0: one conflict in `emitexprs.nra` (kept both sides); the auto-merge of `diagnostics.nra` dropped the `},` of the last E1223 row, restored. `deriveshape.nra` now imports `querysignature.nra`, so `writeDefaultValue` moved into `querysignature.nra` to avoid an import cycle.
