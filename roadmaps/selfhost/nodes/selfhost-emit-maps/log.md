# Log

- Wrote emitmaps.nra against the traits stubs (useEq/useHash return markers).
- Whole-output diff of maps.nra and literal_type_args.nra against Go: every difference is a traits stub marker or the trait function block it omits.
- Fixed: None is `(Option){.tag = N}` without extra parentheses (enumAllocation adds them).
