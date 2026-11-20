# Outcome: manifest-reference

## Built

- `docs/reference/narra-toml.md`, served at `/reference/narra-toml`. It covers `[package]` (`name`, `version`), `[[binary]]` (`name`, `entry`), `[dependencies]` (`git`, `version`, `workspace`), `[overrides]`, `[exports]`, `[link]` (`search`), `[fmt]` (`printWidth`, `tabWidth`, `useTabs`, `trailingComma`), `[workspace]` (`members`) and `narra.lock` (shape and errors). Each key has its type, default, meaning and an example. Each table has an error table by code.
- `docs/reference/README.md` links the page and drops its "Coming" line.

## Deviations

- None.

## Gotchas

- The text output of `narrac` does not print error codes. `narrac check --json` does, in `code`. The page says so.
- `[fmt]` errors have no code (`error narra.toml:4: unknown key indent in [fmt]`). The page says so.
- The loader accepts things no guide chapter or SPEC section mentions:
  - `entry = "app/cli.run"` (package name in front of the module) builds. Only SPEC shows it. The guide writes `"cli.run"`.
  - An unknown key in `[package]` is skipped when its value is a string, an array or an inline table. A bare number is E0409.
  - An unknown table, such as `[nothing]`, is skipped without any check of its lines.
  - Single-quoted strings work, with no escapes.
  - Quoted keys are unquoted in `[exports]` only.
  - `[workspace]` with an unknown key of string type is E0480, and with a number is E0409.
  - `[overrides]` rejects `workspace` (E0483).
- `narrac build` on a package with no `[[binary]]` prints `package app has no [[binary]]; it is a library` with no code.

## Verification

- Compiler built with `sh selfhost/bootstrap.sh <scratch>`. Stage 1 and stage 2 emit identical C.
- `narrac check` and `narrac build` on a scratch package: a manifest with every table builds clean. 30 wrong manifests gave E0406, E0408 to E0416, E0420, E0430 to E0436, E0470 to E0472, E0480, E1000, as listed on the page. A local git dependency built and wrote `narra.lock` as shown. A three-member workspace with `*`, `!` and `workspace = true` built. A bad lock line gave E0437.
- E0443, E0452, E0460 and E0461 were not triggered by hand. Their causes come from `selfhost/src/driver/depcache.nra`, `depresolve.nra` and SPEC section 15.
- `python3 check_prose.py docs/reference/narra-toml.md`: ok.
- `pnpm --dir website build`: passes.

## Size

- Estimated S, actual S.
