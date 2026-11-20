# Outcome: guide-compile-time

## Built

Chapter 17 (`docs/guide/17-proofs-and-compile-time.md`) has a new last section, "Values computed while compiling", before the `Next:` line. It keeps the week-of-working-hours example:

- `@const` functions and their limits (row 166): a table of planned hours computed for a top-level `const`, the same function called at run time, the effects, parameter and call limits, a `narra error` for calling a function that is not `@const`, a `narra error` for a panic during the computation, and the step, memory and call limits in prose.
- "Patterns checked while compiling" (row 75): `Regexes.checked` with a run sample, a `narra error` for a bad pattern, and a `narra error` for a call inside a function.
- "Tagged templates and `Template.raw`" (row 38): the tag call form, a tag with `Array<string>` parts, the hole type error, `Raw` parts, and `Template.raw` feeding `Regexes.checked` in a constant.

## Deviations

- The 10-million-step limit is shown in prose with the error text, not as a `narra error` sample, since a sample would make the guide suite compute for a long time.
- Left out of the section: the `@const` float literal limits, E1057 (a constant read while it is computed), and the newline rule for tagged templates (E1203). They are rare for a learner.
- `docs/guide/README.md` line 66 still lists "`@const` functions, `ensures` lines and refinement types" under "built but not taught". It is outside this node's owned paths; its owner should remove that line.

## Spec against compiler

No claim differs. Checked with the prebuilt narrac: the E1052 message for a call to a non-`@const` function, the step limit message, a panic in a constant with the call chain, `Regexes.checked` with a bad pattern and inside a function, the E1202 hole type error, and `Template.raw` inside `Regexes.checked`. Note for spec-corrections: the SPEC says a bad pattern is E1053 and a use in a function is E1360, but the messages carry no code in the text the guide tests match.

## Stubs left

- None.

## Gotchas

- Error samples are matched by line and column. The first line of a `narra error` sample is the `// error:` comment, so the lines in the message count it.
- The tag error sample needs the tag body to compile; otherwise a second error comes first in the output.

## Proposed learnings

- None.

## Verification

- `NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 559 cases.
- `pnpm --dir website build`: passes.
- Every sample checked with `narrac check` and `narrac run` in a scratch folder outside the repo.
- `check_prose.py` on the chapter: `prose: ok`.

## Size

- Estimated S, actual S.
