# Outcome: traits

## Built

All in `compiler/internal/narra/`.

Syntax (`parser.go`, `ast.go`, `lexer.go`):
- New keywords `trait`, `implement`, `implements`, `derives`.
- `TraitDecl` (`trait Name { function f(a: Self): R }`, signatures only) and `ImplDecl` (`implement Trait for Type { functions }`).
- `StructDecl` and `EnumDecl` gain `Implements` and `Derives` (lists of `*TypeExpr` naming traits).
- `FuncDecl.Bounds` holds constraints, one list per type parameter: `<T: Show & Ordered>`. Constraints on struct and enum type parameters are a parse error.
- `CallExpr.Trait *TraitCall` is set for trait calls the compiler implements.
- `TypeExpr.Func` and `TypeExpr.Result` parse function types `(x: int) => int`, which resolve to `FuncType`. No value of one can be made yet (closures).
- `funcDecl` is split into `funcHeader` plus the body.

Trait model (`check_traits.go`):
- `Trait{Name, Funcs, Self, Decl, Std}`, `TraitFunc{Name, Params, Result, Decl}` as in the contract, plus extra fields. `SelfType = TypeParamType`, one per trait, so `unify` and `substitute` handle Self.
- The standard traits and `enum Ordering derives Show, Ordered { Less Equal Greater }` are Narra source in the prelude (`preludeSrc` in `check.go`). `Option` and `Result` are declared `derives Show, Ordered`.
- `c.implements(t, tr)`: a type parameter implements the traits in its constraint. Primitives implement all standard traits except `void`; `f64` is not `Hashable`. Arrays implement a standard trait when the element does. A struct or enum implements a trait through an `implement` block, or through `derives`, or automatically for `Equatable` and `Hashable`; the last two require every field to implement it. Recursion through arrays (`struct Tree { kids: Array<Tree> }`) counts as implemented.
- Modules: traits are visible per module (`traitState.traits[module][name]`, "" for the prelude, looked up by `c.lookupTrait`). `export trait` adds the trait to `SourceModule.Exports`, and `import { Area } from "app/shape"` imports it (`importTrait`, called from `importPackageModule`). `Alias.Area.f(x)` and `<T: Alias.Area>` are not supported. Per-type trait facts are keyed by module and name (`typeKey`). `c.ownsType` allows an `implement` block only in the module that declares the type, as SPEC section 4 and D5 say.
- `c.implFunc(t, tr, name)` returns the user `FuncDecl` or nil. `Program.Impls` holds the same map (type, `"Trait.function"`) for the emitter.
- Checks: every trait in `implements` has a block or a `derives` entry; a block needs its trait listed in `implements`; a block's signatures match the trait with Self replaced; no blocks for primitives, arrays, function types or prelude types ("declared in another module"); no blocks for generic types (use `derives`); `derives` accepts only the four standard traits, and for a non-generic type every field must implement the trait.
- Pass order in `declareModule`, per module: `declareTraits` after type names, `funcBounds` with each generic function's signature, `resolveTraits` after `buildGeneric`. `implBodies` runs in `checkBodies`.

How constraints are checked:
- Generic bodies are checked once with parameters open. `Trait.f(x)` with `x: T` needs `T` to carry the trait in its constraint, and `==` on a type containing `T` needs `Equatable`.
- `genericCall` calls `c.checkBounds` on the inferred type arguments at each call site, including calls from other generic bodies with open arguments.
- Instances are compiled per type, as before, so in an instance each trait call resolves to a direct call: `CallExpr.Func` = the user implementation, or `CallExpr.Trait` for a compiler-provided one. There is no hidden table parameter.

