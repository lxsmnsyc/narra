---
name: fmt-doc
version: 2
status: frozen
defined_by: fmt-doc
---

# Contract: fmt-doc

## Purpose

The document IR and printer of `narrac fmt`, a pure library in `selfhost/src/fmtdoc.nra` (`import * as Doc from "narrac/fmtdoc"`). fmt-cli and fmt-exprs build documents with it; nothing else in the compiler uses it.

## Interface

```ts
export enum Doc {
  Text(text: string)              // no line breaks inside
  Concat(parts: Array<Doc>)
  Indent(contents: Doc)           // one level deeper for lines inside
  Group(contents: Doc, id: int)   // flat if it fits in the remaining width, else broken; id -1 for none
  Line                            // a space when flat, a newline when broken
  SoftLine                        // nothing when flat, a newline when broken
  HardLine                        // always a newline, and breaks every enclosing group
  IfBreak(broken: Doc, flat: Doc, groupId: int) // by the enclosing group, or by groupId when not -1
  LineSuffix(contents: Doc)       // printed before the next newline (trailing comments)
  BreakParent                     // breaks the enclosing groups
}

export struct Options {
  printWidth: int
  tabWidth: int
  useTabs: bool
}

export function text(value: string): Doc uses alloc
export function concat(parts: Array<Doc>): Doc uses alloc
export function join(separator: Doc, parts: Array<Doc>): Doc uses alloc
export function indent(contents: Doc): Doc uses alloc
export function group(contents: Doc): Doc uses alloc
export function groupWithId(contents: Doc, id: int): Doc uses alloc
export function print(document: Doc, options: Options): string uses alloc
```

Constructors for `Line`, `SoftLine`, `HardLine` and `BreakParent` are the enum variants (`Doc.Line`). The worker may add helpers; the names above stay.

## Behavior

- Printing follows Wadler's "prettier printer" and Prettier's `printDocToString`: a group prints flat when its flat width fits in `printWidth` minus the current column, measured up to the next possible line break after it; otherwise its `Line`s break.
- A group that contains a `HardLine` or `BreakParent` is always broken.
- Width counts characters. Indentation is `tabWidth` spaces per level, or one tab per level with `useTabs` (a tab counts as `tabWidth` for fitting).
- No trailing whitespace on any line; the output ends with exactly one newline when the document ends with a `HardLine`.
- The printer is iterative (an explicit stack), so deep documents do not overflow the C stack.

## Stub

None; fmt-doc lands before its consumers start.

## Contract tests

Unit tests for each rule above, added by fmt-doc (see its brief for how they run).

## Changes

- v1: initial draft.

## Changes

- v2 (2026-11-05, frozen by fmt-doc, PR #75): the helpers declare `uses alloc`, because building an enum value allocates. Added `groupWithId`. Inside `fmtdoc.nra` the exported `print` wraps `printDocument`, because there the name `print` would mean the prelude function. The output ends with a newline only when the document ends with a `HardLine`, so fmt-cli ends every file document with one.
