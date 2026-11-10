# Outcome: value-semantics-docs

## Built

- **SPEC.md section 6**: new subsection "Shared writes" before "Shared results". It states the E0978 rule (what writes, what shares, when sharing ends, which index places are the same, the known gap, where it runs) with three checked samples: an error, its `copy(...)` fix, and two different literal indexes accepted.
- **DECISIONS.md**: D94's status line says every part is built and points to SPEC.md section 6. D81 rule 5 and D27 each gain an "Amended by D94" line.
- **Guide**:
  - Chapter 9: intro and "Why copies are cheap" use `copy(...)`; new section "Writing to a shared value" with an error sample, the fix sample, the rule's details, and the index rule. The last section says copy on write applies to values `copy(...)` shares.
  - Chapters 2 and 4: the prose that said assignment copies now says `copy(...)` does, and that a write after plain assignment is rejected.
  - Chapter 11: "Copies instead of shared references" rewritten around `copy(...)`; the sentence that the compiler cannot warn is replaced by the comparison with JavaScript's shared references and a checked error sample.
  - Guide README: chapter 9 line mentions shared writes.

## Proof

- `--suite guide` and `--suite spec` pass with the new samples (the error samples must match the compiler's message and position).

## Deviations

- None. Chapter 8 had no prose about shared copies (only a sample edited by shared-write-error).

## Stubs left

- None.

## Gotchas

- Chapter 4 still says "You can assign a field only through a `let` name", which chapter 9 contradicts (a `const` struct's fields can be assigned). Not changed here.

## Verification

- See the final report: full `sh scripts/test.sh`.

## Size

- Estimated S, actual S.
