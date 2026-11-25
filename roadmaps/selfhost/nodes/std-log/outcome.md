# Outcome: std-log

## Built

`std/log`, written in Narra at `std/src/log.nra`. It writes leveled lines to standard error, as text or as JSON lines.

Files:

- `std/src/log.nra`: the module.
- `selfhost/src/std/stdlog.nra`, one `registerAll` line and one import in `selfhost/src/std/stdlib.nra`: the registry row.
- `scripts/std/generate.nra`: the module entry for the reference.
- Generated: `selfhost/src/std/stdtext.nra`, `docs/reference/std/log.md`, `docs/reference/std/README.md`.
- `testdata/log_levels.nra` and `.out`: a sample that runs itself as a child and prints the child's stderr.

## Public API

```
export enum Level derives Show, Equatable { Debug, Info, Warn, Error }
export enum Format derives Show, Equatable { Text, Json }
export opaque struct Logger
create(level: Level, format: Format): Logger uses alloc
with(logger, fields: Array<#[string, Value]>): Logger uses alloc   // adds fields, returns a new logger
withLevel(logger, level): Logger
enabled(logger, level): bool
debug / info / warn / error(logger, message: string, fields) uses io, alloc
write(logger, level, message, fields) uses io, alloc
render(logger, level, message, fields): string uses alloc          // the line, no line end, no filter
```

`Value` is from `std/shape`. Text lines read `level=info msg="started" port=8080`: a text value is bare unless it is empty or holds a space, control byte, quote or `=`, and then it is quoted as JSON. JSON lines are one object per line, `level` and `msg` first, then the logger's fields, then the call's fields. A later field with the same key as an earlier one, or as `level` or `msg`, replaces its value, and the key keeps its first position. Both formats follow this rule. A text key that is empty or holds a space, `=`, a quote or a control byte is quoted as JSON, like a value. The `msg` value is always quoted in text. JSON values go through `Json.stringify`, so the escaping is the one of `std/json`. No timestamp is written, so output stays reproducible.

## Deviations

Files outside the node's own path:

- `selfhost/src/std/stdlog.nra`, the registry file for the module.
- `selfhost/src/std/stdlib.nra`, one import and one `registerAll` line.
- `scripts/std/generate.nra`, the module entry.
- Nine module-list goldens that name every std module: `docs/reference/std/README.md` and the eight `.err` files `testdata/errors/async_io_internal.err`, `exports_std_ext.err`, `std_internal_user_import.err`, `std_spec_ext_import.err`, `terminal_internal_import.err`, `unicode_internal_import.err`, `unknown_module.err`, and `testdata/stage2/errors/std_internal_import.err`.

No guide chapter edit (the module is listed against chapter 13 in the reference, as for std/set, but the chapter has no section on it).

## Gotchas

- The compiler embeds std from `stdtext.nra`; after a std edit run `sh scripts/test.sh --suite generated --update` and rebuild the compiler before the sample sees the change.
- Returning a logger that shares a parameter needs `from logger`, and `with` copies the field array first.
