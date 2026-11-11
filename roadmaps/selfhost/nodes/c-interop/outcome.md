# Outcome: c-interop

## Built

- `extern { ... } from "library"` blocks (D95, C interop step 1 part 2). Parsed by `parseExternBlock` / `parseExternMember` (`selfhost/src/parser.nra`) when the identifier `extern` is followed by `{` at the top level, so `extern` stays a plain name elsewhere (no lexer change). Members are signatures with optional `export` and optional `@extern("name")`; type parameters, generators, `uses`, a body, other attributes, an empty block, a missing `from "lib"` and a library name cc would not take are E1000.
- AST (`ast.nra`): `FuncDecl.externBlock: Option<ExternBlock> = None`; `ExternBlock { pos, end, library, libraryPos }`. Each member is a separate top-level `Decl.Func` that carries the same `ExternBlock` value (members of one block share `pos`), with `body: None` and `effects: ["unsafe"]`. `AttributeArgument.text: Option<string>` holds a string argument; `AttributeRow.takesText` marks `@extern` in `KNOWN_ATTRIBUTES`. `@extern` on a function with a body is E1000 (`rejectExternAttribute`).
- Effects: `funcEffects` (`effects.nra`) uses a member's declared effects, so every unsafe rule applies through the existing paths: E0996 outside `unsafe`, `@trusted` wrappers, inferred effect on private callers, generators (`walkGeneratorRun`), trait implementations (E0800), function values (type `(...) => T uses unsafe`, E0119 into a type without it). Probed in `/private/tmp/claude-501/c-interop/p/t2.nra` and `t3.nra`.
- Checks (`selfhost/src/externs.nra`, `checkExternFunctions`, called at the end of `checkEffects`): E1002 for parameter/result types other than numbers, `bool`, `Ptr<T>` (and nothing as result) or a `mutable` parameter; E1001 for a C name that is not a C identifier, is `main`, starts with `narra_`/`lstr_`, is a string literal name (`L<digits>`), has a module hash (`f_3fa9c1e7_x`), or equals a generated function or type name of the program, alone, with a suffix (`<name>_...`) or after a helper's operation (`release_<type>`, `eq_<type>`: any `<word>_<name>`). The unused-parameter check skips members (`reportUnusedLocals`, `check.nra`).
- Emitter (`selfhost/src/emitexterns.nra`): a member keeps its generated hashed C name, so calls and function values are unchanged. `emitReachableBodies` writes no body for a member; `emitPrototypes` adds `emitExternPrototypes`: three `NARRA_C_SYMBOL` macros, then `extern <type> f_<hash>_name(<params>) __asm__(NARRA_C_SYMBOL("cname"));` per reached member (`__USER_LABEL_PREFIX__` gives macOS its `_`). The C identifier never equals the C name, so a header declaring the same function with other types (`size_t strlen(const char*)`, `sqrt` in math.h) does not conflict. Checked with `cc -std=c11 -Wall -fsanitize=address,undefined` on the emitted C of the sample.
- Linking: `EmitResult.libraries` (`emit.nra`) = `collectLibraries`, the libraries of blocks with a reached member, program order, without `c`. `compileToC` (`driver.nra`) returns `CompiledProgram { text, libraries, search }`; `cCompilerArguments` passes `-L<dir>` for `[link] search` (relative to the package directory) before the C file and `-l<lib>` after it, then `-lm` (an `m` block adds nothing). An unreached block links nothing (SPEC section 9).
- `[link] libraries` removed: `Manifest.libraries`, `Package.libraries` and the pkg dump line are gone; any `[link]` key other than `search` is E1000 (`manifest.nra`).
- `std/cstring` (`std/src/cstring.nra`, regenerated into `stdtext.nra`): `export opaque struct CString { text: string }`, `from(text): Option<CString>` (None on a zero byte), `toString(value): string`, `pointer(value): Ptr<u8>`, `read(address: Ptr<u8>): Result<string, string> uses alloc, unsafe`. Primitives `CText.pointer` / `CText.read` in `std/internal/cstring` (rows `CSTRING_FUNCS` in `externs.nra`, emitters `emitCStringCall` in `emitexterns.nra`). The parser accepts `from` as a function name like `of`; `stdType` resolves `Ptr<...>` texts.
- fmt: `printFile` prints consecutive members of one block as `extern { ... } from "lib"` (`printExternBlock`, `isMemberOf`), without the implied `uses unsafe`; the block has trivia key `x<line>:<col>` (a list key for blank lines, `trivia.nra`); `@extern("name")` prints its string. astdump: an `.externBlock` slot with `extern <pos> library="..."` on each member; a string attribute argument dumps as `arg <pos> text="..."`.
- Codes E1000, E1001, E1002 with rows and explanations (`diagnostics.nra`); one new E0990 row (`@extern` without a string).
- Docs: SPEC sections 3 (`std/cstring`), 9 (blocks, effect, types, C names, prototype form and header clashes, linking and tree shaking, memory from C, CString; a checked sample), 12 (`@target` sample as a block, waits for c-audit-target) and 15 (`[link]` = `search` only); D95 amendment in DECISIONS.md; compiler/README.md.
- Tests: `testdata/c_interop.nra` (+ `.out`): `strlen` and `strchr` on a CString from `"c"`, `CStrings.read` of the strchr result, `@extern("abs")`, `sqrt` and `pow` from `"m"`, zero-byte and empty CStrings. `testdata/c_interop_reached.nra`: `cbrt` from `"m"` linked, a block `from "narra_no_such_library"` used only in a dropped Windows branch (builds and runs, so it was not linked). Errors: `c_interop_outside_unsafe` (E0996), `c_interop_name_clash` (E1001: narra_alloc, main, L1, hashed name, E_Ordering, not an identifier), `c_interop_types` (E1002), `c_interop_member_body`, `_member_uses`, `_member_attribute`, `_block_from`, `_library_name`, `_extern_attribute` (E1000). `testdata/manifest/link_libraries` (`[link] libraries` is now an unknown key). AST dump `testdata/astdump/extern_block`; fmt `testdata/fmt/comment_extern_block`. Manifest and pkgdump goldens updated (`manifest/valid`, `pkgdump/link`, `link_bad`, `no_name`, `packages/app`).

