package narra

import "fmt"

// std/internal/...: the primitives behind std modules written in Narra. Only a
// module whose path starts with std/ may import them (check.go), and
// moduleList does not name them.

func init() {
	// probe is a test-only primitive that proves the import rule. It is
	// the same code as Char.code.
	registerModule(&Module{
		Path: "std/internal/probe",
		Funcs: map[string]*Builtin{
			"code": {
				Name: "Probe.code", Params: []Type{TChar}, Result: TInt,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return fmt.Sprintf("((int64_t)%s)", e.argCode(x, 0)), false
				},
			},
		},
	})
}
