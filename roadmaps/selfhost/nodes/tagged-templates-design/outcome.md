# Outcome: tagged-templates-design

## Built

- D114 in `DECISIONS.md`, "Proposed, waiting for the maintainer".
- SPEC section 2, Literals: a "Tagged templates" bullet and sample (fence `text` until built), a row in "Differences from TypeScript", `std/template` in the module list of section 3. All marked not built yet.
- `selfhost/src/diagnostics.nra` header: `E1200-E1209  tagged templates`.
- `nodes/tagged-templates/brief.md` refined into a full brief (size M, not split). Codes E1200-E1203 go to it; E1204-E1209 stay free.

Answers:

1. Signature: `tag(parts, values: Array<T>)`, one `T` per tag; holes of mixed types convert at the hole. Rejected: a `Value` per hole (hidden `Shape.toValue` call), one tag function per type (overloading), positional value parameters (fits one hole count; can come later).
2. Raw: the tag's first parameter type picks. `Array<string>` gets decoded parts with the plain template escapes; `Template.Raw` (`distinct type Raw = Array<string>` in new `std/template`) gets the source text, no escape checked. `Template.raw` is TypeScript's `String.raw`, `@const`.
3. Effects and `@const`: no new rule. Tag effects count at the use site plus `alloc`; an `@const` tag runs in a top-level constant under D105.
4. std tags: only `Template.raw`. No JSON tag. Shell and terminal left to std-design.
5. Tooling: the template is an ordinary `CallExpr` whose first argument is a new leaf `TemplateParts`, so hover, definition, references, callers and holes work as for calls and templates; fmt copies the text after the tag with no space; astdump marks the call `tagged`.

## Deviations

- None.

## Stubs left

- None. The SPEC sample is a `text` fence; the build node switches it to `ts`.

## Gotchas

- The parser must keep escape errors of a tagged template, not report them: only the checker knows whether the tag takes raw parts.
- `parseTemplate` drops empty parts today; the tagged form keeps them.
- `std/template` is a new std source module, so `stdtext.nra` is regenerated, which batch-14 also does.

## Questions for the maintainer

1. Accept one value type per tag (`Array<T>`) and explicit conversion at holes? Recommended: yes. Positional value parameters can be added later.
2. Per-site compile-time checks of the parts (a regex or SQL tag that rejects a bad pattern where it is written, with run-time holes)? Recommended: not now; D105 runs `@const` only for top-level constants. A later design could let a tag name an `@const` check over its parts.
3. Raw parts through the `Template.Raw` parameter type, in a new module `std/template`? Recommended: yes.
4. E1203 for a template on the line after a name (TypeScript reads it as tagged)? Recommended: yes, it replaces a vaguer message for the same code.

## Proposed learnings

- A new syntax form that desugars to an existing node plus a leaf avoids a second arm in every walker | 33 template arms in 24 files | selfhost/src | no

## Verification

- `NARRAC=/private/tmp/claude-501/tagged-templates-design/out/narrac sh scripts/test.sh --suite sources`: PASS, 1 case.
- `... --suite spec`: PASS, 62 cases.

## Size

- Estimated S, actual S.
