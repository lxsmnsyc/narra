# Log: profile-compiler-2

- 2026-11-10: branch spike/profile-compiler-2 reset to feat/batch-5 at 9f8ee5f1. Reusing the scripts in nodes/profile-compiler/.
- build.sh: fixed point holds, 164,332 lines of stage 2 C (was 184,091 at 55fb951), seed current.
- Instructions 2.622/2.635/2.656G, peak 128 MB. time.sh under the bench lock: release 0.254-0.259 s, -g same, noinline ~0.40 s. Load 5-10.
- Sampled opt, opt2, noinline (30 s each, lock held). The old analyze.py put fast-alloc's narra_block_new/free in the caller's category (allocation read 1.8%); copied analyze.py and allocators.py here with those names added.
- Categories: allocation 39.6% -> 4.1%, rc 9.3% -> 24.0% (same ms), compiler code 50%.
- Whole Checker/Emitter retains in 25 tail-loop functions (`_held`): 9.1% opt, 9.4% opt2. tail-loop-borrow is top.
- lstr_concat: 0 samples, 7 call sites in the C. Map loops: all in cold functions. Both follow-ups unsupported for the compiler.
- walkExpression 8.8% self at -O2: added split.sh (noinline on chosen functions). Split run: walkExpression self 4.6%, valuesAt map 2.4%, scope map lookups 4.0%. No single effect helper dominates.
- Other finds: startBody id walk 3.5%, lexPunctuation 5.7% (noinline), mapFunctions Str.replace 2.0%, moduleHash 1.1%, teardown 6.3%.
- Wrote outcome.md.
- Review fixes: added `analyze.py --whole-struct` and committed its output (whole-struct-opt/opt2, callers-cType-opt). Allocation 4.1% marked a lower bound (inlined narra_block_free counts as rc at -O2). Scope-map proposal now names the undo log, narrow|N keys and declaration depth (size M). valuesAt proposal now skips bodies without a FunctionStore. Teardown skip keyed on NARRA_CHECK_LEAKS, emit/build only. Wall-time gap corrected to 0.025 s.
