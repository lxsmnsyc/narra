# narra.toml

`narra.toml` marks the top of a package and holds its settings. This page lists every table and key, its type, its default, and the error code for a wrong value. The last part covers `narra.lock`.

`narrac` reads a small part of TOML: tables (`[name]`), arrays of tables (`[[name]]`), `key = value` lines, comments that start with `#`, and these values:

- A string in double quotes (`"text"`) or single quotes (`'text'`). A double-quoted string takes the escapes `\n`, `\t`, `\r`, `\"`, `\\` and `\uXXXX`.
- An array of strings, which may continue over several lines.
- An inline table, in `[dependencies]` and `[overrides]` only. A trailing comma is an error.

A line outside these forms is an error: E0405 for `[[name]` without the closing `]]`, E0406 for `[name` without `]`, E0408 for a line in a known table with no `=`. A table the loader does not know is skipped. A key the loader does not know in `[package]` is skipped when its value can be read as one of the values above. `[link]` and `[workspace]` reject unknown keys.

An error in `narra.toml` names the file and line, as in `narra.toml:3:1: version "1.0" is not MAJOR.MINOR.PATCH`. The text output does not show the code. `narrac check --json` prints it in the `code` field, and `narrac explain CODE` prints the message.

## Tables

| Table | Needed | Holds |
| --- | --- | --- |
| [`[package]`](#package) | yes, except in a workspace root | The name and version. |
| [`[[binary]]`](#binary) | no | One program to build. |
| [`[dependencies]`](#dependencies) | no | Packages to download. |
| [`[overrides]`](#overrides) | no | Replacement repositories, in the root only. |
| [`[exports]`](#exports) | no | The modules other packages may import. |
| [`[link]`](#link) | no | Where the C compiler looks for libraries. |
| [`[fmt]`](#fmt) | no | Options for `narrac fmt`. |
| [`[workspace]`](#workspace) | no | The member packages of a workspace. |

## package

```toml
[package]
name = "app"
version = "0.1.0"
```

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `name` | string | none | The package name. It is the first part of every import path of the package. Required. |
| `version` | string | `""` | `MAJOR.MINOR.PATCH`, three numbers in the range of `int`. Needed when another package depends on this one. |

Errors:

| Code | Cause |
| --- | --- |
| E0410 | `name` is not a string, or not letters, digits and `_` that start with a letter. |
| E0411 | `name` is `std`, which names the standard library. |
| E0412 | `version` is not a string. |
| E0414 | There is no `name`. A workspace root with no `[package]` table needs none. |
| E0431 | `version` is not `MAJOR.MINOR.PATCH`, for example `"1.0"`. |
| E0409 | A value is not a string, an array or an inline table (a bare number, for example), or a string is not closed. This holds for any key of a known table. |

## binary

```toml
[[binary]]
name = "app"
entry = "cli.run"
```

Each `[[binary]]` table makes one program. A package with no `[[binary]]` is a library: `narrac build` stops with `package app has no [[binary]]; it is a library`. Several tables make several programs.

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `name` | string | none | The name of the executable. Required. |
| `entry` | string | none | The module and the function that starts the program, as `"module.function"`. The module path is relative to `src/`. The form `"app/cli.run"`, with the package name in front, also works. Required. |

The function is `export function run(args: Array<string>): int`, or the same with `async`. Its result is the exit code.

Errors:

| Code | Cause |
| --- | --- |
| E0412 | `name` or `entry` is not a string. |
| E0415 | A table has no `name` or no `entry`. |
| E0416 | Two tables have the same `name`. |
| E0420 | The module has no function with the name in `entry`. `narrac build` reports this. |

In a workspace, two members that each have a binary of one name are E0488 when you build the root or run that name.

## dependencies

```toml
[dependencies]
json = { git = "https://example.com/narra-json", version = "1.2.0" }
core = { workspace = true }
```

A dependency is a git repository tagged with its versions. The key is the package name and the first part of its import paths.

| Key of the inline table | Type | Meaning |
| --- | --- | --- |
| `git` | string | A URL or a local path of the repository. |
| `version` | string | The lowest version the package accepts, as `MAJOR.MINOR.PATCH`. The tag is `v` followed by the version. |
| `workspace` | boolean | `true` names a member of the same workspace. It takes no other key, and nothing is downloaded. |

`git` and `version` are both required. `narrac build` resolves the dependencies, writes `narra.lock` and loads the sources of each one. A package may import only itself, `std` and the packages in its own `[dependencies]`.

Errors:

| Code | Cause |
| --- | --- |
| E0430 | `git` or `version` is missing. |
| E0431 | `version` is not `MAJOR.MINOR.PATCH`. |
| E0432 | The table has a key other than `git` and `version`, or repeats one. |
| E0433 | The name is `std` or the name of this package. |
| E0434 | Two dependencies have one name. |
| E0435 | The value is not an inline table of strings, or it has a trailing comma. |
| E0436 | The name is not letters, digits and `_` starting with a letter. |
| E0483 | `workspace` has a value other than `true`, or another key stands beside it. |
| E0484 | `workspace = true` names a package that is not a member, or the package is not in a workspace. |
| E0450 | A git dependency has the name of a workspace member, or two packages require one name from two repositories. |
| E0463 | A package imports a package it does not list in `[dependencies]`. |
| E0462 | A dependency has no `.nra` file in `src/`. |

The errors E0440 to E0458 come from fetching and resolving. Four examples:

- E0440: git is not on `PATH`.
- E0442: the repository has no tag for the version.
- E0444: there is no cache folder.
- E0453: the manifest of a checkout names another package.

The cache folder is `$NARRA_CACHE`, or `$HOME/.cache/narra` when the variable is not set.

## overrides

```toml
[overrides]
json = { git = "https://example.com/my-json-fork", version = "1.2.0" }
```

An override has the same form as a dependency, without `workspace`. It replaces the repository and the version of that package wherever any package in the build asks for it. Only the root `narra.toml` of a build may have this table.

Errors:

| Code | Cause |
| --- | --- |
| E0430, E0431, E0432, E0433, E0434, E0435, E0436 | The same as for a dependency. |
| E0483 | An entry has `workspace`. |
| E0454 | A manifest other than the root has `[overrides]`. |
| E0455 | An entry names a package that nothing requires. |
| E0456 | An entry names a version with no tag in its repository. |

## exports

```toml
[exports]
"json" = "src/index.nra"
"json/stream" = "src/streaming/reader.nra"
```

Without this table, every file under `src/` is open to importers, by its path. With it, another package may import only the keys. The files of the package itself still import each other by their real paths.

| Part | Type | Meaning |
| --- | --- | --- |
| key | string | The import path other packages write. It is the package name, or starts with the package name and `/`. |
| value | string | The module file. It starts with `src/`, ends with `.nra` and must exist. |

Errors:

| Code | Cause |
| --- | --- |
| E0470 | A key does not start with the package name. |
| E0471 | A value is not a string, is not a path such as `src/index.nra`, or names a file the package does not have. |
| E0472 | Two entries have one key. |
| E0473 | Another package imports a path that is not a key. The message lists the keys. |
| E0474 | A package imports a module under `std/internal/` or `std/ext/`. |

## link

```toml
[link]
search = ["/opt/homebrew/lib"]
```

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `search` | array of strings | `[]` | Folders the C compiler searches for libraries (`-L`). A relative path starts at the package folder. |

The libraries themselves come from the `from` of the `extern` blocks in the code, not from this table.

Errors:

| Code | Cause |
| --- | --- |
| E0413 | `search` is not an array of strings. |
| E1000 | The table has a key other than `search`. |

## fmt

```toml
[fmt]
printWidth = 80
tabWidth = 4
useTabs = false
trailingComma = "none"
```

Only `narrac fmt` reads this table. A file takes the `[fmt]` of the nearest `narra.toml` above it. A member of a workspace with no `[fmt]` takes the one of the workspace root. A `narra.toml` with neither `[package]` nor `[fmt]` is skipped. Without any table, the defaults apply.

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `printWidth` | integer, 1 or more | `100` | The line length the formatter stays within. |
| `tabWidth` | integer, 1 or more | `2` | The columns of one level of indentation. |
| `useTabs` | boolean | `false` | Indent with tabs instead of spaces. |
| `trailingComma` | `"all"` or `"none"` | `"all"` | Whether a split list ends with a comma. |

An unknown key or a wrong value stops `narrac fmt` with a message such as `error narra.toml:4: unknown key indent in [fmt]`. This error has no code.

## workspace

```toml
[workspace]
members = ["packages/*", "tools/lint", "!packages/scratch"]
```

A `narra.toml` with this table makes its folder a workspace root. Each member is a package with its own `narra.toml`. All members share one `narra.lock` at the root.

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `members` | array of strings | none | Member folders, relative to the root. `*` stands for one whole folder name. An entry that starts with `!` removes the folders it matches. Required. |

A root that also has a `[package]` table is a member of its own workspace. A package whose folder no entry lists is standalone.

Errors:

| Code | Cause |
| --- | --- |
| E0480 | `members` is missing, is not an array of strings, or an entry is not a relative path with `*` only as a whole segment and no `..`. Another key in the table is also E0480. |
| E0481 | A member has no `narra.toml`, or has a `[workspace]` of its own. |
| E0482 | Two members have one name. |
| E0485 | A `narra.lock` is inside a member. |
| E0488 | Two members have a binary of one name. |
| E0489 | A flag such as `-dependents` names a path that is not a member folder. |
| E0490 | Members depend on each other in a cycle through `workspace = true`. |

## narra.lock

`narrac build` writes `narra.lock` next to `narra.toml` when the package has dependencies. In a workspace it sits at the root. Do not edit it, and commit it. A build that changes nothing leaves the file as it was.

```toml
# narra.lock: written by narrac. Do not edit.

[[package]]
name = "json"
git = "https://example.com/narra-json"
version = "1.2.0"
tag = "v1.2.0"
commit = "e1a35bc9764c4a9942dfdbeed1dd37fa72867774"
content = "ce02a3c70c6646dffa850079de004723b3e03a6c"
```

Each `[[package]]` table has these keys, in this order, and the tables are sorted by name:

| Key | Meaning |
| --- | --- |
| `name` | The package name. |
| `git` | The repository. It is not empty. |
| `version` | The selected version, `MAJOR.MINOR.PATCH`. |
| `tag` | The tag the version came from: `v<version>`, or `<name>/v<version>` in a repository that tags each package. |
| `commit` | The full commit hash of the tag, 40 or 64 lowercase hex digits. |
| `content` | The git tree hash of that commit, in the same form. |

Errors:

| Code | Cause |
| --- | --- |
| E0437 | A line is not `[[package]]` or `key = "value"`, or a value is not well formed. |
| E0438 | An entry has an unknown, missing or repeated key, or a `tag` that does not belong to its `version`. |
| E0439 | Two entries have one name. |
| E0443 | The cached copy has no recorded hash, or another tree hash than the lock records. |
| E0452 | The repository has other hashes than the lock records, for example after someone moved a tag. |
| E0460 | The lock names a dependency that is not in the cache. `narrac lsp` reports this. |
| E0461 | The lock has no entry for a dependency. `narrac lsp` reports this. |

When the lock and the manifests name different repositories for a package, the build prints `notice: <name> was locked from <old URL> and now resolves from <new URL>` and rewrites the lock. This is not an error.
