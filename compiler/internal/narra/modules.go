package narra

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Packages and the module graph. See roadmaps/selfhost/contracts/module-graph.md.

// Binary is a [[binary]] entry of narra.toml.
type Binary struct {
	Name  string
	Entry string // "cli.run": module path inside the package, then the function
}

// LinkConfig is the [link] table of narra.toml. It is read but not used yet.
type LinkConfig struct {
	Libraries []string
	Search    []string
}

// SourceModule is one module of a package: one file under src/.
// The contract calls it Module; that name is taken by std modules (builtins.go).
type SourceModule struct {
	Path    string // "app/text"
	Files   []*File
	Exports map[string]Decl // declarations with Exported set, by name
}

type Package struct {
	Name     string
	Dir      string
	Modules  []*SourceModule // dependency order, imports first
	Binaries []Binary
	Link     LinkConfig
	single   bool // loaded from one .nra file, without narra.toml
}

func newSourceModule(path string, f *File) *SourceModule {
	m := &SourceModule{Path: path, Files: []*File{f}, Exports: map[string]Decl{}}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *FuncDecl:
			if d.Exported {
				m.Exports[d.Name] = d
			}
		case *StructDecl:
			if d.Exported {
				m.Exports[d.Name] = d
			}
		case *EnumDecl:
			if d.Exported {
				m.Exports[d.Name] = d
			}
		case *TraitDecl:
			if d.Exported {
				m.Exports[d.Name] = d
			}
		}
	}
	return m
}

// declName returns the name of a function, struct or enum declaration.
func declName(d Decl) string {
	switch d := d.(type) {
	case *FuncDecl:
		return d.Name
	case *StructDecl:
		return d.Name
	case *EnumDecl:
		return d.Name
	case *TraitDecl:
		return d.Name
	}
	return ""
}

// declares reports whether the module declares a top-level name, exported or not.
func (m *SourceModule) declares(name string) bool {
	for _, f := range m.Files {
		for _, d := range f.Decls {
			if declName(d) == name {
				return true
			}
		}
	}
	return false
}

// Load reads a package directory, or a single .nra file as a package with
// one module and one binary whose entry is `run`.
func Load(path string) (*Package, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return LoadPackage(path)
	}
	return loadFile(path)
}

func loadFile(path string) (*Package, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	base := filepath.Base(path)
	stem := strings.TrimSuffix(base, ".nra")
	f, err := Parse(string(src))
	if err != nil {
		return nil, withFile(err, base)
	}
	f.Path = base
	return &Package{
		Name:     stem,
		Dir:      filepath.Dir(path),
		Modules:  []*SourceModule{newSourceModule(stem, f)},
		Binaries: []Binary{{Name: stem, Entry: stem + ".run"}},
		single:   true,
	}, nil
}

// LoadPackage reads narra.toml and every file under src/, and orders the
// modules so that each comes after the modules it imports.
func LoadPackage(dir string) (*Package, error) {
	text, err := os.ReadFile(filepath.Join(dir, "narra.toml"))
	if err != nil {
		return nil, err
	}
	pkg, err := parseManifest(string(text))
	if err != nil {
		return nil, err
	}
	pkg.Dir = dir
	var paths []string
	err = filepath.WalkDir(filepath.Join(dir, "src"), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".nra") {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("narra.toml: package %s has no .nra files in src/", pkg.Name)
	}
	sort.Strings(paths)
	byPath := map[string]*SourceModule{}
	cnames := map[string]string{}
	var errs Diagnostics
	var mods []*SourceModule
	for _, p := range paths {
		rel, _ := filepath.Rel(filepath.Join(dir, "src"), p)
		rel = filepath.ToSlash(strings.TrimSuffix(rel, ".nra"))
		display := "src/" + rel + ".nra"
		src, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		f, err := Parse(string(src))
		if err != nil {
			errs = append(errs, diagnosticsOf(withFile(err, display))...)
			continue
		}
		f.Path = display
		m := newSourceModule(pkg.Name+"/"+rel, f)
		if other, dup := cnames[moduleHash(m.Path)]; dup {
			errs = append(errs, Diagnostic{Code: "E0000", File: display, Message: fmt.Sprintf("modules %q and %q have the same C name hash; rename one file", other, m.Path)})
		}
		cnames[moduleHash(m.Path)] = m.Path
		byPath[m.Path] = m
		mods = append(mods, m)
	}
	if len(errs) > 0 {
		return nil, errs
	}
	pkg.Modules, errs = orderModules(mods, byPath)
	if len(errs) > 0 {
		return nil, errs
	}
	return pkg, nil
}

