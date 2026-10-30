package narra

import (
	"fmt"
	"strings"
)

// The AST dump is a fixed text form of a parsed file. `narra parse <file>`
// prints it, and the printer in selfhost/src/astdump.nra prints the same
// format, so the Go parser and the Narra parser can be diffed.
//
// Lines
//
//	A node is one line, indented by two spaces per depth:
//	    <kind> <line>:<col> <attr>...
//	A slot line names a group of children and has no position:
//	    .<slot> <attr>...
//	Children follow their node or slot line, one level deeper. The dump ends
//	with a newline. Node ids and everything the checker sets (types,
//	resolved names, Origin, Module) are not printed.
//
// Attributes
//
//	Attributes follow the position in the order listed below, separated by a
//	space. Flags print their bare name when true and nothing when false.
//	    name="text"     a string: escaped like the lex dump (`\` as `\\`,
//	                    newline, tab and CR as `\n \t \r`, other control
//	                    characters as `\u{<hex>}`), plus `"` as `\"`
//	    name=A,B        a list of identifiers joined with `,`, only when
//	                    non-empty
//	    end=<line>:<col> a position
//	String attributes always print, even when empty, except where noted.
//
// Slots and children
//
//	A slot prints only when it has a child (or, for lists, one element).
//	A node whose children are one list or one expression prints them
//	directly under its line, without a slot line. Those nodes say "direct"
//	below.
//
// Nodes (attributes in order; slots in order)
//
//	file                         .imports  .decls
//	import L:C                   alias="" names=a,b path=""      (alias and
//	                             names print only when set)
//	func L:C                     name="" export typeParams= effectParams=
//	                             uses=  .bound param="T" (repeated, only for
//	                             parameters with bounds; children are type
//	                             nodes)  .params  .result  .body
//	const L:C                    name="" export   .type  .value
//	struct L:C                   name="" export opaque typeParams=
//	                             .implements  .derives  .fields
//	enum L:C                     name="" export typeParams=
//	                             .implements  .derives  .variants
//	distinct L:C                 name="" export opaque   .base
//	test L:C                     name=""   .body
//	trait L:C                    name="" export   .funcs
//	impl L:C                     .trait  .type  .funcs
//	variant L:C                  name=""   direct: field nodes
//	field L:C                    name=""   .type  .default
//	param L:C                    name="" mutable escaping   .type
//	type L:C                     name="" (only when set) func tuple uses=
//	                             .args  .result
//	block L:C                    direct: statements
//	var L:C                      let name="" (name only when set)
//	                             .pattern  .type  .value
//	assign L:C                   .target  .value
//	exprstmt L:C                 direct: the expression
//	return L:C                   direct: the value, if any
//	if L:C                       .cond  .then  .else (a block or an if)
//	while L:C                    .cond  .body
//	for L:C                      name=""   .iter  .body
//	break L:C, continue L:C
//	int L:C                      text=""
//	float L:C                    text=""
//	string L:C                   value="" end=L:C
//	char L:C                     value="" (the character)
//	bool L:C                     value=true|false
//	template L:C                 direct: the parts, string pieces and
//	                             expressions in order
//	ident L:C                    name=""
//	field-expr L:C               name=""   direct: the target
//	is L:C                       .subject  .pattern
//	index L:C                    .target  .index
//	map L:C                      .key and .val, alternating, one pair per
//	                             entry
//	call L:C                     .callee  .args (each `arg [mutable]`, no
//	                             position, with the argument expression
//	                             directly under it)
//	unary L:C                    op=""     direct: the operand
//	binary L:C                   op=""     .left  .right
//	structlit L:C                type="" (only when set)  .spread  .fields
//	                             (each `init L:C name=""` with the value
//	                             directly under it)
//	array L:C                    tuple       direct: the elements
//	match L:C                    .subject  .arms
//	arm L:C                      .pattern  .guard  .body (an expression) or
//	                             .block
//	arrow L:C                    .params  .result  .body (an expression) or
//	                             .block
//	pat-wildcard L:C
//	pat-name L:C                 name=""
//	pat-ctor L:C                 enum="" (only when set) name="" bare
//	                             direct: the argument patterns. `bare` means
//	                             no parentheses: `Circle`, not `Circle()`.
//	pat-struct L:C               direct: `fieldpat L:C name=""` lines with
//	                             the pattern directly under each
//	pat-tuple L:C                direct: the element patterns
//	pat-lit L:C                  direct: the literal expression
//
// Positions
//
//	Lines and columns start at 1. Columns count characters (runes), not
//	bytes. Positions are those the Go parser records, which are not always
//	the first token of the node:
//	  - a declaration's pos is its keyword (`const`, `function`), not `export`
//	  - a call's pos is its `(`; an index's pos is its `[`
//	  - a field-expr's pos is the member name
//	  - a binary's pos is the operator; an `is` expression's pos is `is`
//	  - a struct literal with a type name has the name's pos
//	  - every part of a template (string pieces and the expressions inside
//	    `${...}`) takes the template's pos, and its string pieces carry the
//	    template's `end`
//	  - a shorthand pattern field `{ a }` makes a pat-name at the field's pos
//	`pat-ctor` prints `bare` when the Go parser left Args nil (`Circle`) and
//	not when it set an empty list (`Circle()`). That distinction relies on
//	the parser keeping nil and empty apart.
//
// Order in the file is source order. Nothing in the dump depends on map
// iteration order.
//
// Errors
//
//	When parsing fails the result is the single line
//	`error <line>:<col>: <message>` and ok is false; `narra parse` then
//	exits 1, as `narra lex` does.