How equality is generated (`emit_traits.go`):
- `==` and `!=` on non-primitive types call `<mangle(t)>__eq(a, b)`. Primitives keep the old C.
- `e.useEq/useHash/useCmp/useShow(t)` return the C name and queue its definition. `e.traitFuncs`, called from `Emit` after the string literals and before function bodies, writes all prototypes then all definitions. Definitions can queue more.
- Structs compare field by field; enums compare the tag, then the variant's fields; arrays compare length, then elements. Hash mixes fields with `narra_hash_mix` from an FNV seed. Compare is lexicographic (enums by variant index first). Show writes `Point { x: 1, y: 2 }`, `Some(1)`, `[1, 2]`, with quoted strings and chars.
- With a user implementation, the generated function is a wrapper: `cmp` returns `f_impl(a, b).tag - 1`.
- `Ordered.compare` through the compiler gives `(E_Ordering){.tag = narra_ordering(cmp(a, b))}`.

Where C names are made:
- `traitFuncName(t, op)` in `emit_traits.go`: every generated function, `mangle(t) + "__" + op`, so `S_tp_geo__Circle__eq`. `mangle` uses `CName()`, so the module is in the name. Prelude types keep no prefix: `E_Ordering__show`.
- `primTraitFunc` in `emit_traits.go`: the runtime names for primitives from the contract table.
- `implFuncName(tr, fn, t)` in `check_traits.go`: the `FuncDecl.Name` of a function in an `implement` block, `<Type>__0<cModule(trait module)><Trait>__<fn>`. Its `Module` is the type's module, so `funcCName` gives `f_tp_geo__Square__0tp_shape__Area__area` or `f_tp_geo__Square__0Show__show`. The `0` segment keeps it apart from generic instances (no identifier or mangle output starts with a digit). Trait names and trait function names must pass `cSafeSegment` (no `__`, no leading or trailing `_`), which makes the names injective; see the change request.

Runtime (`runtime.h`): `narra_eq_*`, `narra_hash_*`, `lstr_hash`, `narra_cmp_*`, `narra_show_bool/char/string`, plus `narra_hash_mix`, `narra_ordering`, `lstr_append_c`, `lstr_append_owned`.

## Deviations

