# Log: const-params-pointer

- Base: feat/batch-4 at 9501a3c. Base compiler built with `sh selfhost/bootstrap.sh`.
- Checker already rejects the two in-call aliasing cases: `f(mutable outer, outer.inner)` is E0319, and a closure argument that assigns (or passes as `mutable`) a variable that another argument reads is E0614, for a field too. Probes in the scratchpad, not committed; `testdata/errors/mutable_twice`, `closure_borrow_field_arg` and `closure_borrow_mutable_arg` cover them.
- Limit 16 bytes, from the profile-compiler outcome. Size is the C layout with alignment, cached per struct and enum in `finish`.
- Branch compiler before the rebase: fixed point holds, `NARRAC=<branch> sh scripts/test.sh` PASS (5843 cases and bootstrap).
- `narrac emit selfhost`, instructions retired, 3 runs: base 3.147 G, 3.106 G, 3.098 G; branch 2.634 G, 2.620 G, 2.621 G. Peak memory 128 MB on both.
- Bench programs: only particles changes its C (`move` takes `const Particle*`); clang inlines it and the instructions are the same.
- Mutants, each one copy removed, run on the three samples with `-sanitize`:
  - tail (parameter points at the jump's temporary): `params_pointer_tail` prints the same name on every step, then heap-use-after-free.
  - generator (frame keeps the caller's pointer): `params_pointer_escape` prints `bb b-tag 2` instead of `aa a-tag 1`, then heap-use-after-free.
  - closure (environment keeps the caller's pointer): `params_pointer_escape` prints `ff/f-tag/4` instead of `ee/e-tag/3`, then an UndefinedBehaviorSanitizer error.
  - The other samples pass under each mutant, as expected.
- Rebased on feat/batch-4 5ba04ec (perceus-reuse), no conflicts. Compiler 3.159 G to 2.636 G (median of 3). Review fixes: two renames in emitdecls.nra. Final `sh scripts/test.sh`: PASS, 5853 cases and bootstrap.
