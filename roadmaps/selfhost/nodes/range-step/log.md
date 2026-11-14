# Log: range-step

## Proof first

- Base: feat/batch-15 at d64eee66, compiler built from a clean `git archive` with `sh selfhost/bootstrap.sh`.
- Finding: `Ranges.stepped` already existed and returned `Option<Range>` (None for step 0). D119 changes it to return `Range`.
- Command: `narrac run testdata/range_step.nra` (the new sample).
- Result on base: `range_step.nra:28:31: expected Range, found Option<Range>` (and the same on lines 29 to 35); no output.
