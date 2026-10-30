package narra

import (
	"reflect"
	"runtime"
	"sort"
	"strconv"
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
//	annotated constant (module by module, in declaration order), checks the
//	Map keys and the entry function of each binary (entryFunc, checkEntry),
//	then finish. Bodies are not checked.
//
// Lines
//
//	Each line starts with its kind. Attributes follow; a type always comes
//	last, because it can hold spaces. A type is printed as Type.String()
//	prints it, except that a struct, enum or distinct type of a package
//	module is written <module>.<Name> (app/text.Token, Pair<app/a.Token>),
//	so two modules that declare the same name dump apart. Prelude types
//	(Option, Ordering, tuples) keep their plain name. Sub-lines are indented
//	by two spaces per level.
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
//	      bound <T> <Trait>
//	    const module="<m>" <name> <type>|?
//	    order module="<m>" <type>
//	    trait module="<m>" [std] <name>
//	      function <name>
//	        param <name> [mutable] [escaping] <type>
//	        result <type>
//	    impl module="<m>" listed <Trait> <name>
//	    impl module="<m>" derived <Trait> <name>
//	    impl module="<m>" block <Trait> <name>
//	      function <name>
//	        param ... and result ... (as for a trait)
//	    impl module="<m>" holds=<true|false> <Trait> <type>
//	    module "<m>"
//	      import "<path>" package|std [alias=<A>] [names=<a>,<b>]
//	      scope type <name> <type>
//	      scope generic <name> <module>.<name>
//	      scope func <name> <module>.<name>
//	      scope const <name> <module>.<name>
//	      scope trait <name> <module>.<name>
//	      scope builtin <name> <Builtin.Name>
//	      scope alias <A> "<path>"
//	    entry "<binary>" <module>.<function>|none
//
// Rules
//
//   - Modules come in order: the prelude (module "") first, then the
//     package modules in load order. Inside a module, struct, enum,
//     distinct, generic, func and const lines follow the declarations in
//     source order. A non-generic struct or enum is a `type` line; a generic
//     one is a `generic` line. Tests print nothing here.
//   - A `bound` line sits under the `func` of a generic function, one per
//     constraint, in the order they are written: the type parameter, then
//     the trait name.
//   - A `trait` line sits where the trait is declared. Its `function`
//     lines show the signature with Self printed as "Self". `std` marks the
//     prelude's Equatable, Hashable, Ordered and Show.
//   - Right after a struct or enum (its `type` or `generic` block) come one
//     `impl ... listed` line per trait of its `implements` list and one
//     `impl ... derived` line per trait of its `derives` list, as written.
//     An `implement` block prints `impl ... block` where it is declared,
//     with the functions it defines (the names written in the source).
//   - After the `order` lines come the `holds` lines: for each type of the
//     order, in that order, one line for each of Equatable, Hashable,
//     Ordered and Show with the answer of implements(type, trait).
//   - `default` marks a field that has a default value. `opaque` marks an
//     opaque struct or distinct type (and the instances of an opaque generic
//     struct).
//   - A const prints its declared type, resolved after all modules are
//     declared, or `?` when it has none (the body checker infers it).
//   - After the modules come the generic instances and tuple types created
//     while declaring, as `type struct`, `type enum` and `type tuple`
//     blocks, without those that still hold a type parameter. The blocks are
//     sorted by their first line (byte order, stable).
//   - A `module` block follows the declarations of each package module (not
//     the prelude). Its `import` lines list the module's imports in source
//     order: `package` for a module of the package, `std` for a registered
//     std module. Its `scope` lines are the names the module sees, by kind
//     in the order above and by name (byte order) within a kind: own and
//     imported types (without the primitives and the prelude's types, which
//     every module sees), generics, functions, constants, traits (own and
//     imported; the prelude's are not listed), std functions imported by
//     name, and namespace aliases. An imported name shows the module that
//     declares it.
//   - One `entry` line per [[binary]] of narra.toml follows the last module:
//     the exported function its entry names, or `none` when there is none
//     (CheckPackage leaves such a binary without an entry; the build
//     reports it).
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
// Builtins
//
//	BuiltinDump prints the std registry in the same style, for
//	TestSelfhostBuiltins (`narrac builtins` prints the Narra registry). The
//	prelude functions (module "") come first, then each std module in the
//	order moduleList names them; inside a module, functions by export name.
//	    builtin module="<path>" export=<name> <Builtin.Name>
//	      param [mutable] <type>      one per Builtin.Params entry
//	      mutable <bool>,<bool>       Builtin.Mutable when Params is nil
//	      result <type>               when Builtin.Result is set
//	      effects <e>,<e>             when there are effects
//	      callbacks <i>,<j>           Builtin.Callbacks
//	      check <name>                when Builtin.Check is set
//	The check name is the Go function (checkArrayPush), or `fixed` for an
//	inline closure; every inline closure calls checkFixed. The Narra dump
//	also prints the signature of a `fixed` builtin as `fixed param <type>`
//	and `fixed result <type>` lines, which the test checks by compiling a
//	call and then removes.
//
// Owners
//
//	Each line kind belongs to one roadmap node; a later node appends its own
//	kinds and documents them here. selfhost-check-decls: type, field,
//	variant, base, generic, func, param, result, const, order, error.
//	selfhost-check-traits: trait, impl, bound, and function (the signature
//	line inside trait and impl). selfhost-check-imports: module, import,
//	scope, entry and builtin, and the module in type names.

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
	c.flushMapKeys()
	// The entry functions, as CheckPackage finds and checks them.
	entries := make([]*FuncDecl, len(pkg.Binaries))
	for i, bin := range pkg.Binaries {
		if fd := c.entryFunc(bin); fd != nil {
			saved := c.enter(c.funcMod[fd])
			c.checkEntry(fd)
			c.enter(saved)
			entries[i] = fd
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
		c.dumpScope(&b, c.mods[m.Path])
	}
	for i, bin := range pkg.Binaries {
		target := "none"
		if fd := entries[i]; fd != nil {
			target = fd.Module + "." + fd.Name
		}
		b.WriteString("entry " + quoteDump(bin.Name) + " " + target + "\n")
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
		b.WriteString("order module=" + quoteDump(typeModule(t)) + " " + dumpTypeString(t) + "\n")
	}
	for _, t := range c.prog.Types {
		c.dumpHolds(&b, t)
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
			c.dumpListed(b, mod, s.path, name)
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
				b.WriteString(dumpTypeString(p.T) + "\n")
			}
			b.WriteString("  result " + dumpTypeString(d.ResultT) + "\n")
			for i, tp := range c.funcTP[d] {
				for _, tr := range c.traits.bounds[tp] {
					b.WriteString("  bound " + d.TypeParams[i] + " " + tr.Name + "\n")
				}
			}
		case *TraitDecl:
			dumpTrait(b, mod, d.T)
		case *ImplDecl:
			c.dumpImpl(b, mod, d)
		case *ConstDecl:
			t := "?"
			if ct, ok := constTypes[d]; ok {
				t = dumpTypeString(ct)
			}
			b.WriteString("const" + mod + d.Name + " " + t + "\n")
		}
	}
}

