# Outcome: unsafe-ptr

## Built

- `unsafe { ... }` blocks (D95, C interop step 1 part 1). Parsed in `parseStmtBody` (`parser.nra`) when the identifier `unsafe` is followed by `{`, so `unsafe` stays an effect name and needs no lexer change. AST: `Block.isUnsafe: bool = false` (`ast.nra`), carried as `Stmt.BlockStmt`, so every existing pass treats the block as a nested scope with no new case. fmt prints `unsafe ` before it (`printBlockStatement`, `fmt.nra`); astdump adds the `unsafe` flag to the `block` line.
- `Ptr<T>`: `Type.TPtr(element)` (`types.nra`), key `R(<element>)`, source spelling `Ptr<T>`, C id `internedName("R", "Ptr", [element])`, C type `<element C type>*` (`cnames.nra cType`/`cachedCType`). Cases added to `unify`, `templateEdges`, `substitute`, `shapeMismatch`, `containsTypeParameter`, `containsOpen`, `containsParam`, `computeImplements` (no std trait), `mapKeyProblem` and the decls dump. A Ptr is plain: not counted, not tracked, not a contents type, 8 bytes in the layout tables (the `_` defaults).
- The type `Ptr` resolves only in a module that imports `std/ptr`, by alias or by name (`importsPointers`, `selfhost/src/unsafecode.nra`, called from `resolveType` in `resolve.nra`). The element is checked once every struct has its fields: `resolvePointerType` queues it in `Checker.pointerElements` (`PointerCheck` in `types.nra`) and `flushPointerElements` runs at the start of `flushMapKeys` (`traits.nra`). Allowed: number types, `bool`, `char`, `Ptr`, and structs (tuples included) whose fields are all such values. Everything else, type parameters included, is E0998.
- `std/ptr` (`selfhost/src/stdptr.nra`: rows `POINTER_FUNCS` and checks; `selfhost/src/emitpointers.nra`: emitters): `alloc(count): Ptr<T>` (`uses alloc, unsafe`; `T` from the expected type; `narra_alloc` plus `memset` to zero, so the leak check counts it), `free(pointer)` (`if (p) narra_free(p);`), `read(pointer): T` (`(*(p))`), `write(pointer, value)` (`*p = value;`, the pointer evaluated first), `offset(pointer, count): Ptr<T>` (`((p) + (n))`); these five have the `unsafe` effect. `null(): Ptr<T>` (`((T*)NULL)`) and `isNull(pointer): bool` are safe.
- The unsafe effect (`effects.nra`): the reporting walk collects each call's own effects apart (`walkCall` / `walkCallEffects`) and reports E0996 `<callee> is unsafe; call it inside an \`unsafe { }\` block` when the call has `unsafe` and `EffectState.unsafeDepth` is 0. A function holding an `unsafe` block gets the effect through those calls: exported functions must write `uses unsafe` (E0800 as before), private ones infer it.
- `@trusted`: a `KNOWN_ATTRIBUTES` row without arguments (`parser.nra`). `analyseFunction` drops `unsafe` from a trusted function's inferred set (`stopUnsafeEffect`), and `checkTrusted` (`unsafecode.nra`) reports E0997 for `@trusted` on a function without an `unsafe` block (an error, not a warning) or with `uses unsafe`.
- Codes E0996, E0997, E0998 with rows and explanations in `diagnostics.nra`.
- Docs: SPEC.md section 9 rewritten (unsafe blocks, the effect, `@trusted`, `Ptr<T>` and what it may point at, pointer values as data, std/ptr, a sample; `@layout` and C-callable functions marked "not built yet"), `std/ptr` in the section 3 module list, a D95 amendment in DECISIONS.md, compiler/README.md ("Unsafe code" section, "Not supported yet" updated).
- Tests: `testdata/unsafe_ptr.nra` (+ `.out`): alloc/write/read/offset/free in `unsafe`, an exported `uses unsafe` function called from an `unsafe` block, two `@trusted` wrappers called without one, a struct with a `Ptr` field, an `Array<Ptr<u8>>`, a linked pair of structs, `Ptr<Ptr<int>>`, `Ptr.free` of a null pointer; every allocation freed, so it runs clean under the leak check and the sanitizers. Errors: `testdata/errors/unsafe_outside_block` (E0996 on `Ptr.read`), `unsafe_effect_caller` (E0996 on a private function with an unsafe block), `unsafe_undeclared` (E0800 for an exported function), `unsafe_trusted_no_block` and `unsafe_trusted_uses` (E0997), `unsafe_ptr_element` (E0998 for `string`, a struct with a string field, a type parameter, an array), `unsafe_ptr_calls` (E0998 for `Ptr.alloc` without an expected type and a non-pointer argument). AST dump `testdata/astdump/unsafe_block`, fmt `testdata/fmt/comment_unsafe_block`.

