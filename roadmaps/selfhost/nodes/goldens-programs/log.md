# Log: goldens-programs

- Frozen with a throwaway Go test (deleted): the 16 lex and 82 parse inputs, the 20 habit cases (+ `.meta`: first code and fix or nofix). Goldens made by the Go `narra` binary: `lex`, `parse`, `check --json`, `pkg`.
- Go prints dump errors on stdout, narrac on stderr, same text (`error L:C: ...`). The suite compares stdout on exit 0 and stderr on exit 1.
- Manifest: 8 error folders + `valid` (the TestManifest sample) under `testdata/manifest`; goldens from Go `pkg .`.
- Std-test programs are formatted by `narrac fmt`; the want text keeps Go's message with the file name changed from `prog.nra` to the sample name. Positions did not move.
- `cnames/stable-under-edits` copies both files to `prog.nra` in separate folders: the module hash comes from the file name (Go's test used `prog.nra` too).
- Go `TestSourcesFormatted` skipped testdata/lex, parse, habits (inputs invalid on purpose); same rule in the Narra `sources` suite.
