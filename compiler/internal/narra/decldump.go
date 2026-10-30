package narra

import (
	"sort"
	"strings"
)

// The declaration dump is a fixed text form of what the declaration passes of
// the checker decide. `narra decls <path>` prints it, and
// selfhost/src/declsdump.nra (`narrac decls <path>`) prints the same format,
// so the Go checker and the Narra checker can be diffed.
//
// Input
//
//	path is a package directory or a single .nra file, as Load accepts. The
//	dump runs newPackageChecker (the prelude) and declareSourceModule for
//	every module in load order, then resolves the declared type of each
//	annotated constant (module by module, in declaration order), then
//	finish. Bodies are not checked.
//
// Lines
//
//	Each line starts with its kind. Attributes follow; a type, printed as
//	Type.String() prints it, always comes last, because it can hold spaces.
//	Sub-lines are indented by two spaces per level.
//	    type struct module="<m>" [opaque] <type>
//	      field <name> [default] <type>
//	    type enum module="<m>" <type>
//	      variant <name>
//	        field <name> <type>
//	    type distinct module="<m>" [opaque] <name>
//	      base <type>
//	    type tuple module="" <type>
//	      field <index> <type>
//	    generic struct module="<m>" params=<T>,<U> [opaque] <name>
//	      field ... (as for type struct; types may name the parameters)
//	    generic enum module="<m>" params=<T> <name>
//	      variant ... (as for type enum)
//	    func module="<m>" [typeParams=<T>,<U>] <name>
//	      param <name> [mutable] <type>
//	      result <type>
//	    const module="<m>" <name> <type>|?
//	    order module="<m>" <type>
//
// Rules
//
//   - Modules come in order: the prelude (module "") first, then the
//     package modules in load order. Inside a module, struct, enum,
//     distinct, generic, func and const lines follow the declarations in
//     source order. A non-generic struct or enum is a `type` line; a generic
//     one is a `generic` line. Traits, implementations and tests print
//     nothing here.
//   - `default` marks a field that has a default value. `opaque` marks an
//     opaque struct or distinct type (and the instances of an opaque generic
//     struct).
//   - A const prints its declared type, resolved after all modules are
//     declared, or `?` when it has none (the body checker infers it).
//   - After the modules come the generic instances and tuple types created
//     while declaring, as `type struct`, `type enum` and `type tuple`
//     blocks, without those that still hold a type parameter. The blocks are
//     sorted by their first line (byte order, stable).
//   - Last come the `order` lines: Program.Types after finish, the order in
//     which C defines the structs and enums.
//
// Errors
//
//	A failed load or check prints only the errors, one line each in report
//	order, as FormatError formats them, with the prefix "error ":
//	    error <file>:<line>:<col>: <message>
//	and the command exits 1. File is "prelude" for the prelude.
//
// Owners
//
//	Each line kind belongs to one roadmap node; a later node appends its own
//	kinds and documents them here. selfhost-check-decls: type, field,
//	variant, base, generic, func, param, result, const, order, error.
//	selfhost-check-traits adds trait, impl and bound; selfhost-check-imports
//	adds module, import and builtin.

// DeclDump loads the package at path, runs the declaration passes and returns
// the dump described above.
func DeclDump(path string) (dump string, ok bool) {
	pkg, err := Load(path)
	if err != nil {
		return errorLines(err), false
	}
	c, concrete, err := newPackageChecker(pkg)
	if err != nil {
		return errorLines(err), false
	}
	for _, m := range pkg.Modules {
		concrete = append(concrete, c.declareSourceModule(m)...)
	}
	constTypes := map[*ConstDecl]Type{}
	for _, m := range pkg.Modules {
		c.enter(c.mods[m.Path])
		for _, f := range m.Files {
			for _, d := range f.Decls {
				if cd, ok := d.(*ConstDecl); ok && cd.Type != nil {
					constTypes[cd] = c.resolveType(cd.Type)
				}
			}
		}
	}
	c.finish(concrete)
	if len(c.errs) > 0 {
		return errorLines(Diagnostics(c.errs)), false
	}
	var b strings.Builder
	c.dumpModule(&b, c.mods[""], c.prelude, constTypes)
	for _, m := range pkg.Modules {
		for _, f := range m.Files {
			c.dumpModule(&b, c.mods[m.Path], f, constTypes)
		}
	}
	var blocks []string
	for _, t := range c.instTypes {
		if !containsTypeParam(t) {
			var ib strings.Builder
			dumpType(&ib, t)
			blocks = append(blocks, ib.String())
		}
	}
	sort.SliceStable(blocks, func(i, j int) bool { return firstLine(blocks[i]) < firstLine(blocks[j]) })
	for _, s := range blocks {
		b.WriteString(s)
	}
	for _, t := range c.prog.Types {
		b.WriteString("order module=" + quoteDump(typeModule(t)) + " " + t.String() + "\n")
	}
	return b.String(), true
}

