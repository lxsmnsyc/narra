# Log

- std/bytes is a compiler builtin (selfhost/src/std/stdbytes.nra, runtime std_bytes.c). std modules extend builtins through `std/src/ext__<module>.nra` (like ext__char), so the new functions are Narra source in `std/src/ext__bytes.nra`. No compiler or runtime change.
- Variable shift counts need a local checked against 0..63 (E-shift rule), so the decoders copy the count into a local first.
- stdtext.nra regenerated with `sh scripts/test.sh --suite generated --update`, then the std reference with `sh scripts/std-reference.sh`.
