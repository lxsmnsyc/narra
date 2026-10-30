package narra

import (
	"fmt"
	"strings"
)

// Top-level constants. A constant whose value is a literal, an operator on
// constants or a fieldless constructor is emitted at each use, so it is a C
// constant. Any other constant is built on its first use: it has a static
// variable, a flag and a getter that returns the value without a new
// reference. The variable holds one reference until the program ends, when
// narra_consts_release frees it so the leak check stays quiet. SPEC section
// 10 says modules have no initialization code, so nothing runs at startup.

// staticConst reports whether x can be emitted at each use.
func staticConst(x Expr) bool {
	switch x := x.(type) {
	case *IntLit, *FloatLit, *StringLit, *CharLit, *BoolLit:
		return true
	case *TemplateLit:
		return len(x.Parts) == 0
	case *Ident:
		if x.Const != nil {
			return staticConst(x.Const.Value)
		}
		return x.Ctor != nil
	case *FieldExpr:
		if x.Const != nil {
			return staticConst(x.Const.Value)
		}
		return x.Ctor != nil
	case *UnaryExpr:
		return staticConst(x.Operand)
	case *BinaryExpr:
		return staticConst(x.Left) && staticConst(x.Right)
	}
	return false
}

// constRef emits a read of constant d. The value is borrowed.
func (e *Emitter) constRef(d *ConstDecl) (string, bool) {
	if staticConst(d.Value) {
		return e.expr(d.Value)
	}
	return e.temp(d.T, constGetCName(d)+"()"), false
}

// lazyConsts are the constants that need a getter.
func (e *Emitter) lazyConsts() []*ConstDecl {
	var out []*ConstDecl
	for _, d := range e.prog.Consts {
		if !staticConst(d.Value) {
			out = append(out, d)
		}
	}
	return out
}

// constProtos writes the variables and getter prototypes.
func (e *Emitter) constProtos(b *strings.Builder) {
	for _, d := range e.lazyConsts() {
		fmt.Fprintf(b, "static %s %s;\nstatic int %s_ready;\nstatic %s %s(void);\n", d.T.CName(), constCName(d), constCName(d), d.T.CName(), constGetCName(d))
	}
}

// constBodies writes the getters and narra_consts_release.
func (e *Emitter) constBodies(b *strings.Builder) {
	lazy := e.lazyConsts()
	for _, d := range lazy {
		e.fn = &FuncDecl{Name: "const", ResultT: TVoid, Module: d.Module}
		e.file = e.prog.Files[d.Module]
		e.line("static %s %s(void) {", d.T.CName(), constGetCName(d))
		e.indent++
		e.funcScope(nil)
		e.line("if (!%s_ready) {", constCName(d))
		e.indent++
		e.beginStmt()
		code, owned := e.expr(d.Value)
		v := e.consume(code, owned, d.T)
		e.line("%s = %s;", constCName(d), v)
		e.line("%s_ready = 1;", constCName(d))
		e.endStmt()
		e.indent--
		e.line("}")
		e.line("return %s;", constCName(d))
		e.indent--
		e.line("}")
		e.line("")
	}
	e.line("static void narra_consts_release(void) {")
	e.indent++
	for i := len(lazy) - 1; i >= 0; i-- {
		d := lazy[i]
		if r := release(d.T, constCName(d)); r != "" {
			e.line("if (%s_ready) { %s }", constCName(d), r)
		}
	}
	e.indent--
	e.line("}")
	e.line("")
	b.WriteString(e.out.String())
	e.out.Reset()
}
