# Change request: fmt-trivia v1 to v2

- From node: fmt-trivia
- Status: open

## Problem

The implementation differs from `contracts/fmt-trivia.md` v1 in these ways. The contract file is not edited.

- The types `CommentKind`, `Placement`, `Comment`, `Attached` and `Trivia` are exported from `narrac/trivia` (`selfhost/src/trivia.nra`), not from `narrac/parser`. The lexer must build `Comment`, and importing it from the parser would be an import cycle. `parseWithTrivia` stays in `parser.nra`, and `lexWithComments` is in `lexer.nra`.
- Extra anchor key `b<line>:<col>` for a block (the `{` position). A comment in an empty block is dangling on the block.
- Extra `blankBefore` key `c<index>`: set when a blank line comes before comment `index`.
- `blankBefore` for a node means a blank line between the node and what comes before it: its last leading comment, else the previous sibling (with that sibling's trailing comments). So a first list item can have it, when a blank line follows its leading comment. Only statements, declarations, fields, variants, arms and imports get it, not expressions.
- Keys use the node's `pos` field, not its first token. A declaration's key is its keyword (after `export`); a labeled loop's key is the label; `return c ? a : b` uses the position of the converted `if`.
- `ends` and the ranges used for attachment of a parenthesized expression include its parentheses.
- Parentheses and comments inside `${...}` of a template string are not recorded (the part is lexed and parsed on its own). Type-parameter bounds, import names and map entries have no anchors.
- `parenthesized` holds expression ids only; Narra has no parenthesized patterns.
- A comment in an empty call `f(/* c */)` is trailing on the callee, not dangling on the call.

## Proposed change

Freeze the interface as implemented: types in `narrac/trivia`, `parseWithTrivia` in `narrac/parser`, the `b` and `c` keys, and the blankBefore meaning above. `narrac trivia [--check] <file>` prints the side tables (format in `selfhost/src/triviadump.nra`).

## Affected nodes

fmt-cli, fmt-exprs (printers), language-service (doc comments).

## Workaround in place

None needed; the implementation is the proposal.

## Decision

