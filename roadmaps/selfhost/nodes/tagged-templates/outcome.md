# Outcome: tagged-templates

## Built

- D114 tagged templates. `` tag`a${x}b` `` parses as `CallExpr(tag, [TemplateParts, ArrayLit([x])])`, with the call at the template's position.
- AST (`selfhost/src/ast.nra`): leaf `Expr.TemplateParts(id, pos, parts: Array<TemplatePart>)`, `struct TemplatePart { raw, decoded: Result<string, TemplatePartError>, pos }`, `struct TemplatePartError { code, message, pos }`, fact `SourceFacts.taggedCalls: Map<int, bool>` (call ids), and `isTaggedArguments(arguments)`, the structural test fmt, astdump and the checker use.
- Parser (`selfhost/src/parser.nra`): `splitTemplate` is the one splitting loop for plain and tagged templates (a plain template still fails at its first bad escape and takes its StringLit ids in the old order, so its AST and C do not change); `parseTaggedTemplate` keeps every part, empty ones too, with the first escape error unreported. In `parsePostfix` a template on the same line as a name or dotted name is tagged; after any other expression or a literal it is E1200. A template that starts a line after a line ending with a name is E1203 (`isTemplateAfterNameLine`).
- `selfhost/src/taggedtemplates.nra`: `checkTag` (E1201, before the ordinary call check, from `checkCall` in `calls.nra`), `checkTemplateParts` (type `Template.Raw` when the parameter is `Raw` of `std/template`, else `Array<string>` and the first stored escape error with its own code), `checkTemplateValues` (each value against `T`, E1202 naming the value from 1; a generic tag takes `T` from the first value), `isRawParts`, and `emitTemplateParts` (one array of static literals). The values array is found by `Checker.taggedValues`, set by `checkTag` and cleared when the array is checked.
- `consteval.nra`: `evaluateTemplateParts`, so `const DIGITS = Template.raw`\d+\${x}`` is a static literal in the C (checked in the emitted C of `tagged_raw.nra`: the getter retains `L11`, no call to `raw`).
- `std/src/template.nra`: `export distinct type Raw = Array<string>` and `@const export function raw(parts: Raw, values: Array<string>): string uses alloc`. Row `TEMPLATE_FUNCS` in `selfhost/src/stdlib.nra`; `stdtext.nra` regenerated.
- fmt (`fmtexprs.nra`, `printCall`/`printTaggedTemplate`): the tag, no space, then the template text as written. astdump: `call ... tagged`, then `template-parts` with one `part line:column raw="..."` line per part (documented in the header of `astdump.nra`). typesdump: `template-parts`.
- Review fixes: a void or `panic(...)` value is E1202 ("value N of the tagged template is void"), unless its own check already failed; a field of a local (`tags.tag`, through plain struct fields) is checked for E1201 like a name; an open `T` comes from the first value that is not `None`, `[]` or `Map {}`; a void local as the tag adds no second error; `` tag<int>`x` `` is one E1200 ("a tag takes no type arguments"). After E1201 the call takes the expected type, or the tag's result, so no "found void" follows.
- Codes E1200-E1203 with `narrac explain` texts in `diagnostics.nra`. SPEC section 2: sample fence is `ts` (moved to column 0 so the spec suite reads it) with a context block; "not built yet" removed from the bullet, the differences row and the module list. `compiler/README.md`: a Supported line.
- Tests: `testdata/tagged_{basic,raw,generic,effects}.nra`, 17 `testdata/errors/tagged_*`, `testdata/astdump/tagged_templates`, `testdata/fmt/expr_tagged_templates`, `testdata/query/tagged_templates` (def, refs, callers, type-of on the tag and inside nested holes), `testdata/packages/tagged_modules` (dotted tag and named import from another module, raw parts).

## Proof

- base: feat/batch-13 at 7df2cd63, `narrac run testdata/tagged_basic.nra` exits 1: `tagged_basic.nra:30:17: expected ")", found "a${name}b"`.
- branch: the same command prints the expected output; the four samples pass in the debug build (AddressSanitizer, leak check, copy check) and the release build.

## Deviations

- Files outside the brief's lists, each a needed arm or row: exhaustive `Expr` matches in `trivia.nra`, `idcheck.nra`, `check.nra` (two id/pos lists and the `taggedValues` field of `Checker`), `audit.nra`, `inlining.nra`, `emitasyncspill.nra`, `fmtprinter.nra`, `querysites.nra`; one table and one `registerAll` line in `stdlib.nra`.
- `TemplatePart.decoded` uses a new `TemplatePartError` (code, message, pos) and not the lexer's `LexError`: `lexer.nra` imports `ast.nra`, so `ast.nra` cannot import it.
- The fmt golden is `testdata/fmt/expr_tagged_templates.*`: the fmt suite reads only files with the `expr_`/`stmt_`/... prefixes, so a `tagged_*` name would never run.
- The dotted tag from another module is a package (`testdata/packages/tagged_modules`), since a single-file sample cannot import a package module.
- The checker finds tagged calls by `isTaggedArguments` (first argument is a `TemplateParts`), not by the `taggedCalls` fact; the fact is filled for tooling.
- `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` list `std/template` among the supported modules now.
- The astdump part line carries its position as well as `raw`.

## Stubs left

- None.

## Gotchas

- `std/template` is a new std source module: whichever batch merges second regenerates `stdtext.nra` (batch-14 also edits it).
- A tag's mutability is known only for named functions; a function value's type carries no `mutable`, so E1201 for a mutable parameter applies to declared functions.
- A builtin (std function written in C) as a tag gets no E1201; the ordinary call check reports it.
- The seed does not compile tagged templates; `selfhost/src` and seed-compiled std must not use them until the batch's seed update.

## Proposed learnings

- A desugared form that a later pass must recognise can be found by the shape of its arguments (first argument is the new leaf), so walkers need no fact map | `isTaggedArguments` serves fmt, astdump and the checker | selfhost/src | no

## Verification

- `sh scripts/test.sh` (from the seed, rebased on feat/batch-13 14af8001): see the final report.
- `narrac emit selfhost` of the base tree with the base compiler and with the branch compiler: identical C. The branch's own tree differs only by `std/template` in `stdtext.nra`.

## Size

- Estimated M, actual M (about 750 lines of compiler and std, plus tests).
