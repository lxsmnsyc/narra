package narra

import (
	"sort"
	"strings"
)

// The package dump is a fixed text form of a loaded package. `narra pkg
// <path>` prints it, and selfhost/src/modules.nra (`narrac pkg <path>`)
// prints the same format, so the Go loader and the Narra loader can be
// diffed. The dump does not repeat the syntax tree; `narra parse` covers it.
//
// Input
//
//	path is a package directory or a single .nra file, as Load accepts.
//
// Lines
//
//	package "<name>"
//	binary "<name>" entry="<entry>" module="<module>" fn="<function>"
//	libraries "<lib>" "<lib>"...
//	search "<dir>" "<dir>"...
//	module "<module path>" file="<file>"
//	  exports <name>,<name>...
//	  import <line>:<col> "<path>"
//
// Rules
//
//   - The lines come in this order: package, the binaries in manifest order,
//     libraries, search, then the modules in load order (imports first).
//     libraries and search print only when non-empty.
//   - Strings are quoted as in the AST dump (quoteDump).
//   - module and fn of a binary are the resolution of its entry
//     (Package.entryModule): module is "<package>/<module>" when that module
//     exists, otherwise the module part as written; fn is the part after the
//     last dot. An entry without a dot has module="" and fn=entry. The dump
//     does not check that the module or function exists.
//   - file is the path shown in diagnostics, relative to the package
//     ("src/text.nra"); for a single file it is the base name.
//   - exports lists the exported function, const, struct, enum, trait and
//     distinct names of the module in byte order, without duplicates; the
//     line prints only when non-empty. imports print one line per import, in
//     source order, with the position of the import.
//   - A single file hello.nra is package "hello" with one binary "hello"
//     whose entry is "hello.run".
//
// Errors
//
//	A failed load prints one line per diagnostic, as FormatError does, each
//	with the prefix "error ":
//	    error <file>:<line>:<col>: <message>
//	An error that is not a diagnostic (an unreadable file) prints
//	"error <text>" with the Go error text, which the Narra loader does not
//	reproduce; the comparison tests avoid those inputs. The dump of a failed
//	load is only the error lines, and the command exits 1.

// PkgDump loads the package at path and returns the dump described above.
func PkgDump(path string) (dump string, ok bool) {
	pkg, err := Load(path)
	if err != nil {
		var b strings.Builder
		for _, line := range strings.Split(FormatError(err).Error(), "\n") {
			b.WriteString("error " + line + "\n")
		}
		return b.String(), false
	}
	var b strings.Builder
	b.WriteString("package " + quoteDump(pkg.Name) + "\n")
	for _, bin := range pkg.Binaries {
		mod, fn := pkg.entryModule(bin.Entry)
		b.WriteString("binary " + quoteDump(bin.Name) + " entry=" + quoteDump(bin.Entry) + " module=" + quoteDump(mod) + " fn=" + quoteDump(fn) + "\n")
	}
	dumpList(&b, "libraries", pkg.Link.Libraries)
	dumpList(&b, "search", pkg.Link.Search)
	for _, m := range pkg.Modules {
		file := ""
		if len(m.Files) > 0 {
			file = m.Files[0].Path
		}
		b.WriteString("module " + quoteDump(m.Path) + " file=" + quoteDump(file) + "\n")
		var names []string
		for name := range m.Exports {
			names = append(names, name)
		}
		sort.Strings(names)
		if len(names) > 0 {
			b.WriteString("  exports " + strings.Join(names, ",") + "\n")
		}
		for _, f := range m.Files {
			for _, imp := range f.Imports {
				b.WriteString("  import " + imp.Pos.String() + " " + quoteDump(imp.Path) + "\n")
			}
		}
	}
	return b.String(), true
}

func dumpList(b *strings.Builder, name string, items []string) {
	if len(items) == 0 {
		return
	}
	b.WriteString(name)
	for _, s := range items {
		b.WriteString(" " + quoteDump(s))
	}
	b.WriteString("\n")
}
