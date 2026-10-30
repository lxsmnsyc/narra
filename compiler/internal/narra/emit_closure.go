package narra

import (
	"fmt"
	"strings"
)

// Function values in C (see check_closure.go for the rules).
//
// A function type such as (int) => int is an interned struct (cnames.go),
// `F_Fn1_<hash> { int64_t (*fn)(narra_env*, int64_t); narra_env* env; }`.
// A call passes the environment first: `f.fn(f.env, x)`.
//
// - A named function used as a value is a thunk `w_<hash>_<name>` (thunkCName)
//   with a NULL environment.
// - An arrow function is a C function `k_<hash>_<name>_<n>` (closureCName), where the
//   owner is the declared function (or test) it appears in and n counts
//   from 1 in that owner. Its environment struct is `K...` and the function
//   that releases its captures is `D...`. The last segment is the count, so
//   the names are injective because funcCName is.
// - Without captures the environment is NULL. A closure passed straight to
//   a non-escaping parameter has its environment on the caller's stack with
//   a count of -1, which retain and release ignore; its captures are
//   borrowed. Any other closure has a heap environment with count 1 that
//   owns a reference to each capture and is freed when the count reaches 0.
// - Function values are reference counted through their environment, like
//   strings: copying one retains env, dropping one releases it.

const envRuntime = `
// Closure environments. A count below zero marks an environment on the
// caller's stack, which is never counted or freed.
typedef struct narra_env { int64_t rc; void (*drop)(struct narra_env*); } narra_env;
static void narra_env_retain(narra_env* e) {
  if (e && e->rc >= 0) e->rc++;
}
static void narra_env_release(narra_env* e) {
  if (e && e->rc >= 0 && --e->rc == 0) {
    e->drop(e);
    narra_free(e);
  }
}
`

type closureEmit struct {
	counts map[string]int
	thunks map[string]bool
	envs   strings.Builder // environment structs and drop functions
	protos strings.Builder
	defs   strings.Builder
}

func (e *Emitter) cl() *closureEmit {
	if e.closures == nil {
		e.closures = &closureEmit{counts: map[string]int{}, thunks: map[string]bool{}}
	}
	return e.closures
}

// funcTypeDefs writes the environment runtime and the struct of every
// function type. Called after all types are declared and before any is
// defined, since structs hold function values inline.
func (e *Emitter) funcTypeDefs(b *strings.Builder) {
	if len(e.prog.FuncTypes) == 0 {
		return
	}
	b.WriteString(envRuntime)
	for _, ft := range e.prog.FuncTypes {
		fmt.Fprintf(b, "typedef struct %s %s;\n", ft.CName(), ft.CName())
	}
	for _, ft := range e.prog.FuncTypes {
		params := []string{"narra_env*"}
		for _, p := range ft.Params {
			params = append(params, p.CName())
		}
		fmt.Fprintf(b, "struct %s { %s (*fn)(%s); narra_env* env; };\n", ft.CName(), ft.Result.CName(), strings.Join(params, ", "))
	}
}

// closureFuncs writes the environments, thunks and arrow functions made
// while the function bodies were emitted.
func (e *Emitter) closureFuncs(b *strings.Builder) {
	if e.closures == nil {
		return
	}
	b.WriteString("\n// Closures\n")
	b.WriteString(e.closures.envs.String())
	b.WriteString(e.closures.protos.String())
	b.WriteString(e.closures.defs.String())
}

// funcValueCode returns a named function as a value.
func (e *Emitter) funcValueCode(fd *FuncDecl, ft *FuncType) (string, bool) {
	name := thunkCName(funcCName(fd))
	st := e.cl()
	if !st.thunks[name] {
		st.thunks[name] = true
		params := []string{"narra_env* env"}
		var args []string
		for i, p := range fd.Params {
			params = append(params, fmt.Sprintf("%s p%d", p.T.CName(), i))
			args = append(args, fmt.Sprintf("p%d", i))
		}
		sig := fmt.Sprintf("static %s %s(%s)", fd.ResultT.CName(), name, strings.Join(params, ", "))
		fmt.Fprintf(&st.protos, "%s;\n", sig)
		ret := "return "
		if fd.ResultT == TVoid {
			ret = ""
		}
		fmt.Fprintf(&st.defs, "%s {\n  (void)env;\n  %s%s(%s);\n}\n\n", sig, ret, funcCName(fd), strings.Join(args, ", "))
	}
	return fmt.Sprintf("((%s){.fn = %s, .env = NULL})", ft.CName(), name), false
}

