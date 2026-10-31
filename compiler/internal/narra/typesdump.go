package narra

import (
	"fmt"
	"strings"
)

// The types dump is a fixed text form of what the body checker decides: the
// type of every expression and pattern, what each name resolves to, and the
// locals. `narra types <path>` prints it, and selfhost/src/typesdump.nra
// (`narrac types <path>`) prints the same format, so the Go checker and the
// Narra checker can be diffed.
//
// Input
//
//	path is a package directory or a single .nra file, as Load accepts. The
//	dump runs what `narra check` runs: the declaration passes and the
//	bodies of each module in load order, the pending generic function
//	instances, the Map keys, the entry functions and finish
//	(checkPackageBodies), then the effect pass (checkEffects) when the
//	bodies have no errors.
//
// Blocks
//
//	One block per checked body, in the order the checks start. A body that
//	starts while another is being checked (a constant checked at its first
//	use) comes before the rest of the outer body's block, because blocks are
//	listed when they start and printed after all checks. Headers:
//	    body module="<m>" func <name>
//	    body module="<m>" func <name> targs=<type>,<type>
//	    body module="<m>" test "<name>"
//	    body module="<m>" impl <Trait> <Type>.<function>
//	    body module="<m>" const <name>
//	    body module="<m>" default <Type>.<field>
//	<m> is the module whose names the body sees (the declaring module). A
//	func with targs is an instance of a generic function, checked after
//	the bodies of every module, in discovery order. A generic function is
//	also checked once with its type parameters left open, as a plain func
//	block. A default belongs to a struct field; <Type> is the struct as the
//	decls dump writes it, or <module>.<Name> for a generic struct.
//
// Lines
//
//	Inside a block, one line per node, indented by two spaces, in pre-order
//	with children in source order (the walk of selfhost/src/idcheck.nra):
//	parameters, then the statements; in a statement, its local, then its
//	pattern, then its expressions.
//	    <line>:<col> <kind> [attr...] <type>
//	    <line>:<col> local <name> [mutable] <type>
//	<kind> is the node name the AST dump prints (int, ident, call,
//	field-expr, structlit, pat-name, ...), at the position the AST dump
//	prints (a call is at its `(`). The type is written as in the decls dump
//	(module-qualified), or `-` when the checker gave the node no type (the
//	callee of a call, a node an error skipped). A `local` line stands for
//	each declared local, at its declaring position: a parameter, a
//	variable, a loop variable, later a pattern binding. Attributes, in this
//	order:
//	    local=<name>              a name that resolves to a local
//	    const=<module>.<name>     a top-level constant
//	    func=<module>.<name>      a function (an instance names its generic function)
//	    func=<Type>.<Trait>.<f>   an implement block function, called as Trait.f
//	    ctor=<Enum>.<Variant>     an enum constructor, or the variant a pattern names
//	    builtin=<Builtin.Name>    a std function (Array.push, print)
//	    narrow=<Variant>          a field read through a narrowed enum value
//	    trait=<Trait>.<function>  a trait function the compiler implements
//	    value                     a call through a function value
//	    stack                     an arrow function passed to a non-escaping parameter
//	    captures=<local>,<local>  the locals an arrow function captures, in order of first use
//	    distinct=<module>.<Name>  T.from or T.toInt on a distinct type
//	A namespace alias that the checker resolved (Alias.NAME of a package
//	module) has the type `module <path>`.
//
// Effects
//
//	After the last block, one line per function body that is not a generic
//	instance (kind func or impl), in the order of the blocks:
//	    effects module="<m>" <kind> <name> <effect>,<effect>
//	The set is the one callers see: the declared one for an exported
//	function (effect variables included), the inferred one for a private
//	function, all three for an entry function. Sorted, `-` when empty.
//
// Errors
//
//	A failed load or check prints only the errors, as the decls dump does:
//	    error <file>:<line>:<col>: <message>
//	in report order after dedup, and the command exits 1.
//
// Owners
//
//	selfhost-check-bodies: the block headers, local, the expression lines
//	of literals, names, fields, index, unary, binary, struct and array
//	literals and plain calls, and the attributes local, const, func (plain
//	functions) and builtin. selfhost-check-calls: the impl and targs
//	headers, the attributes ctor, trait and distinct, func on generic
//	functions, package-module functions and implement block functions, and
//	the `module <path>` type of Alias in Alias.Enum.Variant.
//	selfhost-check-patterns: pattern lines, narrow, and pattern-binding
//	locals (a `local` line right after the binding's pat-name line).
//	selfhost-check-closures: value, stack, captures, func on function
//	values and Alias.f, arrow lines and their parameter locals.
//	selfhost-check-builtins: the builtin= attribute and the type of every
//	std call whose check function is named in the registry (Array.push,
//	Option.map, Test.eq and the like). selfhost-check-effects: the effects
//	lines and the effect errors (selfhost/src/effects.nra).

