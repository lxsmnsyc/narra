package narra

import (
	"fmt"
	"strconv"
	"strings"
)

type ptrType struct{ t Type }

func (p *ptrType) String() string { return p.t.String() + "*" }

func (p *ptrType) CName() string { return p.t.CName() + "*" }

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
		if x.Func != nil {
			return e.funcValueCode(x.Func, x.Type().(*FuncType))
		}
	case *FieldExpr:
		if x.Func != nil {
			return e.funcValueCode(x.Func, x.Type().(*FuncType))
		}
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
				return e.temp(TInt, fmt.Sprintf("narra_neg(%s, %s)", v, e.where(x.P))), false
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
		return e.isCond(x), false
	case *CallExpr:
		return e.call(x)
	case *StructLit:
		if x.Spread != nil {
			return e.spreadLit(x)
		}
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
		if x.Tuple != nil {
			return e.tupleLit(x)
		}
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
	case *ArrowExpr:
		return e.arrow(x)
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
	if _, prim := lt.(*PrimType); !prim && (x.Op == "==" || x.Op == "!=") {
		return e.eqCode(x.Op, lt, l, r)
	}
	if lt == TInt {
		fns := map[string]string{"+": "narra_add", "-": "narra_sub", "*": "narra_mul", "/": "narra_div", "%": "narra_rem", "<<": "narra_shl", ">>": "narra_shr"}
		if f, ok := fns[x.Op]; ok {
			return e.temp(TInt, fmt.Sprintf("%s(%s, %s, %s)", f, l, r, w))
		}
		wraps := map[string]string{"+%": "narra_wrap_add", "-%": "narra_wrap_sub", "*%": "narra_wrap_mul"}
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
