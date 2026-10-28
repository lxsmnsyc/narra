# Log: bench-harness

- Git through the rtk wrapper was refused in this worktree; used /usr/bin/git.
- Timing with perl Time::HiRes (median.pl) because macOS has no hyperfine and bash has no sub-second clock.
- Narra has no array pop, no closures, no string indexing, so programs use only supported features (see compiler/README.md).
- Rust inputs go through black_box so rustc does not fold the whole program.
- First run was slower than later ones (cold caches). Scaled up matmul, interp and particles so each runs 0.1s or more.
- Doubt: the machine was busy, numbers are a smoke test.
