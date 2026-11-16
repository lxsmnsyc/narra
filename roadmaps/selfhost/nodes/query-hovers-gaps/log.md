# Log: query-hovers-gaps

## Proof before the fix

Built narrac from the branch base and ran `narrac query --package DIR hovers src/main.nra` on the brief's sample. Wrong lines:

    src/main.nra:4:3 void                 (print, line 4: the call's result)
    src/main.nra:5:3 void                 (print, line 5)
    src/main.nra:5:16 string              (toString inside the template: the call's result)
    (no line for Int at 5:12)

Right lines outside the template:

    src/main.nra:3:16 module std/int
    src/main.nra:3:20 (int) => string uses alloc

Causes:

- The alias use (`Alias.member`) was found by scanning lexer tokens. A template string is one token, so nothing inside `${...}` was seen. It now reads the FieldExpr sites, which the site collector builds for template parts too.
- A callee with no function ref (a builtin) fell back to the call's type. It now looks the name up in the std registry: the prelude, or a name imported from a std module.

Maintainer's addition: a struct literal had no hover on the type name or the field names. Added from the literal's site.

Notes: `Shapes.Point { ... }` does not parse (a struct literal takes a bare type name), and a variant constructor takes no labels (E0332), so neither has a case.
