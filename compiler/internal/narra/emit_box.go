package narra

import (
	"fmt"
	"strings"
)

// Heap values: Box<T> and the fields of enum variants.
//
// An enum with at least one variant that has fields is a C struct
// { int32_t tag; Q* u; }, where Q (payloadCName) holds the reference count
// and a union with one struct per variant. A variant without fields has
// u == NULL and costs no allocation. The value is complete before Q is, so
// an enum can hold itself. Copying the value shares Q; retain and release
// count it.
//
// A Box<T> is a pointer to { rc, f_value }.

func boxName(b *BoxType) string { return strings.TrimSuffix(b.CName(), "*") }

// typedefHeap writes the forward declarations of the heap structs.
func (e *Emitter) typedefHeap(b *strings.Builder) {
	for _, bt := range e.prog.Boxes {
		fmt.Fprintf(b, "typedef struct %s %s;\n", boxName(bt), boxName(bt))
	}
	for _, t := range e.prog.Types {
		if et, ok := t.(*EnumType); ok && isBoxedEnum(et) {
			fmt.Fprintf(b, "typedef struct %s %s;\n", payloadCName(et), payloadCName(et))
		}
	}
}

// heapDefs writes the heap structs, after every value type is complete.
func (e *Emitter) heapDefs(b *strings.Builder) {
	for _, t := range e.prog.Types {
		et, ok := t.(*EnumType)
		if !ok || !isBoxedEnum(et) {
			continue
		}
		fmt.Fprintf(b, "struct %s { int64_t rc; int32_t tag; union {", payloadCName(et))
		for _, v := range et.Variants {
			if len(v.Fields) == 0 {
				continue
			}
			b.WriteString(" struct {")
			for _, f := range v.Fields {
				fmt.Fprintf(b, " %s %s;", f.T.CName(), fieldName(f.Name))
			}
			fmt.Fprintf(b, " } %s;", variantName(v))
		}
		b.WriteString(" }; };\n")
	}
	for _, bt := range e.prog.Boxes {
		fmt.Fprintf(b, "struct %s { int64_t rc; %s %s; };\n", boxName(bt), bt.Elem.CName(), fieldName("value"))
	}
}

func (e *Emitter) boxProtos(b *strings.Builder) {
	for _, bt := range e.prog.Boxes {
		n := boxName(bt)
		fmt.Fprintf(b, "static void %s_retain(%s* a);\nstatic void %s_release(%s* a);\n", n, n, n, n)
	}
}

func (e *Emitter) boxFuncs(b *strings.Builder, bt *BoxType) {
	n := boxName(bt)
	fmt.Fprintf(b, `
static %[1]s* %[1]s_new(%[2]s v) {
  %[1]s* a = narra_alloc(sizeof(%[1]s));
  a->rc = 1; a->f_value = v;
  return a;
}
static void %[1]s_retain(%[1]s* a) { a->rc++; }
static void %[1]s_drop(void* p) {
  %[1]s* a = p;
  %[3]s
  narra_free(a);
}
static void %[1]s_release(%[1]s* a) {
  if (--a->rc == 0) narra_drop(a, %[1]s_drop);
}
`, n, bt.Elem.CName(), release(bt.Elem, "a->f_value"))
}

// enumAlloc emits the allocation of a variant's payload and returns the C
// value. fields are the already consumed field values, in order.
func (e *Emitter) enumAlloc(v *Variant, fields []string) string {
	et := v.Enum
	if len(fields) == 0 {
		return fmt.Sprintf("((%s){.tag = %d})", et.CName(), v.Index)
	}
	e.tempN++
	p := fmt.Sprintf("q%d", e.tempN)
	e.line("%s* %s = narra_alloc(sizeof(%s));", payloadCName(et), p, payloadCName(et))
	e.line("%s->rc = 1; %s->tag = %d;", p, p, v.Index)
	for i, f := range v.Fields {
		e.line("%s->%s.%s = %s;", p, variantName(v), fieldName(f.Name), fields[i])
	}
	return fmt.Sprintf("((%s){.tag = %d, .u = %s})", et.CName(), v.Index, p)
}
