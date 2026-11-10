# Outcome: std-spec-functions

## Built

None of the functions existed on the base (feat/batch-6 at 0ee9a101). `Arrays.entries` was already built by an earlier node and is not part of this one.

- A builtin std module can now have a Narra part: `std/src/ext__<name>.nra` has the module path `std/ext/<name>`. `addStdImports` (`selfhost/src/modules.nra`) loads it whenever a file imports `std/<name>`. `extensionScope` and `memberScope` (`selfhost/src/scopes.nra`) find it. `Alias.name(...)` (`checkDottedCall`), `Alias.NAME` (`checkField`), top-level const reads and `import { name } from "std/<name>"` look in the builtins first and then in this part. The generator of `stdtext.nra` (`selfhost/tests/src/generated.nra`) maps `__` in a file name to `/` in the path.
- `Float.PI` (`std/src/ext__float.nra`, a const).
- `Int.compare` (returns `Ordering`), `Int.divmod` (`Option<#[int, int]>`), `Int.floorDiv` (`Option<int>`) in `std/src/ext__int.nra`. Choice: both division functions return `Option` (`None` for a zero divisor), so they need no proof once proven-operands lands. `MIN / -1` wraps (D92).
- `Str.find(text, needle): Option<int>` (`ext__string.nra`), the same as `Str.indexOf`; `ByteIndex` is not built.
- `Arrays.map` (effect variable) and `Arrays.slice` (`Option<Array<T>>`, `None` outside the array, copies) in `ext__array.nra`.
- `Arrays.withCapacity(count)`: a builtin (`checkArrayWithCapacity` in `stdarray.nra`, emitter case in `emitexprs.nra`), calls the array's `_new`; a negative count reserves nothing. Without an expected array type it reports E0302.
- Checked first: `Arrays.map` did not exist; `Int.truncate` is not added (the spec means `Float.truncate`, which exists; no spec text asks for `Int.truncate`).
- `Str.copy` and `Arrays.copy` are not added. SPEC.md now says `copy(...)` where it named them (section 7, Slices share memory). SPEC.md also documents the Narra parts.
- Samples: `testdata/std_spec_int`, `std_spec_array`, `std_spec_string` (+ `.out`), `testdata/errors/std_spec_with_capacity`.
- `selfhost/tests/spec-skip.txt`: the `Float.PI` and `Int.compare` lines are gone; the `std/array has no map` line keeps only its other reason (the sample imports `app/text`).

## Deviations

- Review fixes: `import { PI } from "std/float"` (with or without `as`) works; `Arrays.withCapacity` cuts the reserve to 2^20 elements; `std/ext/...` imports from outside `std/` are E0426; one helper, `extensionPath` in `modules.nra`, builds the companion path.
- Files outside the owned and shared paths: `selfhost/tests/src/generated.nra` is under `selfhost/tests/src/**` (shared); `selfhost/src/stdtext.nra` is generated. Nothing else.
- `Float.PI` is a source const, not a checker feature: no module constants were added to the builtin checker.
- The seed is not regenerated. `selfhost/src` does not call the new functions. `sh selfhost/bootstrap.sh` takes the stage 3 path (stage 1 C differs from stage 2) until the next seed update.
- `Arrays.slice` copies; the spec says a slice is a view (a recorded difference already for `Str.slice`).

## Stubs left

- A companion export with the same name as a builtin of its module is not reported: the builtin wins silently. Checking it needs a new diagnostic code, so it is left for a later node.

## Gotchas

- A builtin module and a source module of the same path cannot be mixed in the checker (`sourceScope` makes every lookup go to the source). That is why the Narra part has its own path.
- Inside `std/src/ext__array.nra`, `Arrays` is the builtin module; call `Arrays.withCapacity` and `Arrays.push` there.

## Proposed learnings

- Add a function to a builtin std module in Narra when it needs no primitive: `std/src/ext__<name>.nra` | no registry row, emitter or C needed | std/src, selfhost/src | no

## Verification

- `sh selfhost/bootstrap.sh <dir>`: fixed point holds (stage 2 and stage 3 identical).
- `NARRAC=<built> sh scripts/test.sh --suite spec` and `--suite generated --update`: pass.
- `sh scripts/test.sh`: pass (269 s).

## Size

- Estimated M, actual M.
