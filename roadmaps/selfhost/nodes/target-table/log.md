# Log

- Rows, flag, targets command and E1600/E1601 done; seed lacks Os.Wasi so hostOs/hostArch became Option.
- Goldens of four build-targets err files updated for the new rows.
- Fast check: PASS, 9783 cases.
- Criterion 1 on 2026-11-15: `CC=<sh wrapper: exec zig cc "$@"> narrac build --target Linux-X86_64-musl -o h testdata/build-targets/cross_hello.nra` gave a static x86-64 ELF; `docker run --rm -i --platform linux/amd64 alpine sh -c 'cat > /h && chmod +x /h && /h' < h` printed `cross` (the program's text, exit 0). A bind mount of /private/tmp was not visible to docker, so the binary went in over stdin.