// ASTDump parses src and returns the dump described above.
func ASTDump(src string) (dump string, ok bool) {
	f, err := Parse(src)
	if err != nil {
		return "error " + err.Error() + "\n", false
	}
	d := &astDumper{}
	d.file(f)
	return d.b.String(), true
}

type astDumper struct {
	b     strings.Builder
	depth int
}

func quoteDump(s string) string {
	return `"` + strings.ReplaceAll(escapeTokenText(s), `"`, `\"`) + `"`
}

// open prints a line and runs body one level deeper.
func (d *astDumper) open(text string, body func()) {
	for i := 0; i < d.depth; i++ {
		d.b.WriteString("  ")
	}
	d.b.WriteString(text)
	d.b.WriteByte('\n')
	if body != nil {
		d.depth++
		body()
		d.depth--
	}
}

// node prints a node line and runs body one level deeper.
func (d *astDumper) node(kind string, pos Pos, attrs []string, body func()) {
	text := kind + " " + pos.String()
	for _, a := range attrs {
		text += " " + a
	}
	d.open(text, body)
}

// slot prints a slot line and runs body one level deeper.
func (d *astDumper) slot(name string, attrs []string, body func()) {
	text := "." + name
	for _, a := range attrs {
		text += " " + a
	}
	d.open(text, body)
}

func flag(name string, on bool) []string {
	if on {
		return []string{name}
	}
	return nil
}

func strAttr(name, v string) []string { return []string{name + "=" + quoteDump(v)} }

func listAttr(name string, vs []string) []string {
	if len(vs) == 0 {
		return nil
	}
	return []string{name + "=" + strings.Join(vs, ",")}
}

func join(parts ...[]string) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func (d *astDumper) file(f *File) {
	d.open("file", func() {
		if len(f.Imports) > 0 {
			d.slot("imports", nil, func() {
				for _, i := range f.Imports {
					var alias []string
					if i.Alias != "" {
						alias = strAttr("alias", i.Alias)
					}
					d.node("import", i.Pos, join(alias, listAttr("names", i.Names), strAttr("path", i.Path)), nil)
				}
			})
		}
		if len(f.Decls) > 0 {
			d.slot("decls", nil, func() {
				for _, decl := range f.Decls {
					d.decl(decl)
				}
			})
		}
	})
}

