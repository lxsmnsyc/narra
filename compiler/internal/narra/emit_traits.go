package narra

import (
	"fmt"
	"strings"
)

// Standard trait functions in C. See roadmaps/selfhost/contracts/trait-model.md.
//
// eq:   bool f(T a, T b)
// hash: uint64_t f(T v)
// cmp:  int f(T a, T b), negative, zero or positive
// show: lstr* f(T v), returns an owned string
//
// Arguments are borrowed. Primitives use runtime.h functions. Arrays, structs
// and enums get a generated function the first time the emitter uses one; it
// calls the user's `implement` function when there is one.

// traitFuncName is where every C name of a generated trait function is made.
// The "__" keeps it apart from the C name of a type such as `Point_eq`:
// no output of mangle is "eq", "hash", "cmp" or "show", and type names
// contain no "__".
func traitFuncName(t Type, op string) string { return mangle(t) + "__" + op }

func primTraitFunc(t *PrimType, op string) string {
	if t == TStr {
		return map[string]string{"eq": "lstr_eq", "hash": "lstr_hash", "cmp": "lstr_cmp", "show": "narra_show_string"}[op]
	}
	if op == "show" && (t == TInt || t == TF64) {
		return "lstr_from_" + t.Name
	}
	return "narra_" + op + "_" + t.Name
}

func stdFuncName(op string, t Type) string {
	if p, ok := t.(*PrimType); ok {
		return primTraitFunc(p, op)
	}
	return traitFuncName(t, op)
}

// The contract's name functions. They only name the function; inside the
// emitter use the use* methods, which also make sure it is defined.
func eqFunc(t Type) string   { return stdFuncName("eq", t) }
func hashFunc(t Type) string { return stdFuncName("hash", t) }
func cmpFunc(t Type) string  { return stdFuncName("cmp", t) }
func showFunc(t Type) string { return stdFuncName("show", t) }

type traitNeed struct {
	op   string
	t    Type
	name string
}

type traitEmit struct {
	needs []traitNeed
	seen  map[string]bool
}

func (e *Emitter) useEq(t Type) string   { return e.useTraitFunc("eq", t) }
func (e *Emitter) useHash(t Type) string { return e.useTraitFunc("hash", t) }
func (e *Emitter) useCmp(t Type) string  { return e.useTraitFunc("cmp", t) }
func (e *Emitter) useShow(t Type) string { return e.useTraitFunc("show", t) }

func (e *Emitter) useTraitFunc(op string, t Type) string {
	name := stdFuncName(op, t)
	if _, prim := t.(*PrimType); prim {
		return name
	}
	if e.traits == nil {
		e.traits = &traitEmit{seen: map[string]bool{}}
	}
	if !e.traits.seen[name] {
		e.traits.seen[name] = true
		e.traits.needs = append(e.traits.needs, traitNeed{op, t, name})
	}
	return name
}

// traitCall emits a call to a standard trait function the compiler implements.
func (e *Emitter) traitCall(x *CallExpr) (string, bool) {
	tc := x.Trait
	var args []string
	for _, a := range x.Args {
		code, _ := e.expr(a.Value)
		args = append(args, code)
	}
	switch stdTraitFuncs[tc.Trait.Name+"."+tc.Func] {
	case "eq":
		return fmt.Sprintf("%s(%s, %s)", e.useEq(tc.Self), args[0], args[1]), false
	case "hash":
		return fmt.Sprintf("((int64_t)%s(%s))", e.useHash(tc.Self), args[0]), false
	case "cmp":
		return fmt.Sprintf("((%s){.tag = narra_ordering(%s(%s, %s))})", x.Type().CName(), e.useCmp(tc.Self), args[0], args[1]), false
	case "show":
		return e.ownedTemp(TStr, fmt.Sprintf("%s(%s)", e.useShow(tc.Self), args[0])), true
	}
	panic("emit: unknown trait function " + tc.Trait.Name + "." + tc.Func)
}

// eqCode emits `==` or `!=` on a struct, enum or array.
func (e *Emitter) eqCode(op string, t Type, l, r string) string {
	code := fmt.Sprintf("%s(%s, %s)", e.useEq(t), l, r)
	if op == "!=" {
		return "(!" + code + ")"
	}
	return code
}

// traitFuncs writes the prototypes and definitions of every trait function
// used so far. Definitions can use more, so the list grows while it runs.
func (e *Emitter) traitFuncs(b *strings.Builder) {
	if e.traits == nil {
		return
	}
	var protos []string
	var defs strings.Builder
	for i := 0; i < len(e.traits.needs); i++ {
		protos = append(protos, e.traitDef(&defs, e.traits.needs[i]))
	}
	b.WriteString("\n// Trait functions\n")
	for _, p := range protos {
		b.WriteString(p + ";\n")
	}
	b.WriteString(defs.String())
}

