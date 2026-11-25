# Log: std-semver

- Based branch feat/std-semver on feat/batch-24.
- Read SPEC section 14: a dependency version is a minimum written MAJOR.MINOR.PATCH, and the resolver picks the highest minimum with no major rule. So `satisfies` means "at least the requirement".
- Wrote std/src/semver.nra, the registry row selfhost/src/std/stdsemver.nra, a ModuleInfo in scripts/std/generate.nra, testdata/std_semver.{nra,out}.
- Regenerated stdtext.nra and the std reference with `scripts/test.sh --suite generated --update`.
- Iterated on the module in a scratch package because std text is embedded in the compiler (a rebuild takes minutes).
