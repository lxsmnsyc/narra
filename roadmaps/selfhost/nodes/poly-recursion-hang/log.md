# Log: poly-recursion-hang

- Base af35fdcb (feat/batch-12), narrac built with `sh selfhost/bootstrap.sh`. Function case (`grow<T>(v: T)` calling `grow(Grow { value: v })`) and trait case (`implement Show for Grow<T>` showing `Grow { value: self }`): `narrac check` killed at the 20 s cap, no output, both.
- Cause: every call makes an instance with a type one level deeper; the 10000-instance limit (E0310) is reached only after minutes because each instance costs more than the last.
- Fix: all instances go through `addFuncInstance`; it now reports E0310 when the type arguments are deeper than 64 levels (`maxDepth` in `generics.nra`). The queue stops there, so the chain ends. No new code.
- Branch: both cases report E0310 in under a second.
