# Log: selfhost-emit-match

- Port of emit_match.go and emit_patterns.go into emitmatch.nra. Go's closure `body func(arm)` of matchArms became a flag (isValue) plus two arm functions.
- Dead end: importing declaredLocal from emitexprs makes an import cycle (emitexprs imports emitmatch). emitmatch has its own private copy.
- Go's `e.addLocal(&Local{Name: "subject"})` has no BodyInfo local in Narra. Added `ScopeFrame.held` (defaulted field) in emitstate.nra; releaseLocals releases it after the locals.
- Removing the match stub exposed `equality (selfhost-emit-traits)` in packages/opaque_range, which the match stub used to hide. Added it to the traits skip file.
- Checked first diffs against Go on enums, patterns_*, options, generics, Tuple: every difference starts at a std, traits or closures marker, or in the literal pool (see outcome).