func dumpType(b *strings.Builder, t Type) {
	switch t := t.(type) {
	case *StructType:
		if t.Tuple {
			b.WriteString("type tuple module=\"\" " + dumpTypeString(t) + "\n")
		} else {
			b.WriteString("type struct module=" + quoteDump(t.Module) + " " + opaqueFlag(t.Opaque) + dumpTypeString(t) + "\n")
		}
		dumpFields(b, "  ", t.Fields)
	case *EnumType:
		b.WriteString("type enum module=" + quoteDump(t.Module) + " " + dumpTypeString(t) + "\n")
		dumpVariants(b, t.Variants)
	case *DistinctType:
		b.WriteString("type distinct module=" + quoteDump(t.Module) + " " + opaqueFlag(t.Opaque) + t.Name + "\n")
		b.WriteString("  base " + dumpTypeString(t.Base) + "\n")
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
		b.WriteString(dumpTypeString(f.T) + "\n")
	}
}

func opaqueFlag(on bool) string {
	if on {
		return "opaque "
	}
	return ""
}

// dumpListed prints the `implements` and `derives` lists of a struct or enum.
func (c *Checker) dumpListed(b *strings.Builder, mod, path, name string) {
	info := c.traits.types[typeKey{path, name}]
	if info == nil {
		return
	}
	for _, te := range info.implExprs {
		b.WriteString("impl" + mod + "listed " + te.Name + " " + name + "\n")
	}
	for _, te := range info.derivExprs {
		b.WriteString("impl" + mod + "derived " + te.Name + " " + name + "\n")
	}
}