### For c-audit-target

- **How blocks are represented**: there is no block node in `File.declarations`; a block is the run of consecutive `Decl.Func` members whose `externBlock` is `Some` with the same `pos`. Find all members with `funcDecl.externBlock is Some(block)`; group by `block.pos`. `findExternBlock(checker, program)` (`externs.nra`) gives the block of a `ProgramFunc.PlainFunc`.
- **Where libraries are collected**: `collectLibraries(checker, reached)` in `emitexterns.nra`, from `ReachableBodies.functions` (the reached flags) in `emitProgram`; it returns `EmitResult.libraries`, which `compileToC` puts in `CompiledProgram.libraries` for `cCompilerArguments`. A block a target does not compile adds nothing as long as no reached code calls it.
- **How `@target` should attach to a block**: `parseFile` calls `parseExternBlock` only when `extern {` starts the declaration; attributes before it currently go through `parseDecl`, which reads them and then fails with E0019 on `extern`. Read the attribute list first in `parseFile` (or in `parseDecl` before testing `isExternBlockStart`), store it on `ExternBlock` (for example `attributes: Array<Attribute>`), and copy the `@target` attribute into each member's `FuncDecl.attributes` too: `resolveFunctionTarget` (`buildconditions.nra`) and the E0991 use check then work per member unchanged, and a member the target lacks is not reached, so its library is not linked. `parseExternMember` rejects any member attribute but `@extern` (E1000); keep that, `@target` belongs on the block. fmt must then print the block's attributes before `extern` (from `ExternBlock`, not from the members), and the member printer must skip the copied `@target`. The trivia record of the block (`x` key) should start at the first `@`, as `parseDecl` does for functions.
- **Audit**: members are `Decl.Func` with `externBlock`; `unsafe` blocks are `Block.isUnsafe`; `@trusted` is `findAttribute(declaration, "trusted")`.

## Deviations

