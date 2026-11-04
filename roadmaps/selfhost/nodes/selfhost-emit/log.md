# Log: selfhost-emit

- The skip files were already gone (deleted by the earlier emit nodes). Removed the skip machinery from the harness: any `unsupported:` or non-zero exit fails.
- Added `selfhost` as an input. It compared byte-identical on the first run, so no divergence fixes were needed.
- Deleted `reportStdWork`, its call and import in `emit.nra`, and `hasWork`, `bodyNeeds`, `typeNeeds`, `refNeeds` in `emitstate.nra`.
