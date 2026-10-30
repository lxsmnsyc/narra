package narra

import (
	"fmt"
	"strings"
)

func arrName(a *ArrayType) string { return strings.TrimSuffix(a.CName(), "*") }

func (e *Emitter) typeDef(b *strings.Builder, t Type) {
	switch t := t.(type) {
	case *StructType:
		fmt.Fprintf(b, "struct %s {", t.CName())
		if len(t.Fields) == 0 {
			b.WriteString(" char unused;")
		}
		for _, f := range t.Fields {
			fmt.Fprintf(b, " %s %s;", f.T.CName(), fieldName(f.Name))
		}
		b.WriteString(" };\n")
	case *EnumType:
		if isBoxedEnum(t) {
			fmt.Fprintf(b, "struct %s { int32_t tag; %s* u; };\n", t.CName(), payloadCName(t))
		} else {
			fmt.Fprintf(b, "struct %s { int32_t tag; };\n", t.CName())
		}
	}
}

func fieldName(n string) string { return "f_" + n }

func variantName(v *Variant) string { return "v_" + v.Name }

// retain and release return C statements for one value of type t.
func retain(t Type, v string) string {
	switch t := t.(type) {
	case *PrimType:
		if t == TStr {
			return fmt.Sprintf("lstr_retain(%s);", v)
		}
		if t == TStringBuilder {
			return fmt.Sprintf("narra_sb_retain(%s);", v)
		}
	case *ArrayType:
		return fmt.Sprintf("%s_retain(%s);", arrName(t), v)
	case *BoxType:
		return fmt.Sprintf("%s_retain(%s);", boxName(t), v)
	case *MapType:
		return fmt.Sprintf("%s_retain(%s);", mapName(t), v)
	case *FuncType:
		return fmt.Sprintf("narra_env_retain((%s).env);", v)
	case *StructType, *EnumType:
		if needsRC(t) {
			return fmt.Sprintf("%s(%s);", opFuncName("retain", t), v)
		}
	}
	return ""
}

func release(t Type, v string) string {
	switch t := t.(type) {
	case *PrimType:
		if t == TStr {
			return fmt.Sprintf("lstr_release(%s);", v)
		}
		if t == TStringBuilder {
			return fmt.Sprintf("narra_sb_release(%s);", v)
		}
	case *ArrayType:
		return fmt.Sprintf("%s_release(%s);", arrName(t), v)
	case *BoxType:
		return fmt.Sprintf("%s_release(%s);", boxName(t), v)
	case *MapType:
		return fmt.Sprintf("%s_release(%s);", mapName(t), v)
	case *FuncType:
		return fmt.Sprintf("narra_env_release((%s).env);", v)
	case *StructType, *EnumType:
		if needsRC(t) {
			return fmt.Sprintf("%s(%s);", opFuncName("release", t), v)
		}
	}
	return ""
}

func (e *Emitter) rcFuncs(b *strings.Builder, t Type) {
	if !needsRC(t) {
		return
	}
	if et, ok := t.(*EnumType); ok {
		// The drop function runs when the count reaches zero. Releasing a
		// child there queues it (narra_drop) instead of recursing, so the
		// depth of a structure does not use the C stack.
		fmt.Fprintf(b, "static void %s(void* p) {\n  %s* q = p;\n  switch (q->tag) {\n", opFuncName("drop", t), payloadCName(et))
		for _, v := range et.Variants {
			var stmts []string
			for _, f := range v.Fields {
				if s := release(f.T, "q->"+variantName(v)+"."+fieldName(f.Name)); s != "" {
					stmts = append(stmts, s)
				}
			}
			if len(stmts) > 0 {
				fmt.Fprintf(b, "  case %d: %s break;\n", v.Index, strings.Join(stmts, " "))
			}
		}
		b.WriteString("  default: break;\n  }\n  narra_free(q);\n}\n")
	}
	for _, op := range []string{"retain", "release"} {
		fmt.Fprintf(b, "static void %s(%s v) {\n", opFuncName(op, t), t.CName())
		gen := retain
		if op == "release" {
			gen = release
		}
		switch t := t.(type) {
		case *StructType:
			for _, f := range t.Fields {
				if s := gen(f.T, "v."+fieldName(f.Name)); s != "" {
					fmt.Fprintf(b, "  %s\n", s)
				}
			}
		case *EnumType:
			if op == "retain" {
				b.WriteString("  if (v.u) v.u->rc++;\n")
				break
			}
			fmt.Fprintf(b, "  if (v.u && --v.u->rc == 0) narra_drop(v.u, %s);\n", opFuncName("drop", t))
		}
		b.WriteString("}\n")
	}
}

func (e *Emitter) arrayFuncs(b *strings.Builder, a *ArrayType) {
	n := arrName(a)
	et := a.Elem.CName()
	rel := release(a.Elem, "a->items[i]")
	ret := retain(a.Elem, "c->items[i]")
	fmt.Fprintf(b, `
static %[1]s* %[1]s_new(int64_t cap) {
  %[1]s* a = narra_alloc(sizeof(%[1]s));
  a->rc = 1; a->len = 0; a->cap = cap;
  a->items = narra_alloc(sizeof(%[2]s) * (size_t)(cap > 0 ? cap : 1));
  return a;
}
static void %[1]s_retain(%[1]s* a) { a->rc++; }
static void %[1]s_release(%[1]s* a) {
  if (--a->rc == 0) {
    for (int64_t i = 0; i < a->len; i++) { %[3]s }
    narra_free(a->items);
    narra_free(a);
  }
}
// Returns an array the caller owns alone, copying it when it is shared.
static %[1]s* %[1]s_unique(%[1]s* a) {
  if (a->rc == 1) return a;
  %[1]s* c = %[1]s_new(a->cap);
  memcpy(c->items, a->items, sizeof(%[2]s) * (size_t)a->len);
  c->len = a->len;
  for (int64_t i = 0; i < c->len; i++) { %[4]s }
  a->rc--;
  return c;
}
static void %[1]s_push(%[1]s** p, %[2]s v) {
  %[1]s* a = %[1]s_unique(*p);
  if (a->len == a->cap) {
    a->cap = a->cap ? a->cap * 2 : 4;
    a->items = narra_realloc(a->items, sizeof(%[2]s) * (size_t)a->cap);
  }
  a->items[a->len++] = v;
  *p = a;
}
static %[2]s %[1]s_get(%[1]s* a, int64_t i, const char* where) {
  narra_bounds(i, a->len, where);
  return a->items[i];
}
static %[2]s* %[1]s_at(%[1]s** p, int64_t i, const char* where) {
  *p = %[1]s_unique(*p);
  narra_bounds(i, (*p)->len, where);
  return &(*p)->items[i];
}
static %[1]s* %[1]s_filled(%[2]s v, int64_t n, const char* where) {
  if (n < 0) narra_panic(where, "Array.filled with a negative count");
  %[1]s* a = %[1]s_new(n);
  for (int64_t i = 0; i < n; i++) { a->items[i] = v; %[5]s }
  a->len = n;
  return a;
}
`, n, et, rel, ret, retain(a.Elem, "a->items[i]"))
}

// Function bodies.
