# Outcome: guide-targets

## Built

- `docs/guide/19-building-for-other-systems.md`, replacing the stub. Sections: code that depends on the system (`Build.os`, `Build.arch`, `Build.debug`, build conditions, `Build.hostTarget()`), functions that exist on one system (`@target`, use rules, `check --all-targets`, reading the C of another target), building for another system (`--target`, cross `CC` with a `zig cc` wrapper, `-musl` rows, rejected combinations), WebAssembly with WASI, Windows notes, debug and release builds with the cost model, and `std/simd`.
- `docs/guide/18-tools.md`: the paragraph on `std/build` and `--all-targets` is now one sentence that links to chapter 19.

## Deviations

- Audit rows covered: 168, 169, 64, 167, 84. `narrac targets`, `narrac emit` and `-emit-c` stay in chapter 18 (written by guide-tools-gaps); chapter 19 links to them and adds `--target -emit-c` and `emit` under `NARRA_HOST_TARGET`.
- The chapter's `narra` samples avoid output that depends on the host or on the build: macOS and Linux give the same text, and `Build.debug` is a `narra fragment` because the guide suite runs each sample in debug and release.
- Cross builds, the `@target` error text, the WASI check, sizes and the generated C are `sh` or `narra fragment` blocks that I ran by hand (below).
- Windows behavior (row 64) is a list taken from the specification. I could not run a Windows program, so `Path.join("C:\\a", "..\\b")` giving `C:/b` is not checked here. The other-system result (`C:\a/..\b`) was run on the host.
- Not taught: `NARRA_HOST_TARGET` is named only in passing; the TCC limit of `std/simd` and per-type lane rules are left out as too niche.

## Spec against compiler

- None found. Every claim I could run matched: E1600 and E1601 text, `test --target` rejected with exit 2, E0991 text for `@target` and for `Process.run` on WASI, `hostTarget()` giving `None` on WASI. SPEC.md says `build -sanitize --target` and unknown names are errors but not the exit code; both exit 2.
- `docs/guide/README.md` lines 61 to 69 still list "SIMD vectors", "build conditions" and "building for another system" as built but not taught. That is outside this node; the README node should update it.

## Stubs left

- None.

## Gotchas

- The guide suite runs every `narra` sample in debug and release, so a sample cannot print `Build.debug`.
- The error text of a host check names the host's row (`MacOS-Arm64` on a Mac), so the `@target` error is a fragment with a hand-run `sh` block.
- Chapter 16 links to this chapter's `@target` text without an anchor; chapter 18 links to the anchor `#functions-that-exist-on-one-system`.

## Verification

- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 559 cases.
- plain-writing checker: one long-sentence warning left (line 290, a list item of 28 words).
- `pnpm --dir website build`: passes.
- Run by hand with the prebuilt compiler on an arm64 Mac, `CC` a script calling `zig cc`: `build --target` for Windows-X86_64 (`file`: PE32+), Linux-Arm64-musl (`file`: ELF aarch64, static), Linux-X86_64-musl and Wasi-Wasm32; the WASI module under `wasmtime run -W exceptions=y` (prints `list files with ls`; `hostTarget()` gives `None`); `build -emit-c` for Windows against `emit` on the host (`"dir"` present only in the Windows C); `check` and `check --all-targets` on a bad `@target` use; `check --target Wasi-Wasm32` on `Process.run`; unknown target, `-sanitize` with a target and `test --target`; WASI debug and release sizes; `emit` of the generic and `mutable` example.
- Not run: any Windows binary, any Linux binary (docker could not see the scratch folder).

## Size

- Estimated M, actual M.
