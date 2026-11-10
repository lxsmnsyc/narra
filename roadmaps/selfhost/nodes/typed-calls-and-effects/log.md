# Log: typed-calls-and-effects

- Base 81638c3 (feat/batch-4). Built narrac with `sh selfhost/bootstrap.sh`.
- The `Box` program does not check clean: it reports E0800. Array element, map value, nested field and call-result field calls report it too.
- `narrac query type-of` reports `effects: ["alloc"]` on the call. Nothing to fix; wrote outcome as already fulfilled.
