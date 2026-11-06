package narra

// A local or parameter that is never read is an error (SPEC.md section 2).
// A name that is `_` or starts with `_` is unused on purpose; the hidden
// names the parser makes for `for` destructuring start with `__`. The entry
// function `run` has a fixed signature, so its parameter is exempt.

// reportUnused reports the locals in c.declared[start:] that no expression
// read, then forgets them. It reports nothing when the body already has an
// error: a name used in an expression that failed to check would look unused.
func (c *Checker) reportUnused(start, errs int) {
	locals := c.declared[start:]
	c.declared = c.declared[:start]
	if len(c.errs) > errs {
		return
	}
	for _, l := range locals {
		if l.Used || l.Exempt || l.Name == "" || l.Name[0] == '_' {
			continue
		}
		unused := "_" + l.Name
		fix := Fix{Title: "Rename to " + unused, Edits: []Edit{{Start: l.NamePos, End: l.NamePos, Text: "_"}}}
		if l.Shorthand {
			// `{ host }` becomes `{ host: _host }`; renaming alone would change the field.
			end := Pos{Line: l.NamePos.Line, Col: l.NamePos.Col + len(l.Name)}
			fix = Fix{Title: "Rename to " + unused, Edits: []Edit{{Start: end, End: end, Text: ": " + unused}}}
		}
		if l.Param {
			c.errorfFix(l.Pos, []Fix{fix}, "parameter %q is never read; remove it, or name it %q to mark it as unused on purpose", l.Name, unused)
		} else {
			c.errorfFix(l.Pos, []Fix{fix}, "%q is never read; remove it, or name it %q to mark it as unused on purpose", l.Name, unused)
		}
	}
}