var stdTraitKeys = map[string]string{"eq": "Equatable.equals", "hash": "Hashable.hash", "cmp": "Ordered.compare", "show": "Show.show"}

func (e *Emitter) traitDef(b *strings.Builder, n traitNeed) string {
	ct := n.t.CName()
	var proto string
	switch n.op {
	case "eq":
		proto = fmt.Sprintf("static bool %s(%s a, %s b)", n.name, ct, ct)
	case "hash":
		proto = fmt.Sprintf("static uint64_t %s(%s v)", n.name, ct)
	case "cmp":
		proto = fmt.Sprintf("static int %s(%s a, %s b)", n.name, ct, ct)
	case "show":
		proto = fmt.Sprintf("static lstr* %s(%s v)", n.name, ct)
	}
	fmt.Fprintf(b, "%s {\n", proto)
	if impl := e.prog.Impls[n.t][stdTraitKeys[n.op]]; impl != nil {
		call := funcCName(impl)
		switch n.op {
		case "eq":
			fmt.Fprintf(b, "  return %s(a, b);\n", call)
		case "hash":
			fmt.Fprintf(b, "  return (uint64_t)%s(v);\n", call)
		case "cmp":
			// Ordering's tags are Less = 0, Equal = 1, Greater = 2.
			fmt.Fprintf(b, "  return %s(a, b).tag - 1;\n", call)
		case "show":
			fmt.Fprintf(b, "  return %s(v);\n", call)
		}
		b.WriteString("}\n")
		return proto
	}
	switch n.op {
	case "eq":
		e.eqDef(b, n.t)
	case "hash":
		e.hashDef(b, n.t)
	case "cmp":
		e.cmpDef(b, n.t)
	case "show":
		e.showDef(b, n.t)
	}
	b.WriteString("}\n")
	return proto
}

func variantField(v *Variant, f *Field) string {
	return "u." + variantName(v) + "." + fieldName(f.Name)
}

func (e *Emitter) eqDef(b *strings.Builder, t Type) {
	switch t := t.(type) {
	case *ArrayType:
		fmt.Fprintf(b, "  if (a->len != b->len) return false;\n")
		fmt.Fprintf(b, "  for (int64_t i = 0; i < a->len; i++) {\n    if (!%s(a->items[i], b->items[i])) return false;\n  }\n", e.useEq(t.Elem))
		b.WriteString("  return true;\n")
	case *StructType:
		parts := []string{"true"}
		for _, f := range t.Fields {
			fn := fieldName(f.Name)
			parts = append(parts, fmt.Sprintf("%s(a.%s, b.%s)", e.useEq(f.T), fn, fn))
		}
		fmt.Fprintf(b, "  return %s;\n", strings.Join(parts, " && "))
	case *EnumType:
		b.WriteString("  if (a.tag != b.tag) return false;\n  switch (a.tag) {\n")
		for _, v := range t.Variants {
			if len(v.Fields) == 0 {
				continue
			}
			var parts []string
			for _, f := range v.Fields {
				fn := variantField(v, f)
				parts = append(parts, fmt.Sprintf("%s(a.%s, b.%s)", e.useEq(f.T), fn, fn))
			}
			fmt.Fprintf(b, "  case %d: return %s;\n", v.Index, strings.Join(parts, " && "))
		}
		b.WriteString("  default: return true;\n  }\n")
	}
}

func (e *Emitter) hashDef(b *strings.Builder, t Type) {
	b.WriteString("  uint64_t h = UINT64_C(14695981039346656037);\n")
	switch t := t.(type) {
	case *ArrayType:
		fmt.Fprintf(b, "  for (int64_t i = 0; i < v->len; i++) h = narra_hash_mix(h, %s(v->items[i]));\n", e.useHash(t.Elem))
	case *StructType:
		for _, f := range t.Fields {
			fmt.Fprintf(b, "  h = narra_hash_mix(h, %s(v.%s));\n", e.useHash(f.T), fieldName(f.Name))
		}
	case *EnumType:
		b.WriteString("  h = narra_hash_mix(h, (uint64_t)v.tag);\n  switch (v.tag) {\n")
		for _, vr := range t.Variants {
			if len(vr.Fields) == 0 {
				continue
			}
			fmt.Fprintf(b, "  case %d:", vr.Index)
			for _, f := range vr.Fields {
				fmt.Fprintf(b, " h = narra_hash_mix(h, %s(v.%s));", e.useHash(f.T), variantField(vr, f))
			}
			b.WriteString(" break;\n")
		}
		b.WriteString("  default: break;\n  }\n")
	}
	b.WriteString("  return h;\n")
}

