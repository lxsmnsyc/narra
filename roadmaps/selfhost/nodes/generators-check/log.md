# Log: generators-check

- 2026-11-05: started. Nothing exists on main (no `yield` keyword, no Generator type).
- Go side: lexer keyword, FuncDecl.Generator, YieldStmt, GeneratorType (Effects, Binder, Origin), resolveSignature, yield/return/for-of checks, expectGenerator (deferred for inferred types), effect pass changes, Unemitted stub in driver.go. Error tests and dumps written.
- Selfhost port done; narra vs narrac dumps match on all generator files (84). Full fast check running.
- generator_emit moved from testdata/errors to TestGeneratorEmitStub (TestErrorCodes needs checker diagnostics). go test needs -timeout above 10m on this loaded machine.
- Full fast check passes (vet + test, -timeout 40m).
- Review fix: bound call results (Bound / boundGenerators) so effect variables never escape by name; 4 more error tests and bound.nra dump.
- Review fix: bound call results (Bound / boundGenerators) so effect variables never escape by name; 4 more error tests and bound.nra dump.
