# Outcome: attributes

## Built

- AST (`ast.nra`): `FuncDecl.attributes: Array<Attribute> = []` replaces `FuncDecl.target`. `Attribute { pos, name, hasParentheses, arguments: Array<AttributeArgument> }`; `AttributeArgument` is the old `TargetArgument` (`pos, name, typeName, typePos, variant`). `findAttribute(funcDecl, name): Option<Attribute>` looks one up by name. Each name appears once per declaration.
- Parser (`parser.nra`): `parseDecl` reads attributes in a loop (one per line, any order) with `parseAttribute`. The table `KNOWN_ATTRIBUTES: Array<AttributeRow { name, takesArguments }>` lists the known names. An unknown name is E0990 "unknown attribute @x; the known attributes are @target" (built from the table). A repeated name, or attributes before a non-function, is E0990 "`@name` is written once, before a function declaration" (same text as before for `@target`). `addAttributes` stores the list on the `FuncDecl`.
- Checker: `resolveFunctionTarget` (`buildconditions.nra`) gets its attribute with `findAttribute(funcDecl, "target")`. Behavior is unchanged.
- fmt: `printAttributes` prints each attribute on its own line, with its own dangling comments (`printAttribute`, trivia key `@<line>:<column>` as before); parentheses are printed when `hasParentheses` is set. astdump: `dumpAttributeSlots` writes one slot per attribute named after it (`.target` for `@target`), same output as before.
- Diagnostics: E0990 rows made generic (`@%s`), E1000 to E1099 range row added ("C interop and unsafe code").

## How a later node adds an attribute

1. Add a row to `KNOWN_ATTRIBUTES` in `parser.nra` (`takesArguments: false` for `@trusted`, which then prints and dumps without parentheses). The unknown-attribute message lists it by itself; update the E0990 message in tests that list the names.
2. Check it where it matters with `findAttribute(funcDecl, "trusted")`, reporting with a code from the new range.
3. `@extern("name")` takes a string argument: `parseAttribute` parses only `name = Enum.Variant` arguments today, so extend `AttributeArgument` (for example a string value field) and the argument loop for that row, and print it in `printAttribute` and `dumpAttributeSlots`.
4. `@target` on extern blocks: `parseDecl` rejects attributes before any declaration but a function (`addAttributes`); allow the new declaration kind there and store the list on its node.

## Deviations

- None. Only the `shared:` paths were edited (`ast.nra`, `parser.nra`, `astdump.nra`, `fmt.nra`, `buildconditions.nra`, `diagnostics.nra`, two E0990 goldens).

## Stubs left

- Argument syntax other than `name = Enum.Variant` (c-interop).

## Gotchas

- A new row means the E0990 "known attributes" text changes in the unknown-attribute goldens.
- The seed was not regenerated; the compiler source does not use any new language feature.

## Verification

- See the report of the worker: `sh selfhost/bootstrap.sh` and `sh scripts/test.sh`.

## Size

- Estimated S, actual S.

Landed on feat/batch-8 as ea7a0f41 (reviewed inline by the orchestrator: a behavior-preserving refactor).
