# Log: generators

- 2026-11-05: Go side first. Design choices:
  - Every generator value is `narra_gen*` (runtime.h), not one C struct per element type. One C type means no per-T list in the checker of either compiler; resume takes `void* out`.
  - In a generator body every C variable is a frame field: `temp`, locals, loop counters, the match done flag (`cVar` in emit_generator.go). Expression-level C locals (block results, envs, out-param temps) stay C locals: no yield can be inside an expression.
  - What is live at each yield is recorded (`liveValues`, the same list unwindAll releases) and drop and clone switch on the state. So fields are not zeroed when their scope ends (brief item 5 differs; same behavior).
  - Generator bodies never go through `function`, so no tail-call detection runs on them (rule 8).
- Go samples pass under ASan and the leak check (generators, generators_heap, generators_values).
