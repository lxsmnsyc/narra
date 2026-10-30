# Log: hashed-cnames

- Checked first: nothing of the work existed (names used `__` mangling).
- Decision: the registry of C names is a process-wide table keyed by the name, with the key holding the module path relative to the package. A per-compilation table would need a context passed into `CName()`. Same relative path in two programs gives the same name and key, so no false collisions in tests.
- Decision: helper functions put the operation first (`retain_<type>`) because a suffix (`<type>_retain`) can equal a user type named `Point_retain`. The cnames_modules sample uses such a type.
- Decision: module hash input is the path without the package name, so a package renamed or built from another directory keeps its names.
- Dead end: first run failed with a collision panic because the key held the full module path; fixed by using the relative path.
- Removed tests outside owned paths: errors/traits_func_underscore, errors/traits_name_underscore, packages/errors/dunder_name, packages/errors/dunder_module. Replaced by packages/cnames_modules.
- Doubt: names are computed on each `CName()` call (no cache). Cost is small now; a selfhost port may want a per-type cache.
