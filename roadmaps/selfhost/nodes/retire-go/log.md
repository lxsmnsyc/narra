# Log: retire-go

- Dumped the Go runtime strings with a throwaway Go test into selfhost/runtime/ (12 files); regenerated runtimetext.nra and stdtext.nra from them with the new `generated` suite: byte-equal except the header comment.
- Ported check-codes.py to the `codes` suite; broke two sites on purpose (wrong code, E0000), both reported, restored.
- Deleted compiler/cmd, internal, go.mod; moved testdata; seed regenerated last (header names the last commit that changed selfhost/).
- Bash 3.2 on macOS: an empty array with set -u needs the `${a[@]+"${a[@]}"}` form (linux-check.sh).
