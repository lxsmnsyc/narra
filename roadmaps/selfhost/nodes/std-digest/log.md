# Log: std-digest

- Based on feat/batch-24. Wrote std/src/digest.nra from scratch (u32 wrapping arithmetic), registered the module as std-semver did.
- Std text is embedded: after editing std/src, regenerate stdtext and rebuild the compiler (bootstrap) before a sample sees the change.
- `--update` does not create a missing .out; the sample output was written with `narrac run`.
- Goldens updated: `--suite errors --update` and `--filter err_unknown_module --update` (declsdump).
- Unused loop variables in std source are errors; name them `_index`.
