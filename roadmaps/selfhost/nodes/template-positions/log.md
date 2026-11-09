# Log: template-positions

- Base 6924411 (origin/feat/batch-3), narrac built with `sh selfhost/bootstrap.sh <dir>`.
- Proof on base: `narrac check testdata/errors/template_position_type.nra` reports both errors at `3:9`, the backtick, although the bad parts are at 3:19 and 3:38.
- Fix: `lexFrom` in lexer.nra, `parseTemplate` places the tokens of each `${...}`; token ends go to `facts.templateTokenEnds` for narrac query.
- Branch: same command reports `3:19`, `3:38`, `3:38`.
- Text parts of a template keep the template's position; querysites skips their sites so the template site and the expression sites win.
- Full `sh scripts/test.sh` with the branch narrac: PASS.
