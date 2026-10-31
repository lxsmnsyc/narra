# Log: selfhost-names-exported

- Wrote a token-aware renamer (skips string, char and template text, renames code inside `${}` and most comments), applied one map to all of selfhost/src, fixed one shadowing clash by hand (`containsOpen`), then `narra check selfhost` passed.
- Literal diff against origin/main: 0 files differ. Test counts before and after identical (5247 subtests).