### For c-interop

- **Ptr's C form**: `Ptr<T>` is `T*` with `T` the element's C type (`uint8_t*` for `Ptr<u8>`, `int64_t*` for `Ptr<int>`, `S_xxx*` for a struct, `uint8_t**` for `Ptr<Ptr<u8>>`). There is no `Ptr<void>` (`void` cannot be written as a type); use `Ptr<u8>` for C's `void*`, or add a `void` element if a signature needs the exact C type. A `Ptr` is passed and returned by value, so an extern signature can use `cType` as is.
- **How the unsafe effect is checked**: give every extern function the `unsafe` effect (as `FuncDecl.effects`, or as the effect set `funcEffects` returns for it). Nothing else is needed: `walkCall` reports E0996 for a call outside an `unsafe` block from the call's own effects, whatever the callee is (a `Callee` with a declaration, a `Ref.Builtin` row, or a function value whose type lists `uses unsafe`). The `via` of the `unsafe` effect names the callee in the message (`callee.label` for declared functions).
- **How `@trusted` is read**: `findAttribute(declaration, "trusted")` on the `FuncDecl` (attributes outcome). The effect pass checks it in `analyseFunction`, after walking the body: `checkTrusted(checker, declaration, attribute, state.sawUnsafeBlock)` and `stopUnsafeEffect(out)`. `narra audit` can list `@trusted` functions with the same lookup and `unsafe` blocks with `Block.isUnsafe` (they are `Stmt.BlockStmt` statements).

## Deviations

- `selfhost/src/stdptr.nra` (new, not in `owns:`): the std/ptr rows and checks. The checks sit above `exprs.nra`, while `unsafecode.nra` must sit below `resolve.nra` and `traits.nra` (both import it), so one module cannot hold both without an import cycle; `selfhost/CLAUDE.md` names `src/std<module>.nra` as the place for a std module's rows and checks.
- `selfhost/src/stdlib.nra` and `selfhost/src/stdcalls.nra` (not in `shared:`): one `registerAll` line and one dispatch `if`, which `selfhost/CLAUDE.md` requires for a new std module.
- `Ptr.null` and `Ptr.isNull` added beside the five functions of D95; they touch no memory and are safe (no `unsafe` effect). c-interop needs both for C functions that take or return NULL.

## Review round 1

Fixes from the fresh review (second commit):

