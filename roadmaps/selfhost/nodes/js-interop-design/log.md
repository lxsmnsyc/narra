# Log: js-interop-design

## 2026-11-15

- Branch `spike/js-interop-design`, reset to `feat/batch-16` at 9f4017d7.
- Debate folder `roadmaps/selfhost/debate/js-interop/`, ledger from `debate.py init --max-rounds 2 --roadmap roadmaps/selfhost`. Full debate (two subagents) because the brief asks for a goal debate and the answer decides whether four held M nodes are built. Two rounds, not three: the questions are narrow and the proofs settle the large feasibility claims.
- Proof T1 (scratch `/private/tmp/claude-501/js-interop-design/t1`, about 9 minutes): the seed builds for wasm32-wasi with zig after three runtime patches. See `debate/js-interop/evidence/T1-narrac-wasm.md`. The two patches beyond the known process block were found by running it: `Env.args` reads `/proc/self/cmdline` off macOS, and the compiler writes output through `/dev/stdout`. Both are runtime work for wasi-runtime or a new child.
- Proof T2 (scratch `/private/tmp/claude-501/js-interop-design/t2`): downloaded `@yowasp/clang@22.0.0-git20542-10` from npm into the scratch folder only (109 MB). Nothing installed outside the scratch folders. See `evidence/T2-clang-wasm.md`.
- Load average reached 48 during T2 (other workers building), so T2 times are upper bounds. Builds waited while load was above 30.
- Debate run with two subagents (proponent, opponent), two rounds, both sealed. Round 1: the proponent scoped npm to narrac only. The judge added J17 (the goal covers any Narra program, as esbuild and oxc are programs), and round 2 widened P8 to a package template.
- Judge measurements after round 1: a fresh instance per edit is 93 to 325 ms in V8 (J15). narrac.wasm runs with wasm exceptions turned off (J14). Warm clang in wasm takes 0.47 s on hello and 2.9 s on regex, and the runtime header is about 0.4 s of that (J18). J18 rejected the opponent's precompiled-header alternative.
- Stopped at `debate.py status` = ask-user: Q1 (Run in the page or on a server) and Q2 (esbuild form or oxc form) go to the maintainer. D123, debate.md and outcome.md give the recommended answers.
- No compiler code changed, so `sh scripts/test.sh` was not run.

## 2026-11-15, after the maintainer's answers

- Q1: park the REPL. Q2: "ideally it should be either", so both the command form and the Node-API form, each with native packages per platform and a wasm fallback.
- Proof T3 (scratch `/private/tmp/claude-501/js-interop-design/t3`, about 10 minutes, nothing installed): hand-written Node-API glue around emitted C builds into a 51 KB `.node`, and a call takes 0.18 µs. See `debate/js-interop/evidence/T3-node-addon.md`.
- D123 rewritten for both forms. The ledger records the answers as user rulings (P7 rejected, P1 upheld, REPL claims open). outcome.md has the revised children, the amendments for the C chain briefs, and two new questions (emnapi, synchronous exports).
