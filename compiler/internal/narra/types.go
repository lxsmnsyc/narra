package narra

import (
	"strings"
)

type Type interface {
	String() string
	// CName is the C type used for values of this type.
	CName() string
}

type PrimType struct{ Name string }

func (t *PrimType) String() string { return t.Name }
func (t *PrimType) CName() string {
	switch t.Name {
	case "int":
		return "int64_t"
	case "f64":
		return "double"
	case "bool":
		return "bool"
	case "char":
		return "uint32_t"
	case "string":
		return "lstr*"
	case "StringBuilder":
		return "narra_sb*"
	case "void":
		return "void"
	}
	return "void"
}

var (
	TInt  = &PrimType{"int"}
	TF64  = &PrimType{"f64"}
	TBool = &PrimType{"bool"}
	TChar = &PrimType{"char"}
	TStr  = &PrimType{"string"}
	TVoid = &PrimType{"void"}
	// TStringBuilder is the type of std/string's builder: a reference
	// counted buffer. It implements no standard trait.
	TStringBuilder = &PrimType{"StringBuilder"}
)

type Field struct {
	Name    string
	T       Type
	Default Expr
}

type StructType struct {
	Name   string
	Module string // declaring module path; "" for the prelude
	Fields []*Field
	Args   []Type // type arguments of a generic instance
	Tuple  bool   // Tuple<A, B>: fields "0", "1", ... and Args are the element types
	Opaque bool   // declared `opaque`: only the declaring module sees the fields
}

func (t *StructType) String() string {
	if t.Tuple {
		s := typeArgsString(t.Args)
		return "#[" + s[1:len(s)-1] + "]"
	}
	return t.Name + typeArgsString(t.Args)
}

func (t *StructType) CName() string { return structCName(t) }

func (t *StructType) Field(name string) *Field {
	for _, f := range t.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

type Variant struct {
	Name   string
	Index  int
	Fields []*Field
	Enum   *EnumType
}

type EnumType struct {
	Name     string
	Module   string // declaring module path; "" for the prelude
	Variants []*Variant
	Args     []Type // type arguments of a generic instance
}

func (t *EnumType) String() string { return t.Name + typeArgsString(t.Args) }
func (t *EnumType) CName() string  { return enumCName(t) }

func typeArgsString(args []Type) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.String()
	}
	return "<" + strings.Join(parts, ", ") + ">"
}

// TypeParamType stands for a type parameter while a generic declaration is
// checked on its own.
type TypeParamType struct {
	Name string
	ID   int // distinguishes parameters with the same name in different declarations
}

func (t *TypeParamType) String() string { return t.Name }
func (t *TypeParamType) CName() string  { return "void*" }

func containsTypeParam(t Type) bool {
	switch t := t.(type) {
	case *TypeParamType:
		return true
	case *FuncType:
		for _, p := range t.Params {
			if containsTypeParam(p) {
				return true
			}
		}
		return containsTypeParam(t.Result)
	case *ArrayType:
		return containsTypeParam(t.Elem)
	case *GeneratorType:
		return containsTypeParam(t.Elem)
	case *BoxType:
		return containsTypeParam(t.Elem)
	case *MapType:
		return containsTypeParam(t.Key) || containsTypeParam(t.Val)
	case *StructType:
		for _, a := range t.Args {
			if containsTypeParam(a) {
				return true
			}
		}
	case *EnumType:
		for _, a := range t.Args {
			if containsTypeParam(a) {
				return true
			}
		}
	}
	return false
}

func (t *EnumType) Variant(name string) *Variant {
	for _, v := range t.Variants {
		if v.Name == name {
			return v
		}
	}
	return nil
}

type ArrayType struct{ Elem Type }

// BoxType is Box<T>: one value on the heap with a reference count
// (emit_box.go). Box.new makes one and .value reads it.
type BoxType struct{ Elem Type }

func (t *BoxType) String() string { return "Box<" + t.Elem.String() + ">" }
func (t *BoxType) CName() string  { return boxCName(t) + "*" }

// isBoxedEnum reports whether the enum keeps its variant fields on the heap.
// An enum with no fields stays an inline tag.
func isBoxedEnum(t *EnumType) bool {
	for _, v := range t.Variants {
		if len(v.Fields) > 0 {
			return true
		}
	}
	return false
}

// MapType is Map<K, V>: a hash map that keeps insertion order (std_map.go).
type MapType struct{ Key, Val Type }

func (t *MapType) String() string { return "Map<" + t.Key.String() + ", " + t.Val.String() + ">" }
func (t *MapType) CName() string  { return mapCName(t) + "*" }

func (t *ArrayType) String() string { return "Array<" + t.Elem.String() + ">" }
func (t *ArrayType) CName() string  { return arrayCName(t) + "*" }

// RangeType is Range: the bounds of a counting loop, a plain value of two ints
// (narra_range, defined in the std/range runtime). Use the TRange singleton.
type RangeType struct{}

// TRange is the one Range type.
var TRange = &RangeType{}

func (t *RangeType) String() string { return "Range" }
func (t *RangeType) CName() string  { return "narra_range" }

