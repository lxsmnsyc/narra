package langc

import "strings"

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
)

type Field struct {
	Name    string
	T       Type
	Default Expr
}

type StructType struct {
	Name   string
	Fields []*Field
}

func (t *StructType) String() string { return t.Name }
func (t *StructType) CName() string  { return "S_" + t.Name }

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
	Variants []*Variant
}

func (t *EnumType) String() string { return t.Name }
func (t *EnumType) CName() string  { return "E_" + t.Name }

func (t *EnumType) Variant(name string) *Variant {
	for _, v := range t.Variants {
		if v.Name == name {
			return v
		}
	}
	return nil
}

type ArrayType struct{ Elem Type }

func (t *ArrayType) String() string { return "Array<" + t.Elem.String() + ">" }
func (t *ArrayType) CName() string  { return "A_" + mangle(t.Elem) + "*" }

// Range values exist only as the subject of a for loop.
type RangeType struct{}

func (t *RangeType) String() string { return "Range" }
func (t *RangeType) CName() string  { return "void" }

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
func (t *FuncType) CName() string { return "void*" }

func mangle(t Type) string {
	switch t := t.(type) {
	case *PrimType:
		return t.Name
	case *StructType:
		return "S_" + t.Name
	case *EnumType:
		return "E_" + t.Name
	case *ArrayType:
		return "A_" + mangle(t.Elem)
	}
	return "x"
}

func sameType(a, b Type) bool {
	if a == b {
		return true
	}
	aa, ok1 := a.(*ArrayType)
	bb, ok2 := b.(*ArrayType)
	if ok1 && ok2 {
		return sameType(aa.Elem, bb.Elem)
	}
	return false
}

func isNumeric(t Type) bool { return t == TInt || t == TF64 }

// needsRC reports whether values of the type hold reference-counted data.
func needsRC(t Type) bool {
	switch t := t.(type) {
	case *PrimType:
		return t == TStr
	case *ArrayType:
		return true
	case *StructType:
		for _, f := range t.Fields {
			if needsRC(f.T) {
				return true
			}
		}
	case *EnumType:
		for _, v := range t.Variants {
			for _, f := range v.Fields {
				if needsRC(f.T) {
					return true
				}
			}
		}
	}
	return false
}

// Local is a variable or parameter.
type Local struct {
	Name    string
	T       Type
	Mutable bool // let, or a mutable parameter
	Param   bool
	ByRef   bool // a mutable parameter, passed as a C pointer
	CName   string
	Used    bool
	Pos     Pos
}
