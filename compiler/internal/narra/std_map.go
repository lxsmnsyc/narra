package narra

import (
	_ "embed"
	"fmt"
	"strings"
)

// Map<K, V>: a hash map that keeps insertion order.
//
// The checker side (mapOf, mapKeyProblem, the index and literal checks) and
// the emitter side (struct, functions, loops) are both in this file. The C
// for one map type is the template in runtime_map.h.
//
//   - m[k] is Option<V>. m[k] = v stores. Map.remove(mutable m, k) reports
//     whether k was there. Map.size(m) is the count.
//   - Map.keys(m) and Map.entries(m) are special-cased in for ... of, as
//     Array.values is: the loop holds the map, so a write to the variable in
//     the body copies and the loop sees the old entries. Outside a loop they
//     build an Array of the keys, or of #[key, value] tuples.
//   - Copy on write: a map is shared by copying the variable, and _unique
//     copies it before a write when the count is above one.

//go:embed runtime_map.h
var runtimeMapH string

func init() {
	registerNamespace("Map", map[string]*Builtin{
		"size": {
			Check: checkMapSize,
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				return fmt.Sprintf("(%s)->len", e.argCode(x, 0)), false
			},
		},
		"remove": {
			Check: checkMapRemove, Mutable: []bool{true, false},
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				mt := x.Args[0].Value.Type().(*MapType)
				k := e.argCode(x, 1)
				p := e.lvaluePtr(x.Args[0].Value)
				return e.temp(TBool, fmt.Sprintf("%s_remove(%s, %s)", mapName(mt), p, k)), false
			},
		},
		"keys": {
			Check: checkMapIter, Effects: []string{"alloc"},
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) { return e.mapToArray(x, false) },
		},
		"entries": {
			Check: checkMapIter, Effects: []string{"alloc"},
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) { return e.mapToArray(x, true) },
		},
	})
}

func mapName(m *MapType) string { return strings.TrimSuffix(m.CName(), "*") }

// userMap reports whether the module declares its own type named Map, which
// then hides the builtin.
func (c *Checker) userMap() bool { return c.generics["Map"] != nil || c.types["Map"] != nil }

func (c *Checker) mapOf(key, val Type) *MapType {
	id := typeID(key) + "|" + typeID(val)
	if m, ok := c.maps[id]; ok {
		return m
	}
	m := &MapType{Key: key, Val: val}
	c.maps[id] = m
	return m
}

// mapKeyProblem explains why t cannot be a map key, or returns "". A key
// needs Equatable and Hashable. A user Equatable without a user Hashable
// keeps the field-based hash, so equal keys could hash differently: both or
// neither must be written.
func (c *Checker) mapKeyProblem(t Type, seen map[Type]bool) string {
	if seen[t] {
		return ""
	}
	seen[t] = true
	switch t := t.(type) {
	case *PrimType:
		if t == TF64 {
			return "float keys are not allowed"
		}
		if t == TVoid {
			return "void has no value"
		}
	case *ArrayType:
		return c.mapKeyProblem(t.Elem, seen)
	case *BoxType:
		return c.mapKeyProblem(t.Elem, seen)
	case *FuncType:
		return "a function is not hashable"
	case *MapType:
		return "a map is not hashable"
	case *StructType, *EnumType:
		for _, f := range fieldsOf(t) {
			if p := c.mapKeyProblem(f.T, seen); p != "" {
				return p
			}
		}
		if info := c.traits.types[typeKeyOf(t)]; info != nil {
			eq, hash := c.stdTrait("Equatable"), c.stdTrait("Hashable")
			if (info.impls[eq] != nil) != (info.impls[hash] != nil) {
				return "it implements one of Equatable and Hashable but not the other"
			}
		}
	}
	return ""
}

// checkMapKey reports a key type that cannot be hashed. Keys that still hold
// type parameters are checked when the generic is used.
func (c *Checker) checkMapKey(pos Pos, key Type) {
	if containsTypeParam(key) {
		return
	}
	if p := c.mapKeyProblem(key, map[Type]bool{}); p != "" {
		c.errorf(pos, "%s cannot be a map key: %s", key, p)
	}
}

// mapIndex checks m[k]: Option<V>, or V when it is the target of an assignment.
func (c *Checker) mapIndex(e *IndexExpr, mt *MapType) Type {
	c.expectType(e.Index.position(), mt.Key, c.expr(e.Index, mt.Key))
	if e == c.assigning {
		e.MapSet = true
		return mt.Val
	}
	return preludeEnum(c, "Option", mt.Val)
}

