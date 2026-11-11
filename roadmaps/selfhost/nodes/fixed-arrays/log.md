# Log: fixed-arrays

## Fix: narrowing ends for names assigned in the rest of a condition

Found by the fresh review (probes b02, b03, b12, c01 in /private/tmp/claude-501/review-fixed-arrays/p/). `narrowByBinary` (operands.nra) narrowed by the left side of `&&` (and of a negated `||`) and kept that fact although a closure in the right side, run later, assigns the name. Fix: after narrowing by the left side, end the narrowing of every name the right side assigns (`collectAssignedInExpr`, moved from exprs.nra to operands.nra with its helpers, since exprs.nra imports operands.nra), then narrow by the right side.

Sample `testdata/errors/divisor_assigned_in_condition.nra` (if, while, `||` exit; no FixedArray):

- feat/batch-8 compiler (055408e0, bootstrap of the base): `narrac check` exits 0, no error: the three divisions are accepted and divide by zero at run time.
- after the fix (`/private/tmp/claude-501/fixed-arrays/fixbuild/narrac`): exit 1, E0979 at 15:20, 30:23 and 46:18. Review probe c01 also gives E0979 at 8:28 and 12:28.
- Targeted suites with the fix build: errors, programs, sources, codes, spec, guide pass.
