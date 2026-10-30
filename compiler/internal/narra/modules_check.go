package narra

import (
	"fmt"
	"strings"
)

// modScope holds the names visible in one module: its own declarations,
// the names it imports, and the prelude.
type modScope struct {
	path     string
	file     string // for diagnostics
	src      *SourceModule
	types    map[string]Type
	generics map[string]*genericType
	funcs    map[string]*FuncDecl
	modules  map[string]string // namespace import alias -> module path
	imported map[string]*Builtin
}

// enter makes s the current module and returns the previous one.
func (c *Checker) enter(s *modScope) *modScope {
	prev := c.mod
	c.mod = s
	c.types, c.generics, c.funcs, c.modules, c.imported = s.types, s.generics, s.funcs, s.modules, s.imported
	return prev
}

func (c *Checker) newScope(path, file string, src *SourceModule) *modScope {
	s := &modScope{
		path:     path,
		file:     file,
		src:      src,
		types:    map[string]Type{"int": TInt, "i64": TInt, "f64": TF64, "bool": TBool, "char": TChar, "string": TStr},
		generics: map[string]*genericType{},
		funcs:    map[string]*FuncDecl{},
		modules:  map[string]string{},
		imported: map[string]*Builtin{},
	}
	// Every module sees the prelude.
	if p := c.mods[""]; p != nil {
		for n, t := range p.types {
			s.types[n] = t
		}
		for n, g := range p.generics {
			s.generics[n] = g
		}
	}
	c.mods[path] = s
	return s
}

// CheckPackage checks every module of a package in dependency order. Each
// module sees only its own declarations, its imports and the prelude, and
// uses the signatures of the modules it imports.
func CheckPackage(pkg *Package) (*Program, error) {
	prelude, err := Parse(preludeSrc)
	if err != nil {
		return nil, fmt.Errorf("prelude: %v", err)
	}
	c := &Checker{
		prog:      &Program{Entries: map[string]*FuncDecl{}, Files: map[string]string{}},
		funcTP:    map[*FuncDecl][]*TypeParamType{},
		funcInsts: map[string]*FuncDecl{},
		arrays:    map[string]*ArrayType{},
		boxes:     map[string]*BoxType{},
		maps:      map[string]*MapType{},
		mods:      map[string]*modScope{},
		funcMod:   map[*FuncDecl]*modScope{},
		pkg:       pkg,
		traits:    newTraitState(),
	}
	c.enter(c.newScope("", "prelude", nil))
	concrete := c.declareModule(prelude)
	for _, m := range pkg.Modules {
		file := ""
		if len(m.Files) > 0 {
			file = m.Files[0].Path
		}
		c.enter(c.newScope(m.Path, file, m))
		c.prog.Files[m.Path] = file
		for _, f := range m.Files {
			concrete = append(concrete, c.declareModule(f)...)
		}
		for _, f := range m.Files {
			c.checkBodies(f)
		}
	}
	c.checkPending()
	c.flushMapKeys()
	for _, b := range pkg.Binaries {
		if fd := c.entryFunc(b); fd != nil {
			saved := c.enter(c.funcMod[fd])
			c.checkEntry(fd)
			c.enter(saved)
			c.prog.Entries[b.Name] = fd
		}
	}
	c.finish(concrete)
	if len(c.errs) == 0 {
		c.checkEffects()
	}
	if len(c.errs) > 0 {
		return nil, Diagnostics(c.errs)
	}
	return c.prog, nil
}

// entryFunc finds the exported function a binary names, or nil.
func (c *Checker) entryFunc(b Binary) *FuncDecl {
	mod, fn := c.pkg.entryModule(b.Entry)
	s := c.mods[mod]
	if s == nil || s.src == nil {
		return nil
	}
	fd, _ := s.src.Exports[fn].(*FuncDecl)
	return fd
}

// checkEntry reports an entry function with the wrong signature.
func (c *Checker) checkEntry(e *FuncDecl) {
	ok := len(e.Params) == 1 && sameType(e.Params[0].T, &ArrayType{Elem: TStr}) && !e.Params[0].Mutable && e.ResultT == TInt
	if !ok {
		c.errorf(e.Pos, "the entry function must be `export function %s(args: Array<string>): int`", e.Name)
	}
}

// importPackageModule handles an import of a module of this package. It
// returns false when the path is not one, so the caller tries std modules.
func (c *Checker) importPackageModule(imp *ImportDecl) bool {
	s := c.mods[imp.Path]
	if s == nil || s.src == nil {
		if !c.pkg.single && strings.HasPrefix(imp.Path, c.pkg.Name+"/") {
			c.errorf(imp.Pos, "unknown module %q; package %s has no file src/%s.nra", imp.Path, c.pkg.Name, strings.TrimPrefix(imp.Path, c.pkg.Name+"/"))
			return true
		}
		return false
	}
	if imp.Alias != "" {
		c.modules[imp.Alias] = imp.Path
	}
	for _, n := range imp.Names {
		d := c.exported(imp.Pos, s, n)
		switch d := d.(type) {
		case *FuncDecl:
			if _, dup := c.funcs[n]; dup {
				c.errorf(imp.Pos, "function %q is declared twice", n)
			}
			c.funcs[n] = d
		case *TraitDecl:
			c.importTrait(imp.Pos, n, d)
		case *StructDecl, *EnumDecl:
			if g := s.generics[n]; g != nil && s.types[n] == nil {
				if c.types[n] != nil || c.generics[n] != nil {
					c.errorf(imp.Pos, "type %q is declared twice", n)
				}
				c.generics[n] = g
			} else {
				c.declareType(imp.Pos, n, s.types[n])
			}
		}
	}
	return true
}

