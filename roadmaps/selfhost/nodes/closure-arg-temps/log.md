# Log

- Base (feat/batch-4, 81638c3): `narrac emit testdata/closure_arg_temps.nra` holds `base + 1` in a second temporary (`int64_t t11 = t10;`) before the call.
- Branch: the temporary is gone; the call takes `t10` directly.
- A closure that writes a variable another argument reads is rejected by the checker, so the sample reads `base + 1` and the closure writes a different variable.