- Generators: a `for` loop or `yield*` over a generator whose body has the `unsafe` effect is E0996 outside an `unsafe` block (`walkGeneratorRun`, new E0996 format `%s runs unsafe code; iterate it inside an \`unsafe { }\` block`). Other effects of a generator were already charged there and reach E0800 for an exported function; the review's e04b passes because `run` is an entry function, which may use every effect.
- Traits: an implement block member may use `unsafe` only when its trait function declares it (E0800, two new formats, `checkImplementationEffects` in `effects.nra`), and a `Trait.f(value)` call on a type parameter now has the effects its trait function declares (`addTraitCallEffects`). `@trusted` is allowed on implement block members (`parseAttributeList` in `parseImplDecl`; any other attribute there is E0990 with a new format); fmt and astdump print member attributes.
- The subset rule is enforced for `unsafe` only. For `alloc` and `io` it broke 21 cases (the guide's traits chapter, a SPEC sample, `testdata/traits.nra`, `tree_shake.nra`, `call_type_args.nra`, `packages/traits`, and the existing `effects_impl_through` tests, which expect impl effects to flow through direct calls): implementations of pure trait functions such as `Show.show` allocate. Recorded in compiler/README.md "Not supported yet".
- E0996 names the callee as written (`g`, `first`, `Ptr.read`) before falling back to the effect's origin.
- `Ptr<Count>` for `distinct type Count = int` is allowed: `findPointeeProblem` looks through a distinct type to its base.
- SPEC section 9 says the type is written `Ptr<T>` whatever the alias (`P.Ptr<int>` is an error).
- astdump gives an unsafe block the position of its `unsafe` keyword (`Block.keywordPos`).

## Inbox

- Decide whether an implementation's `alloc` and `io` must be a subset of its trait function's effects, as `unsafe` now is. Enforcing it means adding `uses alloc` to trait functions such as `Show.show` and `Label.label` in the guide, SPEC, samples and `testdata/packages/traits`, and changing `effects_impl_through`.

## Stubs left

- None in this node. `extern` blocks, `CString` and `narra audit` are node c-interop; `@layout` and C-callable functions are c-abi-export.

## Gotchas

- `Ptr<T>` is rejected for a type parameter, so a generic helper over pointers cannot be written yet; it needs a bound meaning "no counted fields".
- `Ptr.free` calls `narra_free`, which is right only for memory from `Ptr.alloc`. Memory a C function allocated must go back through C's `free` (an extern call); passing it to `Ptr.free` corrupts the allocator in non-sanitizer builds. c-interop should say so where it documents extern allocation.
- An `unsafe` block is a statement, not an expression: `let value = 0` before it, assign inside.
- A function value of a private function with an inferred `unsafe` effect: calls through it are charged by `valueEffects` like other effects; only a written function type with `uses unsafe` is guaranteed to carry it into another module.
- What extern calls inherit (c-interop): every extern function has the `unsafe` effect, so its call needs an `unsafe` block through `walkCall`, whether it is called directly, from a private function (whose inferred effect then reaches its callers), from a generator body (a `for` over that generator needs a block, `walkGeneratorRun`), or from an implement block member (E0800 unless the trait function declares `uses unsafe`, or the member is `@trusted`). A function value of an extern function carries `unsafe` only if its type is written with `uses unsafe`.
- fixed-arrays adds a `Type` variant to the same exhaustive matches (`buildKey`, `typeString`, `typeId`, `cType`) and appends E0994/E0995 at the same place in `diagnostics.nra`; the second to land keeps both.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/unsafe-ptr/build`: stage 1 and stage 2 emit identical C; the seed is older than the source, as expected mid-batch.
- `NARRAC=<build>/narrac sh scripts/test.sh --filter unsafe --update`: 51 cases, all pass; goldens written.
- `NARRA_CHECK_LEAKS=1 narrac run testdata/unsafe_ptr.nra`: the expected output, exit 0.
- `sh scripts/test.sh` (log `/private/tmp/claude-501/unsafe-ptr/full.log`): PASS in 284 s, exit code 0, 8 shards and the bootstrap job. The first full run failed three goldens whose text lists names (`std/ptr` in the unknown-module list: `testdata/errors/unknown_module.err`, `testdata/declsdump/imports/err_unknown_module.dump`; `@trusted` in the known-attribute list: `testdata/errors/target_attribute_unknown.err`), rewritten with `--update`, and the sanitized fixed point, because `effects.nra` was edited while that run was going; a rebuilt sanitized compiler emits the same C, and the final run passes.

- Review round 1: `sh scripts/test.sh` (log `/private/tmp/claude-501/unsafe-ptr/full2.log`): PASS in 220 s, exit code 0. Review probes in /private/tmp/claude-501/review-unsafe-ptr/p/ (e04, e04c, e04d, e04f, e19, e08, e08b, e08c, e08d, t02) give the expected errors or pass.

## Size

- Estimated M, actual M.

Landed on feat/batch-8 as 90204ffc.