- Files outside `shared:`/`owns:`: `selfhost/src/stdtypes.nra` (one `Ptr` head so a registry row can name `Ptr<u8>`), `selfhost/src/stdlib.nra` (one import and one `registerAll` line for `CSTRING_FUNCS`), `selfhost/src/emitstd.nra` (one dispatch `if` for `CText.`), `selfhost/src/emitexterns.nra`/`externs.nra` are owned.
- `CString` is used as `import * as CStrings from "std/cstring"` with `import { CString } from "std/cstring"`, not `CString.from(...)`: a module alias may not share a type's name (existing rule, as with `Cells`/`Cell`). `Ptr` gets around this with a special case in `resolveType`; doing the same for CString seemed not worth a second special case. SPEC and D95 amendment say so.
- Extra beyond the brief: `export` on a member; `CStrings.pointer` and `CStrings.read` (the "way back"); `[link]` rejects unknown keys (E1000) so the removed `libraries` is an error instead of being ignored; `from` accepted as a function name.
- The extern checks (E1001, E1002) run at the end of the effect pass (`checkEffects`), the one pass both `narrac build`/`check` (`checkForTarget`) and `narrac emit` (`emitter.nra`) run; `body.nra`, where the other end-of-package checks sit, is not shared. So they show only when the bodies have no errors.
- `[link]` rejects unknown keys with E1000 because the manifest has no unknown-key code; `[link]` is the only table that rejects unknown keys (`[package]` and `[[binary]]` ignore them).
- `[link] search` reaching cc has no automated test: the test image cannot build a library without tools beyond cc and sh. Verified by hand with `CC` set to a logging wrapper (see Verification).

## Stubs left

- `@target` on extern blocks and `narra audit` (c-audit-target). `@layout` and C-callable functions (c-abi-export).

## Gotchas

- A member's E1001/E1002 errors appear only after every body checks clean.
- `CStrings.pointer` must not outlive the CString, and C must not write through it (the bytes are a Narra string, possibly a shared literal).
- Generated C uses `__asm__` labels; a compiler without GNU extensions would reject them (clang and gcc accept them with `-std=c11`). Not run on Linux gcc here; glibc's own `__REDIRECT` uses the same form.
- `Ptr.free` must not free memory a C function allocated (SPEC says so).
- New std module `std/cstring` and `std/internal/cstring` change the lists of std modules in goldens that print them (unknown-module messages).

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/c-interop/b2`: stage 1 and stage 2 emit identical C; seed older than source, as expected mid-batch.
- `NARRA_CHECK_LEAKS=1 narrac run testdata/c_interop.nra` and `c_interop_reached.nra`: expected output, exit 0. Emitted C of the probe compiled with `cc -std=c11 -Wall -fsanitize=address,undefined`: no conflict diagnostics, runs clean.
- `[link] search`: package in `/private/tmp/claude-501/c-interop/pkg` with `search = ["lib", "/opt/narra-example/lib"]`, built with `CC=cc-wrapper.sh`: cc got `-L/private/tmp/claude-501/c-interop/pkg/lib -L/opt/narra-example/lib <out.c> -lm`.
- `--filter c_interop --update`, `--filter extern --update`, `--suite spec`, `--suite generated --update`: pass.
- `sh scripts/test.sh` (log /private/tmp/claude-501/c-interop/full2.log): PASS in 482 s, exit code 0, 8 shards and the bootstrap job. The first full run failed three goldens that list names (`@extern` in the known-attribute list: `testdata/errors/target_attribute_unknown.err`; `std/cstring` in the unknown-module list: `testdata/errors/unknown_module.err`, `testdata/declsdump/imports/err_unknown_module.dump`), rewritten with `--update`.

## Review round 1

- E1001 also rejects `<word>_<generated type name>` (the helpers `release_`, `retain_`, `drop_`, `credit_`, `uncredit_`, `eq_` and the other trait helpers put the operation first); case added to `testdata/errors/c_interop_name_clash.nra` (`release_E_Option_e088c655`), the E1001 explanation and SPEC.
- SPEC section 9: variadic C functions cannot be declared (undefined behavior with a fixed prototype; macOS arm64 passes variadic arguments on the stack); they wait for C-side wrappers (c-abi-export). A C predicate returning `int` is declared `i32`, not `bool`.
- `emitExternPrototypes` strips the `static ` prefix by its text.
- `sh scripts/test.sh` (log /private/tmp/claude-501/c-interop/full3.log): PASS in 186 s, exit code 0.

## Size

- Estimated M, actual M.