func dumpTrait(b *strings.Builder, mod string, tr *Trait) {
	std := ""
	if tr.Std {
		std = "std "
	}
	b.WriteString("trait" + mod + std + tr.Name + "\n")
	for _, tf := range tr.Funcs {
		dumpSignature(b, tf.Name, tf.Decl)
	}
}

// dumpImpl prints an `implement` block. The block's functions were renamed
// (implName), so their source names come from the implementation table.
func (c *Checker) dumpImpl(b *strings.Builder, mod string, d *ImplDecl) {
	t := c.traits.implType[d]
	b.WriteString("impl" + mod + "block " + d.Trait.Name + " " + declaredName(t) + "\n")
	for _, fd := range d.Funcs {
		name := fd.Name
		for key, f := range c.traits.implFuncs[t] {
			if f == fd {
				name = key[strings.Index(key, ".")+1:]
			}
		}
		dumpSignature(b, name, fd)
	}
}

func dumpSignature(b *strings.Builder, name string, fd *FuncDecl) {
	b.WriteString("  function " + name + "\n")
	for _, p := range fd.Params {
		b.WriteString("    param " + p.Name + " ")
		if p.Mutable {
			b.WriteString("mutable ")
		}
		if p.Escaping {
			b.WriteString("escaping ")
		}
		b.WriteString(dumpTypeString(p.T) + "\n")
	}
	b.WriteString("    result " + dumpTypeString(fd.ResultT) + "\n")
}

// dumpHolds prints implements(t, trait) for the four standard traits.
func (c *Checker) dumpHolds(b *strings.Builder, t Type) {
	for _, n := range []string{"Equatable", "Hashable", "Ordered", "Show"} {
		holds := "false"
		if c.implements(t, c.stdTrait(n)) {
			holds = "true"
		}
		b.WriteString("impl module=" + quoteDump(typeModule(t)) + " holds=" + holds + " " + n + " " + dumpTypeString(t) + "\n")
	}
}

// dumpTypeString prints t as Type.String does, except that a struct, enum or
// distinct type declared in a package module is written <module>.<Name>, so
// two modules that both declare Token dump apart.
func dumpTypeString(t Type) string {
	switch t := t.(type) {
	case *StructType:
		if t.Tuple {
			return "#[" + dumpTypeList(t.Args) + "]"
		}
		return qualifiedName(t.Module, t.Name) + dumpTypeArgs(t.Args)
	case *EnumType:
		return qualifiedName(t.Module, t.Name) + dumpTypeArgs(t.Args)
	case *DistinctType:
		return qualifiedName(t.Module, t.Name)
	case *ArrayType:
		return "Array<" + dumpTypeString(t.Elem) + ">"
	case *BoxType:
		return "Box<" + dumpTypeString(t.Elem) + ">"
	case *MapType:
		return "Map<" + dumpTypeString(t.Key) + ", " + dumpTypeString(t.Val) + ">"
	case *FuncType:
		return "(" + dumpTypeList(t.Params) + ") => " + dumpTypeString(t.Result)
	}
	return t.String()
}

func qualifiedName(module, name string) string {
	if module == "" {
		return name
	}
	return module + "." + name
}

func dumpTypeList(ts []Type) string {
	parts := make([]string, len(ts))
	for i, t := range ts {
		parts[i] = dumpTypeString(t)
	}
	return strings.Join(parts, ", ")
}

func dumpTypeArgs(ts []Type) string {
	if len(ts) == 0 {
		return ""
	}
	return "<" + dumpTypeList(ts) + ">"
}

