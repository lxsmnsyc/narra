# Outcome: generic-field-defaults

## Built

- Defaults on fields of generic structs whose type uses a type parameter (`items: Array<T> = []`, `value: Option<T> = None`, `slots: Array<Option<T>> = [None, None]`), also on structs with bounded parameters (`Counted<T: Show>`).
- The E0107 check in `buildGeneric` (`selfhost/src/resolve.nra`) and its row in `diagnostics.nra` are gone.
- Open check: `checkGenericDefaults` (`body.nra`) now checks these defaults with `T` open, once, at the declaration. When it reports an error it sets `Checker.failedDefaults[generic]`.
- Instance check: `fillInstance` calls `queueInstanceDefaults` (resolve.nra), which queues a concrete struct instance in `Checker.pendingInstanceDefaults` when a defaulted field uses a parameter. `checkPending` drains the queue before and between function instances (so late instances are covered). `checkInstanceDefaults` checks those fields with the instance's types and environment into bodies keyed `default|<plainTypeKey of the instance>|<field>` (`instanceDefaultKey` in `check.nra`). An instance of a generic whose open check failed is skipped, so nothing is reported per instance.
- Emitter: `defaultBodyKey` (`emitexprs.nra`) returns the instance key for a field whose declared type in the generic uses a type parameter, and the shared key for the others.
- `derives Shape` needed no change: the generated decoder builds a struct literal, which takes the instance default.
- Samples: `testdata/generic_field_defaults.nra` (int, string, user struct, `Bag<Bag<int>>`, a literal inside a generic function, a spread literal, a plain `count: int = 0` beside the `T` defaults, bounded `Counted<T: Show>`), `testdata/generic_field_defaults_shape.nra` (Shape decode without the field). Errors: `testdata/errors/generic_field_default_type.nra` (`count: T = 0`, reported once with two instances), `testdata/errors/generic_field_default_constant.nra` (E0106).
- SPEC "Fields and defaults" states the rule; the README "Not supported yet" line is removed.

## Deviations

- `selfhost/src/check.nra` is shared; the new `Checker` fields are `failedDefaults` and `pendingInstanceDefaults`, plus `instanceDefaultKey` and an updated `findBody` comment.
- README line 93 (language service) is kept as the brief says.

## Stubs left

- None.

## Gotchas

- A new instance of a generic struct made after `checkPending` ends gets no default body, and the emitter would report "field default body". Instances are made only in the checker today.
- Instances whose arguments still contain a type parameter are not queued.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh <scratch>/out`: fixed point reached.
- `NARRAC=<built> sh scripts/test.sh --filter generic_field_default`: `PASS, 20 cases` (debug with sanitizers and leak check, and release).
- `testdata/declsdump/bad_defaults.dump` golden rewritten (it held the E0107 text).
- `sh scripts/test.sh`: `test.sh: PASS, 7778 cases in 290s`.

## Size

- Estimated M, actual M.
