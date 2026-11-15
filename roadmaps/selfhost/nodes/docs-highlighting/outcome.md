# Outcome: docs-highlighting

## What was built

Every fence of the guide, install guide and spec is highlighted by one rule per language, and the build checks it.

- `website/src/highlight/fences.ts`: reads a fence info string. `narra file PKG PATH` takes its language from the extension of PATH.
- `website/src/highlight/toml.ts`, `lua.ts`, `sh.ts`: new twinkleplop languages. `sh.ts` replaces `@twinkleplop/bash`.
- `website/src/highlight/narra.ts`: type arguments are punctuation, effect names after `uses` are class names, `\u{...}` is one escape, `function*` has a keyword star, a call with type arguments is a function, a tagged template tag is a function.
- `website/src/highlight/textmate.ts`: reads the TextMate grammar as token classes.
- `website/scripts/check-grammar.ts`: fence checks (no language, unknown language, no highlighter), token checks for narra, toml, lua and sh, and a character-by-character comparison of the Narra samples with the TextMate grammar. New examples cover generics, effects, escapes and tagged templates.
- `website/plugins/markdown.ts`: uses the new languages and `readFenceLanguage`.
- Docs: `SPEC.md:657` is `narra fragment`; four `text` fences of `docs/guide/12-tools.md` are `json` or `narra fragment`.

The log lists every inconsistency with its fix.

## Deviations

- `@twinkleplop/toml` exists on npm, but `pnpm add` was refused here, so toml, lua and sh are written in this repo. `package.json` and `pnpm-lock.yaml` are unchanged, and `@twinkleplop/bash` stays listed although no code imports it. Remove it with `pnpm remove @twinkleplop/bash` when convenient. There is no `@twinkleplop/lua`.
- The comparison with the TextMate grammar has a short allow list for the grammar's own mistakes (log.md). The grammar is outside this node's paths.
- Fence info strings of `narra file ... narra.toml` stay, because `selfhost/tests/src/guide.nra` reads them.

## Gotchas

- The twinkleplop grammar merges neighbors of one class, so `>(` is one token. Rules that look at `>` and `(` apart need a claim producer, as `promoteGenericCalls` does.
- A `<` directly after a name opens type arguments and ends at the matching `>` or the line end. A comparison written without spaces (`a<b`) would be read as type arguments; the docs write spaces.
- `check-grammar.ts` imports `LANGUAGES` from `plugins/markdown.ts`, so a language added to one list and not the other fails the build.
- A new fence language needs an entry in `FENCE_LANGUAGES` (`fences.ts`) and in `LANGUAGES` (`plugins/markdown.ts`).

## CSS classes for the design node

No new class. The new languages use the existing token classes: `identifier`, `class_name`, `function`, `keyword`, `constant`, `string`, `string_escape`, `number`, `boolean`, `operator`, `punctuation`, `comment`, `type`, `template`, `decorator`. Meaning in `sh`: command is `function`, flag and shell word are `keyword`, variable is `constant`, path or URL is `identifier`. In `toml`: key is `identifier`, table name is `class_name`. Style `class_name` and `constant` for readability next to `function` and `keyword` in the shell and TOML blocks.

## Checks

`pnpm --dir website build` and `pnpm --dir website typecheck` pass. 462 fences, 326 tokenized, 287 Narra samples compared with the TextMate grammar, 0 problems.
