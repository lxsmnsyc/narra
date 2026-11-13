---
id: generic-field-defaults
title: "Defaults for generic struct fields"
status: done
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone:
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent: 180000
owns:
  - roadmaps/selfhost/nodes/generic-field-defaults/**
  - testdata/generic_field_defaults*
  - testdata/errors/generic_field_default*
shared:
  - selfhost/src/resolve.nra
  - selfhost/src/body.nra
  - selfhost/src/check.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/deriveshape.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---

# Defaults for generic struct fields

## Goal

A field of a generic struct whose type uses a type parameter can have a default: `items: Array<T> = []`, `value: Option<T> = None`, `slots: Array<Option<T>> = [None, None]`. The default is checked once in the declaration with `T` open, so it must hold for every `T` (`count: T = 0` is the usual type error there), and it is built per instance with that instance's types.

Error on the current compiler (built from the seed at `57d55087`):

```ts
struct Bag<T> {
  items: Array<T> = []
  count: int = 0
}

export function run(_args: Array<string>): int uses io, alloc {
  const bag: Bag<int> = { count: 1 }
  print(bag.count)
  return 0
}
```

`gen1.nra:2:3: a default for a field whose type uses a type parameter is not supported by this compiler yet` (E0107, `buildGeneric` in `resolve.nra:993`), then E0224 `missing field "items" of Bag<int>` at the literal, because the default is dropped. `value: Option<T> = None` gives the same pair.

Plan: drop the E0107 check, check such defaults in the generic declaration with `T` open, check them again per struct instance into a body of that instance, and let the emitter take the instance's body for those fields.

## Context

- SPEC section 4 "Structs", "Fields and defaults" (line 432): a default is a compile-time constant (section 10). SPEC has no rule for generic fields; the README lists them under "Not supported yet" (line 207). The README line 93 (language service) already mentions a generic struct default with an error; keep that behavior.
- Today: `buildGeneric` (`resolve.nra:973`) reports E0107 and sets the default to `None`. Defaults of fields without a type parameter are checked once per generic in `checkGenericDefaults` (`body.nra`, through `pendingDefaults` in `check.nra:528`), with an open environment from `makeEnv`, into one body keyed `default|<module>.<name>|<field>` (`checkFieldDefaults`, `body.nra:407`; the key format is described at `check.nra:2411`).
- Emitter: `emitStructLiteral` (`emitexprs.nra:1018`) emits a missing field's default inside its default body, found with `defaultBodyKey` (`emitexprs.nra:1011`), which is shared by every instance. That is right for `count: int = 0` but not for `None` of `Option<T>`, whose C type differs per instance.
- Instances: `fillInstance` (`resolve.nra:1070`) substitutes the field types of each struct instance. Generic function instances are checked per instance in `checkPending` (`body.nra:550`) with the instance's environment and a body key `inst|N`; follow that model.
- Steps:
  1. Remove the E0107 check and its row in `diagnostics.nra` (no deprecation path).
  2. The open check: `checkGenericDefaults` already checks every default of the generic with `T` open; it now includes the fields that use `T`. A default that does not hold for every `T` reports the normal error there, at the declaration.
  3. The instance check: when an instance is filled and the generic's open check had no error, queue each default whose field type uses a type parameter; check it with the instance's environment into a body keyed `default|<instance type key>|<field>`. Defaults of fields without a type parameter keep their one shared body.
  4. The emitter takes the instance key for a field whose declared type in the generic uses a type parameter, and the shared key otherwise.
- `deriveshape.nra` reads field defaults to decode a missing field. A generic struct with `derives Shape` and an `Option<T>` default must decode a missing entry as that default.

## Acceptance

- [ ] Samples (`testdata/generic_field_defaults*`): `items: Array<T> = []`, `value: Option<T> = None` and `slots: Array<Option<T>> = [None, None]` left out of a literal for `T` of `int`, `string` and a user struct; one generic with both a `T` field default and a plain `count: int = 0`; a nested generic (`Bag<Bag<int>>`); a literal inside a generic function where `T` is that function's parameter; a spread literal `{ ...bag, count: 2 }`; `derives Shape` on a generic struct with an `Option<T>` default, decoded from JSON without the field.
- [ ] Every sample is clean under the leak check and sanitizers.
- [ ] Error tests (`testdata/errors/generic_field_default*`): `count: T = 0` (the type error at the default, reported once, not per instance); `lookup: Map<string, T> = Map {}`, not a constant (E0106, as for a non-generic field).
- [ ] E0107 is gone from the code table and the source; the `codes` suite passes.
- [ ] SPEC.md "Fields and defaults" states that a default of a field whose type uses a type parameter must type-check for every type argument and is built per instance. The README line is removed.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one removed check, a second set of default bodies keyed by instance, and one key choice in the emitter. The risk is instances made late, inside generic function instances, before the bodies are emitted.

## Notes for the worker

- Open design question, with the recommended answer:
  - **Once with `T` open, or per instance?** Both, with different roles. The open check is the rule: errors appear once at the declaration, and a default cannot work for `Bag<int>` and fail for `Bag<string>`. The per-instance check only gives the emitter concrete types, and should report nothing new. Rejected: per-instance checking alone, which reports errors at each use and lets a default depend on the type argument.
- Instances can be made while checking other instance bodies (`checkPending` loops until no more are queued). Make sure the per-instance default bodies are checked before emitting starts, including for instances queued late.
- No new diagnostic code is needed.