func (c *Checker) mapLit(e *MapLit, want Type) Type {
	var key, val Type
	if mt, ok := want.(*MapType); ok {
		key, val = mt.Key, mt.Val
	}
	for i := range e.Keys {
		kt := c.expr(e.Keys[i], key)
		if key == nil {
			key = kt
		}
		c.expectType(e.Keys[i].position(), key, kt)
		vt := c.expr(e.Vals[i], val)
		if val == nil {
			val = vt
		}
		c.expectType(e.Vals[i].position(), val, vt)
	}
	if key == nil || val == nil {
		c.errorf(e.P, "the type of Map {} is unknown; add a type annotation")
		return TVoid
	}
	c.checkMapKey(e.P, key)
	return c.mapOf(key, val)
}

func (c *Checker) mapArg(e *CallExpr, i int) *MapType {
	t := c.expr(e.Args[i].Value, nil)
	mt, ok := t.(*MapType)
	if !ok {
		c.errorf(e.P, "%s needs a map, found %s", e.Builtin, t)
		return nil
	}
	return mt
}

func checkMapSize(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 1) {
		return TVoid
	}
	c.mapArg(e, 0)
	return TInt
}

func checkMapRemove(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	if !e.Args[0].Mutable {
		c.errorf(e.P, "Map.remove changes the map; write Map.remove(mutable m, key)")
	}
	mt := c.mapArg(e, 0)
	c.checkMutableArg(e.Args[0].Value)
	if mt == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	c.expectType(e.Args[1].Value.position(), mt.Key, c.expr(e.Args[1].Value, mt.Key))
	c.checkExclusive(e, []int{0})
	return TBool
}

// checkMapIter checks Map.keys and Map.entries.
func checkMapIter(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 1) {
		return TVoid
	}
	mt := c.mapArg(e, 0)
	if mt == nil {
		return TVoid
	}
	if e.Builtin == "Map.keys" {
		return c.arrayOf(mt.Key)
	}
	return c.arrayOf(c.tupleOf([]Type{mt.Key, mt.Val}))
}

// ---- emitter ----

// mapFuncs writes the C for one map type. It runs before the trait functions
// are written, because it asks for the key's eq and hash.
func (e *Emitter) mapFuncs(b *strings.Builder, m *MapType) {
	r := strings.NewReplacer(
		"@M@", mapName(m),
		"@K@", m.Key.CName(),
		"@V@", m.Val.CName(),
		"@EQ@", e.useEq(m.Key),
		"@HASH@", e.useHash(m.Key),
		"@REL_KEY_I@", release(m.Key, "m->keys[i]"),
		"@REL_VAL_I@", release(m.Val, "m->vals[i]"),
		"@RET_KEY_C@", retain(m.Key, "c->keys[i]"),
		"@RET_VAL_C@", retain(m.Val, "c->vals[i]"),
		"@REL_KEY_ARG@", release(m.Key, "k"),
		"@REL_KEY_E@", release(m.Key, "m->keys[e]"),
		"@REL_VAL_E@", release(m.Val, "m->vals[e]"),
	)
	b.WriteString("\n")
	b.WriteString(r.Replace(runtimeMapH))
}

func (e *Emitter) mapTypedefs(b *strings.Builder) {
	for _, m := range e.prog.Maps {
		fmt.Fprintf(b, "typedef struct %s %s;\n", mapName(m), mapName(m))
	}
}

func (e *Emitter) mapStructs(b *strings.Builder) {
	for _, m := range e.prog.Maps {
		fmt.Fprintf(b, "struct %s { int64_t rc; int64_t len; int64_t used; int64_t cap; int64_t mask; %s* keys; %s* vals; uint64_t* hashes; uint8_t* alive; int64_t* idx; };\n",
			mapName(m), m.Key.CName(), m.Val.CName())
	}
}

// mapProtos declares what the reference-counting functions of other types call.
func (e *Emitter) mapProtos(b *strings.Builder) {
	for _, m := range e.prog.Maps {
		n := mapName(m)
		fmt.Fprintf(b, "static void %s_retain(%s* a);\nstatic void %s_release(%s* a);\n", n, n, n, n)
		fmt.Fprintf(b, "static %[1]s* %[1]s_new(int64_t n);\nstatic int64_t %[1]s_lookup(%[1]s* m, %[2]s k);\n", n, m.Key.CName())
		fmt.Fprintf(b, "static void %[1]s_set(%[1]s** p, %[2]s k, %[3]s v);\nstatic bool %[1]s_remove(%[1]s** p, %[2]s k);\n", n, m.Key.CName(), m.Val.CName())
	}
}

