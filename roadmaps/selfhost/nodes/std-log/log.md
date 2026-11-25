# Log: std-log

- Branch feat/std-log reset to feat/batch-24.
- Wrote std/src/log.nra, registry row selfhost/src/std/stdlog.nra, ModuleInfo in scripts/std/generate.nra.
- Sample testdata/log_levels.nra runs itself as a child and prints the child's stderr, so stderr has a golden.
- Compiler embeds std from stdtext.nra, so a std edit needs a regenerate and a rebuild. Iterated the module as a copy in a scratch package.
