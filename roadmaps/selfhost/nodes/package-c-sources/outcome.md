# Outcome: package-c-sources

## Built

- `extern { ... } from "./file.c"` (or `"../file.c"`): a C file of the package, compiled and linked. The path is relative to the folder of the `.nra` file that holds the block, like an import path (`findSourceFolder` in `csources.nra` gives that folder from the module path); `[c]` paths stay relative to the package folder. The result must be inside the package, `..` and symbolic links both (E1661). The links are read with one `sh` call that prints real paths (`REAL_PATHS_SCRIPT`, `findEscapingLink`), since `std/fs` has no `realpath` and adding one means editing the runtime and a seed update; without `sh` only the lexical check applies. `isCFilePath` lives in `syntax/ast.nra` for the parser, `csources.nra` and `emitexterns.nra`. A bare name is still a system library (`-l`). A path that does not end in `.c` is E1000 (parser, `selfhost/src/syntax/parser.nra`). `ExternBlock.library` holds the path as written; there is no new AST node, so astdump, fmt and the checker needed no case.
- Manifest tables `[c]` and `[c.os.<MacOS|Linux|Windows|Wasi>]`, each with `sources` (patterns, `*` in a file name, `!` removes), `include` (`-I`) and `defines` (`-D`). Parsed in `selfhost/src/driver/manifest.nra` into `Manifest.cTables` (`CTable`); an unknown key or system is E1000. `DependencyPackage.cTables` carries it for dependencies.
- `selfhost/src/driver/csources.nra`: `PackageC` (name, folder, tables), `normalizeInside` (the inside-the-package rule), `checkSources`, `checkExternFiles` (load time, E1660 and E1661), `listSources` (pattern expansion for a system), `collectUnits` (the `CUnit`s of a build: every package's `[c] sources` plus the reached `from` files, once each, each with the include folders and defines of its own package).
- `selfhost/src/driver/cobjects.nra`: `buildObjects` compiles each unit with the same compiler and base flags as the emitted C (`cBaseArguments` in `driver.nra`, split out of `cCompilerArguments`) plus `-I`/`-D`, and returns object paths. E1662 names the package, the file and the first error line of `cc`.
- `Package.cPackages` (`modules.nra`) lists the root package first, then each dependency. `CompiledProgram.cFiles` (from `EmitResult.cFiles`, `emitexterns.nra` `collectCFiles`) lists reached `from` paths with their module. Link line: emitted C, objects, `-l` libraries, `-lm`.
- E1660 missing file or pattern with no match, E1661 path outside the package, E1662 C compile error, in `selfhost/src/driver/diagnostics.nra`.
- SPEC sections 9 and 15, `compiler/README.md` and the guide chapter 16 ("C files in a package", with a runnable example) describe it.

### Cache layout and key

- Folder: `<cache root>/objects/`, where the root is `$NARRA_CACHE` or `$HOME/.cache/narra` (`findCacheRoot`, now exported from `depcache.nra`). With no root, `<scratch>/objects` inside the build's temporary folder.
- First key: sha256 of `compiler \n first line of "compiler --version" \n flags \n package/relative path \n file text`. The version is read once per build, only when a C file is to compile. `<firstKey>.deps` lists the headers (one path per line) that the first compile found with `-MMD`; the rule is parsed with its escapes (`\ `, `\#`, `$$`, backslash-newline).
- Object: `<secondKey>.o`, where secondKey is sha256 of the first key, the names in the include folders, the folder of the file and the folder of each header, and the text of each header in the `.deps` file. A missing `.deps` file or object is a miss and compiles; a header that cannot be read is a miss too (and the object is then not cached). The object and then the `.deps` file are copied to a staging name and renamed. Old objects stay. Adding a file to one of those folders recompiles every file of that package that uses the folder.
- `NARRA_C_LOG=1` prints `compiled <package>/<file>` to stderr for each real compile; the tests prove cache hits with it. Relative header paths in the `.deps` file are relative to the working directory of the build, so a build from another folder misses.

## Proof

Not a fix node.

## Deviations

- Added tests to the `commands` variant of `selfhost/tests/src/packages.nra` (an environment parameter and a `count` step), which the brief lists as shared. Existing commands cases now also set `NARRA_CACHE` to a folder in the case copy.
- `removeTempDirectory` in `driver.nra` lost its `removeC` parameter and removes the whole folder.
- Checks for E1660 of `[c] sources` run at load only for the base `[c]` table; a `[c.os.*]` table is checked when building for that system.
- The `count` step of the first version was replaced by the `NARRA_C_LOG` compile marker.
- A new E1000 row ("not a C file") and two E1000 rows for `[c]` keys and systems were added, beside the three reserved codes.

## Stubs left

- `from "<package>/<path>"` and `[c] git`: node c-packages.

## Gotchas

- For c-packages: `collectUnits` takes the packages in `Package.cPackages`; a fetched C package adds a `PackageC` whose `directory` is the upstream checkout (the `[c]` paths are relative to it). `createUnit` builds the include and define flags per owner, so a dependent's `-I` from a C package belongs in `createUnit` (add the includes of the packages it depends on), never the defines.
- `findOwner` maps a block's module path (`<package>/<path>`) to its package by the first path part; `from "<package>/<path>"` needs a separate resolution before that.
- A C function named `narra_*` clashes with the runtime names (E1001), so test C functions use another prefix.
- The seed is not involved: no feature the seed lacks is used. Seed update: not needed.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh out`: stage 1 and stage 2 emit identical C.
- `NARRAC=out/narrac sh scripts/test.sh --suite packages`, `errors`, `codes`, `guide`, `spec`, `sources`: pass.
- `sh scripts/test.sh`: see the final report of the worker.

## Size

- Estimated M, actual M.
- Risk estimated medium, measured medium: the link line and temp folder handling changed for every build.
