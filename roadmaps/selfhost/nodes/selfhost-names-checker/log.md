# Log: selfhost-names-checker

- 2026-10-31: baseline on main 54773bb: TestSelfhostBodies compared 143, skipped 202; Decls 249 pass / 84 skip; Traits 270 / 63; Imports 311 / 32.
- Renamed struct fields first: Field.t -> fieldType, ConstInfo.t -> declaredType, ImplInfo.t -> forType, GenericEdge.to/t -> target/instance. Then `c` -> `checker` in every owned file where c is the Checker.
- Renames done with a token-aware script (skips strings, template text and field accesses) scoped per function, with `narra check selfhost` after each file.
- Order: types, builtins, stdlib, main, check, declare, generics, resolve, scopes, modules, declsdump, traits, then the body modules (only `id` -> `nodeId` and `op` -> `operator` had slipped there).
- Function renames: envOf -> makeEnv, exportsOf -> collectExports, tupleOf -> findTupleType, typeTraitsOf -> findTypeTraits, fieldsOf -> listFields, implementsOf -> computeImplements; importDecl -> importDeclaration (the local name took it).
- Mid-way parity run: TestSelfhostBodies compared 143, skipped 202, same as main.
