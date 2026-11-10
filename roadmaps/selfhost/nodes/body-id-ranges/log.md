# Log: body-id-ranges

- Branch feat/body-id-ranges reset to feat/batch-5 (91233c52).
- Both compilers built with `sh selfhost/bootstrap.sh <dir>` (stage 3 binary, the stage-2 runtime). Base = 91233c52, branch = base plus the change.
- Measure: `/usr/bin/time -l <narrac> emit selfhost`, instructions retired, both compilers emitting the same input (the branch's selfhost), alternating base/branch.
- Load 8 to 16 throughout. Run-to-run spread on one side: about 4% (1739 to 1850 M on base).
- Round 1, 8 pairs: branch lower in 5; medians 1773 M base, 1764 M branch.
- Round 2, 15 pairs (a few runs lost their counter line): medians 1758 M base (15 runs), 1745 M branch (11 runs); minima 1739 and 1718.
- Round 3, 10 complete pairs: branch lower in 7; mean difference 18 M (about 1.0%).
- Emitted C: base compiler and branch compiler emit byte-identical C for the branch's selfhost (cmp, 167554 lines).
- Fixed point holds (stage 2 = stage 3).
