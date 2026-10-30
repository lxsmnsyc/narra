package narra

import (
	"fmt"
	"strings"
)

// parseManifest reads the narra.toml keys the compiler uses: [package] name,
// [[binary]] name and entry, and [link] libraries and search. It reads a
// subset of TOML: tables, arrays of tables, `key = value` with basic or
// literal strings and arrays of strings, and comments. Other tables, such as
// [build], are skipped without being parsed. [dependencies] is an error.
func parseManifest(text string) (*Package, error) {
	pkg := &Package{}
	table := ""
	lines := strings.Split(text, "\n")
	fail := func(line int, format string, args ...any) error {
		return Diagnostic{Code: "E0000", File: "narra.toml", Pos: Pos{Line: line, Col: 1}, Message: fmt.Sprintf(format, args...)}
	}
	for i := 0; i < len(lines); i++ {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(lines[i]))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[[") {
			if !strings.HasSuffix(line, "]]") {
				return nil, fail(lineNo, "expected `]]`")
			}
			table = "[[" + strings.TrimSpace(line[2:len(line)-2]) + "]]"
			if table == "[[binary]]" {
				pkg.Binaries = append(pkg.Binaries, Binary{})
			}
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, fail(lineNo, "expected `]`")
			}
			table = strings.TrimSpace(line[1 : len(line)-1])
			if table == "dependencies" {
				return nil, fail(lineNo, "[dependencies] is not supported by this compiler yet")
			}
			continue
		}
		known := table == "package" || table == "link" || table == "[[binary]]"
		eq := strings.Index(line, "=")
		if eq < 0 {
			if !known {
				continue
			}
			return nil, fail(lineNo, "expected `key = value`")
		}
		key := strings.TrimSpace(line[:eq])
		raw := strings.TrimSpace(line[eq+1:])
		// An array may continue over several lines.
		for strings.HasPrefix(raw, "[") && !closedArray(raw) && i+1 < len(lines) {
			i++
			raw += " " + strings.TrimSpace(stripComment(lines[i]))
		}
		if !known {
			continue
		}
		val, err := parseValue(raw)
		if err != nil {
			return nil, fail(lineNo, "%s: %v", key, err)
		}
		str, isStr := val.(string)
		list, isList := val.([]string)
		switch table + "." + key {
		case "package.name":
			if !isStr || !validPackageName(str) {
				return nil, fail(lineNo, "package name must be a string of letters, digits and `_`, starting with a letter")
			}
			if str == "std" {
				return nil, fail(lineNo, "package name \"std\" is reserved for the standard library")
			}
			pkg.Name = str
		case "[[binary]].name", "[[binary]].entry":
			if !isStr {
				return nil, fail(lineNo, "%s must be a string", key)
			}
			b := &pkg.Binaries[len(pkg.Binaries)-1]
			if key == "name" {
				b.Name = str
			} else {
				b.Entry = str
			}
		case "link.libraries", "link.search":
			if !isList {
				return nil, fail(lineNo, "%s must be an array of strings", key)
			}
			if key == "libraries" {
				pkg.Link.Libraries = list
			} else {
				pkg.Link.Search = list
			}
		}
	}
	if pkg.Name == "" {
		return nil, fail(0, "[package] needs a name")
	}
	seen := map[string]bool{}
	for _, b := range pkg.Binaries {
		if b.Name == "" || b.Entry == "" {
			return nil, fail(0, "each [[binary]] needs a name and an entry such as \"cli.run\"")
		}
		if seen[b.Name] {
			return nil, fail(0, "two [[binary]] entries are named %q", b.Name)
		}
		seen[b.Name] = true
	}
	return pkg, nil
}

func validPackageName(s string) bool {
	if s == "" || !(s[0] >= 'a' && s[0] <= 'z' || s[0] >= 'A' && s[0] <= 'Z') {
		return false
	}
	for _, ch := range s {
		if !(ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9') {
			return false
		}
	}
	return true
}

// stripComment removes a `#` comment that is not inside a string.
func stripComment(line string) string {
	var quote byte
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case quote != 0:
			if ch == '\\' && quote == '"' {
				i++
			} else if ch == quote {
				quote = 0
			}
		case ch == '"' || ch == '\'':
			quote = ch
		case ch == '#':
			return line[:i]
		}
	}
	return line
}

// closedArray reports whether the brackets of an array value are balanced.
func closedArray(s string) bool {
	depth := 0
	var quote byte
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case quote != 0:
			if ch == '\\' && quote == '"' {
				i++
			} else if ch == quote {
				quote = 0
			}
		case ch == '"' || ch == '\'':
			quote = ch
		case ch == '[':
			depth++
		case ch == ']':
			depth--
		}
	}
	return depth == 0
}

// parseValue reads a string or an array of strings.
func parseValue(s string) (any, error) {
	if strings.HasPrefix(s, "[") {
		var out []string
		rest := strings.TrimSpace(s[1:])
		for {
			if strings.HasPrefix(rest, "]") {
				if strings.TrimSpace(rest[1:]) != "" {
					return nil, fmt.Errorf("unexpected text after the array")
				}
				return out, nil
			}
			str, n, err := parseString(rest)
			if err != nil {
				return nil, err
			}
			out = append(out, str)
			rest = strings.TrimSpace(rest[n:])
			if strings.HasPrefix(rest, ",") {
				rest = strings.TrimSpace(rest[1:])
			} else if !strings.HasPrefix(rest, "]") {
				return nil, fmt.Errorf("expected `,` or `]` in the array")
			}
		}
	}
	str, n, err := parseString(s)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(s[n:]) != "" {
		return nil, fmt.Errorf("unexpected text after the string")
	}
	return str, nil
}

// parseString reads a basic ("...") or literal ('...') string at the start
// of s and returns its value and length.
func parseString(s string) (string, int, error) {
	if s == "" || (s[0] != '"' && s[0] != '\'') {
		return "", 0, fmt.Errorf("expected a string")
	}
	quote := s[0]
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		ch := s[i]
		if ch == quote {
			return b.String(), i + 1, nil
		}
		if ch == '\\' && quote == '"' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case '"', '\\':
				b.WriteByte(s[i])
			default:
				return "", 0, fmt.Errorf("unsupported escape \\%c", s[i])
			}
			continue
		}
		b.WriteByte(ch)
	}
	return "", 0, fmt.Errorf("unterminated string")
}