func (e *Emitter) cmpDef(b *strings.Builder, t Type) {
	b.WriteString("  int c = 0;\n")
	switch t := t.(type) {
	case *ArrayType:
		fmt.Fprintf(b, "  for (int64_t i = 0; i < a->len && i < b->len; i++) {\n    if ((c = %s(a->items[i], b->items[i])) != 0) return c;\n  }\n", e.useCmp(t.Elem))
		b.WriteString("  return a->len < b->len ? -1 : (a->len > b->len ? 1 : 0);\n")
		return
	case *StructType:
		for _, f := range t.Fields {
			fn := fieldName(f.Name)
			fmt.Fprintf(b, "  if ((c = %s(a.%s, b.%s)) != 0) return c;\n", e.useCmp(f.T), fn, fn)
		}
	case *EnumType:
		b.WriteString("  if (a.tag != b.tag) return a.tag < b.tag ? -1 : 1;\n  switch (a.tag) {\n")
		for _, v := range t.Variants {
			if len(v.Fields) == 0 {
				continue
			}
			fmt.Fprintf(b, "  case %d:", v.Index)
			for _, f := range v.Fields {
				fn := variantField(v, f)
				fmt.Fprintf(b, " if ((c = %s(a.%s, b.%s)) != 0) return c;", e.useCmp(f.T), fn, fn)
			}
			b.WriteString(" break;\n")
		}
		b.WriteString("  default: break;\n  }\n")
	}
	b.WriteString("  return c;\n")
}

// showDef writes text such as `Point { x: 1, y: 2 }`, `Some(1)` and `[1, 2]`.
// Strings and chars inside are quoted.
func (e *Emitter) showDef(b *strings.Builder, t Type) {
	app := func(indent, c string) { fmt.Fprintf(b, "%slstr_append_c(&s, %s);\n", indent, cString(c)) }
	switch t := t.(type) {
	case *ArrayType:
		b.WriteString("  lstr* s = lstr_from_cstr(\"[\");\n")
		b.WriteString("  for (int64_t i = 0; i < v->len; i++) {\n")
		b.WriteString("    if (i > 0) lstr_append_c(&s, \", \");\n")
		fmt.Fprintf(b, "    lstr_append_owned(&s, %s(v->items[i]));\n  }\n", e.useShow(t.Elem))
		app("  ", "]")
	case *StructType:
		if t.Tuple {
			// Tuple(1, "a"), as the literal is written.
			b.WriteString("  lstr* s = lstr_from_cstr(\"Tuple(\");\n")
			for i, f := range t.Fields {
				if i > 0 {
					app("  ", ", ")
				}
				fmt.Fprintf(b, "  lstr_append_owned(&s, %s(v.%s));\n", e.useShow(f.T), fieldName(f.Name))
			}
			app("  ", ")")
			break
		}
		if len(t.Fields) == 0 {
			fmt.Fprintf(b, "  lstr* s = lstr_from_cstr(%s);\n", cString(t.Name+" {}"))
			break
		}
		fmt.Fprintf(b, "  lstr* s = lstr_from_cstr(%s);\n", cString(t.Name+" { "))
		for i, f := range t.Fields {
			sep := ", "
			if i == 0 {
				sep = ""
			}
			app("  ", sep+f.Name+": ")
			fmt.Fprintf(b, "  lstr_append_owned(&s, %s(v.%s));\n", e.useShow(f.T), fieldName(f.Name))
		}
		app("  ", " }")
	case *EnumType:
		b.WriteString("  lstr* s = NULL;\n  switch (v.tag) {\n")
		for _, vr := range t.Variants {
			fmt.Fprintf(b, "  case %d:\n", vr.Index)
			if len(vr.Fields) == 0 {
				fmt.Fprintf(b, "    s = lstr_from_cstr(%s);\n    break;\n", cString(vr.Name))
				continue
			}
			fmt.Fprintf(b, "    s = lstr_from_cstr(%s);\n", cString(vr.Name+"("))
			for i, f := range vr.Fields {
				if i > 0 {
					app("    ", ", ")
				}
				fmt.Fprintf(b, "    lstr_append_owned(&s, %s(v.%s));\n", e.useShow(f.T), variantField(vr, f))
			}
			app("    ", ")")
			b.WriteString("    break;\n")
		}
		b.WriteString("  default: s = lstr_from_cstr(\"?\"); break;\n  }\n")
	}
	b.WriteString("  return s;\n")
}
