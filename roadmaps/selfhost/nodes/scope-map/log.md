# Log: scope-map

- Base: feat/batch-5 at 48823a59. Command: `/usr/bin/time -l <narrac> emit selfhost` (instructions retired), same source tree for both compilers.
- Quiet-ish base, 7 runs: 2.622e9 to 2.635e9 (0.5% spread), one outlier 2.674e9. Load average 3 to 4.
- Load rose to 10-17 from other workers. Paired runs (base, branch, alternating), load 10:
  base 2715, 2667, 2712, 2658, 2656, 2672, 2673 M; branch 2668, 2653, 2688, 2622, 2625, 2620, 2633 M.
  Branch is lower in 7 of 7 pairs, by 14 to 52 M (median 36 M, about 1.4%). Medians 2672 M vs 2632 M.
- Design: one `localNames` map, plus an undo log (`scopeUndo`, `scopeUndoCount`, `scopeStarts`). popScope restores replaced names. Narrowing keys use the same map; endNarrowing sets -1 and patches the logged previous values. LocalInfo.depth replaces the per-depth scan in isDeclaredBelow.
- Full check: sh scripts/test.sh PASS (5860 cases, 0 failed, 189 s).
