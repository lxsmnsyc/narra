package narra

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// portedTraitErrors are the error formats of the trait passes that the Narra
// checker reports (selfhost/src/traits.nra). TestSelfhostTraits accepts them
// on top of portedDeclErrors.
var portedTraitErrors = []string{
	"%q is declared twice",
	"%s is listed twice in `implements`",
	"derives Shape is not supported by this compiler yet",
	"%s cannot be derived; only Equatable, Hashable, Ordered and Show can. List it in `implements` and write `implement %s for %s`",
	"%s is listed twice in `derives`",
	"unknown trait %q",
	"%s is listed twice in the constraint of %s",
	"%s implements %s but has no `implement %s for %s` block or `derives %s`",
	"%s both derives and implements %s; keep one",
	"cannot derive %s for %s: field %q has type %s, which does not implement %s",
	"trait functions cannot have type parameters in this compiler yet",
	"%s declares function %q twice",
	"a parameter of %s.%s must use Self, so a call can tell which implementation to run",
	"cannot implement %s for %s: the type is declared in another module; wrap it in a struct",
	"implement blocks for generic types are not supported by this compiler yet; use derives",
	"%s does not list %s in `implements`; add `implements %s` to its declaration",
	"%s is implemented twice for %s",
	"functions in an implement block cannot have type parameters in this compiler yet",
	"trait %s has no function %q",
	"function %q is implemented twice",
	"%s.%s for %s must be `%s`",
	"implement %s for %s is missing function %s",
	"%s cannot be a map key: %s",
}

var portedTraitRes = func() []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, f := range append(append([]string{}, portedDeclErrors...), portedTraitErrors...) {
		re := regexp.QuoteMeta(f)
		re = strings.NewReplacer("%q", `"(?:[^"\\]|\\.)*"`, "%s", `.*`, "%d", `-?[0-9]+`).Replace(re)
		out = append(out, regexp.MustCompile("^"+re+"$"))
	}
	return out
}()

// unportedTraitError returns the first error line of a Go dump that neither
// the declaration nor the trait passes of the Narra checker report, or "".
func unportedTraitError(dump string) string {
	for _, line := range strings.Split(strings.TrimSuffix(dump, "\n"), "\n") {
		m := declErrorLine.FindStringSubmatch(line)
		if m == nil {
			return line
		}
		ported := false
		for _, re := range portedTraitRes {
			if re.MatchString(m[1]) {
				ported = true
				break
			}
		}
		if !ported {
			return line
		}
	}
	return ""
}

// traitSkip says why an input is not compared, or "".
func traitSkip(path, goDump string, goOK bool) string {
	if _, err := Load(path); err != nil {
		return "does not load"
	}
	if !goOK {
		if line := unportedTraitError(goDump); line != "" {
			return "unported error: " + line
		}
	}
	return ""
}

// TestSelfhostTraits checks that `narrac decls <path>` prints the same dump
// as `narra decls`, trait, impl and bound lines included, or the same
// errors, on every input of declInputs. Skipped inputs are logged with the
// reason.
func TestSelfhostTraits(t *testing.T) {
	exe := buildSelfhost(t)
	for _, path := range declInputs(t) {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			want, ok := DeclDump(path)
			if reason := traitSkip(path, want, ok); reason != "" {
				t.Skip(reason)
			}
			out, errOut, code := runNarracRaw(t, exe, "decls", path)
			if !ok {
				if code != 1 || out != "" {
					t.Fatalf("exit code %d, want 1; stdout:\n%s\nstderr:\n%s", code, out, errOut)
				}
				if errOut != want {
					t.Fatalf("errors differ: %s", firstDiff(want, errOut))
				}
				return
			}
			if code != 0 {
				t.Fatalf("exit code %d, want 0; stderr:\n%s", code, errOut)
			}
			if want != out {
				t.Fatalf("dump differs: %s", firstDiff(want, out))
			}
		})
	}
}

// TestDeclDumpTraits compares the goldens in testdata/declsdump/traits with
// the Go dump. Each case is a .nra file with a .dump file beside it.
// NARRA_UPDATE=1 rewrites the goldens.
func TestDeclDumpTraits(t *testing.T) {
	cases := declDumpCases(t, filepath.Join(testdata, "declsdump", "traits"))
	if len(cases) == 0 {
		t.Fatal("no declsdump/traits cases")
	}
	for _, path := range cases {
		t.Run(filepath.Base(path), func(t *testing.T) {
			got, ok := DeclDump(path)
			if !ok && !strings.HasPrefix(got, "error ") {
				t.Fatalf("a failed dump must start with `error `: %q", got)
			}
			golden := strings.TrimSuffix(path, ".nra") + ".dump"
			if os.Getenv("NARRA_UPDATE") != "" {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Fatalf("dump differs: %s", firstDiff(string(want), got))
			}
		})
	}
}

// The skip rule must accept the trait formats and reject others.
func TestUnportedTraitError(t *testing.T) {
	ported := "error a.nra:1:2: unknown trait \"X\"\nerror a.nra:3:4: Money is implemented twice for Money\n"
	if got := unportedTraitError(ported); got != "" {
		t.Fatalf("ported errors rejected: %q", got)
	}
	other := "error a.nra:1:2: cannot use == on int\n"
	if got := unportedTraitError(other); got == "" {
		t.Fatal("a body error was accepted")
	}
}
