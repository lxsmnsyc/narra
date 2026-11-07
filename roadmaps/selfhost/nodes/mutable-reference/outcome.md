# Outcome: mutable-reference

## Built

- `DECISIONS.md` D81: `mutable` changes the contents of a value, never the binding. Twelve rules, the two rejected forms, reasons, prior art and cost. Rules 1 to 5 are the brief's plan; rules 6 to 12 are added here (listed below).
- `SPEC.md` section 2 "Bindings" rewritten (`const` fixes the name; copies; read-only pattern bindings and parameters; top-level constants; `@unique` on `const`), and section 6 "Mutable parameters" rewritten (contents types, generics, no whole assignment, arguments, exclusivity for every binding, enums and tuples). Lines that repeated the old rule were brought in line: section 1 overview, section 2 "Functions", "Closures" (D77 amendment) and the TypeScript table row on `const`, section 4 struct field writes and `Box`, section 6 "In-place updates", `@unique` and "Cycles", section 8 `Shared.update`, section 10 top-level `const`.
- `compiler/README.md` "Differences from the spec": the D81 rules are not built yet; the compiler still follows D27.
- `roadmaps/selfhost/nodes/mutable-reference/count.py`: the script behind every count below. Run from the repo root: `python3 roadmaps/selfhost/nodes/mutable-reference/count.py`.

### Rules added beyond the brief

