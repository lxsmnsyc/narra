# Outcome: named-args-syntax

Named arguments parse, dump and format. The checker rejects them for now.

## For named-args-check

- `Arg` in `selfhost/src/ast.nra` has two new fields: `pos: Pos` (first token of the argument: the name, `mutable`, or the value) and `name: Option<string>` (`Some` for `name: value`). `isMutable` follows the colon: `buf: mutable xs`. Written order is kept in `CallExpr.arguments`. No other code builds an `Arg`.
- Parser entry point: `parseCallArguments` in `parser.nra`. An identifier token followed by punct `:` starts a named argument. Positional-after-named is NOT rejected by the parser; the checker must report it.
- Placeholder to remove: `checkCallEndingNarrowing` in `exprs.nra` reports every named argument with code E0213 ("named argument `x` is not supported yet"). E0000 is not in the code table (the `codes` suite fails), so E0213 is borrowed. Replace it with real codes.
- astdump prints `arg name="x" mutable`. fmt prints `name: ` before `mutable `.

## Tests

- `testdata/parse/named_args.{nra,ast}`.
- `testdata/fmt/expr_named_args_w40.{nra,fmt}`. The fmt suite only reads goldens named `expr_*`, `stmt_*` and so on, so the brief's `named_args*` name became `expr_named_args_w40`.

## Not done

- No error golden for the placeholder rejection (it would be deleted by named-args-check).
