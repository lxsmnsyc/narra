# Log: package-c-sources

- Checked first: no C source support existed (`from` took only a library name, `[link]` only `search`, the build cache held only dependency checkouts).
- `from "./x.c"` needs no new syntax node: the parser keeps the path in `ExternBlock.library`. The only parser change accepts a `./` or `../` path that ends in `.c`, so astdump, fmt and the checker's block handling needed no case.
- The path checks run at package load (`checkExternFiles`, `checkSources` in `csources.nra`), so `narrac check` reports E1660 and E1661 without a build. Only the `[c]` table and the `[c.os.<system>]` table of the build system are checked for existence at build time (`listSources` with the system).
- The module of a block's function (`FuncInfo.module`, `<package>/<path>`) gives its package, so the unit gets the flags of its own package.
- Object cache: first key from compiler, flags, package path and file text; the `.deps` file beside it holds the headers of the first compile (`-MMD`); the object is stored under a second key that adds the header texts. A header edit changes the second key only.
- Test harness: the `commands` variant of the packages suite got an environment (`NARRA_CACHE` in the case folder) and a `count :: PATH :: N` step. The cases are `testdata/packages/c-sources/build` and `failure`.
- `removeTempDirectory` now removes the whole scratch folder, since it can hold objects of a build with no cache folder.
