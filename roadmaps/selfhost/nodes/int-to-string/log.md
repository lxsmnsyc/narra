# Log: int-to-string

- Base: feat/batch-4 8bd8699, release build (cc -O2 -DNARRA_RELEASE), 4 runs of `emit selfhost`: 4.158 to 4.200 G instructions retired.
- Branch: 4 runs: 3.941 to 3.957 G.
- Old output captured with the base compiler into testdata/int_to_string.out before the change; the branch compiler prints the same bytes.
- Callers: only `Int.toString` (lstr_from_int) and the sized-integer Show wrappers (NARRA_SIZED) in runtime.h. Nothing else in selfhost/runtime or the emitter.
