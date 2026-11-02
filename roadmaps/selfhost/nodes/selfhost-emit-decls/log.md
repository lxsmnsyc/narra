# Log: selfhost-emit-decls

- Ported emit_types.go and emit_box.go into selfhost/src/emitdecls.nra.
- Until programRecorded is set, every section returns the program-facts marker (node selfhost-emit-program), which already lists all inputs, so the decls skip list is empty.
- Harness reads every node's skip file, so the file stays with only its comment line.
- Smoke run with the guards off and stub cnames: struct retain/release shape matches Go; real byte equality needs emit-program.
