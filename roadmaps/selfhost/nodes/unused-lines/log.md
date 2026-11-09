# Log: unused-lines

Method: two compilers built with `sh selfhost/bootstrap.sh`, base (origin/feat/batch-3, 6924411) and branch. `narrac check selfhost` on the package, wall clock with `time.perf_counter`, 1 warmup, 20 runs each, interleaved (base, branch, base, ...).

- First look, 8 sequential `/usr/bin/time` runs of base: 1.07 to 1.34 s (machine loaded, spread 25%). Not usable.
- Interleaved, 20 runs: base median 0.407 s (mad 0.002, min 0.403, max 0.453). Branch median 0.382 s (mad 0.002, min 0.376, max 0.389). Gain 25 ms, 6.1%, about 12 times the mad. This matches the 7% share the lsp-incremental profile gave for the unused check.
- Change: the source is split into lines on the first unused import or first unused function of a module, not for each module.
