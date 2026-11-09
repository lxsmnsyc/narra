# Outcome: std-json

## Built

- `std/src/shape.nra`: `export enum Value { Null, Bool(value), Int(value), Float(value: f64), Text(value), List(items), Record(fields: Map<string, Value>) }`.
- `std/src/json.nra`: `parse(text): Result<Value, string> uses alloc` and `stringify(value): string uses alloc`. Import as `import * as Json from "std/json"` and `import { Value } from "std/shape"`; constructors are written `Value.Int(...)`, also in patterns.
- Primitive `FloatText.parse(string): f64` in `std/internal/float` (`strtod`; the text must already be a valid number). Rows and emitter in `selfhost/src/stdinternal.nra`, dispatch line in `emitstd.nra`, registration in `stdlib.nra`. Module rows `selfhost/src/stdjson.nra` and `stdshape.nra`.
- `selfhost/src/modules.nra`: a std source module is listed after the std modules it imports (also in single-file packages).
- Samples `testdata/json.nra`, `json_floats.nra`, `json_errors.nra`, `json_rpc.nra` with `.out`.
- SPEC.md (Standard modules, Serialization), compiler/README.md, DECISIONS.md D83.

## Behavior to know

- Errors: `<line>:<column>: <reason>`, column counts characters. Reasons: `expected a value`, `expected , or ]`, `expected , or }`, `expected a string key`, `expected :`, `invalid escape`, `invalid \u escape`, `lone surrogate`, `control character in string`, `invalid number`, `nesting is deeper than 512`, `unexpected character after the value`; input ending early gives `unexpected end of input` at the end position.
- `-0` is `Int(0)`. `1e999` is `Float(inf)`, which `stringify` writes as `null`. An integer outside `int` is a `Float`.
- `stringify` recurses on the depth of the given `Value` (only `parse` caps depth).
- Whole floats below 1e15 print as `100.0`; others use `Float.toString` (`1e+22`, `5e-324`).

## Acceptance

- Samples: done (ASan and leak check via `narrac run -sanitize`, debug and release). Go items (`TestSelfhostEmit`): not applicable.
- SPEC, README, D83: done.
- Full `scripts/test.sh`: see the report.

## Gate decisions

- Kept the brief's design (D83 as written): `Value` in `std/shape`, `parse`/`stringify`. No alternative adopted.
- Float parsing: new `std/internal/float` primitive over `strtod` (alternative: a `Float.parse` in public `std/float`, rejected to avoid new public API); `std/internal/probe` kept (alternative: replace it, rejected because existing samples use it).
- Ordering: fixed in `addStdImports` (post-order) rather than calling `orderModules` for single-file packages (simpler, same result).
- Over-large integers become `Float` (as JSON.parse does) rather than errors.

## Stubs left

- None. Typed `encode`/`decode` belong to shape-trait and shape-derive.

## Size

- Estimated M, actual M.