// GeneratorType is Generator<T>: the frame of a function* call, a value
// (SPEC.md section 8, D78). Effects are the effects its body has when it
// runs, sorted, in the form of a `uses` list (effect variables included).
// They are charged where it runs: a for ... of over it and a yield* of it.
// Binder is the function whose effect variables Effects names, or nil.
// Origin is set for the result type of a private function*, whose effects
// are inferred: the effect pass fills Origin.GenEffects, and Effects is
// unused. Bound is set for the result of a call whose callee's result type
// names the callee's effect variables: the effect pass fills it with the
// effects the arguments bind (bindCallResult), and Effects is unused.
type GeneratorType struct {
	Elem    Type
	Effects []string
	Binder  *FuncDecl
	Origin  *FuncDecl
	Bound   *boundEffects
}

// boundEffects are the effects of one call's generator result, filled by the
// effect pass at call, the call that made the type, and at no other call
// (a generic `id(x)` returns the same type). id tells calls apart in typeID.
type boundEffects struct {
	id      int
	call    *CallExpr
	effects []string
}

// effects is the effect list of the type: the inferred one of a private
// generator or of a call result once the effect pass has run.
func (t *GeneratorType) effects() []string {
	if t.Origin != nil {
		return t.Origin.GenEffects
	}
	if t.Bound != nil {
		return t.Bound.effects
	}
	return t.Effects
}

func (t *GeneratorType) String() string {
	return "Generator<" + t.Elem.String() + ">" + usesSuffix(t.effects())
}

// CName is the C type of every generator value: a pointer to the header of
// its frame (emit_generator.go). The element type is not part of it.
func (t *GeneratorType) CName() string { return "narra_gen*" }

// usesSuffix writes an effect list as it follows a type, ` uses a, b`.
func usesSuffix(effects []string) string {
	if len(effects) == 0 {
		return ""
	}
	return " uses " + strings.Join(effects, ", ")
}

// ModuleType is the type of a namespace import such as `Int`.
type ModuleType struct{ Path string }

func (t *ModuleType) String() string { return "module " + t.Path }
func (t *ModuleType) CName() string  { return "void" }

type FuncType struct {
	Params []Type
	Result Type
}

func (t *FuncType) String() string {
	parts := make([]string, len(t.Params))
	for i, p := range t.Params {
		parts[i] = p.String()
	}
	return "(" + strings.Join(parts, ", ") + ") => " + t.Result.String()
}

// CName is the struct that holds a function value: a C function pointer
// and its environment (emit_closure.go).
func (t *FuncType) CName() string { return funcTypeCName(t) }

func sameType(a, b Type) bool {
	if a == b {
		return true
	}
	aa, ok1 := a.(*ArrayType)
	bb, ok2 := b.(*ArrayType)
	if ok1 && ok2 {
		return sameType(aa.Elem, bb.Elem)
	}
	ma, ok1 := a.(*MapType)
	mb, ok2 := b.(*MapType)
	if ok1 && ok2 {
		return sameType(ma.Key, mb.Key) && sameType(ma.Val, mb.Val)
	}
	ga, ok1 := a.(*GeneratorType)
	gb, ok2 := b.(*GeneratorType)
	if ok1 && ok2 {
		return sameType(ga.Elem, gb.Elem) && ga.Origin == gb.Origin && ga.Bound == gb.Bound && ga.Binder == gb.Binder && strings.Join(ga.Effects, ",") == strings.Join(gb.Effects, ",")
	}
	fa, ok1 := a.(*FuncType)
	fb, ok2 := b.(*FuncType)
	if ok1 && ok2 && len(fa.Params) == len(fb.Params) && sameType(fa.Result, fb.Result) {
		for i := range fa.Params {
			if !sameType(fa.Params[i], fb.Params[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func isNumeric(t Type) bool { return t == TInt || t == TF64 }

// needsRC reports whether values of the type hold reference-counted data.
func needsRC(t Type) bool {
	switch t := t.(type) {
	case *PrimType:
		return t == TStr || t == TStringBuilder
	case *ArrayType, *FuncType, *BoxType, *MapType, *GeneratorType:
		return true
	case *StructType:
		for _, f := range t.Fields {
			if needsRC(f.T) {
				return true
			}
		}
	case *EnumType:
		return isBoxedEnum(t)
	case *DistinctType:
		return needsRC(t.Base)
	}
	return false
}

// Local is a variable or parameter.
type Local struct {
	Name      string
	T         Type
	Mutable   bool // let, or a mutable parameter
	Param     bool
	ByRef     bool // a mutable parameter, passed as a C pointer
	CName     string
	Used      bool
	Pos       Pos
	NamePos   Pos   // where the name is written; Pos is the statement for a const, let or for
	Shorthand bool  // a struct pattern field written without `:`
	Writes    []Pos // assignments `x = v` to the local, which do not read it
	Exempt    bool  // a parameter whose signature is fixed (the entry function)
	// NoEscape marks a function-type parameter without `escaping`: it can
	// only be called or passed to another non-escaping parameter.
	NoEscape bool
}
