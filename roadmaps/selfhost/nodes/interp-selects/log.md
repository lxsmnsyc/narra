# Log: interp-selects

- 2026-11-10: branch spike/interp-selects at feat/batch-5 9f8ee5f1. `sh selfhost/bootstrap.sh`: seed current, so the seed has D99. Load average 4 to 9 all session; timings use `cycles elapsed` and `instructions retired` from `/usr/bin/time -l`, under the shared bench lock.
- Base compiler: `git archive 9501a3cc^` (the commit before option-inline), bootstrapped in scratch. Its `interp` C differs from the current one only in the enum layout. `forms/gen.sh` regenerates both (inline.c, heap.c); the C is not committed.
- Current asm: the inner loop loads tag and payload together (the for-of item is a by-value copy), then `add`, `sub`, two `csel`. Heap form: branches only.
- Reproduced (first run 15 interleaved runs, later runs 11): inline 0.18 s, 0.626 G cycles; heap 0.15 s, 0.519 G cycles. +21% cycles.
- Hand forms (`forms/make.py`): `switch` keeps the selects. Match through a pointer alone, or parameter by pointer alone: selects stay. Both together (`pointer`, no copy from the array to the arm): branches, 0.517 G cycles. An empty asm on each payload binding: same.
- No gcc on this machine (`/usr/bin/gcc` is Apple clang). The selects come from the AArch64 backend early if-conversion: `-mllvm -aarch64-enable-early-ifcvt=false` removes them.
- First random orders used a Weyl hash (i * golden ratio); the predictor learned it. Switched to splitmix64.
- Unpredictable data: 1000000 random Add/Sub ops: inline 0.520 G cycles, pointer 1.675 G, heap 1.712 G. Random over all five ops on 1000000: inline 2.584 G, pointer 2.995 G, heap 3.026 G. The selects are the right call there; branches only win when the predictor learns the order, as in the bench.
- Compiler probe (`forms/compiler.sh`): seed built with and without early if-conversion, same emitted C, 9 runs: 0.789 G against 0.785 G cycles (-0.5%), noise. The compiler pays nothing measurable for selects.
- Conclusion: no form is better on both kinds of data. No node proposed.
- Self-review round 1: outcome rewritten in short sentences, table rebuilt from one consistent run of every form, compiler spread added. Round 2: scripts reread, no findings.
- Review fixes: emitted C dropped from the commit, `forms/gen.sh` added. Its output is byte-identical to the measured C, and `make.py` makes the same 32 files from it.