func errorLines(err error) string {
	var b strings.Builder
	for _, line := range strings.Split(FormatError(err).Error(), "\n") {
		b.WriteString("error " + line + "\n")
	}
	return b.String()
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func typeModule(t Type) string {
	switch t := t.(type) {
	case *StructType:
		return t.Module
	case *EnumType:
		return t.Module
	case *DistinctType:
		return t.Module
	}
	return ""
}

// dumpModule prints the declarations of one file of module s.
func (c *Checker) dumpModule(b *strings.Builder, s *modScope, f *File, constTypes map[*ConstDecl]Type) {
	mod := " module=" + quoteDump(s.path) + " "
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl, *EnumDecl:
			name := declName(d)
			if g := s.generics[name]; g != nil {
				dumpGeneric(b, g)
			} else {
				dumpType(b, s.types[name])
			}
		case *DistinctDecl:
			dumpType(b, d.T)
		case *FuncDecl:
			b.WriteString("func" + mod)
			if len(d.TypeParams) > 0 {
				b.WriteString("typeParams=" + strings.Join(d.TypeParams, ",") + " ")
			}
			b.WriteString(d.Name + "\n")
			for _, p := range d.Params {
				b.WriteString("  param " + p.Name + " ")
				if p.Mutable {
					b.WriteString("mutable ")
				}
				b.WriteString(p.T.String() + "\n")
			}
			b.WriteString("  result " + d.ResultT.String() + "\n")
		case *ConstDecl:
			t := "?"
			if ct, ok := constTypes[d]; ok {
				t = ct.String()
			}
			b.WriteString("const" + mod + d.Name + " " + t + "\n")
		}
	}
}

func dumpType(b *strings.Builder, t Type) {
	switch t := t.(type) {
	case *StructType:
		if t.Tuple {
			b.WriteString("type tuple module=\"\" " + t.String() + "\n")
		} else {
			b.WriteString("type struct module=" + quoteDump(t.Module) + " " + opaqueFlag(t.Opaque) + t.String() + "\n")
		}
		dumpFields(b, "  ", t.Fields)
	case *EnumType:
		b.WriteString("type enum module=" + quoteDump(t.Module) + " " + t.String() + "\n")
		dumpVariants(b, t.Variants)
	case *DistinctType:
		b.WriteString("type distinct module=" + quoteDump(t.Module) + " " + opaqueFlag(t.Opaque) + t.Name + "\n")
		b.WriteString("  base " + t.Base.String() + "\n")
	}
}

func dumpGeneric(b *strings.Builder, g *genericType) {
	names := make([]string, len(g.params))
	for i, p := range g.params {
		names[i] = p.Name
	}
	kind, opaque := "enum", false
	if g.sdecl != nil {
		kind, opaque = "struct", g.sdecl.Opaque
	}
	b.WriteString("generic " + kind + " module=" + quoteDump(g.mod.path) + " params=" + strings.Join(names, ",") + " " + opaqueFlag(opaque) + g.name + "\n")
	dumpFields(b, "  ", g.fields)
	dumpVariants(b, g.variants)
}

func dumpVariants(b *strings.Builder, vs []*Variant) {
	for _, v := range vs {
		b.WriteString("  variant " + v.Name + "\n")
		dumpFields(b, "    ", v.Fields)
	}
}

func dumpFields(b *strings.Builder, indent string, fields []*Field) {
	for _, f := range fields {
		b.WriteString(indent + "field " + f.Name + " ")
		if f.Default != nil {
			b.WriteString("default ")
		}
		b.WriteString(f.T.String() + "\n")
	}
}

func opaqueFlag(on bool) string {
	if on {
		return "opaque "
	}
	return ""
}
