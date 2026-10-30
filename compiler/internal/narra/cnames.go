package narra

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// C names. Every name the emitter writes for a user declaration or a
// composite type is made in this file.
//
// - A declaration is <kind>_<module hash>_<name>: f_3fa9c1e7_show. The
//   prelude (module "") has no hash: f_show, E_Ordering.
// - A composite type (generic instance, tuple, function type, array) is
//   interned: it is named once, and a bigger type refers to its parts by
//   their names. Its name is <kind>_<head>_<hash>, where the hash folds the
//   head, the arity and the names of the arguments in order.
// - The table of names stops the compiler if two different things get one
//   name.

const (
	xxP1 uint32 = 2654435761
	xxP2 uint32 = 2246822519
	xxP3 uint32 = 3266489917
	xxP4 uint32 = 668265263
	xxP5 uint32 = 374761393
)

func rotl32(x uint32, r uint) uint32 { return x<<r | x>>(32-r) }

// le32 reads a little-endian word, whatever the host.
func le32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func xxRound(acc, in uint32) uint32 { return rotl32(acc+in*xxP2, 13) * xxP1 }

// xxh32 is xxHash32 of data with a seed.
func xxh32(data []byte, seed uint32) uint32 {
	n := len(data)
	i := 0
	var h uint32
	if n >= 16 {
		v1 := seed + xxP1 + xxP2
		v2 := seed + xxP2
		v3 := seed
		v4 := seed - xxP1
		for ; i+16 <= n; i += 16 {
			v1 = xxRound(v1, le32(data[i:]))
			v2 = xxRound(v2, le32(data[i+4:]))
			v3 = xxRound(v3, le32(data[i+8:]))
			v4 = xxRound(v4, le32(data[i+12:]))
		}
		h = rotl32(v1, 1) + rotl32(v2, 7) + rotl32(v3, 12) + rotl32(v4, 18)
	} else {
		h = seed + xxP5
	}
	h += uint32(n)
	for ; i+4 <= n; i += 4 {
		h = rotl32(h+le32(data[i:])*xxP3, 17) * xxP4
	}
	for ; i < n; i++ {
		h = rotl32(h+uint32(data[i])*xxP5, 11) * xxP1
	}
	h ^= h >> 15
	h *= xxP2
	h ^= h >> 13
	h *= xxP3
	h ^= h >> 16
	return h
}

func hex8(h uint32) string {
	s := strconv.FormatUint(uint64(h), 16)
	return strings.Repeat("0", 8-len(s)) + s
}

// moduleHash is the hash segment of a module's names. The input is the
// module path relative to the package ("app/text" gives "text"), so it
// does not depend on where the package is on disk. A module with no "/"
// (a single file) hashes its whole path.
func moduleHash(path string) string { return hex8(xxh32([]byte(relModule(path)), 0)) }

