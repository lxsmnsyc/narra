# Log: docs-highlighting

Method: built the site, read the rendered `<pre data-language>` blocks in `website/dist/client`, counted the fences by info string in `docs/guide/*.md`, `docs/install.md` and `SPEC.md`, then compared the TextMate grammar and the site grammar character by character on every Narra fence (687 plus 130 differing runs before the fixes).

## Inconsistencies found

Each line: where, what was wrong, the fix.

1. `docs/guide/01-getting-started.md:98`, `08-modules-and-packages.md:119`, `12-tools.md:9` (`narra file PKG narra.toml`): TOML was highlighted as Narra. Fix: the language of a `narra file` fence comes from the extension of its path (`src/highlight/fences.ts`, used by `plugins/markdown.ts`). The info strings stay, because `selfhost/tests/src/guide.nra` reads them.
2. `08-modules-and-packages.md:311,338,352`, `12-tools.md:205,363`, `SPEC.md:2050` (`toml`): rendered plain, the site had no TOML. Fix: `src/highlight/toml.ts`.
3. `12-tools.md:351` (`lua`): rendered plain. Fix: `src/highlight/lua.ts`.
4. `SPEC.md:657`: a fence with no language holding Narra, rendered plain. Fix: tagged `narra fragment`. A plain `narra` tag would put it into the spec suite, which checks only the exact info `narra`.
5. `12-tools.md:108,302`: JSON printed by `narrac check --json` and `narrac query --json` sat in `text` fences, plain, while `SPEC.md:1998` JSON was colored. Fix: tagged `json`.
6. `12-tools.md:151,288`: Narra source printed by `apply-fix` and `query interface` sat in `text` fences. Fix: tagged `narra fragment`.
7. Every `sh` fence (30): `--json` was an operator plus a name, `-o` the same, `https://github.com/x` five tokens, `src/main.nra:5:17` five tokens, `cd` styled as a builtin but `git`, `narrac` and `sh` plain. Fix: `src/highlight/sh.ts` (command, flag, variable, word as one token each).
8. Narra, 687 characters: the `<` and `>` of type arguments (`Array<int>`, `function f<T: Ordered>`) were operators, the TextMate grammar has punctuation. Fix: a `generic` state in `narra.ts`.
9. Narra, 58 names: effect names after `uses` (`uses io, alloc`, `uses (io, alloc)`) were plain names, the TextMate grammar has class names. Fix: an `effects` state.
10. Narra, 15 escapes: `\u{1F600}` was an escape plus string text, the TextMate grammar colors it whole. Fix: a `unicode_escape` state.
11. Narra, 8 `*` of `function*`: operator, the TextMate grammar has keyword. Fix: keyword.
12. Narra, 5 calls with type arguments (`Toml.decode<Manifest>(`, `firstOr<int>(`, `Channel.create<int>(`): the name was plain, not a function. The library rule cannot see `>` and `(` apart because the grammar merges them. Fix: a claim producer in `narra.ts`.
13. Narra, tagged templates (`quoted\`...\``, `Template.raw\`...\``): the tag was plain. Fix: a rewrite rule.
14. Inline code is not highlighted anywhere. Consistent, left as it is.

## TextMate quirks, left in `editors/textmate` (not my path)

The site is right and the TextMate grammar is wrong in these. The check lists them in `explainDifference` so they do not fail the build:

- A struct field value (`Node { name: node.name, next: Some(copy(second)) }`) and the `from` clause of a function (`: Option<T> from items {`) start a type annotation that runs to the next `,`, `)` or `=`. Inside it `{`, `:` and `>` have no scope, `copy(`, `new(` and `trim(` are plain names, `from`, `typeof` and `None` are names or class names.
- The `<` of type arguments in an expression (`Map<string, int> {}`) is a comparison operator.
- `implement Show for int`: `int` after `for` is a class name.
- `Template.raw\`...\``: the whole dotted tag is a function.

## Build check

`pnpm --dir website build` fails on a fence with no language, an unknown language, a language the site cannot highlight, a token with no class, an unterminated string and a Narra character whose class differs from the TextMate grammar (after the allowed differences). Proof: appending a bare fence and a `yaml` fence to `docs/install.md` printed `has no language` and `is not one the site highlights`; the file was restored.
