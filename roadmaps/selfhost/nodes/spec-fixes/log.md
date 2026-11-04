# Log: spec-fixes

- Branch docs/spec-fixes from origin/main a5751ff.
- Checked a scratch file with the four fixed samples (`parsePair`, `readConfig` with a `ConfigError` enum, the struct match, the `Float` conversions): `narra check` is clean. Scratch types use `f64`, not `float`.
- Line 625 (`@unique ... Array<u8> = Arrays.withCapacity(...)`) left as is: it needs sized numbers (F7) and `Arrays.withCapacity` (F18, std-spec-functions). Not in the brief's list.
