package narra

import (
	"fmt"
	"sort"
	"strings"
)

// Effect checking (SPEC.md section 7). It runs after the type checker, over
// the checked syntax, and does not change types or the generated C.
//
// An effect set maps an effect name to the first thing that caused it. The
// names are `alloc`, `io`, `unsafe`, and the effect variables of the function
// being analysed (`E` in `<effect E>`), which stand for the effects of a
// function the caller passes in.
//
// Each function gets a set from its body: the effects of the builtins it
// calls, of the functions it calls, and of the function values it calls. The
// effects of a private function are the set its body gives (a fixed point, so
// recursion works). An exported function states its set after `uses`, and
// callers use the stated one. The body must stay inside it.
//
// An arrow function has a set of its own. It is added to the enclosing
// function only when something calls it. A call to a function with an effect
// variable replaces the variable by the sets of the arguments passed to the
// parameters whose function type names it.
//
// Function types in the type checker carry no effects, so a function value
// has a known set only when it is a parameter (its declared `uses`), a local
// bound to an arrow or a named function, or an argument written at the call.
// Calls through other values (struct fields, results of calls) count as pure.

type effCause struct {
	pos Pos
	via string // the callee that brought the effect, for the message
}

type effSet map[string]effCause

var knownEffects = map[string]bool{"alloc": true, "io": true, "unsafe": true}

type effectState struct {
	c       *Checker
	fx      map[*FuncDecl]effSet // body effects of every analysed function
	local   map[*Local]effSet    // effects of the function a local holds
	arrow   map[*ArrowExpr]effSet
	entry   map[*FuncDecl]bool
	report  bool
	changed bool
}

type effectWalker struct {
	s   *effectState
	out effSet
}

func (c *Checker) checkEffects() {
	s := &effectState{
		c:     c,
		fx:    map[*FuncDecl]effSet{},
		local: map[*Local]effSet{},
		arrow: map[*ArrowExpr]effSet{},
		entry: map[*FuncDecl]bool{},
	}
	for _, fd := range c.prog.Entries {
		s.entry[fd] = true
	}
	for _, m := range c.pkg.Modules {
		saved := c.enter(c.mods[m.Path])
		s.eachBody(m, func(fd *FuncDecl, test *TestDecl) {
			if fd != nil {
				s.checkSignature(fd)
			}
		})
		c.enter(saved)
	}
	if len(c.errs) > 0 {
		return
	}
	// Grow the sets until nothing changes. Sets only grow, so an unchanged
	// size means an unchanged set, and the loop ends.
	for pass := 0; pass < 100; pass++ {
		s.changed = false
		s.walkAll(false)
		if !s.changed {
			break
		}
	}
	s.walkAll(true)
}

// eachBody calls f for every function with a body in a module, and every test.
func (s *effectState) eachBody(m *SourceModule, f func(*FuncDecl, *TestDecl)) {
	for _, file := range m.Files {
		for _, d := range file.Decls {
			switch d := d.(type) {
			case *FuncDecl:
				if d.Body != nil {
					f(d, nil)
				}
			case *ImplDecl:
				for _, fd := range d.Funcs {
					if fd.Body != nil {
						f(fd, nil)
					}
				}
			case *TestDecl:
				f(nil, d)
			}
		}
	}
}

func (s *effectState) walkAll(report bool) {
	s.report = report
	c := s.c
	for _, m := range c.pkg.Modules {
		saved := c.enter(c.mods[m.Path])
		s.eachBody(m, func(fd *FuncDecl, test *TestDecl) {
			if test != nil {
				w := &effectWalker{s: s, out: effSet{}}
				w.block(test.Body)
				return
			}
			s.analyse(fd)
		})
		c.enter(saved)
	}
}

// analyse walks one function body, stores its set, and in the reporting pass
// checks an exported function against its declaration.
func (s *effectState) analyse(fd *FuncDecl) {
	for _, p := range fd.Params {
		if p.Local != nil && p.Type != nil && p.Type.Func {
			set := effSet{}
			for _, n := range p.Type.Uses {
				set[n] = effCause{pos: p.Pos}
			}
			s.local[p.Local] = set
		}
	}
	w := &effectWalker{s: s, out: effSet{}}
	w.block(fd.Body)
	if len(w.out) != len(s.fx[fd]) {
		s.changed = true
	}
	s.fx[fd] = w.out
	if !s.report || !fd.Exported || s.entry[fd] {
		return
	}
	declared := map[string]bool{}
	for _, n := range fd.Effects {
		declared[n] = true
	}
	for _, name := range sortedEffects(w.out) {
		if declared[name] {
			continue
		}
		cause := w.out[name]
		pos := cause.pos
		if pos == (Pos{}) {
			pos = fd.Pos
		}
		from := ""
		if cause.via != "" {
			from = fmt.Sprintf(" (through `%s`)", cause.via)
		}
		s.c.errorf(pos, "Function `%s` uses `%s`%s but its signature does not declare it; write `uses %s`", fd.Name, name, from, name)
	}
}

