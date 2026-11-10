# Log: profile-compiler

- Branch `spike/profile-compiler` from `feat/batch-4` at 55fb951.
- `xctrace` needs full Xcode, and only the command line tools are installed. Used `/usr/bin/sample`.
- `sample` cannot attach to a 0.6 s process in time, so `build.sh` makes `narrac-loop`: the same stage 2 C with `main` calling the driver `NARRA_REPEAT` times. `NARRA_REPEAT=3` prints the stage 2 C three times, byte for byte (checked with `tail | cmp`).
- The first sample had `<deduplicated_symbol>` frames: the linker folds identical functions. `-Wl,-no_deduplicate` fixes it.
- Wall time, 5 runs: release `narrac` 0.60 s, `narrac-g` 0.60 s, `narrac-loop` 0.60 s per run, `narrac-noinline` 0.83 s.
- Samples: `opt` and `opt2` (`narrac-loop`, 30 s each, about 20,500 main-thread samples each), `noinline` (`narrac-noinline`, 30 s, 21,194 samples). 1-minute load 2.1 to 2.5 on 8 cores during all three. The two `opt` runs agree within 0.5 points on every category and phase.
- At -O2 the runtime helpers (`lstr_new`, `A_Array_*_get`, most `retain_*`) are inlined, so their cost shows in the caller. The `-fno-inline` build keeps their frames. Its shares are used only to split allocation by helper.
- The bash tool refused commands that name the worktree path together with subshells, so every step is a script in this folder, run through a symlink.
