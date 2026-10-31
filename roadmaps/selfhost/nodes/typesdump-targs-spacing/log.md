# Log

- Go: targs now joined with "," over per-type dumpTypeString. Narra: Str.join(parts, ",").
- No existing golden had a space inside a type argument, so no existing golden line changed.
- Added bodies/targs_types.nra with a tuple targ (instances identity and pick); its .dump shows `targs=#[int, string],int`.
- Tuple indexing `x[0]` is not valid Narra; the case destructures with `const #[a, _] = x`.