// checkSignature reports unknown effect names and effect variables that no
// parameter binds.
func (s *effectState) checkSignature(fd *FuncDecl) {
	vars := map[string]bool{}
	for _, v := range fd.EffectParams {
		vars[v] = true
	}
	known := func(pos Pos, n string) bool {
		if knownEffects[n] || vars[n] {
			return true
		}
		s.c.errorf(pos, "unknown effect `%s`; the effects are alloc, io and unsafe, or an effect variable declared as `<effect %s>`", n, n)
		return false
	}
	bound := map[string]bool{}
	var scan func(te *TypeExpr, top bool)
	scan = func(te *TypeExpr, top bool) {
		if te == nil {
			return
		}
		for _, n := range te.Uses {
			if known(te.Pos, n) && top {
				bound[n] = true
			}
		}
		for _, a := range te.Args {
			scan(a, false)
		}
		scan(te.Result, false)
	}
	for _, p := range fd.Params {
		scan(p.Type, true)
	}
	for _, n := range fd.Effects {
		if known(fd.Pos, n) && vars[n] && !bound[n] {
			s.c.errorf(fd.Pos, "effect variable `%s` in the `uses` list of `%s` must appear in the function type of a parameter", n, fd.Name)
		}
	}
}

func sortedEffects(set effSet) []string {
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// funcEffects is the set callers see: the declared one for an exported
// function, the inferred one for a private function.
func (s *effectState) funcEffects(fd *FuncDecl) effSet {
	if fd.Origin != nil {
		fd = fd.Origin
	}
	if s.entry[fd] {
		return effSet{"alloc": {}, "io": {}, "unsafe": {}}
	}
	if fd.Exported {
		set := effSet{}
		for _, n := range fd.Effects {
			set[n] = effCause{}
		}
		return set
	}
	return s.fx[fd]
}

func (w *effectWalker) add(name string, pos Pos, via string) {
	if _, ok := w.out[name]; !ok {
		w.out[name] = effCause{pos: pos, via: via}
	}
}

// valueEffects is the set of a function value written as an expression.
func (s *effectState) valueEffects(x Expr) effSet {
	switch x := x.(type) {
	case *ArrowExpr:
		return s.arrow[x]
	case *Ident:
		if x.Local != nil {
			return s.local[x.Local]
		}
		if x.Func != nil {
			return s.concrete(x.Func)
		}
	case *FieldExpr:
		if x.Func != nil {
			return s.concrete(x.Func)
		}
	}
	return nil
}

// concrete is funcEffects without the effect variables, for a named function
// used as a value.
func (s *effectState) concrete(fd *FuncDecl) effSet {
	o := fd
	if o.Origin != nil {
		o = o.Origin
	}
	vars := map[string]bool{}
	for _, v := range o.EffectParams {
		vars[v] = true
	}
	set := effSet{}
	for n, cause := range s.funcEffects(fd) {
		if !vars[n] {
			set[n] = cause
		}
	}
	return set
}

func (w *effectWalker) block(b *Block) {
	if b == nil {
		return
	}
	for _, st := range b.Stmts {
		w.stmt(st)
	}
}

func (w *effectWalker) stmt(st Stmt) {
	switch st := st.(type) {
	case *Block:
		w.block(st)
	case *VarStmt:
		w.expr(st.Value)
		if st.Local != nil {
			if set := w.s.valueEffects(st.Value); set != nil {
				w.s.local[st.Local] = set
			}
		}
	case *AssignStmt:
		w.expr(st.Target)
		w.expr(st.Value)
		if id, ok := st.Target.(*Ident); ok && id.Local != nil {
			if set := w.s.valueEffects(st.Value); set != nil {
				merged := effSet{}
				for n, c := range w.s.local[id.Local] {
					merged[n] = c
				}
				for n, c := range set {
					merged[n] = c
				}
				w.s.local[id.Local] = merged
			}
		}
	case *ExprStmt:
		w.expr(st.Expr)
	case *ReturnStmt:
		w.expr(st.Value)
	case *IfStmt:
		w.expr(st.Cond)
		w.block(st.Then)
		if st.Else != nil {
			w.stmt(st.Else)
		}
	case *WhileStmt:
		w.expr(st.Cond)
		w.block(st.Body)
	case *ForStmt:
		w.expr(st.Iter)
		w.block(st.Body)
	}
}

func (w *effectWalker) expr(x Expr) {
	switch x := x.(type) {
	case nil:
	case *TemplateLit:
		for _, p := range x.Parts {
			w.expr(p)
		}
		if len(x.Parts) > 1 {
			w.add("alloc", x.P, "template string")
		}
	case *FieldExpr:
		w.expr(x.Target)
	case *IsExpr:
		w.expr(x.Subject)
	case *IndexExpr:
		w.expr(x.Target)
		w.expr(x.Index)
	case *UnaryExpr:
		w.expr(x.Operand)
	case *BinaryExpr:
		w.expr(x.Left)
		w.expr(x.Right)
		if x.Op == "+" && x.T == TStr {
			w.add("alloc", x.P, "string +")
		}
	case *StructLit:
		for _, f := range x.Fields {
			w.expr(f.Value)
		}
		w.expr(x.Spread)
	case *ArrayLit:
		for _, e := range x.Elems {
			w.expr(e)
		}
		if !x.IsTuple && x.Tuple == nil {
			w.add("alloc", x.P, "array literal")
		}
	case *MatchExpr:
		w.expr(x.Subject)
		for _, arm := range x.Arms {
			w.expr(arm.Guard)
			w.expr(arm.Body)
			w.block(arm.Block)
		}
	case *ArrowExpr:
		sub := &effectWalker{s: w.s, out: effSet{}}
		sub.expr(x.Body)
		sub.block(x.Block)
		w.s.arrow[x] = sub.out
	case *CallExpr:
		w.call(x)
	}
}

func (w *effectWalker) call(e *CallExpr) {
	w.expr(e.Callee)
	for _, a := range e.Args {
		w.expr(a.Value)
	}
	s := w.s
	switch {
	case e.Builtin != "":
		if b := builtinsByName[e.Builtin]; b != nil {
			for _, n := range b.Effects {
				w.add(n, e.P, b.Name)
			}
		}
	case e.Func != nil:
		w.userCall(e)
	case e.FuncValue != nil:
		for n, cause := range s.valueEffects(e.Callee) {
			w.add(n, e.P, cause.via)
		}
	case e.Ctor != nil:
		if len(e.Args) > 0 {
			w.add("alloc", e.P, e.Ctor.Name)
		}
	}
}

// userCall adds the effects of a call to a declared function. A variable of
// the callee takes the effects of the arguments bound to it.
func (w *effectWalker) userCall(e *CallExpr) {
	s := w.s
	g := e.Func
	if g.Origin != nil {
		g = g.Origin
	}
	vars := map[string]bool{}
	for _, v := range g.EffectParams {
		vars[v] = true
	}
	for n := range s.funcEffects(g) {
		if !vars[n] {
			w.add(n, e.P, g.Name)
			continue
		}
		for i, p := range g.Params {
			if i >= len(e.Args) || p.Type == nil || !p.Type.Func || !contains(p.Type.Uses, n) {
				continue
			}
			for m := range s.valueEffects(e.Args[i].Value) {
				w.add(m, e.P, g.Name)
			}
		}
	}
	// A private function is inferred, so a function-type parameter without
	// `uses` acts as an implicit effect variable: the call has the effects
	// of the argument.
	if !g.Exported {
		for i, p := range g.Params {
			if i < len(e.Args) && p.Type != nil && p.Type.Func && len(p.Type.Uses) == 0 {
				for m, cause := range s.valueEffects(e.Args[i].Value) {
					w.add(m, e.P, cause.via)
				}
			}
		}
	}
	if !s.report {
		return
	}
	// An exported function's parameter whose function type has no variable
	// accepts only the effects it lists.
	for i, p := range g.Params {
		if i >= len(e.Args) || p.Type == nil || !p.Type.Func {
			continue
		}
		hasVar := false
		for _, n := range p.Type.Uses {
			hasVar = hasVar || vars[n]
		}
		if hasVar || (!g.Exported && len(p.Type.Uses) == 0) {
			continue
		}
		var extra []string
		arg := s.valueEffects(e.Args[i].Value)
		for _, n := range sortedEffects(arg) {
			if !contains(p.Type.Uses, n) {
				extra = append(extra, n)
			}
		}
		if len(extra) > 0 {
			s.c.errorf(e.Args[i].Value.position(), "the argument for `%s` of `%s` uses `%s`, but its function type does not declare it; write `uses %s` in the type, or `uses E` with `<effect E>` on `%s`",
				p.Name, g.Name, strings.Join(extra, "`, `"), strings.Join(extra, ", "), g.Name)
		}
	}
}

func contains(list []string, name string) bool {
	for _, n := range list {
		if n == name {
			return true
		}
	}
	return false
}
