# Outcome: selfhost-names-frontend

## Built

A rename pass over the selfhost front end. Behaviour, dump formats and error messages are unchanged: every string, char and template literal in the owned files is byte-identical to main (checked with a literal extractor).

- `lexer.nra`: `Lexer` fields `src`, `i`, `punctChars` are now `source`, `index`, `punctuationChars`. Locals and parameters are full names (`lexer`, `character`, `builder`, `offset`, `message`, `quote`, `value`). Private functions: `newLexer` → `createLexer`, `size` → `getSize`, `text` → `readText`, `number` → `lexNumber`, `punct` → `lexPunctuation`, `escape` → `lexEscape`, `unicodeEscape` → `lexUnicodeEscape`, `quoted` → `lexQuoted`, `template` → `lexTemplate`, `runeOf` → `convertRune`. Two index `while` loops became for-of over `Ranges.between`.
- `parser.nra`: `Parser` fields `toks`, `i` are now `tokens`, `index`. `p` → `parser`, `t` → `token`, and the same for every other short local. Private functions now use action names: `ok` → `isOk`, `at` → `isAt`, `tok` → `currentToken`, `tokAt` → `peekToken`, `posOf` → `getPos`, `ident` → `expectIdent`, `describe` → `describeToken`, `lexDiag` → `splitLexError`, `exprPos` → `getExpressionPos`, `namedType` → `createNamedType`, `noPos` / `dummyExpr` / `dummyType` / `dummyBlock` / `dummyPattern` → `createEmptyPos` / `createDummyExpr` / ..., `arrowAhead` → `isArrowAhead`, `typeArgsLitAhead` → `isLiteralTypeArgsAhead`, `paramAfterComma` → `isParamAfterComma`. Each Go-named parse function gets a `parse` prefix: `importDecl` → `parseImport`, `funcDecl`, `funcHeader`, `constDecl`, `structDecl`, `enumDecl`, `traitDecl`, `implDecl` → `parseFuncDecl` and so on, `fieldList` → `parseFieldList`, `traitLists` → `parseTraitLists`, `traitNames` → `parseTraitNames`, `funcTypeExpr` → `parseFuncType`, `usesList` → `parseUsesList`, `typeParamsEffects` → `parseTypeParams`, `typeParamsNoBounds` → `parseTypeParamsNoBounds`, `opaqueOrDistinct` → `parseOpaqueOrDistinct`, `ifStmt` → `parseIf`, `binary` → `parseBinary`, `unary` → `parseUnary`, `postfix` → `parsePostfix`, `primary` → `parsePrimary`, `template` → `parseTemplate`, `matchExpr` → `parseMatch`, `mapLit` → `parseMapLit`, `structLit` → `parseStructLit`, `litTypeArgs` → `parseLiteralTypeArgs`, `patternList` → `parsePatternList`, `memberName` → `parseMemberName`. The file header says how the names map to Go.
- `astdump.nra`: `sb` → `builder`, every short local and parameter renamed. Private helpers: `line` / `node` / `slot` → `writeLine` / `writeNode` / `writeSlot`, `posText` → `formatPos`, `quote` → `quoteText`, `strAttr` / `flag` / `listAttr` → `formatStringAttr` / `formatFlag` / `formatListAttr`, `typesSlot` / `typeSlot` / `blockSlot` / `exprSlot` / `patSlot` / `funcsSlot` → `dumpTypesSlot` / `dumpTypeSlot` / `dumpBlockSlot` / `dumpExprSlot` / `dumpPatternSlot` / `dumpFuncsSlot`.
- `idcheck.nra`: id bindings are `nodeId`, the rest are full names. Private walkers: `paramIds`, `fieldIds`, `stmtIds`, `exprsIds`, `armIds`, `patternIds` → `collectParamIds`, `collectFieldIds`, `collectStatementIds`, `collectExpressionListIds`, `collectArmIds`, `collectPatternIds`.
- `manifest.nra`: full names (`lineNumber`, `binaryName`, `qualifiedKey`, `index`, `code`, `builder`). Private: `fail` → `createError`, `closedArray` → `isClosedArray`, `validPackageName` → `isValidPackageName`.
- `ast.nra`: unchanged. See Gotchas.
- `roadmaps/selfhost/CLAUDE.md`: one rule paragraph on readable names.

### Renamed exported names

None. Every exported name in the owned files is used by a module this node does not own, so all keep their names:

- lexer: `lex`, `hex`, `kindName`, `escapeText`, `Token` and its fields (main.nra, modules.nra, parser, astdump).
- parser: `parse`, `quoteGo`, `diagText`, `Diag`, `Parsed` (declare, exprs, stmts, resolve, scopes, traits, modules, manifest, main).
- astdump: `dump` (main). idcheck: `collectIds`, `checkIdList` (main), `funcIds`, `blockIds`, `exprIds` (check.nra).
- manifest: `parseManifest`, `Binary`, `Manifest`, `LoadError` and their fields (modules.nra).

Only parameter names of exported functions changed, and they are not part of the interface.

### Kept for a later pass

Names a later rename can change, with the files that would change with them:
- `funcIds`, `blockIds`, `exprIds` → `collectFuncIds`, `collectBlockIds`, `collectExpressionIds` (check.nra).
- AST fields `id` and `op` (two letters), and short AST fields such as `pos`, `cond`, `iter`, `vals`, `elems`, `stmts`, `decls`, `funcs`, `params`, `args`, `targs`. Every checker module reads them.
- `Pos.col` and `Token.col` / `endCol`, `LoadError.col` (main, modules, every checker module).
- `Diag` → `Diagnostic`, `diagText` → `formatDiagnostic`, `kindName` → `getKindName`.

## Deviations

- `selfhost/CLAUDE.md` (not owned): one line changed, `ok(p)` / `tokAt` → `isOk(parser)` / `peekToken`, so the rule names functions that exist.
- Function-level comments on private functions are kept. Most say which Go function a function ports or why it differs, which the brief asks to keep.

## Stubs left

- None.

## Gotchas

- Narra rejects shadowing, also in a pattern binding: `if (block is Some(block))` fails to build. Give the bound value its own name (`bodyBlock`, `destructured`).
- The parser header now explains the Go mapping: a Go parse function `x` is `parseX` here, and `tok` / `tokAt` / `posOf` are `currentToken` / `peekToken` / `getPos`.

## Proposed learnings

- Check a rename pass by comparing every string, char and template literal with main | a rename must not change printed text, and the dump tests do not cover every error message | selfhost/ | yes

## Verification

- After each file: `go test ./internal/narra -run 'TestSelfhost|TestASTDump|TestTypesDump' -v`. The subtest PASS/SKIP counts per test match main (54773bb): TestSelfhostBodies 144 pass / 202 skip ("compared 143, skipped 202"), TestSelfhostDecls 249/84, TestSelfhostImports 311/32, TestSelfhostTraits 270/63, TestSelfhostLexer 417, TestSelfhostParser 483, TestSelfhostModules 463, TestSelfhostModuleIds 431, TestSelfhostNodeIds 432, TestSelfhostBuiltins 12, TestASTDump 379, TestTypesDump 18.
- Literal check: the string, char and template text of lexer, parser, ast, astdump, idcheck and manifest are identical to origin/main.
- Fast check `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 201 s).

## Size

- Estimated L, actual M. The renames were scripted (an identifier renamer that skips strings), so most of the work was choosing names and checking.
