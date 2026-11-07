# Log: effects-indirect-calls

- Function types in the checker (`FuncType`) carry no effects and are interned by signature, so adding effects to them would touch interning, emission and equality. Chose to follow the written type (`TypeExpr`) in the effect pass instead.
- Go: `typeExprOf` follows a parameter or annotated local, a struct field (looked up in the `StructDecl`), an array or map element, and a call result. `valueEffects` gets cases for field, index and call.
- Selfhost: the same in `effects.nra` (`findWrittenType` and helpers, `EffectState.localTypes`).
- Limit: a local whose type is inferred from something with no written type (a for-of variable, a destructured name, a match binding) stays unknown and counts as pure, as before.
- Effect variables of a generic callee in a call result type are bound through the arguments (`resultEffects`). Causes are dropped there so Go and selfhost print the same message.