func (d *astDumper) typeExprSlot(name string, attrs []string, ts []*TypeExpr) {
	if len(ts) == 0 {
		return
	}
	d.slot(name, attrs, func() {
		for _, t := range ts {
			d.typeExpr(t)
		}
	})
}

func (d *astDumper) oneType(slot string, t *TypeExpr) {
	if t != nil {
		d.slot(slot, nil, func() { d.typeExpr(t) })
	}
}

func (d *astDumper) typeExpr(t *TypeExpr) {
	var name []string
	if t.Name != "" {
		name = strAttr("name", t.Name)
	}
	d.node("type", t.Pos, join(name, flag("func", t.Func), flag("tuple", t.Tuple), listAttr("uses", t.Uses)), func() {
		d.typeExprSlot("args", nil, t.Args)
		d.oneType("result", t.Result)
	})
}

func (d *astDumper) params(ps []*Param) {
	if len(ps) == 0 {
		return
	}
	d.slot("params", nil, func() {
		for _, p := range ps {
			d.node("param", p.Pos, join(strAttr("name", p.Name), flag("mutable", p.Mutable), flag("escaping", p.Escaping)), func() {
				d.oneType("type", p.Type)
			})
		}
	})
}

func (d *astDumper) funcDecl(f *FuncDecl) {
	d.node("func", f.Pos, join(strAttr("name", f.Name), flag("export", f.Exported), listAttr("typeParams", f.TypeParams),
		listAttr("effectParams", f.EffectParams), listAttr("uses", f.Effects)), func() {
		for i, bs := range f.Bounds {
			if len(bs) > 0 && i < len(f.TypeParams) {
				d.typeExprSlot("bound", strAttr("param", f.TypeParams[i]), bs)
			}
		}
		d.params(f.Params)
		d.oneType("result", f.Result)
		d.blockSlot("body", f.Body)
	})
}

func (d *astDumper) funcs(fs []*FuncDecl) {
	if len(fs) == 0 {
		return
	}
	d.slot("funcs", nil, func() {
		for _, f := range fs {
			d.funcDecl(f)
		}
	})
}

func (d *astDumper) fieldDecls(fs []*FieldDecl) {
	for _, f := range fs {
		d.node("field", f.Pos, strAttr("name", f.Name), func() {
			d.oneType("type", f.Type)
			d.exprSlot("default", f.Default)
		})
	}
}

func (d *astDumper) decl(decl Decl) {
	switch n := decl.(type) {
	case *FuncDecl:
		d.funcDecl(n)
	case *ConstDecl:
		d.node("const", n.Pos, join(strAttr("name", n.Name), flag("export", n.Exported)), func() {
			d.oneType("type", n.Type)
			d.exprSlot("value", n.Value)
		})
	case *StructDecl:
		d.node("struct", n.Pos, join(strAttr("name", n.Name), flag("export", n.Exported), flag("opaque", n.Opaque), listAttr("typeParams", n.TypeParams)), func() {
			d.typeExprSlot("implements", nil, n.Implements)
			d.typeExprSlot("derives", nil, n.Derives)
			if len(n.Fields) > 0 {
				d.slot("fields", nil, func() { d.fieldDecls(n.Fields) })
			}
		})
	case *EnumDecl:
		d.node("enum", n.Pos, join(strAttr("name", n.Name), flag("export", n.Exported), listAttr("typeParams", n.TypeParams)), func() {
			d.typeExprSlot("implements", nil, n.Implements)
			d.typeExprSlot("derives", nil, n.Derives)
			if len(n.Variants) > 0 {
				d.slot("variants", nil, func() {
					for _, v := range n.Variants {
						d.node("variant", v.Pos, strAttr("name", v.Name), func() { d.fieldDecls(v.Fields) })
					}
				})
			}
		})
	case *DistinctDecl:
		d.node("distinct", n.Pos, join(strAttr("name", n.Name), flag("export", n.Exported), flag("opaque", n.Opaque)), func() {
			d.oneType("base", n.Base)
		})
	case *TestDecl:
		d.node("test", n.Pos, strAttr("name", n.Name), func() { d.blockSlot("body", n.Body) })
	case *TraitDecl:
		d.node("trait", n.Pos, join(strAttr("name", n.Name), flag("export", n.Exported)), func() { d.funcs(n.Funcs) })
	case *ImplDecl:
		d.node("impl", n.Pos, nil, func() {
			d.oneType("trait", n.Trait)
			d.oneType("type", n.Type)
			d.funcs(n.Funcs)
		})
	default:
		panic(fmt.Sprintf("astdump: unknown declaration %T", decl))
	}
}

