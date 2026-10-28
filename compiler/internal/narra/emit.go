package narra

import (
	_ "embed"
	"fmt"
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
	traits   *traitEmit   // trait functions used so far
	closures *closureEmit // emit_closure.go
}

func Emit(prog *Program, entry *FuncDecl, tests bool) string {
	e := &Emitter{prog: prog}
	var b strings.Builder
	b.WriteString(runtimeH)
	for _, m := range prog.Modules {
		if m.Runtime != "" {
			b.WriteString("\n")
			b.WriteString(m.Runtime)
		}
	}
	b.WriteString("\n// Types\n")
	for _, a := range prog.Arrays {
		fmt.Fprintf(&b, "typedef struct %s %s;\n", arrName(a), arrName(a))
	}
	for _, t := range prog.Types {
		fmt.Fprintf(&b, "typedef struct %s %s;\n", t.CName(), t.CName())
	}
	e.funcTypeDefs(&b)
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
		e.fn = &FuncDecl{Name: "test", ResultT: TVoid, Module: t.Module}
		e.file = prog.Files[t.Module]
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
	e.traitFuncs(&b)
	e.closureFuncs(&b)
	b.WriteString("\n")
	b.WriteString(bodies.String())
	if tests {
		b.WriteString("\nint main(void) {\n")
		for i, t := range prog.Tests {
			fmt.Fprintf(&b, "  test_%d();\n  printf(\"ok - %%s\\n\", %s);\n", i, cString(t.Name))
		}
		fmt.Fprintf(&b, "  printf(\"%d tests passed\\n\");\n  narra_report_leaks();\n  return 0;\n}\n", len(prog.Tests))
	} else if entry != nil {
		at := entry.Params[0].T.(*ArrayType)
		n := arrName(at)
		fmt.Fprintf(&b, `
int main(int argc, char** argv) {
  %[1]s* args = %[1]s_new(argc);
  for (int i = 0; i < argc; i++) {
    args->items[args->len++] = lstr_from_cstr(argv[i]);
  }
  int64_t code = %[2]s(args);
  %[1]s_release(args);
  fflush(stdout);
  narra_report_leaks();
  return (int)code;
}
`, n, funcCName(entry))
	}
	return b.String()
}

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
	return fmt.Sprintf("static %s %s(%s)", fd.ResultT.CName(), funcCName(fd), strings.Join(params, ", "))
}

func (e *Emitter) function(fd *FuncDecl) {
	e.fn = fd
	e.file = e.prog.Files[fd.Module]
	e.line("%s {", e.signature(fd))
	e.indent++
	e.funcScope(fd.Params)
	e.blockInScope(fd.Body)
	e.unwindAll()
	if fd.ResultT != TVoid {
		e.line("narra_panic(%s, \"function ended without returning a value\");", e.where(fd.Pos))
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
