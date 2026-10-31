# Log: selfhost-check-calls

- Baseline on main 4ee7636: TestSelfhostBodies compared 143, skipped 202. Nothing of this node existed yet (every calls arm reported `unsupported:`).
- Layout decision: calls.nra and ctors.nra need expectType, checkArguments and checkBuiltinCall from exprs.nra, so they sit above exprs and exprs reaches them through two new Walker functions (checkCall, checkConstructorName) that BodyWalker implements. instances.nra sits below (check, resolve, generics only).
- Moved the call dispatch (checkCall, checkNamedCall, checkDottedCall) from exprs.nra to calls.nra. checkArguments now takes the expected type per argument (Option), so generic and trait calls reuse it.
- Instance queue: Checker.instanceList, funcInstances (key "<func>|<typeKeys>"), pending, pendingNext. checkPending in body.nra after the module loop.
- Go dump printed the C name of an implement block function on trait calls (`func=m.Cat_Named_name_<hash>`). Changed both printers to `func=<Type>.<Trait>.<function>`. No existing golden had such a line.
- First full TestSelfhostBodies run after the port: compared 184, skipped 161, no mismatch.
- Added testdata/typesdump/calls (6 compared cases incl. a package, 7 error cases incl. a package) and selfhost_calls_test.go. All call cases match Go byte for byte.
- Left: E0425 (import of a constructor) needs ImportName positions in the Narra AST (parser, not owned).