// relModule drops the package name from a module path.
func relModule(path string) string {
	if i := strings.IndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

var cnameTable = struct {
	sync.Mutex
	keys map[string]string
}{keys: map[string]string{}}

// registerCName records what a C name stands for and stops with an internal
// error if the name already stands for something else.
func registerCName(name, key string) string {
	cnameTable.Lock()
	defer cnameTable.Unlock()
	if have, ok := cnameTable.keys[name]; ok && have != key {
		panic(fmt.Sprintf("internal error: C name %s is used for both %q and %q", name, have, key))
	}
	cnameTable.keys[name] = key
	return name
}

// declHead is <module hash>_<name>, or the name alone in the prelude.
func declHead(module, name string) string {
	if module == "" {
		return name
	}
	return moduleHash(module) + "_" + name
}

// declCName names a declaration: kind is "f", "S" or "E".
func declCName(kind, module, name string) string {
	return registerCName(kind+"_"+declHead(module, name), kind+"|"+relModule(module)+"|"+name)
}

// internedName names a composite type from its head and the names of its
// arguments.
func internedName(kind, head string, args []string) string {
	fold := head + "|" + strconv.Itoa(len(args)) + "|" + strings.Join(args, "|")
	name := kind + "_" + head + "_" + hex8(xxh32([]byte(fold), 0))
	return registerCName(name, kind+"|"+fold)
}

// typeID is the C name of a type without a pointer star: the name that
// stands for the type inside a bigger type's name.
func typeID(t Type) string {
	switch t := t.(type) {
	case *PrimType:
		return t.Name
	case *StructType:
		return structCName(t)
	case *EnumType:
		return enumCName(t)
	case *TypeParamType:
		return fmt.Sprintf("P_%s_%d", t.Name, t.ID)
	case *ArrayType:
		return arrayCName(t)
	case *FuncType:
		return funcTypeCName(t)
	}
	return "x"
}

func typeIDs(ts []Type) []string {
	ids := make([]string, len(ts))
	for i, t := range ts {
		ids[i] = typeID(t)
	}
	return ids
}

// typeArgsKey is the key of a list of types in the checker's tables.
func typeArgsKey(ts []Type) string { return strings.Join(typeIDs(ts), "|") }

func structCName(t *StructType) string {
	if t.Tuple {
		return internedName("S", "Tuple"+strconv.Itoa(len(t.Args)), typeIDs(t.Args))
	}
	if len(t.Args) == 0 {
		return declCName("S", t.Module, t.Name)
	}
	return internedName("S", declHead(t.Module, t.Name), typeIDs(t.Args))
}

func enumCName(t *EnumType) string {
	if len(t.Args) == 0 {
		return declCName("E", t.Module, t.Name)
	}
	return internedName("E", declHead(t.Module, t.Name), typeIDs(t.Args))
}

// arrayCName is the struct of an array, without the pointer star.
func arrayCName(t *ArrayType) string {
	return internedName("A", "Array", []string{typeID(t.Elem)})
}

// funcTypeCName is the struct of a function value: the parameters, then
// the result, as arguments of the head Fn<n>.
func funcTypeCName(t *FuncType) string {
	args := append(typeIDs(t.Params), typeID(t.Result))
	return internedName("F", "Fn"+strconv.Itoa(len(t.Params)), args)
}

// funcCName is the C name of a function.
func funcCName(fd *FuncDecl) string { return declCName("f", fd.Module, fd.Name) }

// instanceName is the FuncDecl.Name of a generic function instance.
func instanceName(module, name string, targs []Type) string {
	ids := typeIDs(targs)
	fold := declHead(module, name) + "|" + strconv.Itoa(len(ids)) + "|" + strings.Join(ids, "|")
	inst := name + "_" + hex8(xxh32([]byte(fold), 0))
	registerCName("inst|"+relModule(module)+"|"+inst, fold)
	return inst
}

// implName is the FuncDecl.Name of a function in an `implement` block. The
// function's Module is the type's module. The hash folds the type, the
// trait with its module, and the function, so no separator rule is needed.
func implName(tr *Trait, fn string, t Type) string {
	ty := declaredName(t)
	fold := ty + "|" + relModule(tr.Module) + "|" + tr.Name + "|" + fn
	name := ty + "_" + tr.Name + "_" + fn + "_" + hex8(xxh32([]byte(fold), 0))
	registerCName("impl|"+name, fold)
	return name
}

// opFuncName names a function that the emitter generates for a type, such
// as eq, retain or release. The operation comes first, so no declaration
// name can equal it.
func opFuncName(op string, t Type) string { return op + "_" + typeID(t) }

// closureCName names the C function, environment or drop function of an
// arrow function: kind is "k", "K" or "D". The owner is the C name of the
// enclosing function (f_<hash>_<name>), and n counts its arrows from 1.
func closureCName(kind, owner string, n int) string {
	return fmt.Sprintf("%s%s_%d", kind, owner[1:], n)
}

// thunkCName names the thunk of a named function used as a value.
func thunkCName(owner string) string { return "w" + owner[1:] }

// writeCNameTable writes the comment table that maps each interned name to
// its Narra type, sorted by name.
func writeCNameTable(b *strings.Builder, prog *Program) {
	names := map[string]Type{}
	for _, t := range prog.Types {
		if st, ok := t.(*StructType); ok && (st.Tuple || len(st.Args) > 0) {
			names[st.CName()] = t
		} else if et, ok := t.(*EnumType); ok && len(et.Args) > 0 {
			names[et.CName()] = t
		}
	}
	for _, a := range prog.Arrays {
		names[arrayCName(a)] = a
	}
	for _, ft := range prog.FuncTypes {
		names[ft.CName()] = ft
	}
	keys := make([]string, 0, len(names))
	for k := range names {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, "/* %s = %s */\n", k, names[k].String())
	}
}
