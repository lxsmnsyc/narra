# Narra TextMate grammar

`narra.tmLanguage.json` is a standard TextMate grammar with the scope name `source.narra`. Any tool that reads TextMate JSON grammars can load it: VS Code, Shiki, and others.

It covers comments (`//`, `///`, `/* */`), annotations (`@const`, `@export(C)`), strings, characters, template strings with `${...}` holes, tagged templates, numbers (`0xFF`, `0b1010`, `1_000`, `1.0e-9`), keywords, `uses` effect lists, type annotations after `:`, and `function*` and `async function`. The keyword list comes from `selfhost/src/lexer.nra`.

## VS Code

Use the grammar in a local extension. Create a folder with a `package.json` like this and copy `narra.tmLanguage.json` next to it:

```json
{
  "name": "narra",
  "version": "0.0.0",
  "engines": { "vscode": "^1.80.0" },
  "contributes": {
    "languages": [{ "id": "narra", "extensions": [".nra"], "aliases": ["Narra"] }],
    "grammars": [
      { "language": "narra", "scopeName": "source.narra", "path": "./narra.tmLanguage.json" }
    ]
  }
}
```

Then copy the folder into `~/.vscode/extensions/` and restart VS Code.

## Shiki

Pass the grammar as a custom language. Shiki takes the parsed JSON with its `name` and `scopeName`:

```js
import { createHighlighter } from "shiki"
import narra from "./narra.tmLanguage.json" with { type: "json" }

const highlighter = await createHighlighter({
  themes: ["github-dark"],
  langs: [{ ...narra, name: "narra" }],
})
const html = highlighter.codeToHtml(code, { lang: "narra", theme: "github-dark" })
```

## Check

`check.cjs` tokenizes every `narra` fence in `docs/guide/*.md` and `SPEC.md` with `vscode-textmate` and `vscode-oniguruma`. It fails on a token with an `invalid` scope (an unterminated string or character, an unknown escape) and on a sample that ends inside an open string, template or comment. To check other files, pass them as arguments (`.md` files by their `narra` fences, `.nra` files whole).

```sh
cd editors/textmate
npm install
npm run check
```

The check is not part of `scripts/test.sh`, which needs only `cc` and `sh`. The website build runs it.
