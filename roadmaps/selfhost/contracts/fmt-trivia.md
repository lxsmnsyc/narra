---
name: fmt-trivia
version: 2
status: frozen
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

## Changes

- v2 (2026-11-05, frozen by fmt-trivia, PR #76): the change request from fmt-trivia is accepted. Where the text above differs, these rules win. `narrac trivia [--check] <file>` prints the side tables (format in `selfhost/src/triviadump.nra`).

- The types `CommentKind`, `Placement`, `Comment`, `Attached` and `Trivia` are exported from `narrac/trivia` (`selfhost/src/trivia.nra`), not from `narrac/parser`. The lexer must build `Comment`, and importing it from the parser would be an import cycle. `parseWithTrivia` stays in `parser.nra`, and `lexWithComments` is in `lexer.nra`.
- Extra anchor key `b<line>:<col>` for a block (the `{` position). A comment in an empty block is dangling on the block.
- Extra `blankBefore` key `c<index>`: set when a blank line comes before comment `index`.
- `blankBefore` for a node means a blank line between the node and what comes before it: its last leading comment, else the previous sibling (with that sibling's trailing comments). So a first list item can have it, when a blank line follows its leading comment. Only statements, declarations, fields, variants, arms and imports get it, not expressions.
- Keys use the node's `pos` field, not its first token. A declaration's key is its keyword (after `export`); a labeled loop's key is the label; `return c ? a : b` uses the position of the converted `if`.
- `ends` and the ranges used for attachment of a parenthesized expression include its parentheses.
- Parentheses and comments inside `${...}` of a template string are not recorded (the part is lexed and parsed on its own). Type-parameter bounds, import names and map entries have no anchors.
- `parenthesized` holds expression ids only; Narra has no parenthesized patterns.
- A comment in an empty call `f(/* c */)` is trailing on the callee, not dangling on the call.
