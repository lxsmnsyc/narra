# Log: cache-type-names-rest

- Started from feat/batch-5 at 1ee1a90a. Not already done: 88 uncached cnames call sites in emit*.nra and std*.nra.
- Base and branch compilers built the same way: `sh selfhost/bootstrap.sh <dir>` (the resulting compiler equals the build of stage2.c). The first base build was contaminated by edits made while it ran; rebuilt from an archive of 1ee1a90a.
- feat/batch-5 moved on (loop-unique-hoist, 0a681165) while I worked; the base stays at 1ee1a90a, my branch base.
- Both compilers emit the branch's `selfhost`: byte-identical C (cmp).
- Pairs, `/usr/bin/time -l narrac emit selfhost`, instructions retired, 9 alternating base/branch pairs, load average about 10 (1 min) on 8 cores:
  base 1.774 1.759 1.746 1.746 1.743 1.742 1.741 1.742 1.756 G; branch 1.698 1.686 1.686 1.685 1.683 1.681 1.681 1.691 1.708 G. Branch won 9 of 9, about 3.4% fewer; base spread 1.9%.