// dumpScope prints the `module` block of a package module: its imports in
// source order, then the names its scope holds that the prelude does not
// give every module, by kind and then by name.
func (c *Checker) dumpScope(b *strings.Builder, s *modScope) {
	b.WriteString("module " + quoteDump(s.path) + "\n")
	for _, f := range s.src.Files {
		for _, imp := range f.Imports {
			kind := "std"
			if m := c.mods[imp.Path]; m != nil && m.src != nil {
				kind = "package"
			}
			b.WriteString("  import " + quoteDump(imp.Path) + " " + kind)
			if imp.Alias != "" {
				b.WriteString(" alias=" + imp.Alias)
			}
			if len(imp.Names) > 0 {
				b.WriteString(" names=" + strings.Join(imp.Names, ","))
			}
			b.WriteString("\n")
		}
	}
	prelude := c.mods[""]
	var lines []string
	for n, t := range s.types {
		if pt, ok := prelude.types[n]; !ok || pt != t {
			lines = append(lines, "type "+n+" "+dumpTypeString(t))
		}
	}
	writeScopeLines(b, lines)
	lines = nil
	for n, g := range s.generics {
		if prelude.generics[n] != g {
			lines = append(lines, "generic "+n+" "+qualifiedName(g.mod.path, g.name))
		}
	}
	writeScopeLines(b, lines)
	lines = nil
	for n, fd := range s.funcs {
		lines = append(lines, "func "+n+" "+qualifiedName(fd.Module, fd.Name))
	}
	writeScopeLines(b, lines)
	lines = nil
	for n, d := range s.consts {
		lines = append(lines, "const "+n+" "+qualifiedName(d.Module, d.Name))
	}
	writeScopeLines(b, lines)
	lines = nil
	for n, tr := range c.traits.traits[s.path] {
		lines = append(lines, "trait "+n+" "+qualifiedName(tr.Module, tr.Name))
	}
	writeScopeLines(b, lines)
	lines = nil
	for n, bi := range s.imported {
		lines = append(lines, "builtin "+n+" "+bi.Name)
	}
	writeScopeLines(b, lines)
	lines = nil
	for a, p := range s.modules {
		lines = append(lines, "alias "+a+" "+quoteDump(p))
	}
	writeScopeLines(b, lines)
}

func writeScopeLines(b *strings.Builder, lines []string) {
	sort.Strings(lines)
	for _, l := range lines {
		b.WriteString("  scope " + l + "\n")
	}
}

// BuiltinDump prints the builtin registry, the signature data of the
// standard library, in the `builtin` format of the doc comment above.
func BuiltinDump() string {
	var b strings.Builder
	dumpBuiltins(&b, "", preludeFuncs)
	for _, p := range stdModuleOrder() {
		dumpBuiltins(&b, p, stdModules[p].Funcs)
	}
	return b.String()
}

// stdModuleOrder lists the std module paths in the order moduleList names
// them (TestStdModuleOrder keeps the two in step).
func stdModuleOrder() []string {
	var paths []string
	seen := map[string]bool{}
	for _, p := range []string{"std/int", "std/float", "std/string"} {
		if stdModules[p] != nil {
			paths = append(paths, p)
			seen[p] = true
		}
	}
	for _, p := range registeredModules {
		if !seen[p] {
			paths = append(paths, p)
		}
	}
	return paths
}

func dumpBuiltins(b *strings.Builder, path string, funcs map[string]*Builtin) {
	names := make([]string, 0, len(funcs))
	for n := range funcs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		bi := funcs[n]
		b.WriteString("builtin module=" + quoteDump(path) + " export=" + n + " " + bi.Name + "\n")
		for i, p := range bi.Params {
			b.WriteString("  param ")
			if i < len(bi.Mutable) && bi.Mutable[i] {
				b.WriteString("mutable ")
			}
			b.WriteString(dumpTypeString(p) + "\n")
		}
		if bi.Params == nil && bi.Mutable != nil {
			flags := make([]string, len(bi.Mutable))
			for i, m := range bi.Mutable {
				flags[i] = strconv.FormatBool(m)
			}
			b.WriteString("  mutable " + strings.Join(flags, ",") + "\n")
		}
		if bi.Result != nil {
			b.WriteString("  result " + dumpTypeString(bi.Result) + "\n")
		}
		if len(bi.Effects) > 0 {
			b.WriteString("  effects " + strings.Join(bi.Effects, ",") + "\n")
		}
		if len(bi.Callbacks) > 0 {
			idx := make([]string, len(bi.Callbacks))
			for i, k := range bi.Callbacks {
				idx[i] = strconv.Itoa(k)
			}
			b.WriteString("  callbacks " + strings.Join(idx, ",") + "\n")
		}
		if bi.Check != nil {
			b.WriteString("  check " + checkKind(bi.Check) + "\n")
		}
	}
}

// checkKind names a builtin's Check function: the Go function name for a
// named one (checkArrayPush), and "fixed" for an inline closure. Every
// inline closure calls checkFixed with a fixed signature; TestSelfhostBuiltins
// confirms the signature the Narra registry holds for each one.
func checkKind(f func(*Checker, *CallExpr, Type) Type) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.Index(name, "."); i >= 0 {
		name = name[i+1:]
	}
	if strings.Contains(name, ".") {
		return "fixed"
	}
	return name
}
