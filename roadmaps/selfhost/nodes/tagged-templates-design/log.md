# Log: tagged-templates-design

2026-11-13, base `feat/batch-13` at b23bbd4f. Compiler built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/tagged-templates-design/out` (fixed point, seed current).

## Checks on the base

- `` tag`a${name}b` `` with `tag(parts: Array<string>, values: Array<string>)`: `t1.nra:9:12: expected ")", found "a${name}b"`. The syntax is free today.
- `` `a\db` ``: `unknown escape \d` (E0008), reported by the parser. The lexer only skips `\X` pairs and keeps the raw source (`lexTemplate`, `skipTemplateBody` in `selfhost/src/lexer.nra`); `parseTemplate` in `selfhost/src/parser.nra` decodes. So raw text is available to a tagged form without a lexer change, but escape errors must move out of the parser for tagged templates, because the tag's type picks raw or decoded.
- `const x = tag` then `` `a` `` on the next line: `this expression does nothing`. Narra already rejects it; D114 gives it E1203 with a message that names the TypeScript reading.
- `parseTemplate` drops empty text parts (`pending > 0`); the tagged form must keep them so there are `values + 1` parts.

## Counts

- `rg -c "Expr.TemplateLit" selfhost/src`: 33 arms in 24 files (effects, closures, sharedwrites, emitasyncspill, consteval, operands, inlining, audit and others). A new expression kind with children would need an arm in each.
- `Expr.StringLit` is named in 19 files, `Expr.CharLit` in 18: most walkers treat leaves through a default arm. So a leaf `TemplateParts` inside an ordinary `CallExpr` is the cheaper AST.

## Other inputs

- D12 (one call form, no varargs, no overloading), D105 (`@const` only in top-level constants), D84 (`Shape`/`Value`), SPEC section 4 distinct types (a distinct type over `Array<int>` exists in `testdata/distinct_generic_base.nra`), `Show` trait for generic tags.
- std-design (batch-14) item 7 says std/terminal decides tags "if D114 adds them". D114 leaves terminal and shell to it.

## Suites

- `sources`: PASS, 1 case. `spec`: PASS, 62 cases (the new SPEC sample uses a `text` fence, so it is not checked until the build node switches it to `ts`; `selfhost/tests/spec-skip.txt` is outside this spike's paths).
