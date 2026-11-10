# Outcome: fast-alloc

## Built

- `selfhost/runtime/runtime.h`: `narra_alloc`, `narra_free` and `narra_realloc` use size-class free lists. Blocks of up to 512 bytes, header included, come from 32 lists in steps of 16 bytes, carved from 1 MiB chunks (`narra_carve`, `narra_block_new`, `narra_block_free`, state in `narra_small`). Larger blocks use `malloc`. An 8-byte header holds the class, 0 for a `malloc` block. Chunks are linked through their first word and never returned to the system.
- `narra_realloc`: a small block that still fits stays in place; otherwise it moves to another class or to `malloc`. A `malloc` block grows with `realloc` and stays a `malloc` block when it shrinks. `narra_realloc(NULL, n)` allocates without counting, as before.
- Plain `malloc` and `free` when AddressSanitizer is on (`__SANITIZE_ADDRESS__` or `__has_feature(address_sanitizer)`), or when `NARRA_PLAIN_MALLOC` is defined. `-sanitize` builds take that path.
- `narra_live` is counted outside the `#ifdef`, the same in both paths.
- `NARRA_HEAP_LOCAL` is empty today; defining it as `_Thread_local` gives one heap per thread.
- `selfhost/src/runtimecore.nra` regenerated (`--suite generated --update`).
- `DECISIONS.md`: a paragraph under D100 with the design and the gain.

## Proof

Performance. Apple M2, 8 GB, macOS 27. Compiler from `selfhost/bootstrap.sh`, stage2.c built with `cc -O2 -DNARRA_RELEASE`, `/usr/bin/time -l <narrac> emit selfhost`.

- base: feat/batch-4 `dfb4803`. 7 runs at load 4.2: 9.29 G instructions, 0.80 s user, 135.8 MB peak.
- branch: 7 runs at load 27: 5.34 G instructions, 1.20 s user, 149.7 MB peak.
- base and branch alternating, 9 runs each, load 25 to 46: base 9.35 G, 1.48 s, 133.2 MB; branch 5.33 G, 1.12 s, 150.0 MB. 43% fewer instructions, 24% less user time, 13% more peak memory.
- `bench/run.sh -n 9`, wall seconds:

| program | base (load 4.0 to 3.4) | branch (load 31 to 8.8) |
| --- | --- | --- |
| fib | 0.113 | 0.192 |
| sieve | 0.265 | 0.530 |
| matmul | 0.113 | 0.167 |
| strings | 0.265 | 0.252 |
| interp | 0.426 | 0.495 |
| particles | 0.189 | 0.133 |
| compiler | 0.620, 129 MB | 0.541, 145 MB |

  The branch run is load-bound: fib does not allocate and still went from 0.11 to 0.19 s. Instructions retired per program, median of 5, both compilers at load 7 to 8:

| program | base | branch |
| --- | --- | --- |
| fib | 1.213 G | 1.213 G |
| sieve | 2.441 G | 2.207 G (-9.6%) |
| matmul | 2.460 G | 2.455 G |
| strings | 1.243 G | 1.168 G (-6.1%) |
| interp | 4.895 G | 4.909 G (+0.3%) |
| particles | 4.215 G | 2.616 G (-37.9%) |

## Deviations

- None. Only `runtime.h`, the generated `runtimecore.nra`, `DECISIONS.md` and this node's folder changed. The seed is not regenerated: the orchestrator does the two seed updates.

## Stubs left

- None.

## Gotchas

- Blocks are 8-byte aligned, not 16 (the header is 8 bytes). Nothing in the generated C needs more today. A future type that needs 16-byte alignment (SIMD, `long double`) needs a 16-byte header or aligned classes.
- The suites run the leak check only in `-sanitize` builds, which keep plain `malloc`. The free-list path with the leak check is covered by hand: `NARRA_CHECK_LEAKS=1` on the free-list compiler (`emit selfhost`, `fmt --check selfhost/src`) exits 0. The counting code is the same line in both paths.
- No new sample: the existing suites already allocate, grow and free across every size class with the free lists. Every test runs the non-sanitized narrac, and `emit selfhost` (bootstrap fixed point) uses all 32 classes and the `malloc` path, with 494789 realloc moves between classes (scratch histogram). Release-mode samples in `programs` also run on the free lists.
- Memory freed into a class stays in that class. A program that frees many blocks of one size and then allocates another size keeps both.
- A pointer from `narra_alloc` must never reach plain `free` or `realloc`, and a raw `malloc` pointer must never reach `narra_free` or `narra_realloc`. Today no code mixes them (checked: `std_*.c`, `stdfs.nra`, `emitdecls.nra`). ASan cannot catch such a mix, since ASan builds use plain `malloc` everywhere.

## Proposed learnings

- Keep `narra_alloc` blocks and raw `malloc` buffers apart | the free-list path crashes or corrupts on a mix and ASan builds cannot see it | selfhost/runtime | no

## Verification

- `NARRAC=<base narrac> sh scripts/test.sh --suite generated --update`: PASS.
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C (184193 lines).
- Scratch C test of the allocator (sizes 0 to 1199, growth to 1300 bytes, shrink, free and reuse): live 0 at -O2 and under UBSan with free lists; the ASan build selects plain `malloc`, live 0.
- `sh scripts/test.sh`: PASS in 116 s at load 15 to 8, 8 shards with 5774 passed and 0 failed, and the sanitized bootstrap job passed.

## Size

- Estimated S, actual S.
