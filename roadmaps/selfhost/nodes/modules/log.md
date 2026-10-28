# Log: modules

- Checked main (f84c1ef): no package support, only std imports. Nothing of the node exists yet.
- Contract conflict: `type Module` already names a std module (builtins contract v1, frozen). The package module type is `SourceModule` here. Change request written.
- Plan: one Checker, per-module name tables swapped in and out (`modScope`), so the rest of the checker is unchanged. The prelude is checked once as module "". Generic instances are resolved in the scope of the module that declares them.
- Baseline C for every testdata file dumped before the change, to diff after renaming.
- Single file = package `<stem>`, module `"<stem>"`. After the change the C is identical modulo the `<stem>__` prefix. Multi-line errors now print the file on every line.
- Sample package `testdata/packages/app` ran on the first try, including generic instances across modules that use a private type of their module.
- Error packages: private import, private namespace call, cycle, unknown module, missing entry. Fast check passes.
- Review on PR #4: `ctor` and `structLit` matched types by name (a crash and a wrong type accepted); C names could collide through `__`; "contains itself" named the wrong file. Fixed with `namesEnum`, a `Module` check, the `__` rule and `typeErrorf`. Also rejected package `std` and duplicate binary names. Six new error packages. The `contains_itself.nra` pin moves from 0:0 to 2:1.
