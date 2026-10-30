# Log: strings-std

- Positions are `int` (no distinct type yet); slices copy.
- StringBuilder is a prim type with COW on shared append.
- Fixed `builtinCall` ordering so `Mutable` builtins work.
- Fast check passes; strings2 sample clean under ASan and leak check.
