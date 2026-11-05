# Change request: query-output v1 to v2

- From node: query-cli
- Status: open

## Problem

The v1 draft was written before `narrac query` existed. The built command (PR #88, `selfhost/src/query.nra`, goldens in `compiler/testdata/query/*/queries.txt`) needs some rules the draft does not have:

- The draft does not say what `type` and `effects` mean at each kind of position.
- Nodes inside a template string have no positions of their own. The parser gives every node inside `${...}` the template's position (`parseTemplate` in `selfhost/src/parser.nra`).
- A load error with no file has no location.
- The draft lists kinds that `def` never produces.
- It leaves open what happens with an ambiguous bare name, how callers are counted, and whether tests count as callers.

## Proposed change

Exact edits to `contracts/query-output.md`:

1. Under "Interface", after the line about `NAME` and `FUNCTION`, add:

   > - `FILE` must be the package-relative path exactly as diagnostics print it (`src/text.nra`). Other spellings (`./src/text.nra`, an absolute path) find no module.

2. Replace the line

   > - `kind` is one of `function`, `struct`, `enum`, `variant`, `trait`, `const`, `distinct`, `local`, `parameter`.

   with

   > - `kind` is one of `function`, `struct`, `enum`, `variant`, `trait`, `const`, `distinct`. `local` and `parameter` are reserved for a later position-based `def`; no command prints them yet.

3. Replace the line

   > - `effects` is sorted and empty for a pure expression.

   with

   > - `type-of` reports the innermost expression, local name (parameter, `const`, `let`, loop variable, pattern binding), or function or constant name at the position. `location` is the range of the token or name it found.
   > - `type` is the type of the value at the position. In a function type, `uses` comes from the declaration of a named function (the declared effects of an exported function, the inferred ones of a private one), from the annotation of a parameter or local, or from the body of an arrow function. Any other function-typed value (an unannotated local, a struct field, an array element, a call's result) shows no `uses`, and calling it reports only the effects the effect pass records.
   > - `effects` is what evaluating the expression at the position does, sorted. A name (a local or a declared function) evaluates with no effects. A call has its call effects. On the name of a called local or declared function, the answer is about the name: with `return f(text)`, `type-of` on `f` gives `(string) => int uses alloc` with `effects: []`, and on the `(` it gives `int` with the call's effects.
   > - Provisional: a builtin or constructor name in a call answers for the call (its type and effects), because the checker records nothing for that name. A later version may answer with the name's function type instead. Consumers must not rely on the call's answer there.
   > - A position inside a template string that has `${...}` parts is an error: `positions inside template strings are not supported yet`. The template's opening backtick answers for the whole template.

4. After the line "Lists are in file then position order; ...", add:

   > - A `refs` use or a `callers` entry inside a template string has the template's whole range as its location and an extra field `"inTemplate": true`. The field is absent otherwise.
   > - `refs` and `callers` need a name that finds one declaration. A bare name that finds several exits 1 with `"<name>" names <qualified>, <qualified>; write a qualified name`. `def` lists every match.
   > - `refs` finds functions, constants and enum variants. Any other kind exits 1.
   > - `callers` has one entry per call site: a function that calls twice is listed twice. The location is the callee's name in the call, or the whole template when `inTemplate` is set. Test bodies count as callers, named `<module>.test "<test name>"`. `recursion` counts only functions and implement block functions.
   > - `interface` prints each signature from the source as written: type names keep the spelling they have there (aliases and qualified names included), field defaults and bodies are left out, opaque types hide their fields or base, and implement blocks show only through a type's `implements` list.
   > - Nested function types with `uses` are ambiguous as text (in `(() => int uses io) => int uses alloc`, a reader cannot tell from the text alone which function each `uses` belongs to). A consumer that needs the structure must not parse `type` or `signature`.

5. Replace the paragraph "Errors with `--json`: ..." with

   > Errors with `--json`: exit 1 and one object `{"error": "<message>"}`, or for load and check errors `{"errors": [<diagnostic objects as narra check --json>]}`. Until the self-hosted checker has diagnostic codes, `code` is the placeholder `E0000` and `range` is empty (start equals end) at the error's position. A load error with no file, such as a package directory that does not exist, leaves out `file` and `range`. A usage error prints the usage text on stderr and exits 2, also with `--json`.

6. Under "Changes", add:

   > - v2 (from the query-cli change request): the type-of rule for `type` and `effects`, `inTemplate`, diagnostics without file and range, the `E0000` and empty-range placeholders, `local` and `parameter` reserved, the ambiguity, call site and test caller rules, and notes on `interface` text, nested `uses` text and the FILE spelling.

## Affected nodes

language-service (hover, definition, references), lint and narra-skills read the JSON form.

## Workaround in place

None needed. `narrac query` already behaves as proposed; the goldens in `compiler/testdata/query/` pin it. The parser fix for template positions has been sent to the inbox by the orchestrator. Once it lands, `type-of` inside templates can answer, and `inTemplate` entries can get exact ranges.

## Decision

<Filled in by the orchestrator.>