- **Contents types (rule 1, widened).** Besides structs, `Array` and `Map`: `FixedArray`, `Box<T>`, `StringBuilder`, `Generator<T>`, `Coroutine<In, Out>`, an `opaque struct`, and a distinct type over any of them. Needed: `selfhost/src` has 53 `mutable` parameters of type `StringBuilder`, `std/src/coroutine.nra` takes `mutable` generators and coroutines, and SPEC uses `b.value = v` and `Net.read(mutable conn)`.
- **Rule 6.** Bindings without a keyword are read-only views: a plain parameter and a `match` or `is` pattern binding cannot have fields assigned or be passed as `mutable`. Bindings with `const` or `let` (declarations, `for ... of`, destructuring) own a copy and follow rule 4.
- **Rule 7.** A top-level `const` stays fixed at every depth (E0290 unchanged). Tests `errors/const_mutable_arg.nra`, `const_assign_field.nra` and `const_top_assign.nra` keep their meaning.
- **Rule 8.** A struct inside an enum payload or a tuple is not a place. The field holding the enum or tuple is replaced whole.
- **Rule 9, the generic bound (answers the brief's question).** No bound fits: no Narra trait guarantees a contents type, and none of the four trees has a bare generic `mutable` parameter. `mutable value: T` is a compile error for any bare type parameter. Generic instances (`mutable items: Array<T>`) are fine. In a trait, `mutable value: Self` is allowed and every `implement` of the trait must be for a contents type, checked at the `implement` block (the two testdata traits that do this implement them for structs).
- **Rule 10 (amends D77).** A non-escaping closure that changes a captured local `const` captures it by reference and borrows it for the call (E0614); a `const` it only reads stays a copy. Any other closure cannot change a captured value.
- **Rule 11.** E0319, E0929, E0609 and E0915 apply to every binding passed as `mutable`, `const` included.
- **Rule 12.** `Shared.update(hits, (count) => count + 1)`: the callback returns the new value. The old SPEC form used a `mutable int` arrow parameter, which breaks rules 1 and 2 and E0028 already rejects.

## Counts

Trees: `selfhost/src`, `std/src`, `compiler/testdata` (every `.nra`), and the `narra` blocks of `docs/guide/*.md`.

### `mutable` parameters by type

| Tree | struct | Array | Map | StringBuilder | Generator / Coroutine | Self | int | Option | function type | bare generic |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| selfhost/src | 635 | 46 | 22 | 53 | 0 | 0 | 3 | 0 | 0 | 0 |
| std/src | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 |
| compiler/testdata | 17 | 11 | 2 | 1 | 1 | 2 | 23 | 8 | 3 | 0 |
| docs/guide | 10 | 3 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 |

Rejected by D81: 3 in selfhost, 0 in std, 34 in testdata, 2 in the guide. Bare generic: 0 everywhere.

### Parameters of a rejected type, and the rewrite

selfhost/src (3, all `int`):

- `selfhost/src/patterns.nra:358` `findMissingCases(mutable limit: int)`, a budget shared down the recursion (decremented at `:364`, passed on at `:391`, `:414`; caller `selfhost/src/match.nra:591-592`). Rewrite: a struct `struct CaseBudget { remaining: int }` passed as `mutable budget: CaseBudget`, `budget.remaining = budget.remaining - 1`. Example 1 below.
- `selfhost/src/declsdump.nra:118` `dumpModule(mutable nextImpl: int)`, passed on at `:143`; caller `:29-31` (`let nextImpl = 0`, checked at `:36`).
- `selfhost/src/declsdump.nra:267` `dumpImpl(mutable nextImpl: int)`, incremented at `:276`. Rewrite for both: `dumpImpl` returns `nextImpl + 1` and `dumpModule` returns the index after its last block (`nextImpl = dumpImpl(...)` in the `Decl.Impl` arm, which becomes a block arm), or a `struct ImplCursor { next: int }` passed as `mutable`. The return form is shorter here: one caller each.

compiler/testdata (34). Most are fixtures for other rules (exclusivity, narrowing, closures); each keeps its rule by switching to a struct counter, for example `struct Counter { n: int }` and `bump(mutable counter: Counter)` with `counter.n = counter.n + 1`:

- `int`, written as a counter or output (23): `compiler/testdata/closure_borrow.nra:21`, `match_block_values_exclusive.nra:5`, `mutable_read_ok.nra:3`, `mutable_read_ok.nra:12`, `ternary_mutable.nra:5`, `unused_mutable_output.nra:3`, `diagnostics/mutable_read_conflict_fix.nra:2`, `typesdump/bodies/err_statements.nra:7`, `typesdump/closures/err_closures.nra:42`, `errors/closure_borrow_mutable_param.nra:2`, `errors/closures_capture_mutable.nra:2`, `errors/match_block_values_mutable_conflict.nra:2`, `errors/mutable_read_conflict_block_arg.nra:2`, `errors/mutable_read_conflict_block_operand.nra:2`, `errors/mutable_read_conflict_loop.nra:2`, `errors/mutable_read_conflict_match.nra:2`, `errors/mutable_read_conflict_nested.nra:2`, `errors/mutable_read_conflict_operand.nra:2`, `errors/mutable_read_conflict_template.nra:4`, `errors/mutable_read_conflict_twice.nra:2`, `errors/mutable_twice.nra:2` (two parameters), `errors/ternary_mutable_conflict.nra:2`.
- `Option<int>` (8), the narrowing tests: `errors/let_narrowing_mutable_arg.nra:2`, `errors/let_narrowing_same_statement_arg.nra:2`, `_const.nra:2`, `_for.nra:4`, `_match.nra:2`, `_param.nra:2` and `:7`, `errors/mutable_read_conflict_option.nra:2`. Rewrite: hold the option in a struct field (`struct Slot { value: Option<int> }`), narrow `slot.value`, and pass `mutable slot`; the narrowing must still end at the call.
- function type (3), the E0609 tests: `typesdump/closures/err_closures.nra:11` and `:15`, `errors/closures_callee_mutable.nra:6`. Rewrite: a struct with a function field, as `errors/closures_field_callee_mutable.nra` already does.
- Also changed by D81 though the type is allowed: `unused_mutable_output.nra` tests that a `mutable` parameter only assigned counts as used; with no whole assignment the output becomes a field write (`out.total = 3`), and README line "a `mutable` parameter that is only assigned counts as used" needs the same rewording.

docs/guide (2): `docs/guide/09-memory-and-mutable.md:399` `reporter(mutable score: int)` (the E0613 sample; use a struct, or show it on an array) and `docs/guide/11-narra-for-typescript-programmers.md:608` `addOne(mutable count: int)` (keep `nextCount` at `:612`, drop `addOne`, or count in a struct field).

### Assignments to a whole `mutable` parameter

- selfhost/src (3): `selfhost/src/patterns.nra:364` (`limit`, int, above), `selfhost/src/declsdump.nra:276` (`nextImpl`, int, above), `selfhost/src/modules.nra:237` `sortStrings`: `items = sorted` on an `Array<string>`. Rewrite: sort in place, the body of `selfhost/src/querygraph.nra:8` (which is already an in-place insertion sort with the same name and signature). Example 2 below.
- std/src: 0.
- compiler/testdata (28): 19 of the `int` parameters and 7 of the `Option` parameters above assign the parameter (all but `typesdump/bodies/err_statements.nra:7`, `typesdump/closures/err_closures.nra:42`, `errors/closure_borrow_mutable_param.nra:2`, `errors/closures_capture_mutable.nra:2` and `errors/let_narrowing_same_statement_param.nra:7`, which only read or capture it), plus `errors/closures_callee_mutable.nra:7` (`f = ...`, function type) and `errors/closures_field_callee_mutable.nra:11` (`h = ...` on a struct; rewrite to `h.run = ...`). `python3 count.py` prints the exact lines.
- docs/guide (1): `docs/guide/11-narra-for-typescript-programmers.md:609` (`count = count + 1`).

### Generic `mutable` parameters

- Bare type parameter: 0 in all four trees.
- Generic instances: `std/src/coroutine.nra` (`mutable g: Generator<T> uses E`, `mutable g: Coroutine<In, Out> uses E`), allowed. The selfhost `Array<...>` and `Map<...>` parameters are concrete.
- `Self` in a trait: `compiler/testdata/declsdump/traits/basic.nra:4` (implemented for struct `Rect`) and `declsdump/traits/err_signature.nra:2` (struct `A`), allowed.

### Tests and samples whose error goes away

A local `const` can now be changed through its contents, so these error samples become valid programs or lose a line:

- `compiler/testdata/errors/box_assign_const.nra` (`const b`, `b.value = 6`): becomes valid; delete or move to a run test.
- `compiler/testdata/typesdump/stdcalls/err_arrays_maps.nra:13` (`Arrays.push(mutable frozen, 3)` on a `const`): line 4 of `err_arrays_maps.dump` goes away.
- `docs/guide/02-values-and-types.md:430-438` ("A `const` array cannot change at all") and `docs/guide/11-narra-for-typescript-programmers.md:267-275` ("A Narra `const` is deep"): become valid.
- Still errors, with new hints: `docs/guide/09-memory-and-mutable.md:25` and `11-narra-for-typescript-programmers.md:976` (field write on a plain parameter; the hint "declare it with `let`" no longer fits), `09-memory-and-mutable.md:146` (top-level constant, unchanged).

### Guide chapters for mutable-bindings to rewrite

- `docs/guide/09-memory-and-mutable.md` (the main one): `:7`, `:20-45` (parameter hint), `:141-178` ("Add to the starting kit": a `const` local can now take the shield, the top-level constant still cannot; `:157` states the old argument rule), `:342-410` ("Closures and `let`", the `int` sample at `:399`), and every sample that uses `let` only to pass as `mutable`.
- `docs/guide/02-values-and-types.md`: `:29-32` stays (reassignment), `:77` (when to use `let`), `:425-440` (the deep-`const` sample).
- `docs/guide/03-functions-and-closures.md:309` (closures capture `const` "since they cannot change").
- `docs/guide/11-narra-for-typescript-programmers.md`: `:267-275` (`const` is deep), `:304-330` (copies; now the only difference from TypeScript `const`), `:576`, `:603-632` (`addOne`), `:940-991` (the comparison list at `:966-971` and the plain-parameter sample).
- `docs/guide/08-modules-and-packages.md` uses `mutable tallies: Array<Tally>` only: still valid, no change needed.

## Before and after

Example 1, an `int` budget becomes a struct field (`selfhost/src/patterns.nra:354`, `selfhost/src/match.nra:591`):

```ts
// before
export function findMissingCases(checker: Checker, rows: Array<Array<Pattern>>, types: Array<Type>, mutable limit: int): Array<Array<string>> uses alloc {
  if (limit <= 0) {
    return []
  }
  if (Arrays.length(rows) == 0) {
    limit = limit - 1
    ...
let limit = 10
const missing = findMissingCases(checker, rows, [subjectType], mutable limit)

// after
struct CaseBudget {
  remaining: int
}

export function findMissingCases(checker: Checker, rows: Array<Array<Pattern>>, types: Array<Type>, mutable budget: CaseBudget): Array<Array<string>> uses alloc {
  if (budget.remaining <= 0) {
    return []
  }
  if (Arrays.length(rows) == 0) {
    budget.remaining = budget.remaining - 1
    ...
const budget = CaseBudget { remaining: 10 }
const missing = findMissingCases(checker, rows, [subjectType], mutable budget)
```

Example 2, a whole assignment becomes an in-place sort (`selfhost/src/modules.nra:225`):

```ts
// before
function sortStrings(mutable items: Array<string>) uses alloc {
  let sorted: Array<string> = []
  for (const item of Arrays.values(items)) {
    Arrays.push(mutable sorted, item)
    ...
  }
  items = sorted                    // error under D81
}

// after (the body of selfhost/src/querygraph.nra:8)
function sortStrings(mutable items: Array<string>) {
  for (const start of Ranges.between(1, Arrays.length(items))) {
    let index = start
    while (index > 0 && items[index] < items[index - 1]) {
      const swap = items[index]
      items[index] = items[index - 1]
      items[index - 1] = swap
      index = index - 1
    }
  }
}
```

The caller side gains too: `let paths: Array<string> = []` at `selfhost/src/modules.nra` (and the other `let` locals that are only pushed to) can become `const`. At least 399 of the 1016 `let` locals in `selfhost/src` are never reassigned by name (per-file name scan, a lower bound). Changing them is optional; nothing forces it.

## Deviations

- Rule 1's type list is wider than the brief's (struct, Array, Map): the std and selfhost pass `StringBuilder`, generators and coroutines as `mutable`, and SPEC has `Box`, `FixedArray` and opaque std handles. Without them `Str.append` and `Co.resume` would break.
- SPEC sections beyond 2 and 6 were edited where they restated the old rule (sections 1, 4, 8, 10), including `Shared.update` in section 8.
- SPEC.md code blocks are not compiled by any test (`TestGuide` reads only `docs/guide`), so no sample needed an "unbuilt" marker. The guide was not edited: its samples are compiled against today's compiler, and mutable-bindings rewrites them with the build.

## Stubs left

- None. Nothing is built; node mutable-bindings builds D81.

## Gotchas

- `Local.Mutable` in the Go checker means two things that D81 splits: "can be reassigned" (`let`) and "contents can change" (`const`, `let`, `mutable` parameter, `for ... of` binding). Sites: `compiler/internal/narra/check_stmt.go:331-356` (AssignStmt, E0227 on `!root.Mutable`), `check_call.go:604-616` (`checkMutableArg`, E0315), `check_closure.go:142`, and `emit_closure.go:155-185` (`captureValue` and `envStruct` capture by pointer exactly when `Mutable`; rule 10 needs a by-reference flag per capture). E0290 in `check_const.go:138-142` stays for top-level constants. Selfhost: `selfhost/src/exprs.nra:1714` (E0315 text), `:1769`, `:1779` (E0290), `selfhost/src/stmts.nra:351` (E0227).
- Narrowing: a field path of a `const` can now change (`hero.weapon is Some(w)` then `hero.weapon = None`), so narrowing of `const` field paths must end at writes and `mutable` passes, as it does for `let` (#68 logic).
- The unused check treats a `mutable` parameter that is only assigned as used (README). With no whole assignment that case disappears; keep the rule for field writes.
- Pattern bindings are passed by pointer today (D50); rule 6 keeps them read-only so no write can reach the subject.
- New diagnostics need codes (none reserved here): parameter of a non-contents type, whole assignment to a `mutable` parameter, bare generic `mutable` parameter, `implement` of a trait with `mutable Self` for a non-contents type, an escaping closure changing a captured value. E0227 and E0315 messages need new hints (they say "declare it with `let`"), and E0315 now fires only for plain parameters, pattern bindings and top-level constants.
- DECISIONS.md jumps from D79 to D81; D80 is reserved for another node.

## Proposed learnings

- None.

## Verification

- `python3 roadmaps/selfhost/nodes/mutable-reference/count.py`: the counts above.
- `cd compiler && go vet ./... && go test -count=1 ./internal/narra -run 'TestGuide|TestSpec|TestErrorCodes'`: pass (98 s). No `TestSpec` exists and no test compiles SPEC.md blocks. The full suite was not run: no Go, `.nra` or testdata file changed.
- `check_prose.py` (plain-writing) on the new text: no errors, warnings for long sentences and semicolons only.

## Size

- Estimated S, actual S.
