package narra

// Syntax tree. The checker fills in the Type fields of expressions and the
// resolved declarations of names.

type File struct {
	Path    string // source path relative to the package directory, for diagnostics
	Imports []*ImportDecl
	Decls   []Decl
}

type ImportDecl struct {
	Pos   Pos
	Alias string // import * as Alias from "path"
	Names []string
	Path  string
}

type Decl interface{ declNode() }

type TypeExpr struct {
	Pos  Pos
	Name string
	Args []*TypeExpr
	// A function type `(a: A, b: B) => R` sets Func; Args are the parameter
	// types and Result the result type.
	Func   bool
	Result *TypeExpr
}

type Param struct {
	Pos     Pos
	Name    string
	Mutable bool
	Type    *TypeExpr
	T       Type
	Local   *Local
	// Escaping marks a function-type parameter that may be stored,
	// returned or captured (`escaping f: () => void`).
	Escaping bool
}

type FuncDecl struct {
	Pos        Pos
	Name       string
	TypeParams []string
	Bounds     [][]*TypeExpr // trait constraints, one list per type parameter
	Exported   bool
	Params     []*Param
	Result     *TypeExpr
	Effects    []string
	Body       *Block
	ResultT    Type
	Module     string // declaring module path, set by the checker
}

type FieldDecl struct {
	Pos     Pos
	Name    string
	Type    *TypeExpr
	Default Expr
}

type StructDecl struct {
	Pos        Pos
	Name       string
	TypeParams []string
	Exported   bool
	Implements []*TypeExpr
	Derives    []*TypeExpr
	Fields     []*FieldDecl
	T          *StructType
}

type VariantDecl struct {
	Pos    Pos
	Name   string
	Fields []*FieldDecl
}

type EnumDecl struct {
	Pos        Pos
	Name       string
	TypeParams []string
	Exported   bool
	Implements []*TypeExpr
	Derives    []*TypeExpr
	Variants   []*VariantDecl
	T          *EnumType
}

type TestDecl struct {
	Pos    Pos
	Name   string
	Body   *Block
	Module string // declaring module path, set by the checker
}

// TraitDecl is `trait Name { function f(a: Self): R ... }`. Funcs have no body.
type TraitDecl struct {
	Pos      Pos
	Name     string
	Exported bool
	Funcs    []*FuncDecl
	T        *Trait
}

// ImplDecl is `implement Trait for Type { ... }`.
type ImplDecl struct {
	Pos   Pos
	Trait *TypeExpr
	Type  *TypeExpr
	Funcs []*FuncDecl
}

func (*FuncDecl) declNode()   {}
func (*StructDecl) declNode() {}
func (*EnumDecl) declNode()   {}
func (*TestDecl) declNode()   {}
func (*TraitDecl) declNode()  {}
func (*ImplDecl) declNode()   {}

type Stmt interface{ stmtNode() }

type Block struct {
	Pos   Pos
	Stmts []Stmt
}

type VarStmt struct {
	Pos     Pos
	Mutable bool // let
	Name    string
	Type    *TypeExpr
	Value   Expr
	Local   *Local
}

type AssignStmt struct {
	Pos    Pos
	Target Expr
	Value  Expr
}

type ExprStmt struct {
	Pos  Pos
	Expr Expr
}

type ReturnStmt struct {
	Pos   Pos
	Value Expr
}

type IfStmt struct {
	Pos  Pos
	Cond Expr
	Then *Block
	Else Stmt // *Block or *IfStmt, or nil
}

type WhileStmt struct {
	Pos  Pos
	Cond Expr
	Body *Block
}

type ForStmt struct {
	Pos   Pos
	Name  string
	Iter  Expr
	Body  *Block
	Local *Local
	Kind  forKind
}

type BreakStmt struct{ Pos Pos }
type ContinueStmt struct{ Pos Pos }

