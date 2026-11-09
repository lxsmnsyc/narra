# Log: sized-f32

- Added TF32 following the sized-integer pattern; the compiler listed each exhaustive match.
- Float literal range check is done on the digit text because Narra has no float parse.
- `Float.from` lives in `stdf32.nra` (module `std/float`) so `stdfloat.nra` stays untouched.
- `scripts/test.sh --update` wrote the new `.err` files; `f32.out` came from `narrac run -sanitize` and was read over.
- Heredocs inside compound shell commands are refused in this worktree; files were written with the editor tools.
