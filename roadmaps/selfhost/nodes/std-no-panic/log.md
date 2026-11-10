# Log: std-no-panic

- Branch renamed to feat/std-no-panic and reset to feat/batch-6 before the first commit.
- Seed check: the seed's tables have the old result types, and the bootstrap compiles `selfhost/src` with the new tables in stage 2. A compiler call site on the old form would fail there. So the compiler calls four helpers in `selfhost/src/textops.nra` (`substring`, `splitString`, `repeatString`, `filledArray`) whose bodies use only builtins that keep their type. The 110 call sites in 41 files stay as they are after the seed update; only the four bodies change.
- Option or Result per function (each has exactly one way to fail, so Option everywhere):
  - Str.slice: Option<string>. Str.split: Option<Array<string>>. Str.repeat: Option<string>.
  - Bytes.at: Option<u8>. Bytes.slice: Option<bytes>.
  - Float.truncate: Option<int> (NaN and out of range are one case for the caller).
  - Arrays.filled: Option<Array<T>>. Ranges.stepped: Option<Range>.
  - Math.clamp: Option<T>. Math.powInt: Option<int>. Random.nextInt: Option<int>.
- F32.truncate keeps its panic (not in the D92 list); the runtime keeps `narra_truncate` for it and adds `narra_float_truncate`.
- The effect checker charged `alloc` for every `Some`, `Ok` and `Err` with a payload. It now charges it only when the enum is stored on the heap (`isHeapEnum`, the emitter's test), so `clamp`, `powInt`, `nextInt` and `shuffle` need no `uses alloc`.
- Review fixes: helpers panic on a negative count and on a position inside a character; `Str.repeat` of an empty text returns at once; `F32.truncate` returns `Option<int>`; the remaining std panics (`Str.sliceFrom`, `Str.charAt`, `Str.nextChar`) are listed in SPEC and D92.
- Maps.update(mutable map, key, initial, change): builtin with a callback argument (check in stdmap, emit in emitmaps). New value is `change(old)`, with `initial` as `old` when the key is missing.
- Compound assignment on a Map index: E0226 with the plain message "only variables, fields and array elements take a compound assignment". The emitter path for it is removed.
- Test sources and sample programs: the runner has `selfhost/tests/src/textops.nra` (panics on None, since the runner computes its own positions). Samples that only needed a fresh string use a private helper at the end of the file. The three typesdump samples keep their builtin calls and wrap them in `Options.withDefault`.
- Removed samples: testdata/panics strings2_inside, strings2_range, strings2_split_empty, map_missing_key (the cases no longer panic or compile).
