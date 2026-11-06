# Log

- Go rule: check_unused.go, Local.Used, reports at the end of each body (func, test, const, field default); E0921 locals, E0922 parameters. Entry run args exempt.
- Cleanup commit: renames by applying the E0921/E0922 fixes with a script; errors/ tests re-fixed with header columns shifted.
- Selfhost: reportUnusedLocals in check.nra (finishBody), BodyInfo.errorCount.
