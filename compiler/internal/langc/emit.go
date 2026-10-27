package langc

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

//go:embed runtime.h
var runtimeH string

// Ownership rules for reference-counted values in generated code:
//   - A local variable owns one reference. It is released when its scope ends.
//   - Parameters are borrowed. The caller keeps them alive for the call.
//   - A call returns an owned reference.
//   - Each expression produces either an owned temporary or a borrowed value.
//     Owned temporaries are released at the end of the statement unless a
//     binding, field or return takes them over.

type tempVar struct {
	name string
	t    Type
}

type scopeFrame struct {
	locals []*Local
	loop   bool // the body scope of a loop: break and continue unwind to here
}

type Emitter struct {
	prog     *Program
	file     string
	lits     strings.Builder
	litN     int
	out      strings.Builder // current function body
	indent   int
	tempN    int
	frames   [][]tempVar // owned temporaries, one list per statement in progress
	scopes   []*scopeFrame
	loopTemp []int // frame depth when each enclosing loop started
	fn       *FuncDecl
}

func Emit(prog *Program, file string, tests bool) string {
	e := &Emitter{prog: prog, file: file}
	var b strings.Builder
	b.WriteString(runtimeH)
	b.WriteString("\n// Types\n")
	for _, a := range prog.Arrays {
		fmt.Fprintf(&b, "typedef struct %s %s;\n", arrName(a), arrName(a))
	}
	for _, t := range prog.Types {
		fmt.Fprintf(&b, "typedef struct %s %s;\n", t.CName(), t.CName())
	}
	for _, t := range prog.Types {
		e.typeDef(&b, t)
	}
	for _, a := range prog.Arrays {
		fmt.Fprintf(&b, "struct %s { int64_t rc; int64_t len; int64_t cap; %s* items; };\n", arrName(a), a.Elem.CName())
	}
	b.WriteString("\n// Reference counting\n")
	for _, t := range prog.Types {
		if needsRC(t) {
			fmt.Fprintf(&b, "static void %s_retain(%s v);\nstatic void %s_release(%s v);\n", t.CName(), t.CName(), t.CName(), t.CName())
		}
	}
	for _, a := range prog.Arrays {
		n := arrName(a)
		fmt.Fprintf(&b, "static void %s_retain(%s* a);\nstatic void %s_release(%s* a);\n", n, n, n, n)
	}
	for _, t := range prog.Types {
		e.rcFuncs(&b, t)
	}
	for _, a := range prog.Arrays {
		e.arrayFuncs(&b, a)
	}
	b.WriteString("\n// Functions\n")
	for _, fd := range prog.Funcs {
		b.WriteString(e.signature(fd) + ";\n")
	}
	var bodies strings.Builder
	for _, fd := range prog.Funcs {
		e.function(fd)
		bodies.WriteString(e.out.String())
		e.out.Reset()
	}
	for i, t := range prog.Tests {
		e.fn = &FuncDecl{Name: "test", ResultT: TVoid}
		e.line("static void test_%d(void) {", i)
		e.indent++
		e.funcScope(nil)
		e.blockInScope(t.Body)
		e.unwindAll()
		e.indent--
		e.line("}")
		bodies.WriteString(e.out.String())
		e.out.Reset()
	}
	b.WriteString("\n// String literals\n")
	b.WriteString(e.lits.String())
	b.WriteString("\n")
	b.WriteString(bodies.String())
	if tests {
		b.WriteString("\nint main(void) {\n")
		for i, t := range prog.Tests {
			fmt.Fprintf(&b, "  test_%d();\n  printf(\"ok - %%s\\n\", %s);\n", i, cString(t.Name))
		}
		fmt.Fprintf(&b, "  printf(\"%d tests passed\\n\");\n  lang_report_leaks();\n  return 0;\n}\n", len(prog.Tests))
	} else if prog.Entry != nil {
		at := prog.Entry.Params[0].T.(*ArrayType)
		n := arrName(at)
		fmt.Fprintf(&b, `
int main(int argc, char** argv) {
  %[1]s* args = %[1]s_new(argc);
  for (int i = 0; i < argc; i++) {
    args->items[args->len++] = lstr_from_cstr(argv[i]);
  }
  int64_t code = f_run(args);
  %[1]s_release(args);
  fflush(stdout);
  lang_report_leaks();
  return (int)code;
}
`, n)
	}
	return b.String()
}

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
		fmt.Fprintf(b, "struct %s { int32_t tag;", t.CName())
		hasFields := false
		for _, v := range t.Variants {
			if len(v.Fields) > 0 {
				hasFields = true
			}
		}
		if hasFields {
			b.WriteString(" union {")
			for _, v := range t.Variants {
				if len(v.Fields) == 0 {
					continue
				}
				b.WriteString(" struct {")
				for _, f := range v.Fields {
					fmt.Fprintf(b, " %s %s;", f.T.CName(), fieldName(f.Name))
				}
				fmt.Fprintf(b, " } %s;", variantName(v))
			}
			b.WriteString(" } u;")
		}
		b.WriteString(" };\n")
	}
}