// exported returns the exported declaration n of module s, or reports why
// it cannot be used and returns nil.
func (c *Checker) exported(pos Pos, s *modScope, n string) Decl {
	if d, ok := s.src.Exports[n]; ok {
		return d
	}
	if s.src.declares(n) {
		c.errorf(pos, "%q is not exported by module %q; add `export` to its declaration", n, s.path)
	} else {
		c.errorf(pos, "module %q has no export %q", s.path, n)
	}
	return nil
}

// moduleCall checks `Alias.name(args)` where Alias names a package module.
func (c *Checker) moduleCall(e *CallExpr, s *modScope, callee *FieldExpr, want Type) Type {
	fd, ok := c.exported(callee.P, s, callee.Name).(*FuncDecl)
	if !ok {
		c.argsLoose(e)
		return TVoid
	}
	if _, generic := c.funcTP[fd]; generic {
		return c.genericCall(e, fd, want)
	}
	e.Func = fd
	c.args(e, fd.Params, func(i int) Type { return fd.Params[i].T })
	return fd.ResultT
}

// qualifiedType resolves a type written as `Alias.Name`.
func (c *Checker) qualifiedType(te *TypeExpr) Type {
	alias, name, _ := strings.Cut(te.Name, ".")
	s := c.mods[c.modules[alias]]
	if _, ok := c.modules[alias]; !ok || s == nil || s.src == nil {
		c.errorf(te.Pos, "unknown module %q in type %s; import it with `import * as %s from \"...\"`", alias, te.Name, alias)
		return TVoid
	}
	switch c.exported(te.Pos, s, name).(type) {
	case *StructDecl, *EnumDecl:
	case nil:
		return TVoid
	default:
		c.errorf(te.Pos, "%s is not a type", te.Name)
		return TVoid
	}
	if g := s.generics[name]; g != nil && s.types[name] == nil {
		if len(te.Args) != len(g.params) {
			c.errorf(te.Pos, "%s takes %d type arguments", te.Name, len(g.params))
			return TVoid
		}
		args := make([]Type, len(te.Args))
		for i, a := range te.Args {
			args[i] = c.resolveType(a)
		}
		return c.instantiate(g, args)
	}
	if len(te.Args) > 0 {
		c.errorf(te.Pos, "type %s is not generic", te.Name)
		return TVoid
	}
	return s.types[name]
}

// genericOf finds a generic type by its declaring module and name.
func (c *Checker) genericOf(module, name string) *genericType {
	if s := c.mods[module]; s != nil {
		return s.generics[name]
	}
	return nil
}

// namesEnum reports whether name, as written in the current module, refers
// to the enum et. Two modules may declare enums with the same name.
func (c *Checker) namesEnum(name string, et *EnumType) bool {
	if t, ok := c.types[name]; ok {
		return t == et
	}
	if g := c.generics[name]; g != nil {
		return g.name == et.Name && g.mod.path == et.Module
	}
	return false
}

// typeName prints t for a message that also prints other. When both print
// the same, the module is added, as in "p/b.Kind".
func typeName(t, other Type) string {
	s := t.String()
	if s != other.String() {
		return s
	}
	switch t := t.(type) {
	case *StructType:
		if t.Module != "" {
			return t.Module + "." + s
		}
	case *EnumType:
		if t.Module != "" {
			return t.Module + "." + s
		}
	}
	return s
}

// typeErrorf reports an error at the declaration of a struct or enum, in
// the file of the module that declares it.
func (c *Checker) typeErrorf(t Type, format string, args ...any) {
	module, name := "", ""
	switch t := t.(type) {
	case *StructType:
		module, name = t.Module, t.Name
	case *EnumType:
		module, name = t.Module, t.Name
	}
	s := c.mods[module]
	if s == nil {
		c.errorf(Pos{}, format, args...)
		return
	}
	pos := Pos{}
	if s.src != nil {
		for _, f := range s.src.Files {
			for _, d := range f.Decls {
				if _, isFunc := d.(*FuncDecl); !isFunc && declName(d) == name {
					pos = declPos(d)
				}
			}
		}
	}
	saved := c.enter(s)
	c.errorf(pos, format, args...)
	c.enter(saved)
}

func declPos(d Decl) Pos {
	switch d := d.(type) {
	case *FuncDecl:
		return d.Pos
	case *StructDecl:
		return d.Pos
	case *EnumDecl:
		return d.Pos
	}
	return Pos{}
}
