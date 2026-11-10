# Log: cli-skip-teardown

- Base: feat/batch-5 at 61199a64. Branch: the same plus skipTeardown in emitter.nra and a call in compileToC (driver.nra) and runEmit.
- No process-exit function exists in std, and adding one needs a seed update, which this node may not do. So the final release is skipped by parking the Checker and the Package in a Cell that holds a value that holds the Cell (a documented identity cycle, never freed). The process exit frees the memory. Under NARRA_CHECK_LEAKS the call returns first, so everything is released and the leak check keeps working.
- Both compilers were built the same way: the seed-built stage compiler emitted selfhost (base source and branch source) and `cc -O2 -DNARRA_RELEASE -std=c11 -w` built each.
- Measure: `/usr/bin/time -l <narrac> emit selfhost` and `CC=true /usr/bin/time -l <narrac> build selfhost -o out` (CC=true skips the C compiler child, so only narrac is counted). 7 alternating base/branch pairs each. Load average 18 to 22 on 8 cores (shared machine), instructions retired are load independent.
- emit, base (M): 1912.8 1886.8 1886.3 1874.8 1884.9 1880.9 1902.4. Branch: 1859.7 1837.6 1843.6 1848.4 1841.3 1862.2 1845.9. Branch won 7 of 7, mean 1889.8 -> 1848.4 (-2.2%). Ranges do not overlap (base min 1874.8, branch max 1862.2).
- build, base (M): 2615.4 2643.4 2627.0 2611.2 2641.0 2616.0 2642.2. Branch: 2583.1 2572.1 2580.4 2585.0 2592.9 2594.7 2596.2. Branch won 7 of 7, mean 2628.9 -> 2586.3 (-1.6%). Ranges do not overlap.
- The profile said 6.3% for the teardown; 2.2% shows. The rest of the teardown is inside emitProgram (the Emitter and its tables), in emit.nra, which this node does not own.
- With NARRA_CHECK_LEAKS=1: `narrac emit selfhost` exits 0, no leak report. Instructions 1951M on the branch against 1943M on base (the extra is the Env.get call).
- Tests: lsp suite, sources suite with NARRAC set to the branch compiler; full `sh scripts/test.sh` at the end.
