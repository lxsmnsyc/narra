package narra

import (
	"fmt"
	"strings"
)

// Generators in C (SPEC section 8, D19, D78). A function* compiles to five C
// definitions, named from its function name (generatorCName):
//
//   - The frame struct Y...: the header narra_gen (runtime.h), then one field
//     per C variable of the body: the parameters, every Narra local and every
//     temporary and loop counter. Nothing of the body lives in a C local
//     across a yield.
//   - The resume function y...(narra_gen* h, void* out): a switch on the
//     state jumps to the label after the yield it stopped at. A yield writes
//     its value through out, sets the state and returns true. Ending the body
//     or a return releases what is live, sets the state to -1 and returns false.
//   - The clone function j...: copies the frame and retains what is live at
//     its state. narra_gen_next calls it to resume a shared generator.
//   - The drop function z...: releases what is live at its state and frees
//     the frame.
//   - The function itself: it allocates the frame, stores and retains the
//     parameters, and runs nothing of the body.
//
// What is live at a state is what a return at that yield would release (the
// statement temporaries and scope locals of unwindAll), recorded per yield.
// So the fields of values whose scope has ended are never read again, and
// they need no zeroing.

// genState is the generator body being emitted.
type genState struct {
	fields []string        // "type name" of each frame field, in first-use order
	seen   map[string]bool // field names
	locals map[*Local]bool // the locals that are frame fields
	elem   Type
	// live lists, for state 0 and each yield after it, the values the frame
	// holds while it waits there.
	live [][]tempVar
}

func (g *genState) field(ctype, name string) {
	if !g.seen[name] {
		g.seen[name] = true
		g.fields = append(g.fields, ctype+" "+name)
	}
}

// cVar declares a C variable: it returns the declaration to write before
// " = value" and the name to use. In a generator body the variable is a
// frame field and the declaration is the field itself.
func (e *Emitter) cVar(ctype, name string) (string, string) {
	if e.gen == nil {
		return ctype + " " + name, name
	}
	e.gen.field(ctype, name)
	return "f->" + name, "f->" + name
}

// declareLocal writes the C declaration of a local holding value.
func (e *Emitter) declareLocal(l *Local, value string) {
	decl := e.declareCounter(l)
	e.line("%s = %s;", decl, value)
}

// declareCounter declares a local without a value and returns the text that
// starts its declaration, for a for-loop header or `decl = value;`.
func (e *Emitter) declareCounter(l *Local) string {
	if e.gen != nil {
		e.gen.locals[l] = true
	}
	decl, _ := e.cVar(l.T.CName(), l.CName)
	return decl
}

// liveValues lists what the body holds now, in the order unwindAll releases it.
func (e *Emitter) liveValues() []tempVar {
	var live []tempVar
	for i := len(e.frames) - 1; i >= 0; i-- {
		for j := len(e.frames[i]) - 1; j >= 0; j-- {
			live = append(live, e.frames[i][j])
		}
	}
	for i := len(e.scopes) - 1; i >= 0; i-- {
		locals := e.scopes[i].locals
		for j := len(locals) - 1; j >= 0; j-- {
			live = append(live, tempVar{e.localRef(locals[j]), locals[j].T})
		}
	}
	return live
}

