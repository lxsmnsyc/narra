# Log: fast-alloc

- 2026-11-10. Branch feat/fast-alloc from feat/batch-4 (dfb4803).
- Checked callers: every `narra_alloc` block is freed with `narra_free`; `narra_realloc` is only used on array items (`emitdecls.nra`), which come from `narra_alloc`. Raw `malloc` buffers in `std_*.c` and `narra_fs_names.items` never reach `narra_free`.
- Base: bootstrap from the seed, stage2.c built with `cc -O2 -DNARRA_RELEASE`. `emit selfhost`, 7 runs at load 4.2: 9.29 G instructions, 0.80 s user, 135.8 MB.
- Built: runtime.h free lists, `--suite generated --update`. Bootstrap reaches the fixed point (stage 2 and stage 3).
- Branch, 7 runs at load 27: 5.34 G instructions, 1.20 s user, 149.7 MB. Alternating base and branch, 9 runs each, load 25 to 46: base 9.35 G, 1.48 s, 133.2 MB; branch 5.33 G, 1.12 s, 150.0 MB.
- Class histogram of `emit selfhost` (scratch instrumentation): every class 1 to 32 and the malloc path used; 494789 realloc moves, 11763 reallocs that stay.
- Leak check on the free-list compiler: `NARRA_CHECK_LEAKS=1` on `emit selfhost` and `fmt --check selfhost/src`, -O2 release and debug builds: exit 0.
- Scratch C test (not committed): sizes 0 to 1199, growth in steps to 1300 bytes, shrink, free and reuse. Free lists at -O2 and under UBSan: live 0. ASan build: picks plain malloc, live 0.
- bench/run.sh -n 9: base at load 4.0 to 3.4, branch at load 31 to 8.8. Wall times on the branch are load-bound (fib, no allocation, 0.11 to 0.19 s). Instructions per program measured instead.
