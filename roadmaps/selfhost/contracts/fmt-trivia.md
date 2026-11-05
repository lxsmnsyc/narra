---
name: fmt-trivia
version: 1
status: draft
defined_by: fmt-trivia
---

# Contract: fmt-trivia

## Purpose

The source details the AST drops, kept in side tables so `narrac fmt` (and later the language service) can see them: comments, blank lines and written parentheses. Built by `selfhost/src/lexer.nra` and `selfhost/src/parser.nra` only when asked for.

## Interface

```ts
// parser.nra
export function parseWithTrivia(source: string, firstId: int): Result<#[Parsed, Trivia], Diagnostic> uses alloc

export enum CommentKind { Line, Block, Doc }   // `//`, `/* */`, `///`
export enum Placement { Leading, Trailing, Dangling }

export struct Comment {
  kind: CommentKind
  text: string        // as written, with the markers
  pos: Pos            // start
  end: Pos            // after the last character
  ownLine: bool       // nothing but whitespace before it on its line
}

export struct Attached {
  comment: int        // index into Trivia.comments
  placement: Placement
}

export struct Trivia {
  comments: Array<Comment>                // source order
  attached: Map<string, Array<Attached>>  // by anchor key, in source order
  blankBefore: Map<string, bool>          // anchors with a blank line before them
  parenthesized: Map<int, bool>           // expression and pattern ids written inside ( )
  ends: Map<string, Pos>                  // anchor key to the position after the node's last token
}
```

Anchor keys: `"e<id>"` for an expression, pattern or parameter (they have node ids). Statements, declarations, type expressions, fields, variants, arms and imports have no id, so their key is a kind letter and the start position: `"s<line>:<column>"` (statement), `"d..."` (declaration), `"t..."` (type), `"f..."` (struct field, field init or field pattern), `"v..."` (variant), `"a..."` (match arm), `"i..."` (import). Adding ids to those nodes would change the id sequence that `narrac pkg --ids` and the dumps compare with Go, so they stay keyed by position.

## Behavior

- `parse` and `lex` are unchanged: same tokens, same ids, same AST, same errors. `parseWithTrivia` returns the same `Parsed` as `parse` for the same input and `firstId`.
- Every comment is attached exactly once, as Prettier does: to the enclosing node's following node as `Leading` when it is on its own line or before the node on the same line, to the preceding node as `Trailing` when it follows it on the same line, and `Dangling` on the enclosing node when the node has no children to attach to (an empty block, list or literal).
- `blankBefore` holds an anchor when at least one empty line (ignoring comment lines that attach to it) separates it from the previous sibling in the same list. The first item of a list never has it.
- `parenthesized` records only the outermost of nested parentheses `((a))`; the printer prints at most one pair.

## Stub

None; fmt-trivia lands before its consumers start.

## Contract tests

`narrac trivia <file>` prints the side tables in a fixed text format; golden files under `compiler/testdata/fmt/trivia_*`, added by fmt-trivia.

## Changes

- v1: initial draft.
