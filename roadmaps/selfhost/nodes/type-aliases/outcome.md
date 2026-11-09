# Outcome: type-aliases

## Built

- `type Name<T> = T` and `export type` in the parser (`AliasDecl`, `Decl.Alias`), checker (`resolve.nra` `resolveAlias`/`useAlias`, `declare.nra`, `scopes.nra` imports, `AliasInfo` in `types.nra`), `fmt`, `astdump`, `declsdump`, `narra query` (interface, kind `alias`) and navigation.
- Codes E0972 (alias refers to itself) and E0973 (wrong type argument count) in `diagnostics.nra`.
- Samples: `testdata/type_alias.nra`, `testdata/packages/type_alias_reexport`, `testdata/errors/type_alias_{self,cycle,arity,message}.nra`, astdump and declsdump samples.
- SPEC.md (section 4 "Type aliases") and compiler/README.md updated.

## Deviations

- Constructors through a generic alias are not available (`Alias.Ctor`). Non-generic aliases and re-exports work.
- `Alias (Type)` in diagnostics is by type, not by what the source wrote: once an alias of a struct, enum or distinct type is used, every message about that type shows it. Aliases of primitives are never shown.

## Stubs left

- None.

## Gotchas

- A local named `alias` clashes in `qualifiedType`; the alias index is `aliasIndex` there.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh`: PASS (130s). The seed was not regenerated; no alias is used in selfhost/src.

## Size

- Estimated M, actual M.