func (*Block) stmtNode()        {}
func (*VarStmt) stmtNode()      {}
func (*AssignStmt) stmtNode()   {}
func (*ExprStmt) stmtNode()     {}
func (*ReturnStmt) stmtNode()   {}
func (*IfStmt) stmtNode()       {}
func (*WhileStmt) stmtNode()    {}
func (*ForStmt) stmtNode()      {}
func (*BreakStmt) stmtNode()    {}
func (*ContinueStmt) stmtNode() {}

type Expr interface {
	exprNode()
	Type() Type
	setType(Type)
	position() Pos
}

type exprBase struct {
	P Pos
	T Type
}

func (e *exprBase) Type() Type     { return e.T }
func (e *exprBase) setType(t Type) { e.T = t }
func (e *exprBase) position() Pos  { return e.P }
func (e *exprBase) exprNode()      {}

type IntLit struct {
	exprBase
	Text string
}

type FloatLit struct {
	exprBase
	Text string
}

type StringLit struct {
	exprBase
	Value string
}

type CharLit struct {
	exprBase
	Value rune
}

type BoolLit struct {
	exprBase
	Value bool
}

type TemplateLit struct {
	exprBase
	Parts []Expr // StringLit pieces and string expressions, in order
}

type Ident struct {
	exprBase
	Name  string
	Local *Local    // set when the name is a local or parameter
	Func  *FuncDecl // set when the name is a function
	Ctor  *Variant  // set when the name is an enum constructor without fields
}

type FieldExpr struct {
	exprBase
	Target Expr
	Name   string
	// Set by the checker when Target names a module or an enum:
	Builtin string   // "Int.toString" and other standard functions
	Ctor    *Variant // Shape.Circle
	// Set when Target is an enum value narrowed to one variant by `is`.
	Narrow *Variant
	Func   *FuncDecl // Alias.f used as a function value
}

// IsExpr is `subject is Pattern`.
type IsExpr struct {
	exprBase
	Subject Expr
	Pattern *Pattern
}

type IndexExpr struct {
	exprBase
	Target Expr
	Index  Expr
}

type Arg struct {
	Mutable bool
	Value   Expr
}

type CallExpr struct {
	exprBase
	Callee  Expr
	Args    []*Arg
	Builtin string // set for standard functions
	Func    *FuncDecl
	Ctor    *Variant
	Trait   *TraitCall // set for a trait function the compiler implements
	// FuncValue is set when the callee is a function value, such as a
	// parameter, a local or a struct field of function type.
	FuncValue *FuncType
}

type UnaryExpr struct {
	exprBase
	Op      string
	Operand Expr
}

type BinaryExpr struct {
	exprBase
	Op          string
	Left, Right Expr
}

type FieldInit struct {
	Pos   Pos
	Name  string
	Value Expr
}

type StructLit struct {
	exprBase
	TypeName string // empty when the type comes from context
	Fields   []*FieldInit
	Struct   *StructType
}

type ArrayLit struct {
	exprBase
	Elems []Expr
}

type Pattern struct {
	Pos      Pos
	Wildcard bool
	Name     string   // constructor name
	Binds    []string // one per constructor field, "_" to skip
	Variant  *Variant
	Locals   []*Local
}

type MatchArm struct {
	Pos     Pos
	Pattern *Pattern
	Guard   Expr
	Body    Expr   // set when the arm is a single expression
	Block   *Block // set when the arm is a block
}

type MatchExpr struct {
	exprBase
	Subject Expr
	Arms    []*MatchArm
	IsValue bool
}

// ArrowExpr is an arrow function, `(x: int): int => x + 1` or
// `(x) => { ... }`. See check_closure.go and emit_closure.go.
type ArrowExpr struct {
	exprBase
	Params  []*Param  // Type is nil when it comes from the expected type
	Result  *TypeExpr // nil when not written
	Body    Expr      // expression body, or nil
	Block   *Block    // block body, or nil
	Closure *Closure  // filled by the checker
}