// checkedBody is one body the checker checked, logged when it starts.
type checkedBody struct {
	module string
	kind   string // func, test, impl, const, default
	name   string // the header text after the kind
	targs  []Type // a generic function instance
	fd     *FuncDecl
	block  *Block
	expr   Expr
}

// logBody lists a body in checking order when TypesDump runs.
func (c *Checker) logBody(b checkedBody) {
	if !c.logBodies {
		return
	}
	b.module = c.mod.path
	c.bodies = append(c.bodies, b)
}

// implSourceName is the name an implement block's function has in the
// source; resolveImpl renames the function to its C name.
func (c *Checker) implSourceName(t Type, fd *FuncDecl) string {
	for key, f := range c.traits.implFuncs[t] {
		if f == fd {
			_, name, _ := strings.Cut(key, ".")
			return name
		}
	}
	return fd.Name
}

// instanceArgs lists the type arguments of an instance in the order of the
// generic function's type parameters.
func instanceArgs(tps []*TypeParamType, env map[string]Type) []Type {
	var out []Type
	for _, p := range tps {
		out = append(out, env[p.Name])
	}
	return out
}

// TypesDump loads the package at path, checks it and returns the dump
// described above.
func TypesDump(path string) (dump string, ok bool) {
	pkg, err := Load(path)
	if err != nil {
		return errorLines(err), false
	}
	c, err := checkPackageBodies(pkg, true)
	if err != nil {
		return errorLines(err), false
	}
	var effects *effectState
	if len(c.errs) == 0 {
		effects = c.checkEffects()
	}
	if len(c.errs) > 0 {
		return errorLines(Diagnostics(c.errs)), false
	}
	d := &typesDumper{implNames: implDumpNames(c)}
	for _, b := range c.bodies {
		d.body(b)
	}
	d.effects(c, effects)
	return d.b.String(), true
}

type typesDumper struct {
	b strings.Builder
	// implNames names each implement block function <Type>.<Trait>.<f>;
	// resolveImpl renamed the declarations to their C names.
	implNames map[*FuncDecl]string
}

// implDumpNames lists the implement block functions by their dump names.
func implDumpNames(c *Checker) map[*FuncDecl]string {
	names := map[*FuncDecl]string{}
	for t, funcs := range c.traits.implFuncs {
		for key, fd := range funcs {
			names[fd] = dumpTypeString(t) + "." + key
		}
	}
	return names
}

func (d *typesDumper) body(b checkedBody) {
	// The header names the body; an instance adds its type arguments.
	header := "body module=" + quoteDump(b.module) + " " + b.kind + " " + b.name
	if b.targs != nil {
		var parts []string
		for _, targ := range b.targs {
			parts = append(parts, dumpTypeString(targ))
		}
		header += " targs=" + strings.Join(parts, ",")
	}
	d.b.WriteString(header + "\n")

	// Then the nodes of the body, parameters first.
	switch {
	case b.fd != nil:
		d.params(b.fd.Params)
		if b.fd.Body != nil {
			d.block(b.fd.Body)
		}
	case b.block != nil:
		d.block(b.block)
	case b.expr != nil:
		d.expr(b.expr)
	}
}