- Contract change request `contracts/trait-model.change-request.md`: adds `e.useEq` and friends, because the v1 package functions cannot make the emitter define the generated functions. The v1 package functions exist and return names only.
- Primitive trait helpers went into `runtime.h`, as the contract says, though the package CLAUDE.md says module runtime C goes in `Module.Runtime`. These are not a module.
- `lexer.go` (not in the brief's lists) gained one line of keywords.
- Modules files (not in the brief's lists), a few lines each: `modules.go` (`TraitDecl` in `newSourceModule` exports and in `declName`, so `__` checks and privacy errors cover traits), `modules_check.go` (`traits: newTraitState()` in `CheckPackage`, a `*TraitDecl` case in `importPackageModule`).
- `runtime.h` and `compiler/README.md` edited (not in the brief's lists).
- Function types compare structurally: `sameType`, `unify`, `substitute`, `containsTypeParam` and `containsOpen` have a `FuncType` case, and `mangle` gives `F<n>__<params>__<result>`, so `Box<(int) => int>` is `S_<mod>__Box__F1__int__int`. `testdata/traits_functypes.nra` checks both; it would fail to compile in C if two signatures shared a name.
- `check_stmt.go` (not in the brief's lists): `this expression has no value` is no longer reported when checking the value already reported an error. This removes the cascade after the constructor ambiguity error and after other failed expressions.
- Function types are parsed and resolved here: `parser.go` `funcTypeExpr`, `TypeExpr.Func` and `TypeExpr.Result`, and the `te.Func` case in `resolveType` that builds a `FuncType`. The orchestrator accepted keeping this. The closures node must build on it rather than add a second parser for function types. It still needs function values, arrow functions and calls through values.
- Every program's C now has the `E_Ordering` type and the new runtime helpers.

## Stubs left

- `Dynamic<Trait>`, `derives Shape`, `Hash.of`: not built.
- `implement` blocks for generic types, constraints on struct and enum type parameters, generic trait functions: errors saying "not supported by this compiler yet".

## Gotchas

- A call that fails a constraint is reported twice: once at the call site by `checkBounds`, and once from inside the instance body, which is still checked.
- A user `implement Equatable` without `implement Hashable` keeps the field-based automatic hash, so values the user calls equal can hash differently. maps must handle this, for example by requiring both or by rejecting such keys.
- `narra_cmp_f64` returns 0 for NaN, while `narra_eq_f64` returns false (IEEE). So `Ordered.compare(nan, x)` is `Equal` but `nan == x` is false.
- `Show.show` on strings and chars quotes them and escapes the quote, `\`, newline, tab and carriage return.

- Constructor lookup is deterministic (`ctor`, `ctorOwners` and `visibleEnums` in `check_call.go`). The expected type decides first. Otherwise the enums visible in the current module fall into three ranks: (1) those the module declares, (2) those it imports by name, (3) the prelude. The first rank with a match wins, so a user or imported `Less` hides `Ordering.Less`. Two enums in the winning rank give "constructor X is ambiguous: A, B all have it; write A.X", following SPEC section 3. Order within a rank is declaration order, then import order, then by name, never map order. A qualified `Enum.X` checks that one enum only.
- `implements` and `derives` are keywords now, and so are `trait` and `implement`.
- `Show.show` on a string returns it in quotes.
- The emitter generates trait functions only for types it was asked about. Use `e.useEq(t)` and friends, not `eqFunc(t)`, when emitting a call.
- The orchestrator suggested allowing an `implement` block in the trait's module too. I kept SPEC D5 (the type's module only). The other case cannot happen anyway: the type must list the trait in `implements`, so the type's module imports the trait's module. Implementing in the trait's module would need the reverse import, which is a cycle.

## Proposed learnings

- Emit calls to standard trait functions through `e.useEq/useHash/useCmp/useShow`, never by building `<type>_eq` by hand | the definition is only emitted when requested | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- `testdata/traits.nra` (a user trait with two implementations, constrained generics `<T: Area>`, `<T: Ordered & Show>`, `<T: Equatable>`) and `testdata/traits_derive.nra` (derived Show and Ordered on structs, enums and a generic struct, `==` on nested structs, enums, options and arrays, automatic hash) run clean under AddressSanitizer, UBSan and the leak check, and in release mode.
- Package `testdata/packages/traits` (`run.out`): trait `Area` in `tp/shape`, implemented in `tp/geo`, used with `<T: Area>`, `<T: Ordered & Show>` and `<T: Equatable>` from `tp/report`, called from `tp/main`. `tp/main` declares its own `Square` with the same traits as `tp/geo`'s, and the C names differ (`f_tp_geo__Square__0Show__show` and `f_tp_main__Square__0Show__show`). An imported `Mood.Less` outranks `Ordering.Less`. It passes in debug (ASan, UBSan, leak check), release and test mode.
- Error packages: `testdata/packages/errors/trait_orphan_impl` (implement for a type imported from another module) and `ctor_imported_ambiguous` (two imported enums share a constructor).
- `testdata/traits_ctor_shadow.nra`: a user enum `Rank` with `Less`, `Equal` and `Greater`, used with no expected type. Emitting it and running it 20 times each gave identical output, and so did 20 runs of the ambiguity error test.
- Review fixes: `testdata/traits_functypes.nra` (function types and their C names), escapes in `testdata/traits.nra`, error tests `testdata/errors/traits_name_underscore.nra` (the review's collision repro), `traits_func_underscore.nra` and `traits_dup_bound.nra`.
- Error tests: `testdata/errors/traits_ctor_ambiguous.nra`, `testdata/errors/traits_missing_impl.nra`, `traits_missing_bound.nra`, `traits_unconstrained.nra`, `traits_foreign_impl.nra`, `traits_foreign_option.nra`, `traits_eq_function.nra`, `traits_not_listed.nra`, `traits_derive_field.nra`, `traits_signature.nra`.

## Size

- Estimated L, actual L.
