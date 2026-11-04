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
	loopTemp []int        // frame depth when each enclosing loop started
	loops    []*loopInfo  // the enclosing loops, parallel to loopTemp
	blocks   []*blockInfo // the block expressions being emitted, innermost last
	fn       *FuncDecl
	// tailCalls are the self calls of the function being emitted that
	// become jumps to its start (findTailCalls).
	tailCalls map[*CallExpr]bool
	traits    *traitEmit   // trait functions used so far
	closures  *closureEmit // emit_closure.go
}

func Emit(prog *Program, entry *FuncDecl, tests bool) string {
	e := &Emitter{prog: prog}
	var b strings.Builder
	b.WriteString(runtimeH)
	for _, m := range prog.Modules {
		// std/test is called only from test blocks, which a non-test build drops.
		if m.Runtime != "" && (tests || m.Path != "std/test") {
			b.WriteString("\n")
			b.WriteString(m.Runtime)
		}
	}
	b.WriteString("\n// Types\n")
	writeCNameTable(&b, prog)
	for _, a := range prog.Arrays {
		fmt.Fprintf(&b, "typedef struct %s %s;\n", arrName(a), arrName(a))
	}
	e.mapTypedefs(&b)
	for _, t := range prog.Types {
		fmt.Fprintf(&b, "typedef struct %s %s;\n", t.CName(), t.CName())
	}
	e.typedefHeap(&b)
	e.funcTypeDefs(&b)
	for _, t := range prog.Types {
		e.typeDef(&b, t)
	}
	for _, a := range prog.Arrays {
		fmt.Fprintf(&b, "struct %s { int64_t rc; int64_t len; int64_t cap; %s* items; };\n", arrName(a), a.Elem.CName())
	}
	e.mapStructs(&b)
	e.heapDefs(&b)
	b.WriteString("\n// Reference counting\n")
	e.boxProtos(&b)
	e.mapProtos(&b)
	for _, t := range prog.Types {
		if needsRC(t) {
			fmt.Fprintf(&b, "static void %s(%s v);\nstatic void %s(%s v);\n", opFuncName("retain", t), t.CName(), opFuncName("release", t), t.CName())
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
	for _, bt := range prog.Boxes {
		e.boxFuncs(&b, bt)
	}
	// The map functions call the key's eq and hash, so they are written
	// after the trait functions.
	var mapb strings.Builder
	for _, m := range prog.Maps {
		e.mapFuncs(&mapb, m)
	}
	b.WriteString("\n// Constants\n")
	e.constProtos(&b)
	b.WriteString("\n// Functions\n")
	for _, fd := range prog.Funcs {
		b.WriteString(e.signature(fd) + ";\n")
	}
	var bodies strings.Builder
	e.constBodies(&bodies)
	for _, fd := range prog.Funcs {
		e.function(fd)
		bodies.WriteString(e.out.String())
		e.out.Reset()
	}
	for i, t := range prog.Tests {
		if !tests {
			break // test blocks are not part of a non-test build
		}
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
	b.WriteString(mapb.String())
	e.closureFuncs(&b)
	b.WriteString("\n")
	b.WriteString(bodies.String())
	if tests {
		emitTestMain(&b, prog)
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
  narra_consts_release();
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
	e.loops = nil
	e.blocks = nil
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
	e.unwindToLoop(len(e.loopTemp) - 1)
}

// loopInfo is an enclosing loop: its label and the C labels that a labeled
// break or continue jumps to. A C label is written only when used.
type loopInfo struct {
	label        string
	pos          Pos
	breakUsed    bool
	continueUsed bool
}

func (l *loopInfo) breakName() string {
	return fmt.Sprintf("narra_break_%d_%d", l.pos.Line, l.pos.Col)
}

func (l *loopInfo) continueName() string {
	return fmt.Sprintf("narra_continue_%d_%d", l.pos.Line, l.pos.Col)
}

// enterLoop records a loop that starts at the current frame depth.
func (e *Emitter) enterLoop(label string, pos Pos) {
	e.loopTemp = append(e.loopTemp, len(e.frames))
	e.loops = append(e.loops, &loopInfo{label: label, pos: pos})
}

func (e *Emitter) leaveLoop() *loopInfo {
	info := e.loops[len(e.loops)-1]
	e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
	e.loops = e.loops[:len(e.loops)-1]
	return info
}

// continueLabel ends a loop body with the label that `continue name` jumps to.
func (e *Emitter) continueLabel() {
	if info := e.loops[len(e.loops)-1]; info.continueUsed {
		e.line("%s:;", info.continueName())
	}
}

// breakLabel follows a loop with the label that `break name` jumps to.
func (e *Emitter) breakLabel(info *loopInfo) {
	if info.breakUsed {
		e.line("%s:;", info.breakName())
	}
}

// labeledExit leaves the loop called label and jumps to its break or
// continue label.
func (e *Emitter) labeledExit(label string, isBreak bool) {
	for k := len(e.loops) - 1; k >= 0; k-- {
		info := e.loops[k]
		if info.label != label {
			continue
		}
		e.unwindToLoop(k)
		if isBreak {
			info.breakUsed = true
			e.line("goto %s;", info.breakName())
		} else {
			info.continueUsed = true
			e.line("goto %s;", info.continueName())
		}
		return
	}
	panic("labeledExit: unknown label " + label)
}

// unwindToLoop releases what the loops inside loop k, and loop k's own body,
// hold. The scope of loop k is the k-th loop scope from the outside.
func (e *Emitter) unwindToLoop(k int) {
	depth := e.loopTemp[k]
	for i := len(e.frames) - 1; i >= depth; i-- {
		for j := len(e.frames[i]) - 1; j >= 0; j-- {
			if s := release(e.frames[i][j].t, e.frames[i][j].name); s != "" {
				e.line("%s", s)
			}
		}
	}
	stop := len(e.loops) - 1 - k // loop scopes to pass before the target's
	for i := len(e.scopes) - 1; i >= 0; i-- {
		e.releaseLocals(e.scopes[i])
		if e.scopes[i].loop {
			if stop == 0 {
				break
			}
			stop--
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
	e.tailCalls = findTailCalls(fd)
	if len(e.tailCalls) > 0 {
		// The function owns its parameters, so a jump can replace them.
		for _, p := range fd.Params {
			if !p.Mutable && needsRC(p.T) {
				e.line("%s", retain(p.T, p.Local.CName))
				e.addLocal(p.Local)
			}
		}
		e.line("narra_tail:;")
	}
	e.blockInScope(fd.Body)
	e.unwindAll()
	e.tailCalls = nil
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

// blockInfo is a block expression being emitted: where a return stores its
// value and jumps to, and how deep the frames and scopes were at its start.
type blockInfo struct {
	result string
	t      Type
	label  string
	frames int
	scopes int
}

// blockExpr emits a block in value position as a C block with a result
// variable. Each return stores its value, releases what the block holds and
// jumps to the label after the block.
func (e *Emitter) blockExpr(x *BlockExpr) (string, bool) {
	return e.blockBody(x.Block, x.Type())
}

// blockBody emits a block as a block expression of type t. A match arm with a
// block in value position uses it too.
func (e *Emitter) blockBody(block *Block, t Type) (string, bool) {
	e.tempN++
	info := &blockInfo{result: fmt.Sprintf("r%d", e.tempN), t: t, label: fmt.Sprintf("narra_block_%d", e.tempN), frames: len(e.frames), scopes: len(e.scopes)}
	e.line("%s %s;", t.CName(), info.result)
	e.line("{")
	e.indent++
	e.blocks = append(e.blocks, info)
	e.pushScope(false)
	e.blockInScope(block)
	e.popScope()
	e.blocks = e.blocks[:len(e.blocks)-1]
	e.indent--
	e.line("}")
	e.line("%s:;", info.label)
	if needsRC(t) {
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{info.result, t})
		return info.result, true
	}
	return info.result, false
}

// blockReturn emits a return inside a block expression.
func (e *Emitter) blockReturn(s *ReturnStmt) {
	info := e.blocks[len(e.blocks)-1]
	e.beginStmt()
	code, owned := e.expr(s.Value)
	v := e.consume(code, owned, s.Value.Type())
	e.line("%s = %s;", info.result, v)
	// Release what the block holds, down to where it started.
	for i := len(e.frames) - 1; i >= info.frames; i-- {
		for j := len(e.frames[i]) - 1; j >= 0; j-- {
			if r := release(e.frames[i][j].t, e.frames[i][j].name); r != "" {
				e.line("%s", r)
			}
		}
	}
	for i := len(e.scopes) - 1; i >= info.scopes; i-- {
		e.releaseLocals(e.scopes[i])
	}
	e.line("goto %s;", info.label)
	e.frames = e.frames[:len(e.frames)-1]
}
