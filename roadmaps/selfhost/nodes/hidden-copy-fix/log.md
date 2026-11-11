# Log

- Base 7a1a817f (feat/batch-9). Debug compiler: `cc -O1 -std=c11 -w` on seed/narrac.c (no -DNARRA_RELEASE), then `narrac0-dbg emit selfhost`, then cc again (narrac1-dbg).
- Problem 1 shows: `NARRA_CHECK_COPIES=1 narrac1-dbg check testdata/hello.nra` prints `array 1` and exits 103; `emit selfhost` exits 103 (`array 953, map 964`).
- Cause: loadFile did `let modules = std.modules` then pushed. `std` (a StdLoad) still holds the array, so the push copies. E0978 does not see it: `std` is never used after the read (the documented known gap, write after last use).
- Reduced to testdata/copy_dead_name.nra: debug copy check reports array 1, exit 103. Not an E0978 case, so a known-gap example in SPEC section 6.
- loadPackage had the same pattern (`withStd`); fixed the same way.
- Problem 2 does not show: NARRA_COPY_GRANT expands to ((void)0) under NARRA_RELEASE (runtime.h), as NARRA_COPY_GRANTED does. The other emit sites use macros too; no `#ifndef` appears in generated code. No change.