// effects writes the effects lines: the set callers see for each function
// body that is not an instance.
func (d *typesDumper) effects(c *Checker, s *effectState) {
	for _, b := range c.bodies {
		if b.fd == nil || b.targs != nil || (b.kind != "func" && b.kind != "impl") {
			continue
		}
		text := "-"
		if names := sortedEffects(s.funcEffects(b.fd)); len(names) > 0 {
			text = strings.Join(names, ",")
		}
		d.b.WriteString(fmt.Sprintf("effects module=%s %s %s %s\n", quoteDump(b.module), b.kind, b.name, text))
	}
}

func typeText(t Type) string {
	if t == nil {
		return "-"
	}
	return dumpTypeString(t)
}

func (d *typesDumper) line(pos Pos, kind string, attrs []string, t Type) {
	d.b.WriteString(fmt.Sprintf("  %d:%d %s", pos.Line, pos.Col, kind))
	for _, a := range attrs {
		d.b.WriteString(" " + a)
	}
	d.b.WriteString(" " + typeText(t) + "\n")
}

func (d *typesDumper) local(l *Local) {
	if l == nil {
		return
	}
	flag := ""
	if l.Mutable {
		flag = " mutable"
	}
	d.b.WriteString(fmt.Sprintf("  %d:%d local %s%s %s\n", l.Pos.Line, l.Pos.Col, l.Name, flag, typeText(l.T)))
}

func (d *typesDumper) params(ps []*Param) {
	for _, p := range ps {
		d.local(p.Local)
	}
}

func (d *typesDumper) block(b *Block) {
	for _, s := range b.Stmts {
		d.stmt(s)
	}
}

func (d *typesDumper) stmt(s Stmt) {
	switch s := s.(type) {
	case *Block:
		d.block(s)
	case *VarStmt:
		d.local(s.Local)
		if s.Pattern != nil {
			d.pattern(s.Pattern)
		}
		d.expr(s.Value)
	case *AssignStmt:
		d.expr(s.Target)
		d.expr(s.Value)
	case *ExprStmt:
		d.expr(s.Expr)
	case *ReturnStmt:
		if s.Value != nil {
			d.expr(s.Value)
		}
	case *IfStmt:
		d.expr(s.Cond)
		d.block(s.Then)
		if s.Else != nil {
			d.stmt(s.Else)
		}
	case *WhileStmt:
		d.expr(s.Cond)
		d.block(s.Body)
	case *ForStmt:
		d.local(s.Local)
		d.expr(s.Iter)
		d.block(s.Body)
	}
}

// funcName names a function for the func attribute: an instance by the
// generic function it was made from, an implement block function by its
// type, trait and name.
func (d *typesDumper) funcName(fd *FuncDecl) string {
	if name, ok := d.implNames[fd]; ok {
		return name
	}
	if fd.Origin != nil {
		fd = fd.Origin
	}
	return qualifiedName(fd.Module, fd.Name)
}

func ctorName(v *Variant) string {
	if v.Enum != nil {
		return v.Enum.Name + "." + v.Name
	}
	return v.Name
}

// refAttrs lists the resolution attributes of a node, in the documented order.
func (d *typesDumper) refAttrs(local *Local, cd *ConstDecl, fd *FuncDecl, ctor *Variant, builtin string) []string {
	var out []string
	if local != nil {
		out = append(out, "local="+local.Name)
	}
	if cd != nil {
		out = append(out, "const="+qualifiedName(cd.Module, cd.Name))
	}
	if fd != nil {
		out = append(out, "func="+d.funcName(fd))
	}
	if ctor != nil {
		out = append(out, "ctor="+ctorName(ctor))
	}
	if builtin != "" {
		out = append(out, "builtin="+builtin)
	}
	return out
}

func (d *typesDumper) exprs(es []Expr) {
	for _, e := range es {
		d.expr(e)
	}
}

