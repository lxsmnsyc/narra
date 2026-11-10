# Log

- Proof: `out/narrac build -o r repro.nra` on feat/batch-4 (e64faa2) exits 139.
- ASan build of the compiler (stage 2 C, `cc -g -fsanitize=address`): stack overflow, `isUseful` (patterns.nra) calling itself.
- Cause: with a wildcard in the vector, `isUseful` tried every constructor even when no row had one. On a recursive enum (`Chain.Link(.., Chain)`) the wildcards kept expanding, so the walk never ended.
- Fix: a wildcard vector head goes to the default matrix unless the rows cover every constructor (`coversEveryConstructor`).
