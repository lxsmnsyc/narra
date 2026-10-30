# Log: selfhost-modules

- Go dump first (`pkgdump.go`, `narra pkg`), then manifest.nra and modules.nra. The Narra side matched on the first run for all 293 inputs.
- Go sorts the full path list (`sort.Strings`), not the walk order. `a.nra` sorts before `a/b.nra` and `B.nra` before `a-b.nra`. Narra collects recursively and sorts relative paths with the same byte order. Cases in `testdata/pkgdump/nested_order`.
- Load errors are `Array<LoadError>`, not one `Diag`: Go reports several (parse errors in several files, several cycles).
- Skipped: the C-name-hash collision check of LoadPackage (needs xxHash32). selfhost-emit needs xxh32 anyway and can do the check.
- OS errors (missing narra.toml, unreadable file) differ in text between Go and C runtime; the comparison avoids them.
- `quoteGo` in parser.nra is now exported (used for `two [[binary]] entries are named %q`).