func (d *astDumper) block(b *Block) {
	d.node("block", b.Pos, nil, func() {
		for _, s := range b.Stmts {
			d.stmt(s)
		}
	})
}

func (d *astDumper) exprSlot(name string, e Expr) {
	if e != nil {
		d.slot(name, nil, func() { d.expr(e) })
	}
}

func (d *astDumper) blockSlot(name string, b *Block) {
	if b != nil {
		d.slot(name, nil, func() { d.block(b) })
	}
}

func (d *astDumper) patSlot(name string, p *Pattern) {
	if p != nil {
		d.slot(name, nil, func() { d.pattern(p) })
	}
}

func (d *astDumper) stmt(s Stmt) {
	switch n := s.(type) {
	case *Block:
		d.block(n)
	case *VarStmt:
		var name []string
		if n.Name != "" {
			name = strAttr("name", n.Name)
		}
		d.node("var", n.Pos, join(flag("let", n.Mutable), name), func() {
			d.patSlot("pattern", n.Pattern)
			d.oneType("type", n.Type)
			d.exprSlot("value", n.Value)
		})
	case *AssignStmt:
		d.node("assign", n.Pos, nil, func() {
			d.exprSlot("target", n.Target)
			d.exprSlot("value", n.Value)
		})
	case *ExprStmt:
		d.node("exprstmt", n.Pos, nil, func() { d.expr(n.Expr) })
	case *ReturnStmt:
		d.node("return", n.Pos, nil, func() {
			if n.Value != nil {
				d.expr(n.Value)
			}
		})
	case *IfStmt:
		d.node("if", n.Pos, nil, func() {
			d.exprSlot("cond", n.Cond)
			d.blockSlot("then", n.Then)
			if n.Else != nil {
				d.slot("else", nil, func() { d.stmt(n.Else) })
			}
		})
	case *WhileStmt:
		d.node("while", n.Pos, nil, func() {
			d.exprSlot("cond", n.Cond)
			d.blockSlot("body", n.Body)
		})
	case *ForStmt:
		d.node("for", n.Pos, strAttr("name", n.Name), func() {
			d.exprSlot("iter", n.Iter)
			d.blockSlot("body", n.Body)
		})
	case *BreakStmt:
		d.node("break", n.Pos, nil, nil)
	case *ContinueStmt:
		d.node("continue", n.Pos, nil, nil)
	default:
		panic(fmt.Sprintf("astdump: unknown statement %T", s))
	}
}