func fieldName(n string) string     { return "f_" + n }
func variantName(v *Variant) string { return "v_" + v.Name }

// retain and release return C statements for one value of type t.
func retain(t Type, v string) string {
	switch t := t.(type) {
	case *PrimType:
		if t == TStr {
			return fmt.Sprintf("lstr_retain(%s);", v)
		}
	case *ArrayType:
		return fmt.Sprintf("%s_retain(%s);", arrName(t), v)
	case *StructType, *EnumType:
		if needsRC(t) {
			return fmt.Sprintf("%s_retain(%s);", t.CName(), v)
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
	case *ArrayType:
		return fmt.Sprintf("%s_release(%s);", arrName(t), v)
	case *StructType, *EnumType:
		if needsRC(t) {
			return fmt.Sprintf("%s_release(%s);", t.CName(), v)
		}
	}
	return ""
}

func (e *Emitter) rcFuncs(b *strings.Builder, t Type) {
	if !needsRC(t) {
		return
	}
	for _, op := range []string{"retain", "release"} {
		fmt.Fprintf(b, "static void %s_%s(%s v) {\n", t.CName(), op, t.CName())
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
			b.WriteString("  switch (v.tag) {\n")
			for _, v := range t.Variants {
				var stmts []string
				for _, f := range v.Fields {
					if s := gen(f.T, "v.u."+variantName(v)+"."+fieldName(f.Name)); s != "" {
						stmts = append(stmts, s)
					}
				}
				if len(stmts) > 0 {
					fmt.Fprintf(b, "  case %d: %s break;\n", v.Index, strings.Join(stmts, " "))
				}
			}
			b.WriteString("  default: break;\n  }\n")
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
  %[1]s* a = lang_alloc(sizeof(%[1]s));
  a->rc = 1; a->len = 0; a->cap = cap;
  a->items = lang_alloc(sizeof(%[2]s) * (size_t)(cap > 0 ? cap : 1));
  return a;
}
static void %[1]s_retain(%[1]s* a) { a->rc++; }
static void %[1]s_release(%[1]s* a) {
  if (--a->rc == 0) {
    for (int64_t i = 0; i < a->len; i++) { %[3]s }
    lang_free(a->items);
    lang_free(a);
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
    a->items = lang_realloc(a->items, sizeof(%[2]s) * (size_t)a->cap);
  }
  a->items[a->len++] = v;
  *p = a;
}
static %[2]s %[1]s_get(%[1]s* a, int64_t i, const char* where) {
  lang_bounds(i, a->len, where);
  return a->items[i];
}
static %[2]s* %[1]s_at(%[1]s** p, int64_t i, const char* where) {
  *p = %[1]s_unique(*p);
  lang_bounds(i, (*p)->len, where);
  return &(*p)->items[i];
}
static %[1]s* %[1]s_filled(%[2]s v, int64_t n, const char* where) {
  if (n < 0) lang_panic(where, "Array.filled with a negative count");
  %[1]s* a = %[1]s_new(n);
  for (int64_t i = 0; i < n; i++) { a->items[i] = v; %[5]s }
  a->len = n;
  return a;
}
`, n, et, rel, ret, retain(a.Elem, "a->items[i]"))
}

// Function bodies.

func (e *Emitter) line(format string, args ...any) {
	e.out.WriteString(strings.Repeat("  ", e.indent))
	fmt.Fprintf(&e.out, format, args...)
	e.out.WriteString("\n")
}

func (e *Emitter) where(p Pos) string {
	return cString(fmt.Sprintf("%s:%d:%d", e.file, p.Line, p.Col))
}

func (e *Emitter) temp(t Type, code string) string {
	e.tempN++
	name := fmt.Sprintf("t%d", e.tempN)
	e.line("%s %s = %s;", t.CName(), name, code)
	return name
}

func (e *Emitter) ownedTemp(t Type, code string) string {
	name := e.temp(t, code)
	if needsRC(t) {
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{name, t})
	}
	return name
}

func (e *Emitter) beginStmt() { e.frames = append(e.frames, nil) }

func (e *Emitter) endStmt() {
	top := e.frames[len(e.frames)-1]
	for i := len(top) - 1; i >= 0; i-- {
		if s := release(top[i].t, top[i].name); s != "" {
			e.line("%s", s)
		}
	}
	e.frames = e.frames[:len(e.frames)-1]
}

// consume turns an expression result into an owned reference that the
// caller takes over.
func (e *Emitter) consume(code string, owned bool, t Type) string {
	if !needsRC(t) {
		return code
	}
	if owned {
		top := e.frames[len(e.frames)-1]
		for i, tv := range top {
			if tv.name == code {
				e.frames[len(e.frames)-1] = append(top[:i:i], top[i+1:]...)
				return code
			}
		}
	}
	name := e.temp(t, code)
	e.line("%s", retain(t, name))
	return name
}

func (e *Emitter) funcScope(params []*Param) {
	e.scopes = []*scopeFrame{{}}
	e.frames = nil
	e.loopTemp = nil
}

func (e *Emitter) pushScope(loop bool) { e.scopes = append(e.scopes, &scopeFrame{loop: loop}) }

func (e *Emitter) popScope() {
	top := e.scopes[len(e.scopes)-1]
	e.releaseLocals(top)
	e.scopes = e.scopes[:len(e.scopes)-1]
}

func (e *Emitter) releaseLocals(s *scopeFrame) {
	for i := len(s.locals) - 1; i >= 0; i-- {
		if r := release(s.locals[i].T, s.locals[i].CName); r != "" {
			e.line("%s", r)
		}
	}
}

func (e *Emitter) addLocal(l *Local) {
	if needsRC(l.T) {
		top := e.scopes[len(e.scopes)-1]
		top.locals = append(top.locals, l)
	}
}

// unwindAll releases every temporary and local, for a return.
func (e *Emitter) unwindAll() {
	for i := len(e.frames) - 1; i >= 0; i-- {
		for j := len(e.frames[i]) - 1; j >= 0; j-- {
			if s := release(e.frames[i][j].t, e.frames[i][j].name); s != "" {
				e.line("%s", s)
			}
		}
	}
	for i := len(e.scopes) - 1; i >= 0; i-- {
		e.releaseLocals(e.scopes[i])
	}
}

// unwindLoop releases the temporaries and locals created inside the
// innermost loop, for a break or continue.
func (e *Emitter) unwindLoop() {
	depth := e.loopTemp[len(e.loopTemp)-1]
	for i := len(e.frames) - 1; i >= depth; i-- {
		for j := len(e.frames[i]) - 1; j >= 0; j-- {
			if s := release(e.frames[i][j].t, e.frames[i][j].name); s != "" {
				e.line("%s", s)
			}
		}
	}
	for i := len(e.scopes) - 1; i >= 0; i-- {
		e.releaseLocals(e.scopes[i])
		if e.scopes[i].loop {
			break
		}
	}
}

func (e *Emitter) signature(fd *FuncDecl) string {
	var params []string
	for _, p := range fd.Params {
		if p.Mutable {
			params = append(params, fmt.Sprintf("%s* %s", p.T.CName(), p.Local.CName))
		} else {
			params = append(params, fmt.Sprintf("%s %s", p.T.CName(), p.Local.CName))
		}
	}
	if len(params) == 0 {
		params = []string{"void"}
	}
	return fmt.Sprintf("static %s f_%s(%s)", fd.ResultT.CName(), fd.Name, strings.Join(params, ", "))
}

func (e *Emitter) function(fd *FuncDecl) {
	e.fn = fd
	e.line("%s {", e.signature(fd))
	e.indent++
	e.funcScope(fd.Params)
	e.blockInScope(fd.Body)
	e.unwindAll()
	if fd.ResultT != TVoid {
		e.line("lang_panic(%s, \"function ended without returning a value\");", e.where(fd.Pos))
		e.line("return (%s){0};", fd.ResultT.CName())
	}
	e.indent--
	e.line("}")
	e.line("")
}

func (e *Emitter) blockInScope(b *Block) {
	for _, s := range b.Stmts {
		e.stmt(s)
	}
}

func (e *Emitter) block(b *Block, loop bool) {
	e.pushScope(loop)
	e.blockInScope(b)
	e.popScope()
}

func (e *Emitter) localRef(l *Local) string {
	if l.ByRef {
		return "(*" + l.CName + ")"
	}
	return l.CName
}

func (e *Emitter) stmt(s Stmt) {
	switch s := s.(type) {
	case *Block:
		e.line("{")
		e.indent++
		e.block(s, false)
		e.indent--
		e.line("}")
	case *VarStmt:
		e.beginStmt()
		code, owned := e.expr(s.Value)
		v := e.consume(code, owned, s.Local.T)
		e.line("%s %s = %s;", s.Local.T.CName(), s.Local.CName, v)
		e.endStmt()
		e.addLocal(s.Local)
	case *AssignStmt:
		e.beginStmt()
		code, owned := e.expr(s.Value)
		v := e.consume(code, owned, s.Value.Type())
		loc := e.lvaluePtr(s.Target)
		t := s.Target.Type()
		if r := release(t, "(*"+loc+")"); r != "" {
			ptr := e.temp(&ptrType{t}, loc)
			e.line("%s", release(t, "(*"+ptr+")"))
			e.line("*%s = %s;", ptr, v)
		} else {
			e.line("*%s = %s;", loc, v)
		}
		e.endStmt()
	case *ExprStmt:
		e.beginStmt()
		if m, ok := s.Expr.(*MatchExpr); ok {
			e.matchStmt(m)
		} else {
			code, _ := e.expr(s.Expr)
			if code != "" && s.Expr.Type() == TVoid {
				e.line("%s;", code)
			}
		}
		e.endStmt()
	case *ReturnStmt:
		e.beginStmt()
		if s.Value != nil {
			code, owned := e.expr(s.Value)
			v := e.consume(code, owned, s.Value.Type())
			r := e.temp(s.Value.Type(), v)
			e.unwindAll()
			e.line("return %s;", r)
		} else {
			e.unwindAll()
			e.line("return;")
		}
		e.frames = e.frames[:len(e.frames)-1]
	case *IfStmt:
		if is, ok := s.Cond.(*IsExpr); ok {
			e.ifIs(s, is)
			return
		}
		e.beginStmt()
		cond, _ := e.expr(s.Cond)
		c := e.temp(TBool, cond)
		e.endStmt()
		e.line("if (%s) {", c)
		e.indent++
		e.block(s.Then, false)
		e.indent--
		if s.Else != nil {
			e.line("} else {")
			e.indent++
			if b, ok := s.Else.(*Block); ok {
				e.block(b, false)
			} else {
				e.stmt(s.Else)
			}
			e.indent--
		}
		e.line("}")
	case *WhileStmt:
		e.line("while (1) {")
		e.indent++
		e.beginStmt()
		cond, _ := e.expr(s.Cond)
		c := e.temp(TBool, cond)
		e.endStmt()
		e.line("if (!%s) break;", c)
		e.loopTemp = append(e.loopTemp, len(e.frames))
		e.block(s.Body, true)
		e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
		e.indent--
		e.line("}")
	case *ForStmt:
		e.forStmt(s)
	case *BreakStmt:
		e.unwindLoop()
		e.line("break;")
	case *ContinueStmt:
		e.unwindLoop()
		e.line("continue;")
	}
}

// ifIs emits `if (x is Pattern) { ... }`. The subject is held for the whole
// statement, because the pattern's names read from it.
func (e *Emitter) ifIs(s *IfStmt, is *IsExpr) {
	et := is.Subject.Type().(*EnumType)
	e.line("{")
	e.indent++
	e.pushScope(false)
	e.beginStmt()
	code, owned := e.expr(is.Subject)
	held := e.consume(code, owned, et)
	subj := e.temp(et, held)
	e.endStmt()
	e.addLocal(&Local{Name: "subject", T: et, CName: subj})
	cond := "true"
	if !is.Pattern.Wildcard {
		cond = fmt.Sprintf("%s.tag == %d", subj, is.Pattern.Variant.Index)
	}
	e.line("if (%s) {", cond)
	e.indent++
	e.pushScope(false)
	e.bindPattern(is.Pattern, subj)
	e.blockInScope(s.Then)
	e.popScope()
	e.indent--
	if s.Else != nil {
		e.line("} else {")
		e.indent++
		if b, ok := s.Else.(*Block); ok {
			e.block(b, false)
		} else {
			e.stmt(s.Else)
		}
		e.indent--
	}
	e.line("}")
	e.popScope()
	e.indent--
	e.line("}")
}

// bindPattern declares the names a pattern binds, reading from subj.
func (e *Emitter) bindPattern(p *Pattern, subj string) {
	for i, l := range p.Locals {
		if l == nil {
			continue
		}
		f := p.Variant.Fields[i]
		e.line("%s %s = %s.u.%s.%s;", l.T.CName(), l.CName, subj, variantName(p.Variant), fieldName(f.Name))
		if r := retain(l.T, l.CName); r != "" {
			e.line("%s", r)
		}
		e.addLocal(l)
	}
}

type ptrType struct{ t Type }

func (p *ptrType) String() string { return p.t.String() + "*" }
func (p *ptrType) CName() string  { return p.t.CName() + "*" }

func (e *Emitter) forStmt(s *ForStmt) {
	call := s.Iter.(*CallExpr)
	e.beginStmt()
	switch s.Kind {
	case forRange:
		var lo, hi string
		if call.Builtin == "Range.upTo" {
			lo = "0"
			c, _ := e.expr(call.Args[0].Value)
			hi = e.temp(TInt, c)
		} else {
			a, _ := e.expr(call.Args[0].Value)
			lo = e.temp(TInt, a)
			b, _ := e.expr(call.Args[1].Value)
			hi = e.temp(TInt, b)
		}
		e.line("for (int64_t %s = %s; %s < %s; %s++) {", s.Local.CName, lo, s.Local.CName, hi, s.Local.CName)
		e.indent++
		e.loopTemp = append(e.loopTemp, len(e.frames))
		e.block(s.Body, true)
		e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
		e.indent--
		e.line("}")
	case forArray:
		at := call.Args[0].Value.Type().(*ArrayType)
		code, owned := e.expr(call.Args[0].Value)
		arr := e.consume(code, owned, at)
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{arr, at})
		e.tempN++
		idx := fmt.Sprintf("i%d", e.tempN)
		e.line("for (int64_t %s = 0; %s < %s->len; %s++) {", idx, idx, arr, idx)
		e.indent++
		e.loopTemp = append(e.loopTemp, len(e.frames))
		e.pushScope(true)
		e.line("%s %s = %s->items[%s];", at.Elem.CName(), s.Local.CName, arr, idx)
		if r := retain(at.Elem, s.Local.CName); r != "" {
			e.line("%s", r)
		}
		e.addLocal(s.Local)
		e.blockInScope(s.Body)
		e.popScope()
		e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
		e.indent--
		e.line("}")
	}
	e.endStmt()
}

// lvaluePtr returns a C pointer to the storage an assignable expression
// names. Arrays on the path are made unique first, so writes never change a
// shared array.
func (e *Emitter) lvaluePtr(x Expr) string {
	switch x := x.(type) {
	case *Ident:
		if x.Local.ByRef {
			return x.Local.CName
		}
		return "&" + x.Local.CName
	case *FieldExpr:
		return fmt.Sprintf("&(%s)->%s", e.lvaluePtr(x.Target), fieldName(x.Name))
	case *IndexExpr:
		at := x.Target.Type().(*ArrayType)
		base := e.lvaluePtr(x.Target)
		idx, _ := e.expr(x.Index)
		i := e.temp(TInt, idx)
		return e.temp(&ptrType{at.Elem}, fmt.Sprintf("%s_at(%s, %s, %s)", arrName(at), base, i, e.where(x.P)))
	}
	panic("not an lvalue")
}

func (e *Emitter) literal(s string) string {
	e.litN++
	name := fmt.Sprintf("L%d", e.litN)
	fmt.Fprintf(&e.lits, "static lstr %s = {-1, %d, %s};\n", name, len(s), cString(s))
	return "(&" + name + ")"
}

func cString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c >= 0x20 && c < 0x7f && c != '?':
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "\\%03o", c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// expr emits an expression and returns its C code and whether that code is
// an owned temporary.
func (e *Emitter) expr(x Expr) (string, bool) {
	switch x := x.(type) {
	case *IntLit:
		v, _ := strconv.ParseInt(x.Text, 0, 64)
		if v == -9223372036854775808 {
			return "INT64_MIN", false
		}
		return fmt.Sprintf("INT64_C(%d)", v), false
	case *FloatLit:
		return x.Text, false
	case *StringLit:
		return e.literal(x.Value), false
	case *CharLit:
		return fmt.Sprintf("%dU", x.Value), false
	case *BoolLit:
		return strconv.FormatBool(x.Value), false
	case *TemplateLit:
		if len(x.Parts) == 0 {
			return e.literal(""), false
		}
		acc, owned := e.expr(x.Parts[0])
		for _, p := range x.Parts[1:] {
			code, _ := e.expr(p)
			acc = e.ownedTemp(TStr, fmt.Sprintf("lstr_concat(%s, %s)", acc, code))
			owned = true
		}
		return acc, owned
	case *Ident:
		if x.Local != nil {
			return e.localRef(x.Local), false
		}
		if x.Ctor != nil {
			return fmt.Sprintf("((%s){.tag = %d})", x.Ctor.Enum.CName(), x.Ctor.Index), false
		}
	case *FieldExpr:
		if x.Ctor != nil {
			return fmt.Sprintf("((%s){.tag = %d})", x.Ctor.Enum.CName(), x.Ctor.Index), false
		}
		target, _ := e.expr(x.Target)
		if x.Narrow != nil {
			return fmt.Sprintf("(%s).u.%s.%s", target, variantName(x.Narrow), fieldName(x.Name)), false
		}
		return fmt.Sprintf("(%s).%s", target, fieldName(x.Name)), false
	case *IndexExpr:
		at := x.Target.Type().(*ArrayType)
		target, _ := e.expr(x.Target)
		idx, _ := e.expr(x.Index)
		return e.temp(at.Elem, fmt.Sprintf("%s_get(%s, %s, %s)", arrName(at), target, idx, e.where(x.P))), false
	case *UnaryExpr:
		v, _ := e.expr(x.Operand)
		switch x.Op {
		case "-":
			if x.Type() == TInt {
				return e.temp(TInt, fmt.Sprintf("lang_neg(%s, %s)", v, e.where(x.P))), false
			}
			return "(-" + v + ")", false
		case "!":
			return "(!" + v + ")", false
		case "~":
			return "(~" + v + ")", false
		}
	case *BinaryExpr:
		return e.binary(x), false
	case *IsExpr:
		code, _ := e.expr(x.Subject)
		if x.Pattern.Wildcard {
			return "true", false
		}
		return fmt.Sprintf("((%s).tag == %d)", code, x.Pattern.Variant.Index), false
	case *CallExpr:
		return e.call(x)
	case *StructLit:
		var parts []string
		given := map[string]bool{}
		for _, fi := range x.Fields {
			code, owned := e.expr(fi.Value)
			v := e.consume(code, owned, fi.Value.Type())
			parts = append(parts, fmt.Sprintf(".%s = %s", fieldName(fi.Name), v))
			given[fi.Name] = true
		}
		for _, f := range x.Struct.Fields {
			if !given[f.Name] {
				code, owned := e.expr(f.Default)
				v := e.consume(code, owned, f.T)
				parts = append(parts, fmt.Sprintf(".%s = %s", fieldName(f.Name), v))
			}
		}
		return e.ownedTemp(x.Struct, fmt.Sprintf("((%s){%s})", x.Struct.CName(), strings.Join(parts, ", "))), true
	case *ArrayLit:
		at := x.Type().(*ArrayType)
		arr := e.ownedTemp(at, fmt.Sprintf("%s_new(%d)", arrName(at), len(x.Elems)))
		for _, el := range x.Elems {
			code, owned := e.expr(el)
			v := e.consume(code, owned, at.Elem)
			e.line("%s->items[%s->len++] = %s;", arr, arr, v)
		}
		return arr, true
	case *MatchExpr:
		return e.matchValue(x)
	}
	panic(fmt.Sprintf("emit: unsupported expression %T", x))
}

func (e *Emitter) binary(x *BinaryExpr) string {
	if x.Op == "&&" || x.Op == "||" {
		l, _ := e.expr(x.Left)
		r := e.temp(TBool, l)
		if x.Op == "&&" {
			e.line("if (%s) {", r)
		} else {
			e.line("if (!%s) {", r)
		}
		e.indent++
		e.beginStmt()
		rc, _ := e.expr(x.Right)
		e.line("%s = %s;", r, rc)
		e.endStmt()
		e.indent--
		e.line("}")
		return r
	}
	l, _ := e.expr(x.Left)
	r, _ := e.expr(x.Right)
	lt := x.Left.Type()
	w := e.where(x.P)
	if lt == TStr {
		switch x.Op {
		case "==":
			return fmt.Sprintf("lstr_eq(%s, %s)", l, r)
		case "!=":
			return fmt.Sprintf("(!lstr_eq(%s, %s))", l, r)
		default:
			return fmt.Sprintf("(lstr_cmp(%s, %s) %s 0)", l, r, x.Op)
		}
	}
	if lt == TInt {
		fns := map[string]string{"+": "lang_add", "-": "lang_sub", "*": "lang_mul", "/": "lang_div", "%": "lang_rem", "<<": "lang_shl", ">>": "lang_shr"}
		if f, ok := fns[x.Op]; ok {
			return e.temp(TInt, fmt.Sprintf("%s(%s, %s, %s)", f, l, r, w))
		}
		wraps := map[string]string{"+%": "lang_wrap_add", "-%": "lang_wrap_sub", "*%": "lang_wrap_mul"}
		if f, ok := wraps[x.Op]; ok {
			return fmt.Sprintf("%s(%s, %s)", f, l, r)
		}
	}
	if lt == TF64 && x.Op == "%" {
		return fmt.Sprintf("fmod(%s, %s)", l, r)
	}
	op := x.Op
	return fmt.Sprintf("(%s %s %s)", l, op, r)
}

func (e *Emitter) call(x *CallExpr) (string, bool) {
	w := e.where(x.P)
	switch {
	case x.Ctor != nil:
		v := x.Ctor
		var parts []string
		for i, a := range x.Args {
			code, owned := e.expr(a.Value)
			c := e.consume(code, owned, v.Fields[i].T)
			parts = append(parts, fmt.Sprintf(".%s = %s", fieldName(v.Fields[i].Name), c))
		}
		init := fmt.Sprintf("((%s){.tag = %d", v.Enum.CName(), v.Index)
		if len(parts) > 0 {
			init += fmt.Sprintf(", .u.%s = {%s}", variantName(v), strings.Join(parts, ", "))
		}
		init += "})"
		return e.ownedTemp(v.Enum, init), true
	case x.Builtin != "":
		return e.builtin(x, w)
	case x.Func != nil:
		var args []string
		for i, a := range x.Args {
			if x.Func.Params[i].Mutable {
				args = append(args, e.lvaluePtr(a.Value))
				continue
			}
			code, _ := e.expr(a.Value)
			args = append(args, code)
		}
		callCode := fmt.Sprintf("f_%s(%s)", x.Func.Name, strings.Join(args, ", "))
		if x.Func.ResultT == TVoid {
			return callCode, false
		}
		return e.ownedTemp(x.Func.ResultT, callCode), true
	}
	panic("emit: unresolved call")
}

func (e *Emitter) builtin(x *CallExpr, w string) (string, bool) {
	arg := func(i int) string {
		code, _ := e.expr(x.Args[i].Value)
		return code
	}
	switch x.Builtin {
	case "print":
		return fmt.Sprintf("lang_print(%s)", arg(0)), false
	case "assert":
		return fmt.Sprintf("lang_assert(%s, %s)", arg(0), w), false
	case "panic":
		return fmt.Sprintf("lang_panic_str(%s, %s)", arg(0), w), false
	case "Int.toString":
		return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_int(%s)", arg(0))), true
	case "Int.abs":
		return e.temp(TInt, fmt.Sprintf("lang_abs(%s, %s)", arg(0), w)), false
	case "Float.toString":
		return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_f64(%s)", arg(0))), true
	case "Float.sqrt":
		return fmt.Sprintf("sqrt(%s)", arg(0)), false
	case "Float.fromInt":
		return fmt.Sprintf("((double)%s)", arg(0)), false
	case "Float.truncate":
		return e.temp(TInt, fmt.Sprintf("lang_truncate(%s, %s)", arg(0), w)), false
	case "Str.length":
		return fmt.Sprintf("(%s)->len", arg(0)), false
	case "Str.concat":
		a := arg(0)
		return e.ownedTemp(TStr, fmt.Sprintf("lstr_concat(%s, %s)", a, arg(1))), true
	case "Str.fromChar":
		return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_char(%s)", arg(0))), true
	case "Str.repeat":
		a := arg(0)
		return e.ownedTemp(TStr, fmt.Sprintf("lstr_repeat(%s, %s, %s)", a, arg(1), w)), true
	case "Array.length":
		return fmt.Sprintf("(%s)->len", arg(0)), false
	case "Array.push":
		at := x.Args[0].Value.Type().(*ArrayType)
		code, owned := e.expr(x.Args[1].Value)
		v := e.consume(code, owned, at.Elem)
		p := e.lvaluePtr(x.Args[0].Value)
		return fmt.Sprintf("%s_push(%s, %s)", arrName(at), p, v), false
	case "Array.filled":
		at := x.Type().(*ArrayType)
		v := arg(0)
		return e.ownedTemp(at, fmt.Sprintf("%s_filled(%s, %s, %s)", arrName(at), v, arg(1), w)), true
	}
	panic("emit: unknown builtin " + x.Builtin)
}

// matchArms emits an if-chain over the arms. body emits each arm's action.
func (e *Emitter) matchArms(m *MatchExpr, body func(arm *MatchArm)) {
	code, _ := e.expr(m.Subject)
	et := m.Subject.Type().(*EnumType)
	subj := e.temp(et, code)
	e.tempN++
	done := fmt.Sprintf("m%d", e.tempN)
	e.line("bool %s = false;", done)
	for _, arm := range m.Arms {
		p := arm.Pattern
		if p.Wildcard {
			e.line("if (!%s) {", done)
		} else {
			e.line("if (!%s && %s.tag == %d) {", done, subj, p.Variant.Index)
		}
		e.indent++
		e.pushScope(false)
		if !p.Wildcard {
			e.bindPattern(p, subj)
		}
		if arm.Guard != nil {
			e.beginStmt()
			g, _ := e.expr(arm.Guard)
			gv := e.temp(TBool, g)
			e.endStmt()
			e.line("if (%s) {", gv)
			e.indent++
		}
		e.line("%s = true;", done)
		body(arm)
		if arm.Guard != nil {
			e.indent--
			e.line("}")
		}
		e.popScope()
		e.indent--
		e.line("}")
	}
}

func (e *Emitter) matchStmt(m *MatchExpr) {
	e.matchArms(m, func(arm *MatchArm) {
		if arm.Block != nil {
			e.block(arm.Block, false)
			return
		}
		e.beginStmt()
		code, _ := e.expr(arm.Body)
		if code != "" && arm.Body.Type() == TVoid {
			e.line("%s;", code)
		}
		e.endStmt()
	})
}

func (e *Emitter) matchValue(m *MatchExpr) (string, bool) {
	t := m.Type()
	e.tempN++
	result := fmt.Sprintf("r%d", e.tempN)
	e.line("%s %s;", t.CName(), result)
	e.matchArms(m, func(arm *MatchArm) {
		e.beginStmt()
		code, owned := e.expr(arm.Body)
		v := e.consume(code, owned, t)
		e.line("%s = %s;", result, v)
		e.endStmt()
	})
	if needsRC(t) {
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{result, t})
		return result, true
	}
	return result, false
}
