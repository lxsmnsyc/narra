# Outcome: std-args

## Built

- `std/args` (alias `Args`), Narra source in `std/src/args.nra` (647 lines over `std/string` and `std/array`), registry row `ARGS_FUNCS` in `selfhost/src/stdargs.nra`, one `registerAll` line in `selfhost/src/stdlib.nra`, regenerated `selfhost/src/stdtext.nra`.
- Types: `Flag { name, short = "", value = "", repeated = false, help = "" }`, `enum Rest { None, Positionals, Raw }`, `Command { name, summary = "", flags = [], positionals = [], rest, commands = [] }`, `Entry { name, value }`, `Parsed { path, flags, positionals, rest }` (fields public), and `enum ArgsError { UnknownFlag(name, closest), MissingValue(flag), UnexpectedValue(flag), Repeated(flag), UnknownCommand(name, closest), MissingCommand(path), MissingPositional(name), ExtraPositional(text), Help(path), BadSpec(reason) }`. All derive `Show` and `Equatable`.
- Functions: `parse(command, arguments): Result<Parsed, ArgsError>`, `has`, `value` (last one wins), `values`, `positional`, `rest`, `commandPath`, `help(command, path): string`, `formatError(error): string`.
- Parse rules: `--name`, `--name=value`, `--name value`, `-x`, `-x value`, `-xvalue`, bundled `-abc`, `--` ends the flags, lone `-` is a positional, `-h`/`--help` on every command (unless the command declares its own `help` or `h`). Flags belong to the command that declares them; a parent's flags are accepted before the subcommand name. Every declared positional is required. A command with subcommands reads its first positional as the subcommand. `Rest.Raw` keeps everything after the last declared positional as written (flags and `--` included). The path starts with the root command's `name`, not `arguments[0]`.
- `BadSpec` (checked on every parse, over every subcommand): an empty flag name, two flags with one long or short name, a short name that is not one character or is `-`, two positionals or two subcommands with one name, positionals beside subcommands, and a subcommand flag with the long name of an enclosing command's flag (review fix: `Parsed.flags` is flat by name, so `tool -v build -V f` with `--verbose` on both levels was a false `Repeated`; the clash is now `BadSpec` rather than per-command entries).
- Hints: Levenshtein distance over bytes of the names without dashes, at most 2, first candidate wins a tie; a name under 3 bytes gets none. Long flags of the current command plus `--help`; subcommand names. Unknown short flags get no hint. `--=v` is `UnknownFlag("--=v", "")`.
- `UnknownFlag` and `MissingValue` name the flag as typed (`-o` or `--output`); `UnexpectedValue` and `Repeated` give the long name without dashes.
- Help: `Usage: <path> [flags] <positional>... [arguments...] <command>`, the summary, `Flags:` in two aligned columns (`-o, --output <path>`, `    --release`, `...` after a repeated flag, the added `-h, --help` last), `Commands:` with summaries. No wrapping.
- Samples: `testdata/args.nra` (21 argument lists over a narrac-like spec plus the readers), `testdata/args_help.nra` (root, `build`, `run`), `testdata/args_errors.nra` (every `ArgsError` case, the `--relase` hint, six `BadSpec` cases, the no-hint cases); error test `testdata/errors/args_flag_missing_name.nra` (`missing field "name" of Flag`, E-code of the existing missing-field check).
- Docs: SPEC section 3 has a `std/args` bullet of its own (removed from the "Not built yet" D118 group); `compiler/README.md` one line; new guide chapter `docs/guide/12-standard-library.md` (tested sample) and its line in `docs/guide/README.md`.
- Goldens with the module list: `testdata/errors/unknown_module.err`, `testdata/declsdump/imports/err_unknown_module.dump` (`std/args` comes last).

## Proof

- base: 4403befe (feat/batch-14), `narrac run testdata/args.nra` with the base release compiler: exit 1, `unknown module "std/args"`.
- branch: the same command prints `testdata/args.out`; the programs suite runs all three samples in debug (AddressSanitizer, leak check, copy check) and release.

## Deviations

- `Command.rest` has no default (D118 says `rest: Rest = Rest.None`). The checker accepts only literals and unqualified constructors as field defaults (`isConstantValue` in `selfhost/src/exprs.nra` has no case for `Rest.None`, E0106), and the module must stay seed-compilable, so a compiler fix would not help before the seed update. Every `Command` literal writes `rest: Rest.None`. Adding the default later breaks no caller. std-terminal's `Style` (`foreground: Color = Color.Default`) hits the same gap.
- `UnknownFlag` and `UnknownCommand` carry a second field, `closest`, because `formatError(error)` takes only the error and the hint needs the candidate names.
- `Parsed` and `Entry` are plain exported structs with public fields, not opaque.
- `--help=x` is `UnexpectedValue("help")`.
- `Args.help(command, path)` with a path that names no subcommand (`["tool", "nope"]`) prints the help of the deepest command it matched (here the root), with no error.
- Guide chapter number 12 created here, as the std-design outcome said the first build would. Other std builds add sections to it.

## Stubs left

- None. narrac still uses its own parser; the move is the proposed node narrac-args (after the batch seed update).

## Gotchas

- A field default of a qualified enum variant (`Rest.None`) is E0106 "a field default must be a constant". Fix in `isConstantValue` (a member expression whose ref is a constructor), then add `rest: Rest = Rest.None` after a seed update.
- A std function that returns texts taken from an `Array<string>` parameter is "the result shares parameter"; `parse` works on `copy(arguments)` so the D118 signature needs no `from`.
- A multi-line result type `Result<\n  A,\n  B,\n>` does not parse in a std source (`expected a name, found ">"`); keep it on one line and let `narrac fmt` decide.
- `narrac0` from the seed has no `std/args` in its embedded std text; to check seed-compilability, copy `args.nra` into a package as a local module (`/private/tmp/claude-501/std-args/seedcheck.sh` did this).

## Proposed learnings

- Field defaults of an enum use only payload-less variants of the prelude until `isConstantValue` takes qualified constructors | E0106 on `Rest.None` | std/src | yes (a sample with a qualified default)

## Verification

- `narrac0` built from `seed/narrac.c` compiles and runs all three samples with `std/args` as a local module; output equals the goldens.
- `NARRAC=<branch release> sh scripts/test.sh --filter args`: PASS, 237 cases. `--suite guide`, `--suite sources`, `--suite generated`: PASS.
- `narrac emit` of the base `selfhost/` tree with the base and branch compilers: byte-identical (223,852 lines).
- Full check `sh scripts/test.sh` after the review fixes, rebased on feat/batch-14 cb834b6e: PASS, 8611 cases in 514s.

## Size

- Estimated M, actual M.
