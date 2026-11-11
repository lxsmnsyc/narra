# Premortem: batch-8

It is a week from now and batch-8 is late or reverted. Why?

1. **A proven index was not in range.** FixedArray indexing has no runtime check, so a wrong proof (a loop variable assigned in the body, an `N` of 0, a const instance whose value the prover did not see, a sized index type that wraps) is out-of-bounds memory access, not a panic. Change: fixed-arrays' samples cover each proof form and its failure; the fresh review probes them adversarially, as the contracts reviews did.
2. **Const generics broke instance naming.** Two instances that differ only in `N` got one C name, or an instance key missed the value, so one body served both lengths. Change: const-generics' acceptance needs two instances with different values in one program; the tree-shake reachable set must key instances the same way.
3. **Raw pointers leaked into reference counting.** A `Ptr<T>` over a heap type, or a `CString` dropped without freeing, bypassed the leak check or freed a live value. Change: unsafe-ptr states which `T` are allowed in SPEC; every unsafe sample runs under the leak check and sanitizers.
4. **Linking differed between macOS and Linux.** `-lm`, symbol names and `from "c"` behave differently between Apple clang and gcc. Change: samples use only libc and libm; the Linux check runs at the end of the batch, as in batch-7.
5. **The two chains conflicted in the parser and checker.** Both add attributes, types and diagnostics. Change: logic in modules of its own, shared files only for dispatch; the second branch to land rebases and reruns the full check (batch-7 lesson: check the rebase's exit code first).

Accepted risk: `@layout` and C-callable functions stay out (step 2, with c-abi-export), so `extern` structs passed by value are not possible in this batch.
