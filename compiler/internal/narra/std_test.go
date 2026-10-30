package narra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunStdTestFailures runs test programs whose std/test assertions fail.
// Each must exit 101 with the message on stderr, in debug (sanitizers and
// leak check) and release builds, and stop before the next test.
func TestRunStdTestFailures(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"eq_int", `test "a" {
  Test.eq(1 + 1, 3)
}`, "panic at prog.nra:4:10: expected 3, got 2"},
		{"eq_struct", `struct P derives Show {
  x: int
  name: string
}

test "a" {
  Test.eq(P { x: 1, name: "a" }, P { x: 2, name: "a" })
}`, `panic at prog.nra:9:10: expected P { x: 2, name: "a" }, got P { x: 1, name: "a" }`},
		{"eq_array", `test "a" {
  const xs = [1, 2]
  Test.eq(xs, [1, 2, 3])
}`, "panic at prog.nra:5:10: expected [1, 2, 3], got [1, 2]"},
		{"ne", `test "a" {
  Test.ne("x", "x")
}`, `panic at prog.nra:4:10: expected a value other than "x"`},
		{"panics_returns", `test "a" {
  Test.panics(() => print("returned"))
}`, "panic at prog.nra:4:14: expected a panic, but the function returned"},
		{"panics_with_other_message", `test "a" {
  Test.panicsWith(() => { panic("boom") }, "bang")
}`, `panic at prog.nra:4:18: expected a panic containing "bang", got "boom"`},
		{"failure_inside_closure", `test "a" {
  const xs = [1]
  Test.panics(() => { const y = xs[0] })
}`, "expected a panic, but the function returned"},
		{"timeout", `test "a" {
  Test.panics(() => { while (true) {} })
}`, "panic at prog.nra:4:14: expected a panic, but the function ran past the 1 s timeout"},
		// The nested closure and the closure around it have the same limit and
		// race: either the outer call fails on its timeout, or it passes and the
		// last call fails. Neither hangs, and the message is the same.
		{"nested_timeout", `function spin() {
  while (true) {}
}

test "a" {
  Test.panics(() => Test.panics(() => spin()))
  Test.panics(() => spin())
}`, "ran past the 1 s timeout"},
	}
	// A short limit for the cases whose closure never returns. The test has 2
	// seconds, so a Test.panics closure gets 1 and its own message comes first.
	t.Setenv("NARRA_TEST_TIMEOUT", "2")
	for _, tc := range cases {
		for _, m := range modes {
			t.Run(tc.name+"/"+m.name, func(t *testing.T) {
				src := filepath.Join(t.TempDir(), "prog.nra")
				text := "import * as Test from \"std/test\"\n\n" + tc.src + "\n\ntest \"never runs\" {\n  print(\"next test\")\n}\n"
				if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
					t.Fatal(err)
				}
				opts := m.opts
				opts.Tests = true
				// The runner prints "FAIL - a: <message> (<position>)" and goes on.
				want := tc.want
				if rest, ok := strings.CutPrefix(want, "panic at "); ok {
					pos, msg, _ := strings.Cut(rest, ": ")
					want = "FAIL - a: " + msg + " (" + pos + ")"
				}
				out, stderr, code := runProgram(t, src, opts)
				if code != 1 || !strings.Contains(out, want) {
					t.Fatalf("want exit 1 and %q, got exit %d\nstdout:\n%s\nstderr:\n%s", want, code, out, stderr)
				}
				if !strings.Contains(out, "next test") || !strings.Contains(out, "1 passed, 1 failed") {
					t.Fatalf("the run did not go on after the failure:\n%s", out)
				}
			})
		}
	}
}

// TestRunStdTestBuilds checks that a file whose tests use std/test still
// builds and runs as a program: test blocks are not part of that build.
func TestRunStdTestBuilds(t *testing.T) {
	src := filepath.Join(t.TempDir(), "prog.nra")
	text := `import * as Test from "std/test"

test "a" {
  Test.eq(1, 1)
}

export function run(args: Array<string>): int {
  print("main")
  return 0
}
`
	if err := os.WriteFile(src, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	out, stderr, code := runProgram(t, src, modes[0].opts)
	if code != 0 || out != "main\n" {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, stderr)
	}
}
