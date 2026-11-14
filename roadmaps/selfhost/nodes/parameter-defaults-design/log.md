# Log: parameter-defaults-design

- 2026-11-13: branch spike/parameter-defaults-design reset to feat/batch-13 (b23bbd4f). No D115 or parameter default decision in DECISIONS.md; D12 rejects optional parameters, D87 leaves defaults to a separate decision.
- Evidence (grep, one-line calls, selfhost/src): checkExpression 50 of 102 calls pass `None` last; argumentText 41 of 57 pass `0`; writeSlot 22 of 24 pass `""`. Calls whose last argument is a literal, selfhost/src and std/src: true 74, false 112, None 68, 0 146, "" 109, [] 31.
- std/src has no option struct for any function (structs: Random, Instant, Duration, Target, a private json Parser). SPEC's `OpenOptions` example names nothing real.
- Field default code to reuse: checkFieldDefaults and checkGenericDefaults (body.nra), isConstantValue (exprs.nra: literals, unary, arrays, constructors), E0106, emitter defaultBodyKey and switchBody (emitexprs.nra).
- Fill points: matchNamedArguments (calls.nra, E0331, E0316), checkArguments (exprs.nra, E0316), closures.nra E0316 for calls through values.
- parseFuncHeader is shared by declarations, trait functions and extern functions; parseArrow and parseFuncType are separate.
- Possible existing gap: effects.nra Expr.StructLit walks only written fields, so an omitted non-empty array field default charges no alloc. Not checked further; noted in the build brief.
- No goal debate: the trailing-constant form follows from D26 and D87 and the counts; the open choices are listed for the maintainer.
- SPEC: no ts sample added (the spec suite checks every ts block and spec-skip.txt is outside this node), so the spec and sources suites were not run.
