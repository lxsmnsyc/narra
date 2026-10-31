package narra

import "strings"

// Import names (SPEC section 3, D72). A namespace alias never equals a type
// name, two imports never share an alias, and `{ a as b }` renames a name.

// pluralAlias is the conventional alias of the std modules of the capitalized
// prelude types, by the type name.
var pluralAlias = map[string]string{
	"Array": "Arrays", "Map": "Maps", "Box": "Boxes",
	"Range": "Ranges", "Option": "Options", "Result": "Results",
}

// importNames prints the names of an import for the dumps: `a`, or
// `a as b` when renamed.
func importNames(names []ImportName) string {
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = n.Name
		if n.Local != n.Name {
			parts[i] += " as " + n.Local
		}
	}
	return strings.Join(parts, ",")
}

// importNameStrings is importNames as a list, for the AST dump.
func importNameStrings(names []ImportName) []string {
	var out []string
	for _, n := range names {
		s := n.Name
		if n.Local != n.Name {
			s += " as " + n.Local
		}
		out = append(out, s)
	}
	return out
}

// bindAlias makes the alias of a namespace import visible. aliases holds
// the aliases the current file has bound. An alias may be bound again by
// another file of the module for the same path, never for another one.
func (c *Checker) bindAlias(imp *ImportDecl, aliases map[string]bool) {
	if imp.Alias == "" {
		return
	}
	if prev, ok := c.modules[imp.Alias]; ok && (prev != imp.Path || aliases[imp.Alias]) {
		c.errorf(imp.Pos, "two imports use the alias %s; give one of them another alias", imp.Alias)
	}
	aliases[imp.Alias] = true
	c.modules[imp.Alias] = imp.Path
}

// aliasFor returns the alias under which the module imports path, or "".
func (c *Checker) aliasFor(path string) string {
	best := ""
	for a, p := range c.modules {
		if p == path && (best == "" || a < best) {
			best = a
		}
	}
	return best
}

// aliasTaken reports whether name is a type, a trait or a prelude namespace
// type visible in the module.
func (c *Checker) aliasTaken(name string) bool {
	if c.typeNamed(name) {
		return true
	}
	_, ns := namespaceModules[name]
	return ns
}

// checkImportNames runs after the module's type names are declared. It
// reports an alias that equals a type name and an imported name that clashes
// with a declaration.
func (c *Checker) checkImportNames(f *File) {
	for _, imp := range f.Imports {
		if imp.Alias == "" || !c.aliasTaken(imp.Alias) {
			continue
		}
		sugg := pluralAlias[imp.Alias]
		if want, ok := namespacePlural(imp.Path); ok {
			sugg = want
		}
		if sugg == "" {
			sugg = imp.Alias + "s"
		}
		var fixes []Fix
		// When the alias is also an enum, `Alias.Ctor(...)` is a constructor
		// and stays. Only the uses of an export of a std module are renamed,
		// and a package module gets no fix.
		mod := stdModules[imp.Path]
		ok := !c.aliasTaken(sugg) && c.modules[sugg] == "" && !f.inTemplate[imp.Alias]
		if c.isEnumName(imp.Alias) && mod == nil {
			ok = false
		}
		if ok {
			edits := []Edit{{Start: imp.AliasPos, End: Pos{imp.AliasPos.Line, imp.AliasPos.Col + len(imp.Alias)}, Text: sugg}}
			for _, u := range f.dotUses[imp.Alias] {
				if c.isEnumName(imp.Alias) && mod.Funcs[u.Member] == nil {
					continue
				}
				edits = append(edits, Edit{Start: u.Pos, End: Pos{u.Pos.Line, u.Pos.Col + len(imp.Alias)}, Text: sugg})
			}
			fixes = []Fix{{Title: "rename the alias to " + sugg, Edits: edits}}
		}
		c.errorfFix(imp.Pos, fixes, "the import alias %s is also a type name; rename it, for example `import * as %s from %q`", imp.Alias, sugg, imp.Path)
	}
	declared := map[string]bool{}
	for _, d := range f.Decls {
		if fd, ok := d.(*FuncDecl); ok {
			declared[fd.Name] = true
		}
	}
	for _, imp := range f.Imports {
		if s := c.mods[imp.Path]; s != nil && s.src != nil {
			continue
		}
		for _, n := range imp.Names {
			_, isAlias := c.modules[n.Local]
			if declared[n.Local] || c.mod.consts[n.Local] != nil || isAlias || c.aliasTaken(n.Local) {
				c.errorf(imp.Pos, "imported name %q is already declared in this module; rename the import with `as`", n.Local)
			}
		}
	}
}

// namespacePlural returns the plural alias of the std module at path.
func namespacePlural(path string) (string, bool) {
	for ns, p := range namespaceModules {
		if p == path {
			return pluralAlias[ns], true
		}
	}
	return "", false
}
