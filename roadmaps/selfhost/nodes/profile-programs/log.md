# Log: profile-programs

- 2026-11-10 15:52: started on `spike/profile-programs` at 9f8ee5f1 (`feat/batch-5`). Load average 4 to 7 the whole session (other spikes), so instructions retired are the main measure; every timed run held the shared bench lock.
- Built narrac with `selfhost/bootstrap.sh` (fixed point, seed current). Built sieve and particles with `narrac build -release -emit-c`, Go and Rust as `bench/run.sh` does.
- Base: sieve 0.259 s, 2.26G instructions (Go 0.150 s / 0.54G, Rust 0.143 s / 0.37G); particles 0.111 s, 2.61G (Go 0.079 s / 1.73G, Rust 0.098 s / 0.72G).
- `sample` (1 ms) on a 10x sieve: 1989 of 2197 user samples in `A_Array_e7fe1e37_at`, which clang did not inline. The function spills 10 registers and checks the stack guard (the `char msg[96]` of `narra_bounds` is inlined into it), about 30 instructions per `marks[j] = 0`.
- `sample` on a 10x particles: all samples in `main`; `_at` and `_get` are inlined there. The asm of the inner loop reloads `len`, `items` and `rc` from the array header each element, tests `rc == 1` and keeps the copy call inside the loop: about 26 instructions per particle.
- Variant 1 (runtime only, `experiments/variant1.py`): `narra_panic` noreturn and cold, the bounds failure out of line, `_unique` split into an inline `rc == 1` test and a noinline `_copy`. Sieve 0.131 s / 0.68G (Rust parity). Particles unchanged. Other programs unchanged; matmul 0.113 to 0.106 s.
- Variants 3 to 5 (`experiments/variant3.py`, hand-edited emitted C): v3 makes the array unique once before the loop and writes by bounds check and store. Particles 0.068 s / 1.51G (faster than Go 0.074). Sieve 0.132 s / 0.54G (same time as v1, memory bound).
- v4 (header fields typed `long` so TBAA separates them from `int64_t` elements): particles 0.063 s / 1.41G. Small, platform dependent; not proposed.
- v5 (len and items cached in C locals): clang vectorizes particles with ld4/st4, half the instructions (0.77G) but 0.098 s, slower than v3 on the M2; Rust's binary shows the same pattern (0.72G, 0.093 s). Not proposed.
- Compiler with variant 1 applied to `stage2.c`: same emitted C, 2.65G to 2.53G instructions (-4.5%), time 0.238 to 0.237 s (noise).
- Note: variant1.py first patched the `narra_panic` text inside the compiler's embedded runtime string as well; limited to the first match.
- Variant 2 (variant 1 plus header fields typed `long`, no hoisting): the sieve inner loop still reloads rc, len and items each element, because the `_copy` call stays in the loop; no gain, so it was not timed. It led to v3 (hoist first) and v4 (v3 plus `long`).
- Review fixes: generators excluded from loop-unique-hoist with the reviewer's repro as a required test; baseline comparison limited to the Narra rows; box and map `_unique` splits marked unmeasured; profile and asm excerpts committed in `evidence/`.
