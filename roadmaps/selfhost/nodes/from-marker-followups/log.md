# Log: from-marker-followups

Base: feat/batch-6 at 465201f5, built with `sh selfhost/bootstrap.sh <dir>`. Probe command: `narrac emit <file>`.

- Part 1 (E0976 wording) reproduced. `struct Row {cells: Array<int>}`, inside `run`: `const pick = (r: Row) => r.cells`. Base: `the arrow function returns a value that shares its parameter "r", which its function type does not allow; return a copy with copy(...)`. A top-level `const f = (...)` is a different error (closures are not constants), so the probe sits in a function.
- Part 2 (E0047) reproduced. `function g(c: Array<int>): Array<int> from {` reports with code E0976 (diagnostics table row and parser call).
- Part 3 (buildLocalName) reproduced. `const MaxValue = 3` then `MaxValue = 5`: base suggests `let maxvalue = MaxValue`; `const TEST` suggests `let test = TEST`.
- Part 4 (recursive struct) reproduced as a missing sample: `struct S { a: S  b: S }` gives E0109 on base without hanging, and no sample covered it.
- Fix 1: sharedresults.nra picks the new message when the arrow has no written result type, the arrow type carries no from, and no expected type asked for a fresh result.
- Fix 3: stdcopy.nra keeps camel case for names with lowercase letters (`maxValue`, `HTTPServer` gives `httpServer`), and appends `Copy` to a keyword result (`testCopy`). The keyword test lexes the candidate name.
- Samples: errors/from_arrow_no_type, const_local_name_mixed, const_local_name_keyword, recursive_struct_twice; none for E0047 (a diagnostics sample needs an autofix, so errors/from_missing_name covers it).
