# Log: tail-loop-borrow

- Check first: `openTailLoop` retained every non-mutable parameter; no pass-through analysis existed. Not fulfilled.
- Base proof at 24aec0d7 (feat/batch-5), `/usr/bin/time -l narrac emit selfhost`, instructions retired, 5 interleaved runs, base binary from `sh selfhost/bootstrap.sh`. Load average 9 to 13 (3 other workers).
  - base: 2621456000, 2590741322, 2566203103, 2556871328, 2561595315 (median 2.567e9)
  - branch: 2175129162, 2065723997, 2058880554, 2053785138, 2064533689 (median 2.064e9)
  - Spread inside each column is under 2% apart from the first run (cold caches); the gain is 19.6%.
- Change: `openTailLoop` skips retain, `_held` and registration for a counted parameter that every tail call passes back as it is; `emitTailJump` skips such a parameter (it is not registered in any scope).
- Stage 2 C shrinks from 164332 to 164026 lines. Stage 2 and stage 3 agree.
- New sample `testdata/tail_loop_borrow.nra`; `params_pointer_tail` and `tail_loops` pass.