// orderModules sorts modules so imports come first, and reports import cycles.
func orderModules(mods []*SourceModule, byPath map[string]*SourceModule) ([]*SourceModule, Diagnostics) {
	const (
		visiting = 1
		done     = 2
	)
	state := map[*SourceModule]int{}
	var order []*SourceModule
	var stack []string
	var errs Diagnostics
	var visit func(m *SourceModule)
	visit = func(m *SourceModule) {
		state[m] = visiting
		stack = append(stack, m.Path)
		for _, f := range m.Files {
			for _, imp := range f.Imports {
				dep := byPath[imp.Path]
				if dep == nil {
					continue // std or unknown; the checker reports unknown modules
				}
				switch state[dep] {
				case visiting:
					start := 0
					for i, p := range stack {
						if p == dep.Path {
							start = i
						}
					}
					cycle := append(append([]string{}, stack[start:]...), dep.Path)
					errs = append(errs, Diagnostic{Code: "E0000", Pos: imp.Pos, File: f.Path, Message: "import cycle: " + strings.Join(cycle, " -> ")})
				case 0:
					visit(dep)
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[m] = done
		order = append(order, m)
	}
	for _, m := range mods {
		if state[m] == 0 {
			visit(m)
		}
	}
	return order, errs
}

// binary finds a binary by name. An empty name selects the only binary.
func (p *Package) binary(name string) (Binary, error) {
	if name != "" {
		for _, b := range p.Binaries {
			if b.Name == name {
				return b, nil
			}
		}
		return Binary{}, fmt.Errorf("narra.toml: package %s has no binary %q", p.Name, name)
	}
	switch len(p.Binaries) {
	case 0:
		return Binary{}, fmt.Errorf("narra.toml: package %s has no [[binary]]; it is a library", p.Name)
	case 1:
		return p.Binaries[0], nil
	}
	var names []string
	for _, b := range p.Binaries {
		names = append(names, b.Name)
	}
	return Binary{}, fmt.Errorf("narra.toml: package %s has several binaries (%s); choose one with -bin", p.Name, strings.Join(names, ", "))
}

// entryModule resolves the module part of a binary's entry. It accepts the
// contract form "cli.run" and the spec form "app/cli.run".
func (p *Package) entryModule(entry string) (mod, fn string) {
	i := strings.LastIndex(entry, ".")
	if i < 0 {
		return "", entry
	}
	mod, fn = entry[:i], entry[i+1:]
	for _, m := range p.Modules {
		if m.Path == p.Name+"/"+mod {
			return m.Path, fn
		}
	}
	return mod, fn
}

// withFile sets the file of diagnostics that have none.
func withFile(err error, file string) error {
	ds := diagnosticsOf(err)
	if ds == nil {
		return err
	}
	for i := range ds {
		if ds[i].File == "" {
			ds[i].File = file
		}
	}
	return ds
}

func diagnosticsOf(err error) Diagnostics {
	switch e := err.(type) {
	case Diagnostics:
		return e
	case Diagnostic:
		return Diagnostics{e}
	}
	return nil
}

// FormatError prints diagnostics as `file:line:col: message`, one per line.
func FormatError(err error) error {
	ds := diagnosticsOf(err)
	if ds == nil {
		return err
	}
	lines := make([]string, len(ds))
	for i, d := range ds {
		if d.File != "" {
			lines[i] = d.File + ":" + d.Error()
		} else {
			lines[i] = d.Error()
		}
	}
	return fmt.Errorf("%s", strings.Join(lines, "\n"))
}
