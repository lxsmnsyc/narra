package narra

import "sort"

// A local or parameter that is never read is an error (SPEC.md section 2).
// A name that is `_` or starts with `_` is unused on purpose; the hidden
// names the parser makes for `for` destructuring start with `__`. The entry
// function of the entry module has a fixed signature, so its parameter is
// exempt. A `let` that is only assigned is unused; a `mutable` parameter that
// is only written is an output the caller sees, so it is used.

// reportUnused reports the locals in c.declared[start:] that no expression
// read, then forgets them. It reports nothing when the body already has an
// error: a name used in an expression that failed to check would look unused.
func (c *Checker) reportUnused(start, errs int) {
	locals := append([]*Local(nil), c.declared[start:]...)
	c.declared = c.declared[:start]
	if len(c.errs) > errs {
		return
	}
	// Report in source order, whatever order the locals were declared in.
	sort.SliceStable(locals, func(a, b int) bool {
		if locals[a].Pos.Line != locals[b].Pos.Line {
			return locals[a].Pos.Line < locals[b].Pos.Line
		}
		return locals[a].Pos.Col < locals[b].Pos.Col
	})
	for _, l := range locals {
		if l.Used || l.Exempt || l.Name == "" || l.Name[0] == '_' {
			continue
		}
		unused := "_" + l.Name
		// The declaration: `{ host }` becomes `{ host: _host }`, because renaming
		// alone would change the field; any other name gets a leading `_`.
		edit := Edit{Start: l.NamePos, End: l.NamePos, Text: "_"}
		if l.Shorthand {
			end := Pos{Line: l.NamePos.Line, Col: l.NamePos.Col + len(l.Name)}
			edit = Edit{Start: end, End: end, Text: ": " + unused}
		}
		fix := Fix{Title: "Rename to " + unused, Edits: []Edit{edit}}
		for _, w := range l.Writes {
			// The fix renames the assignments too, or the file would not compile.
			fix.Edits = append(fix.Edits, Edit{Start: w, End: w, Text: "_"})
		}
		if l.Param {
			c.errorfFix(l.Pos, []Fix{fix}, "parameter %q is never read; remove it, or name it %q to mark it as unused on purpose", l.Name, unused)
		} else {
			c.errorfFix(l.Pos, []Fix{fix}, "%q is never read; remove it, or name it %q to mark it as unused on purpose", l.Name, unused)
		}
	}
}

// isEntryFunc reports whether fd is the entry function of a binary of the
// package: `export function run(args: Array<string>): int` of the entry module.
func (c *Checker) isEntryFunc(fd *FuncDecl) bool {
	if c.pkg == nil {
		return false
	}
	for _, b := range c.pkg.Binaries {
		if c.entryFunc(b) == fd {
			return true
		}
	}
	return false
}