// arrow emits an arrow function's C function and returns the value.
func (e *Emitter) arrow(x *ArrowExpr) (string, bool) {
	clo := x.Closure
	ft := clo.Type
	st := e.cl()
	owner := funcCName(e.fn)
	st.counts[owner]++
	n := st.counts[owner]
	name, envT, drop := closureCName("k", owner, n), closureCName("K", owner, n), closureCName("D", owner, n)
	if len(clo.Captures) > 0 {
		e.envStruct(envT, drop, clo)
	}
	e.closureFunc(name, envT, clo)
	if len(clo.Captures) == 0 {
		return fmt.Sprintf("((%s){.fn = %s, .env = NULL})", ft.CName(), name), false
	}
	e.tempN++
	env := fmt.Sprintf("k%d", e.tempN)
	if clo.Stack {
		parts := []string{".h = {-1, NULL}"}
		for _, l := range clo.Captures {
			parts = append(parts, fmt.Sprintf(".c_%s = %s", l.CName, e.localRef(l)))
		}
		e.line("%s %s = {%s};", envT, env, strings.Join(parts, ", "))
		return fmt.Sprintf("((%s){.fn = %s, .env = &%s.h})", ft.CName(), name, env), false
	}
	e.line("%s* %s = narra_alloc(sizeof(%s));", envT, env, envT)
	e.line("%s->h.rc = 1;", env)
	e.line("%s->h.drop = %s;", env, drop)
	for _, l := range clo.Captures {
		field := fmt.Sprintf("%s->c_%s", env, l.CName)
		e.line("%s = %s;", field, e.localRef(l))
		if r := retain(l.T, field); r != "" {
			e.line("%s", r)
		}
	}
	return e.ownedTemp(ft, fmt.Sprintf("((%s){.fn = %s, .env = &%s->h})", ft.CName(), name, env)), true
}

func (e *Emitter) envStruct(envT, drop string, clo *Closure) {
	st := e.cl()
	fmt.Fprintf(&st.envs, "typedef struct %s { narra_env h;", envT)
	for _, l := range clo.Captures {
		fmt.Fprintf(&st.envs, " %s c_%s;", l.T.CName(), l.CName)
	}
	fmt.Fprintf(&st.envs, " } %s;\n", envT)
	fmt.Fprintf(&st.envs, "static void %s(narra_env* h) {\n  %s* k = (%s*)h;\n  (void)k;\n", drop, envT, envT)
	for _, l := range clo.Captures {
		if r := release(l.T, "k->c_"+l.CName); r != "" {
			fmt.Fprintf(&st.envs, "  %s\n", r)
		}
	}
	st.envs.WriteString("}\n")
}

// closureFunc emits the C function of an arrow function. Captures are read
// from the environment into locals with their usual C names, borrowed.
func (e *Emitter) closureFunc(name, envT string, clo *Closure) {
	st := e.cl()
	fn := clo.Fn
	params := []string{"narra_env* env"}
	for _, p := range fn.Params {
		params = append(params, fmt.Sprintf("%s %s", p.T.CName(), p.Local.CName))
	}
	sig := fmt.Sprintf("static %s %s(%s)", fn.ResultT.CName(), name, strings.Join(params, ", "))
	fmt.Fprintf(&st.protos, "%s;\n", sig)

	savedOut := e.out.String()
	e.out.Reset()
	savedIndent, savedFrames, savedScopes, savedLoop := e.indent, e.frames, e.scopes, e.loopTemp
	e.indent = 0
	e.line("%s {", sig)
	e.indent++
	if len(clo.Captures) == 0 {
		e.line("(void)env;")
	} else {
		e.line("%s* k_env = (%s*)env;", envT, envT)
		for _, l := range clo.Captures {
			e.line("%s %s = k_env->c_%s;", l.T.CName(), l.CName, l.CName)
		}
	}
	e.funcScope(nil)
	e.blockInScope(fn.Body)
	e.unwindAll()
	if fn.ResultT != TVoid {
		e.line("narra_panic(%s, \"function ended without returning a value\");", e.where(fn.Pos))
		e.line("return (%s){0};", fn.ResultT.CName())
	}
	e.indent--
	e.line("}")
	e.line("")
	st.defs.WriteString(e.out.String())
	e.out.Reset()
	e.out.WriteString(savedOut)
	e.indent, e.frames, e.scopes, e.loopTemp = savedIndent, savedFrames, savedScopes, savedLoop
}

// valueCall emits a call through a function value.
func (e *Emitter) valueCall(x *CallExpr) (string, bool) {
	ft := x.FuncValue
	// The call owns a reference to the callee until the statement ends, so
	// an argument that reassigns the variable holding it cannot free the
	// environment before the call. Stack and NULL environments ignore it.
	code, owned := e.expr(x.Callee)
	f := e.consume(code, owned, ft)
	e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{f, ft})
	args := []string{f + ".env"}
	for _, a := range x.Args {
		c, _ := e.expr(a.Value)
		args = append(args, c)
	}
	callCode := fmt.Sprintf("%s.fn(%s)", f, strings.Join(args, ", "))
	if ft.Result == TVoid {
		return callCode, false
	}
	return e.ownedTemp(ft.Result, callCode), true
}