func (d *typesDumper) expr(e Expr) {
	switch e := e.(type) {
	case *IntLit:
		d.line(e.P, "int", nil, e.T)
	case *FloatLit:
		d.line(e.P, "float", nil, e.T)
	case *StringLit:
		d.line(e.P, "string", nil, e.T)
	case *CharLit:
		d.line(e.P, "char", nil, e.T)
	case *BoolLit:
		d.line(e.P, "bool", nil, e.T)
	case *TemplateLit:
		d.line(e.P, "template", nil, e.T)
		d.exprs(e.Parts)
	case *Ident:
		d.line(e.P, "ident", d.refAttrs(e.Local, e.Const, e.Func, e.Ctor, ""), e.T)
	case *FieldExpr:
		attrs := d.refAttrs(nil, e.Const, e.Func, e.Ctor, e.Builtin)
		if e.Narrow != nil {
			attrs = append(attrs, "narrow="+e.Narrow.Name)
		}
		d.line(e.P, "field-expr", attrs, e.T)
		d.expr(e.Target)
	case *IsExpr:
		d.line(e.P, "is", nil, e.T)
		d.expr(e.Subject)
		d.pattern(e.Pattern)
	case *IndexExpr:
		d.line(e.P, "index", nil, e.T)
		d.expr(e.Target)
		d.expr(e.Index)
	case *MapLit:
		d.line(e.P, "map", nil, e.T)
		d.exprs(e.Keys)
		d.exprs(e.Vals)
	case *CallExpr:
		attrs := d.refAttrs(nil, nil, e.Func, e.Ctor, e.Builtin)
		if e.Trait != nil {
			attrs = append(attrs, "trait="+e.Trait.Trait.Name+"."+e.Trait.Func)
		}
		if e.FuncValue != nil {
			attrs = append(attrs, "value")
		}
		if e.Distinct != nil {
			attrs = append(attrs, "distinct="+qualifiedName(e.Distinct.Module, e.Distinct.Name))
		}
		d.line(e.P, "call", attrs, e.T)
		d.expr(e.Callee)
		for _, a := range e.Args {
			d.expr(a.Value)
		}
	case *UnaryExpr:
		d.line(e.P, "unary", nil, e.T)
		d.expr(e.Operand)
	case *BinaryExpr:
		d.line(e.P, "binary", nil, e.T)
		d.expr(e.Left)
		d.expr(e.Right)
	case *StructLit:
		d.line(e.P, "structlit", nil, e.T)
		for _, f := range e.Fields {
			d.expr(f.Value)
		}
		if e.Spread != nil {
			d.expr(e.Spread)
		}
	case *ArrayLit:
		d.line(e.P, "array", nil, e.T)
		d.exprs(e.Elems)
	case *MatchExpr:
		d.line(e.P, "match", nil, e.T)
		d.expr(e.Subject)
		for _, arm := range e.Arms {
			d.pattern(arm.Pattern)
			if arm.Guard != nil {
				d.expr(arm.Guard)
			}
			if arm.Body != nil {
				d.expr(arm.Body)
			}
			if arm.Block != nil {
				d.block(arm.Block)
			}
		}
	case *ArrowExpr:
		var attrs []string
		if e.Closure != nil {
			if e.Closure.Stack {
				attrs = append(attrs, "stack")
			}
			if len(e.Closure.Captures) > 0 {
				names := make([]string, len(e.Closure.Captures))
				for i, l := range e.Closure.Captures {
					names[i] = l.Name
				}
				attrs = append(attrs, "captures="+strings.Join(names, ","))
			}
		}
		d.line(e.P, "arrow", attrs, e.T)
		d.params(e.Params)
		if e.Body != nil {
			d.expr(e.Body)
		}
		if e.Block != nil {
			d.block(e.Block)
		}
	default:
		panic(fmt.Sprintf("typesdump: unknown expression %T", e))
	}
}

var patternKinds = map[PatternKind]string{
	PatWildcard: "pat-wildcard", PatName: "pat-name", PatCtor: "pat-ctor",
	PatStruct: "pat-struct", PatTuple: "pat-tuple", PatLit: "pat-lit",
}

func (d *typesDumper) pattern(p *Pattern) {
	kind, ok := patternKinds[p.Kind]
	if !ok {
		panic(fmt.Sprintf("typesdump: unknown pattern kind %d", p.Kind))
	}
	d.line(p.Pos, kind, d.refAttrs(nil, nil, nil, p.Variant, ""), p.T)
	d.local(p.Local)
	for _, a := range p.Args {
		d.pattern(a)
	}
	for _, f := range p.Fields {
		d.pattern(f.Pattern)
	}
	if p.Lit != nil {
		d.expr(p.Lit)
	}
}
