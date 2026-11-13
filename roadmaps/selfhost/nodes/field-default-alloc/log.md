# Log: field-default-alloc

- Proof on feat/batch-13 (6ae5adcf), release compiler from `git archive`: `narrac check testdata/errors/field_default_alloc.nra` exits 0 with no error. Branch compiler reports the missing `uses alloc`, exit 1.
- Emit comparison: both compilers on the base `selfhost/` give byte-identical C.