// mapLitCode emits Map { k: v, ... }.
func (e *Emitter) mapLitCode(x *MapLit) (string, bool) {
	mt := x.Type().(*MapType)
	m := e.ownedTemp(mt, fmt.Sprintf("%s_new(%d)", mapName(mt), len(x.Keys)))
	for i := range x.Keys {
		kc, ko := e.expr(x.Keys[i])
		k := e.consume(kc, ko, mt.Key)
		vc, vo := e.expr(x.Vals[i])
		v := e.consume(vc, vo, mt.Val)
		e.line("%s_set(&%s, %s, %s);", mapName(mt), m, k, v)
	}
	return m, true
}

// mapGet emits m[k] as an Option<V>.
func (e *Emitter) mapGet(x *IndexExpr) (string, bool) {
	mt := x.Target.Type().(*MapType)
	opt := x.Type().(*EnumType)
	tc, _ := e.expr(x.Target)
	m := e.temp(mt, tc)
	kc, _ := e.expr(x.Index)
	i := e.temp(TInt, fmt.Sprintf("%s_lookup(%s, %s)", mapName(mt), m, kc))
	e.tempN++
	out := fmt.Sprintf("t%d", e.tempN)
	e.line("%s %s;", opt.CName(), out)
	e.line("if (%s >= 0) {", i)
	e.indent++
	v := e.temp(mt.Val, fmt.Sprintf("%s->vals[%s]", m, i))
	if r := retain(mt.Val, v); r != "" {
		e.line("%s", r)
	}
	e.line("%s = %s;", out, e.enumAlloc(variantOf(opt, "Some"), []string{v}))
	e.indent--
	e.line("} else {")
	e.indent++
	e.line("%s = (%s){.tag = %d};", out, opt.CName(), variantOf(opt, "None").Index)
	e.indent--
	e.line("}")
	e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{out, opt})
	return out, true
}

// mapSet emits m[k] = v.
func (e *Emitter) mapSet(target *IndexExpr, value Expr) {
	mt := target.Target.Type().(*MapType)
	code, owned := e.expr(value)
	v := e.consume(code, owned, mt.Val)
	kc, ko := e.expr(target.Index)
	k := e.consume(kc, ko, mt.Key)
	p := e.lvaluePtr(target.Target)
	e.line("%s_set(%s, %s, %s);", mapName(mt), p, k, v)
}

// mapEntryValue is the C code for entry i of map m: the key, or a
// #[key, value] tuple.
func (e *Emitter) mapEntryValue(m, i string, elem Type, entries bool) string {
	if !entries {
		return fmt.Sprintf("%s->keys[%s]", m, i)
	}
	st := elem.(*StructType)
	return fmt.Sprintf("((%s){.%s = %s->keys[%s], .%s = %s->vals[%s]})", st.CName(),
		fieldName("0"), m, i, fieldName("1"), m, i)
}

// mapToArray emits Map.keys(m) or Map.entries(m) outside a loop.
func (e *Emitter) mapToArray(x *CallExpr, entries bool) (string, bool) {
	mt := x.Args[0].Value.Type().(*MapType)
	at := x.Type().(*ArrayType)
	m := e.temp(mt, e.argCode(x, 0))
	arr := e.ownedTemp(at, fmt.Sprintf("%s_new(%s->len)", arrName(at), m))
	e.tempN++
	i := fmt.Sprintf("i%d", e.tempN)
	e.line("for (int64_t %s = 0; %s < %s->used; %s++) {", i, i, m, i)
	e.indent++
	e.line("if (!%s->alive[%s]) continue;", m, i)
	v := e.temp(at.Elem, e.mapEntryValue(m, i, at.Elem, entries))
	if r := retain(at.Elem, v); r != "" {
		e.line("%s", r)
	}
	e.line("%s->items[%s->len++] = %s;", arr, arr, v)
	e.indent--
	e.line("}")
	return arr, true
}

// mapFor emits `for (const x of Map.keys(m))` and Map.entries.
func (e *Emitter) mapFor(s *ForStmt, call *CallExpr, entries bool) {
	mt := call.Args[0].Value.Type().(*MapType)
	code, owned := e.expr(call.Args[0].Value)
	m := e.consume(code, owned, mt)
	e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{m, mt})
	e.tempN++
	idx := fmt.Sprintf("i%d", e.tempN)
	e.line("for (int64_t %s = 0; %s < %s->used; %s++) {", idx, idx, m, idx)
	e.indent++
	e.line("if (!%s->alive[%s]) continue;", m, idx)
	e.loopTemp = append(e.loopTemp, len(e.frames))
	e.pushScope(true)
	e.line("%s %s = %s;", s.Local.T.CName(), s.Local.CName, e.mapEntryValue(m, idx, s.Local.T, entries))
	if r := retain(s.Local.T, s.Local.CName); r != "" {
		e.line("%s", r)
	}
	e.addLocal(s.Local)
	e.blockInScope(s.Body)
	e.popScope()
	e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
	e.indent--
	e.line("}")
}
