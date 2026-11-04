# Log: selfhost-emit-traits

- Ported emit_traits.go into emittraits.nra; recorded traitSelf in calls.nra.
- Deleted the skip file. TestSelfhostEmit passes; whole-output diff against Go is empty on the 11 trait and pattern inputs.
- Fast check passes.
