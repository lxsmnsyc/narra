# Outcome: from-marker-followups

## Built

- `selfhost/src/sharedresults.nra`: an arrow with no written result type, no expected type and a shared parameter now says to write the result type with `from <name>` or return a copy (new E0976 row in `diagnostics.nra`).
- `selfhost/src/parser.nra`, `selfhost/src/diagnostics.nra`: the parse error "`from` needs the name of a parameter" uses E0047.
- `selfhost/src/stdcopy.nra`: `buildLocalName` keeps camel case for names with lowercase letters (`MaxValue` gives `maxValue`) and appends `Copy` when the result is a keyword (`TEST` gives `testCopy`). New helpers `lowerLeadingCapitals`, `buildCamelCase`, `isKeywordName` (lexes the name).
- Samples: `testdata/errors/{from_arrow_no_type,const_local_name_mixed,const_local_name_keyword,recursive_struct_twice}`; E0047 keeps its text, so the existing `testdata/errors/from_missing_name` covers it.

## Proof

- base (465201f5, `narrac emit` on the probes): arrow without result type says "which its function type does not allow"; `from {` reports E0976 and the table has no E0047; `MaxValue` suggests `let maxvalue`, `TEST` suggests `let test`; `struct S { a: S, b: S }` gives E0109 but no sample existed (it does not hang on base).
- branch: the arrow message says "write the result type with `from r`, or return a copy"; E0047; `let maxValue = MaxValue`, `let testCopy = TEST`; the recursive sample passes.

## Deviations

- None. A diagnostics sample for E0047 was dropped: the habits suite requires an autofix there.

## Stubs left

- None.

## Gotchas

- The new arrow message applies only when the arrow type carries no `from` and no expected type asked for a fresh result; the old message stays for an expected function type without a matching `from`.

## Proposed learnings

- None.

## Verification

- full `sh scripts/test.sh`: all suites passed except habits on my extra diagnostics sample, which I removed; habits re-run passes.

## Size

- Estimated S, actual S.
