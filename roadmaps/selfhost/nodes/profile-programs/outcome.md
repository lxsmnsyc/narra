# Outcome: profile-programs

## Built

- No compiler change (spike). This folder holds the findings, `log.md`, excerpts of the profiles and inner-loop asm in `evidence/` (`sieve.sample.txt`, `particles.sample.txt`, `inner-loops.s.txt` for base, v1 and v3), and two scripts that rewrite emitted C to test the proposed fixes: `experiments/variant1.py` (runtime change, any emitted C file) and `experiments/variant3.py` (hand-written emitter output for the two hot loops, applied to variant-1 C of sieve and particles).
- Both gaps come from one construct: an array element write, `A_Array_<T>_at(&array, index, where)` in `selfhost/src/emitdecls.nra` (the `_at`, `_unique` and `narra_bounds` templates), emitted by `indexPointer` in `selfhost/src/emitexprs.nra`.

## Proof

Machine: Apple M2, Apple clang 21.0.0, go1.27.3, rustc 1.98.0, commit 9f8ee5f1. The load average was 3 to 7 during the session (other spikes were running; every timed run held the shared bench lock). Wall times are medians of 15 runs. Against the end-of-batch-4 section of `bench/BASELINE.md`, the Narra base rows are close (sieve 0.246 vs 0.265, particles 0.107 vs 0.108), but Go and some other rows are not (Go sieve 0.142 vs 0.177, strings 0.212 vs 0.238), so compare variants with each other here, not with that section. Instructions retired come from `/usr/bin/time -l` and do not depend on load.

| program | build | seconds | instructions |
| --- | --- | --- | --- |
| sieve | narra base | 0.246 | 2.26G |
| sieve | variant 1 (runtime) | 0.131 | 0.68G |
| sieve | variant 1 + 3 (hoisted unique) | 0.132 | 0.54G |
| sieve | Go | 0.142 | 0.53G |
| sieve | Rust | 0.131 | 0.37G |
| particles | narra base | 0.107 | 2.61G |
| particles | variant 1 | 0.107 | 2.61G |
| particles | variant 1 + 3 | 0.068 | 1.51G |
| particles | Go | 0.074 | 1.72G |
| particles | Rust | 0.093 | 0.72G |

Variant 1 on the other programs: fib 0.112/0.111, matmul 0.113/0.106, strings 0.212/0.212, interp 0.186/0.185 (instructions within 2%). On the compiler (`stage2.c` rewritten, `narrac emit selfhost`): same emitted C, 2.65G to 2.53G instructions (-4.5%), wall time unchanged at 0.24 s.

### sieve: the element write is a call

- Profile: `sample` (1 ms) on sieve repeated 10 times: 1989 of 2197 user samples in `A_Array_e7fe1e37_at`, the rest in `main` and `memset_pattern16` (`Arrays.filled`). Excerpts: `evidence/sieve.sample.txt`; base, v1 and v3 asm of the loop and of `_at` in `evidence/inner-loops.s.txt`.
- Emitted C of the inner loop: `int64_t* t13 = A_Array_e7fe1e37_at(&v_marks_2, t12, "sieve.nra:14:14"); *t13 = INT64_C(0);`.
- Clang did not inline `_at` here. `_at` contains the whole copy path of `_unique` (allocation, memcpy) and the inlined failure path of `narra_bounds` (a `char msg[96]`, `snprintf`, `narra_panic`), so it is large, needs a stack guard, and spills 10 registers. Each `marks[j] = 0` costs a call of about 30 instructions; Go's loop is 6. In particles the same `_at` was inlined, so inlining depends on clang's per-call-site budget and is not reliable.
- `narra_panic` is not marked noreturn, so clang also treats every failure path as one that returns.
- After variant 1 the inner loop is 11 instructions and the time matches Rust. The rest is memory traffic over a 160 MB array, the same as in Go and Rust.

### particles: the copy-on-write check stays in the loop

- Profile: `sample` on particles with 10x the rounds: every user sample in `main` (all helpers inlined), so the cost is in the loop body. The asm of `particles[i] = move(particles[i])` (`_get`, `f_move`, `_at`) is about 26 instructions per particle: it reloads `len`, `items` and `rc` from the array header for each element, tests `rc == 1`, checks bounds twice, and keeps the `_copy` call inside the loop. Excerpts: `evidence/particles.sample.txt`, `evidence/inner-loops.s.txt`.
- The reloads come from that call: a loop with a possible call cannot keep the header in registers. Variant 1 alone does not help particles, since `_at` was already inlined.
- Variant 3 makes the array unique once before the loop (`v_particles_7 = A_Array_186399d5_unique(v_particles_7);`) and writes through `narra_bounds(...); &v_particles_7->items[i]`. The loop then has no call. Clang uses 128-bit loads and stores per particle, 15 instructions per element. Result: 0.068 s, faster than Go.

## Result

- result: met. The brief's suspects: copy-on-write checks are confirmed (in particles, and in sieve through the size they add to `_at`). Bounds checks are not the cost: Go and Rust keep them too, and every variant here keeps them. Float and struct copies are ruled out: `move` takes a `const Particle*`, is inlined, and its struct return becomes register moves.
- numbers: the table above.
- reproduce: at 9f8ee5f1, `sh selfhost/bootstrap.sh out`, then `out/narrac build -release -emit-c sieve.c -o sieve bench/sieve/sieve.nra` (same for particles); `python3 experiments/variant1.py sieve.c sieve-v1.c`; `python3 experiments/variant3.py sieve-v1.c particles-v1.c`; build each with `cc -O2 -DNARRA_RELEASE -std=c11 -w -o X X.c -lm`; time with `perl bench/median.pl 15 ./X` and `/usr/bin/time -l ./X`.

