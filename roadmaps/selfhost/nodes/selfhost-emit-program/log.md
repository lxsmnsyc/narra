# Log: selfhost-emit-program

- Facts recorded at the Go sites: PlainFunc in declareFunc, ImplFunc in resolveImpl, InstanceFunc in addFuncInstance, programConsts at CONST_DONE, local number at both LocalInfo sites, composites through `internComposite` at the resolve, substitute, funcvalues, closures, literals, stdcalls and stdlib sites.
- Body lookup: `Checker.bodyKeys` filled by `keyBody` right after each startBody; findBody reads it.
- `programRecorded` now defaults to true. emit.nra calls the match, std and options probes always (they no longer wait for the facts).
- Hash search: "maqoucq" and "mqddgyw" collide (ac37b99a). Short 7-byte names collide quickly; sequential names did not in 4M tries.
- Skip file deleted; the harness skips a missing skip file.