func (d *astDumper) expr(e Expr) {
	pos := e.position()
	switch n := e.(type) {
	case *IntLit:
		d.node("int", pos, strAttr("text", n.Text), nil)
	case *FloatLit:
		d.node("float", pos, strAttr("text", n.Text), nil)
	case *StringLit:
		d.node("string", pos, join(strAttr("value", n.Value), []string{"end=" + n.End.String()}), nil)
	case *CharLit:
		d.node("char", pos, strAttr("value", string(n.Value)), nil)
	case *BoolLit:
		d.node("bool", pos, []string{fmt.Sprintf("value=%t", n.Value)}, nil)
	case *TemplateLit:
		d.node("template", pos, nil, func() {
			for _, p := range n.Parts {
				d.expr(p)
			}
		})
	case *Ident:
		d.node("ident", pos, strAttr("name", n.Name), nil)
	case *FieldExpr:
		d.node("field-expr", pos, strAttr("name", n.Name), func() { d.expr(n.Target) })
	case *IsExpr:
		d.node("is", pos, nil, func() {
			d.exprSlot("subject", n.Subject)
			d.patSlot("pattern", n.Pattern)
		})
	case *IndexExpr:
		d.node("index", pos, nil, func() {
			d.exprSlot("target", n.Target)
			d.exprSlot("index", n.Index)
		})
	case *MapLit:
		d.node("map", pos, nil, func() {
			for i := range n.Keys {
				d.exprSlot("key", n.Keys[i])
				d.exprSlot("val", n.Vals[i])
			}
		})
	case *CallExpr:
		d.node("call", pos, nil, func() {
			d.exprSlot("callee", n.Callee)
			if len(n.Args) > 0 {
				d.slot("args", nil, func() {
					for _, a := range n.Args {
						text := "arg"
						if a.Mutable {
							text += " mutable"
						}
						d.open(text, func() { d.expr(a.Value) })
					}
				})
			}
		})
	case *UnaryExpr:
		d.node("unary", pos, strAttr("op", n.Op), func() { d.expr(n.Operand) })
	case *BinaryExpr:
		d.node("binary", pos, strAttr("op", n.Op), func() {
			d.exprSlot("left", n.Left)
			d.exprSlot("right", n.Right)
		})
	case *StructLit:
		var tn []string
		if n.TypeName != "" {
			tn = strAttr("type", n.TypeName)
		}
		d.node("structlit", pos, tn, func() {
			d.exprSlot("spread", n.Spread)
			if len(n.Fields) > 0 {
				d.slot("fields", nil, func() {
					for _, f := range n.Fields {
						d.node("init", f.Pos, strAttr("name", f.Name), func() { d.expr(f.Value) })
					}
				})
			}
		})
	case *ArrayLit:
		d.node("array", pos, flag("tuple", n.IsTuple), func() {
			for _, x := range n.Elems {
				d.expr(x)
			}
		})
	case *MatchExpr:
		d.node("match", pos, nil, func() {
			d.exprSlot("subject", n.Subject)
			if len(n.Arms) > 0 {
				d.slot("arms", nil, func() {
					for _, arm := range n.Arms {
						d.node("arm", arm.Pos, nil, func() {
							d.patSlot("pattern", arm.Pattern)
							d.exprSlot("guard", arm.Guard)
							d.exprSlot("body", arm.Body)
							d.blockSlot("block", arm.Block)
						})
					}
				})
			}
		})
	case *ArrowExpr:
		d.node("arrow", pos, nil, func() {
			d.params(n.Params)
			d.oneType("result", n.Result)
			d.exprSlot("body", n.Body)
			d.blockSlot("block", n.Block)
		})
	default:
		panic(fmt.Sprintf("astdump: unknown expression %T", e))
	}
}

func (d *astDumper) pattern(p *Pattern) {
	switch p.Kind {
	case PatWildcard:
		d.node("pat-wildcard", p.Pos, nil, nil)
	case PatName:
		d.node("pat-name", p.Pos, strAttr("name", p.Name), nil)
	case PatCtor:
		var enum []string
		if p.Enum != "" {
			enum = strAttr("enum", p.Enum)
		}
		d.node("pat-ctor", p.Pos, join(enum, strAttr("name", p.Name), flag("bare", p.Args == nil)), func() {
			for _, a := range p.Args {
				d.pattern(a)
			}
		})
	case PatStruct:
		d.node("pat-struct", p.Pos, nil, func() {
			for _, f := range p.Fields {
				d.node("fieldpat", f.Pos, strAttr("name", f.Name), func() { d.pattern(f.Pattern) })
			}
		})
	case PatTuple:
		d.node("pat-tuple", p.Pos, nil, func() {
			for _, a := range p.Args {
				d.pattern(a)
			}
		})
	case PatLit:
		d.node("pat-lit", p.Pos, nil, func() { d.expr(p.Lit) })
	default:
		panic(fmt.Sprintf("astdump: unknown pattern kind %d", p.Kind))
	}
}
