---
name: fmt-printer
version: 1
status: frozen
defined_by: fmt-cli
---

# Contract: fmt-printer

## Purpose

How the statement printer of fmt-cli and the expression printer of fmt-exprs call each other. Narra modules cannot import each other in a cycle, and an arrow function body is a block while a statement holds expressions, so the two meet through function fields of one context struct, the pattern the checker uses for its walkers. Frozen at planning so fmt-cli and fmt-exprs can run at the same time.

## Interface

`selfhost/src/fmtprinter.nra` (written by fmt-cli; fmt-exprs writes the same file from this text if it needs it before fmt-cli lands, and takes fmt-cli's on rebase):

```ts
import * as Doc from "narrac/fmtdoc"
import { Trivia } from "narrac/parser"
import { Expr, Pattern, TypeExpr, Block } from "narrac/ast"

export struct Settings {
  options: Doc.Options
  trailingComma: bool          // true for "all"
}

export struct Printer {
  source: string
  trivia: Trivia
  settings: Settings
  printExpression: (printer: Printer, expression: Expr) => Doc.Doc
  printPattern: (printer: Printer, pattern: Pattern) => Doc.Doc
  printType: (printer: Printer, type: TypeExpr) => Doc.Doc
  printBlock: (printer: Printer, block: Block) => Doc.Doc
}

// The comments attached to an anchor, wrapped around a node's document:
// leading comments before it, trailing ones as a LineSuffix after it.
export function withComments(printer: Printer, anchor: string, document: Doc.Doc): Doc.Doc uses alloc
// Dangling comments of an anchor, or an empty Concat.
export function printDangling(printer: Printer, anchor: string): Doc.Doc uses alloc
// A trailing comma for a broken list, by settings.trailingComma.
export function trailingComma(printer: Printer): Doc.Doc uses alloc
```

`selfhost/src/fmtexprs.nra` (fmt-exprs) exports `printExpression`, `printPattern` and `printType` with the field types above. `selfhost/src/fmt.nra` (fmt-cli) builds the `Printer` and exports `printBlock`.

## Behavior

- Each print function handles the comments and parentheses of the node it prints (`withComments`, `Trivia.parenthesized`), so the caller never does.
- A print function returns a document with no leading or trailing line break; the caller places line breaks between items.

## Stub

Until fmt-exprs lands, fmt-cli fills `printExpression`, `printPattern` and `printType` with a function that copies the node's source text from `Trivia.ends` and its start position. fmt-exprs runs its golden files through `narrac fmt --stdin` once fmt-cli is on the story branch; before that it checks that the package builds and its functions type-check against this interface.

## Contract tests

The golden files of fmt-cli and fmt-exprs under `compiler/testdata/fmt/`.

## Changes

- v1: initial, frozen at planning (2026-11-05).