// generator emits a function* (see the comment at the top of the file).
func (e *Emitter) generator(fd *FuncDecl) {
	e.fn = fd
	e.file = e.prog.Files[fd.Module]
	owner := funcCName(fd)
	frame, resume, clone, drop := generatorCName("Y", owner), generatorCName("y", owner), generatorCName("j", owner), generatorCName("z", owner)
	e.gen = &genState{seen: map[string]bool{}, locals: map[*Local]bool{}, elem: fd.ResultT.(*GeneratorType).Elem}
	e.funcScope(fd.Params)
	e.tailCalls = nil

	// The resume body, written first: it decides the frame fields. The frame
	// owns the parameters, so they are locals of the outer scope.
	e.indent = 1
	for _, p := range fd.Params {
		e.declareCounter(p.Local)
		e.addLocal(p.Local)
	}
	e.gen.live = append(e.gen.live, e.liveValues())
	e.blockInScope(fd.Body)
	e.generatorReturn()
	body := e.out.String()
	e.out.Reset()
	gen := e.gen
	e.gen = nil
	e.indent = 0

	e.line("typedef struct %s { narra_gen h;%s } %s;", frame, fieldList(gen.fields), frame)
	e.line("static bool %s(narra_gen* h, void* out) {", resume)
	e.line("  %s* f = (%s*)h;", frame, frame)
	e.line("  (void)out;")
	e.line("  switch (f->h.state) {")
	e.line("  case 0: break;")
	for n := 1; n < len(gen.live); n++ {
		e.line("  case %d: goto narra_resume_%d;", n, n)
	}
	e.line("  default: return false;")
	e.line("  }")
	e.out.WriteString(body)
	e.line("}")
	e.line("static narra_gen* %s(narra_gen* h) {", clone)
	e.line("  %s* f = narra_alloc(sizeof(%s));", frame, frame)
	e.line("  memcpy(f, h, sizeof(%s));", frame)
	e.line("  f->h.rc = 1;")
	e.stateSwitch(gen.live, retain)
	e.line("  return &f->h;")
	e.line("}")
	e.line("static void %s(void* p) {", drop)
	e.line("  %s* f = p;", frame)
	e.stateSwitch(gen.live, release)
	e.line("  narra_free(f);")
	e.line("}")
	e.line("%s {", e.signature(fd))
	e.line("  %s* f = narra_alloc(sizeof(%s));", frame, frame)
	e.line("  memset(f, 0, sizeof(%s));", frame)
	e.line("  f->h = (narra_gen){1, 0, %s, %s, %s};", resume, clone, drop)
	for _, p := range fd.Params {
		e.line("  f->%s = %s;", p.Local.CName, p.Local.CName)
		if r := retain(p.T, "f->"+p.Local.CName); r != "" {
			e.line("  %s", r)
		}
	}
	e.line("  return &f->h;")
	e.line("}")
	e.line("")
}

func fieldList(fields []string) string {
	var b strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&b, " %s;", f)
	}
	return b.String()
}

// stateSwitch writes a switch on the state that applies op (retain or
// release) to each value live at that state. States with nothing live have
// no case.
func (e *Emitter) stateSwitch(live [][]tempVar, op func(Type, string) string) {
	e.line("  switch (f->h.state) {")
	for n, values := range live {
		var stmts []string
		for _, v := range values {
			if s := op(v.t, v.name); s != "" {
				stmts = append(stmts, s)
			}
		}
		if len(stmts) > 0 {
			e.line("  case %d: %s break;", n, strings.Join(stmts, " "))
		}
	}
	e.line("  default: break;")
	e.line("  }")
}

// generatorReturn ends the generator: a return, or the end of the body.
func (e *Emitter) generatorReturn() {
	e.unwindAll()
	e.line("f->h.state = -1;")
	e.line("return false;")
}

// suspend records what is live, and returns from resume with a value. The
// next resume continues at the label after it.
func (e *Emitter) suspend() {
	n := len(e.gen.live)
	e.gen.live = append(e.gen.live, e.liveValues())
	e.line("f->h.state = %d;", n)
	e.line("return true;")
	e.line("narra_resume_%d:;", n)
}

// yieldStmt emits `yield value` and `yield* other`.
func (e *Emitter) yieldStmt(s *YieldStmt) {
	e.beginStmt()
	t := s.Value.Type()
	code, owned := e.expr(s.Value)
	v := e.consume(code, owned, t)
	if !s.Delegate {
		// The caller of resume owns the value.
		e.line("*(%s*)out = %s;", e.gen.elem.CName(), v)
		e.endStmt()
		e.suspend()
		return
	}
	// yield* resumes its own copy of other and passes each value on.
	e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{v, t})
	e.line("while (narra_gen_next(&%s, out)) {", v)
	e.indent++
	e.suspend()
	e.indent--
	e.line("}")
	e.endStmt()
}

// generatorFor emits `for (const x of g)` over a Generator<T>: the loop
// resumes its own copy of g until it finishes, and each value it gets is
// owned by x.
func (e *Emitter) generatorFor(s *ForStmt) {
	t := s.Iter.Type()
	code, owned := e.expr(s.Iter)
	g := e.consume(code, owned, t)
	e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{g, t})
	e.line("while (1) {")
	e.indent++
	decl := e.declareCounter(s.Local)
	if e.gen == nil {
		e.line("%s;", decl)
	}
	e.line("if (!narra_gen_next(&%s, &%s)) break;", g, e.localRef(s.Local))
	e.enterLoop(s.Label, s.Pos)
	e.pushScope(true)
	e.addLocal(s.Local)
	e.blockInScope(s.Body)
	e.popScope()
	e.continueLabel()
	info := e.leaveLoop()
	e.indent--
	e.line("}")
	e.breakLabel(info)
}
