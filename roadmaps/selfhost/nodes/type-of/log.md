# Log: type-of

- Contextual keyword: parseTypeBody reads `typeof` + name; no lexer change.
- Resolution in resolve.nra (resolveTypeof); declare passes run before consts are checked, so an unannotated same-module const cannot be inferred there (E0982).
- Forward functions resolved from their declaration with resolveSignature; busy keys give E0981.
- Full sh scripts/test.sh passed (5.4k cases).
