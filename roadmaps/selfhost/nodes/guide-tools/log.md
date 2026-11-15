# guide-tools log

- Branch renamed to docs/guide-tools, reset to feat/batch-16 (9f4017d7).
- Not fulfilled yet: docs/guide has no tools chapter; chapter 12 is the TypeScript comparison.
- The guide suite checks code blocks only, not links; links checked by a script over docs/guide.
- Planned list: regions node is dropped (not in SPEC either); async, threads, SIMD and @const are built. Not built: package dependencies (deps-* nodes todo), package exports table, workspaces, C-callable functions (c-abi-export), other platforms (cross-targets, wasi, windows).
- Built narrac at /private/tmp/claude-501/guide-tools/out; captured every shell output in the chapter from it, on a scratch weather package.
- explain E0800 prints the trait rule format (code shared by several messages); used E0003 for the examples.
- Moved the comparison to 13, added a Tools section there; README Planned list rewritten; guide suite PASS 328.
- Self-review round 1: clarified hello.nra comment line, placeholder wording, apply-fix on packages, fmt sample wording, printWidth 60 sentence, recursion wording, -time variance.
- Self-review round 2: no findings. Fast check PASS, 9781 cases.