### Proposed nodes, ranked by value for cost

1. `array-write-cold-paths` (size S, runtime and emitdecls templates only). Mark `narra_panic` noreturn and cold. Split `narra_bounds` into an inline unsigned compare `(uint64_t)i >= (uint64_t)len` and an out-of-line cold `narra_bounds_fail` that formats the message. Split every `_unique` (arrays and boxes in `emitdecls.nra`; maps in `selfhost/runtime/runtime_map.h`) into an inline `rc == 1` test and a noinline `_copy`. The gains below come from splitting only the array `_unique` (`variant1.py` matches `A_Array_*` only); the box and map splits are unmeasured. Expected gain: sieve 0.25 to 0.13 s (-47%, 2.26G to 0.68G instructions, Rust parity), matmul 0.113 to 0.106 s (instructions only -1.6%, so partly noise), compiler -4.5% instructions. Others unchanged. Wrap the GNU attributes in a macro so the C stays valid C11 (`_Noreturn` is C11; `cold` and `noinline` behind `#if defined(__GNUC__)`, which clang and gcc define). Editing `selfhost/runtime/` needs two seed updates (GOAL constraint). Test: every suite, plus an index-out-of-range sample that still prints the same panic.
2. `loop-unique-hoist` (size M, emitter, depends on 1). When a `while` or C-style/range `for` body writes `a[i] = ...` or `a[i].field = ...` to a local array, and the body uses `a` only as `a[expr]` reads, index writes and `Arrays.length(a)` (no copy into another binding, no argument except index use, no closure capturing `a`), and the loop is not inside a `function*` body, emit `a = A_unique(a);` once before the loop and emit the writes as `narra_bounds(i, a->len, where); &a->items[i]` with no `rc` test. The `for (mutable x of items)` loop already does this (`emitReferenceFor` in `emitstmts.nra`); `checkLoopBorrows` in `exprs.nra` has a related "body cannot use the source" walk to reuse. Making the array unique before a loop that runs zero times copies a shared array that would not have been copied: allowed, as the copy is not observable, or guard it with the loop condition. Generators are excluded because a generator frame holds the array across `yield`, and copying the generator (`let h = g`) retains it, so the array reaches rc 2 inside the loop and hoisted writes would change storage both copies share. The compiler refuses `for (mutable ...)` in a `function*` body for the same reason (`selfhost/src/stmts.nra`, E0957). Required test (found by the fresh review): a `function*` that writes `a[i] = 1` in a loop and yields the sum after each write, resumed once, then copied, with both copies resumed. The current compiler prints `1 2 3 2 3`; hand-hoisted C prints `1 2 3 3 3`. The node must keep `1 2 3 2 3`. Expected gain: particles 0.107 to 0.068 s (-36%, 2.61G to 1.51G instructions, ahead of Go 0.074 s); sieve -20% instructions with no time change (memory bound). Also helps any compiler loop that writes to a local array with index writes (not measured).

Not proposed:

- Header fields typed `long` so TBAA separates them from `int64_t` elements (variant 4): particles 0.068 to 0.063 s, but it depends on `int64_t` being `long long` (macOS) and would need the other choice on Linux. Too fragile for 7%.
- Caching `len` and `items` in C locals before the loop (variant 5): clang then vectorizes particles with `ld4`/`st4` (0.77G instructions) but it runs 0.098 s, slower than variant 3 on the M2. Rust's binary shows the same pattern (0.72G, 0.093 s).
- Removing bounds checks: Go and Rust keep them in these loops, and Narra is at or ahead of both without removing them.

## Deviations

- The machine was not idle (load 3 to 7 from the other spikes), so no row goes into `bench/BASELINE.md`. The brief asked for numbers against the end-of-batch-4 section: the Narra base times here are within 0.02 s of it, the Go and strings rows differ by up to 0.035 s, and instruction counts carry the comparison.
- The fixes were tested by rewriting emitted C (scripts in `experiments/`), not by changing the compiler, as a spike should.

## Stubs left

- None.

## Gotchas

- Whether clang inlines `_at` depends on how many call sites a function has (inlined in particles, not in sieve). Judge a runtime helper by its hot-path size after splitting, not by one program.
- A text replacement on emitted C of the compiler also matches the runtime header embedded as a string literal in it (`RUNTIME_H` in `runtimecore.nra`). Replace only the first match, or the compiler emits different C.
- On the M2, fewer instructions did not mean faster here (variant 5, Rust particles). Check wall time as well as instructions.
- The `_get` before `_at` in `a[i] = f(a[i])` checks bounds twice on the same index. Clang does not merge them once the header is reloaded; after `loop-unique-hoist` it costs 2 instructions per element. Not worth its own node.

## Proposed learnings

- Runtime helpers called in hot loops keep the failure and copy paths out of line (noinline/cold), so the hot path inlines | sieve spent 90% of its user samples in a non-inlined `_at` because the copy and panic paths made it big | selfhost/runtime, selfhost/src/emitdecls.nra | no
- Measure performance changes by instructions retired as well as wall time when the machine is shared | load 3 to 7 moved wall times by up to 20% while instruction counts held within 1% | bench | no

## Verification

- `sh selfhost/bootstrap.sh <scratch>/narrac-out`: fixed point, seed current.
- Every variant binary prints the expected answer (sieve 1270607, particles 6996184, and the other bench outputs); the compiler built from variant-1 C emits byte-identical C for `selfhost`.
- `sh scripts/test.sh` not run: this branch changes only files under `roadmaps/selfhost/nodes/profile-programs/`.

## Size

- Estimated M, actual S to M. Both programs pointed to the same construct, so one profile pass covered both.
