---
id: tagged-templates
title: "Tagged template literals"
status: todo
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/tagged-templates/**
  - selfhost/src/taggedtemplates.nra
  - std/src/template.nra
  - testdata/tagged_*.nra
  - testdata/tagged_*.out
  - testdata/errors/tagged_*.nra
  - testdata/astdump/tagged_*
  - testdata/typesdump/tagged_*
  - testdata/fmt/tagged_*
  - testdata/query/tagged_*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/exprs.nra
  - selfhost/src/calls.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/consteval.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/stdtext.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1200-E1209]
provides: []
depends_on:
  - on: tagged-templates-design
    kind: hard
    why: the design gate (D114) fixes the call form, the raw parts type, the codes and the SPEC text this node builds
---

# Tagged template literals

## Goal

Build D114: `` tag`a${x}b` `` is the call `tag(["a", "b"], [x])`. The tag's first parameter type picks decoded parts (`Array<string>`) or source text (`Template.Raw` from a new std source module `std/template`, which also exports the `@const` tag `raw`). Codes E1200-E1203.

## Context

- Source: D114 in `DECISIONS.md` and the "Tagged templates" bullet of SPEC section 2, Literals. Read `roadmaps/selfhost/nodes/tagged-templates-design/outcome.md` first; the maintainer's answers there override this brief.
- Proof first: `tagged_basic.nra` fails on the batch base with `expected ")", found "a${name}b"` (checked at b23bbd4f).
- Today the lexer keeps a template's raw source (`lexTemplate`, `selfhost/src/lexer.nra`), and `parseTemplate` (`selfhost/src/parser.nra`) decodes escapes and splits the `${}` parts. A tag needs both texts, chosen only once the checker knows the tag's type.
- Feature logic goes in `selfhost/src/taggedtemplates.nra`; shared files carry dispatch rows only (batch-13 rule). Other batch-13 builds edit the parser, `ast.nra`, the dumps, fmt and `diagnostics.nra`: rebase and keep both sides.
- `std/src/template.nra` is a new std source module, so `selfhost/src/stdtext.nra` is regenerated (`sh scripts/test.sh --suite generated --update`). batch-14 regenerates the same file on its branch; whichever batch merges second regenerates again.
- Seed rule: do not use tagged templates in `selfhost/src` or seed-compiled std.

## Steps

- [ ] Parser. In `parsePostfix`, a `TokTemplate` right after an `Ident` or a `FieldExpr` chain of names (no newline token between) builds `CallExpr(callId, pos, tag, [], [Arg(TemplateParts), Arg(ArrayLit(values))])`. Share the splitting loop of `parseTemplate`: each part keeps its raw text, its decoded text or the first escape error (code, message, position, not reported yet), and its position. Empty parts are kept, so there are always `values + 1` parts. Record the call id in a new fact, `taggedCalls`, and set `callEnds` as for other calls. A template after any other postfix expression or a literal is E1200. A statement that starts with a template, when the previous line ends with a name, is E1203 (keep the old "this expression does nothing" for every other case).
- [ ] AST. One new leaf, `Expr.TemplateParts(id: int, pos: Pos, parts: Array<TemplatePart>)` with `struct TemplatePart { raw: string, decoded: Result<string, ...>, pos: Pos }` (pick the error type the lexer's `decodeEscape` already returns). It has no children, so walkers with a default arm need nothing; add a case only where the compiler matches leaves exhaustively (`check`, `emitexprs`, `consteval`, `astdump`, `typesdump`, `fmtexprs`, `idcheck`, and any other site the build reports).
- [ ] Checker (`taggedtemplates.nra`, called from the call check for ids in `taggedCalls`). The callee's type must be a function of two non-`mutable` parameters, the first `Array<string>` or the distinct type `Raw` of module `std/template`, the second `Array<T>`, else E1201 at the tag naming the found type. Check each hole with `T` as expected type (generic `T` from the first hole whose type is known); a mismatch is E1202 at the hole, naming its index from 1, the hole type and `T`. With `Array<string>`, report the first stored escape error of any part with its own code (E0005, E0008, the `\u{}` codes). Record raw or decoded per `TemplateParts` id in a side table. Effects: the call's effects plus `alloc` at the template.
- [ ] Emitter: `TemplateParts` emits an array of string literals (raw or decoded per the side table), one array allocation; the values are the ordinary array literal. `await` inside holes follows the existing array-literal spill.
- [ ] Evaluator (`consteval.nra`): `TemplateParts` evaluates to the array of strings, so an `@const` tag in a top-level constant runs (D105).
- [ ] `std/src/template.nra`: `export distinct type Raw = Array<string>` and `export @const function raw(parts: Raw, values: Array<string>): string uses alloc`, joining parts and values in order. Add `std/template` to the module list the registry and SPEC section 3 already have, regenerate `stdtext.nra`.
- [ ] fmt: print the tag, no space, then the template text as written (the same copy plain templates use, `getEndedText`). astdump: the call line gains a `tagged` mark, and `TemplateParts` prints as a `template-parts` node with one line per part (`raw` text, quoted); document it in the doc comment of `astdump.nra`.
- [ ] Language service and `narrac query`: check that hover and go to definition work on the tag, that `callers` and references list tagged uses, and that positions inside holes are right (they come from `templateTokenEnds`). Add a `testdata/query` case.
- [ ] Codes: add E1200-E1203 rows and `narrac explain` texts to `diagnostics.nra`.
- [ ] SPEC: change the fence of the tagged-template sample in section 2 from `text` to `ts` (with a `<!-- context -->` block for `name` and `digits`), drop "not built yet" in the bullet, the differences table row and the module list. Run the `spec` suite.
- [ ] `compiler/README.md`: tagged templates under supported features.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check:
  - `tagged_basic.nra`: a tag with `Array<string>` values; parts with escapes (`\n`, `` \` ``), an empty first and last part, zero holes (`` tag`x` `` gives `["x"]` and `[]`), a dotted tag from another module, a function value as the tag.
  - `tagged_raw.nra`: a `Template.Raw` tag and `Template.raw`: `` Template.raw`\d+\${x}` `` prints `\d+\${x}`; a part spanning two lines; `Template.raw` in a top-level `const`, computed while compiling.
  - `tagged_generic.nra`: `function list<T: Show>(parts: Array<string>, values: Array<T>): string` with `int` holes and with a struct that derives `Show`; a hole `None` against `Array<Option<int>>`.
  - `tagged_effects.nra`: a tag that `uses io` called from a function that declares it; a tag returning a non-string type (a struct).
- [ ] Error tests in `testdata/errors/`, each with its `// error:` line:
  - E1200: `` f()`x` ``, `` items[0]`x` ``, `` "a"`x` ``.
  - E1201: a tag of one parameter, a tag whose first parameter is `string`, a tag with a `mutable` parameter, a tag that is an `int` local.
  - E1202: an `int` hole for an `Array<string>` tag; a generic tag with an `int` hole and a `string` hole.
  - E1203: `const x = tag` then `` `a` `` on the next line.
  - E0008 at the backslash of `` tag`\d` `` for an `Array<string>` tag; the same template with a `Template.Raw` tag compiles.
  - An effect error: a tag that `uses io` called from a function without `io`.
- [ ] `testdata/astdump/tagged_*` and `testdata/fmt/tagged_*` goldens; fmt is stable on them; `tag `x`` formats to `` tag`x` ``.
- [ ] The `spec`, `codes`, `generated` and `sources` suites pass.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: about 600 to 800 lines (parser 120, checker 200, emitter and evaluator 80, std module 30, dumps and fmt 100, tests); one new leaf instead of a new node with children keeps the 33 template walker arms unchanged.

## Notes for the worker

- Do not add an `Expr` kind with children for the tagged form (D114 rejects it): every walker that recurses through templates would need a second arm, and a missed one skips the holes silently.
- The parts are checked by the tag's first parameter type, not by syntax. Do not report escape errors in the parser for a tagged template.
- `Template.Raw` is found by module path and name (`std/template`, `Raw`), as `std/shape`'s `Value` is found for `Shape`.
- No explicit type arguments on a tag (`` tag<int>`x` `` stays a less-than and fails as it does today).
- Maintainer questions open at the gate are in the design outcome; if the maintainer accepts positional value parameters or per-site `@const` checks, they are follow-up nodes, not this one.
