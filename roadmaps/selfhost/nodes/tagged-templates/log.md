# Log: tagged-templates

## Proof first

- Base: `feat/batch-13` at 7df2cd63, compiler from `sh selfhost/bootstrap.sh` (release).
- `narrac run testdata/tagged_basic.nra` exits 1: `tagged_basic.nra:30:17: expected ")", found "a${name}b"`.

## Notes

- Parts exist before this node: none. The lexer already kept the raw template text; `parseTemplate` was the only splitter.
- `parseTemplate` now calls `splitTemplate`. A plain template takes the id of each StringLit when its part ends, as before, so ids and the C of untagged code do not move.
- The parser cannot import `taggedtemplates.nra` (it imports `check.nra`, which other modules that import the parser need), so the parsing lives in `parser.nra`. `exprs.nra`, `calls.nra`, `emitexprs.nra` and `consteval.nra` import `taggedtemplates.nra`; it imports only `check`, `funcvalues`, `scopes`, `cnames`, `emitstate` and `textops`.
- The call's position is the template's backtick, as a plain call's is its `(`.
- After E1201 the call takes the expected type, so `print(tag`x`)` does not add "expected string, found void".
- SPEC: the spec suite reads only fences at column 0, so the sample moved out of the bullet's indent; `quoted` is `@const` so the two top-level constants check (E0192 otherwise).
- `Template.raw` in a top-level constant: the emitted C of `tagged_raw.nra` builds `DIGITS` from the static literal `"\\d+\\${x}"`, with no call to `raw`.
- An `await` in a hole works through the existing array spill (`/private/tmp/claude-501/tagged-templates/awaits.nra`, clean under `-sanitize`).
- Rebased on feat/batch-13 0bab063f (threads-process landed) with no conflict; base compiler rebuilt from that commit for the emit comparison.
