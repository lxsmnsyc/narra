package narra

import "fmt"

// runtimeInt is the C of Int.parse and Int.toHex.
const runtimeInt = `
// Parses s in the given base into *out. On failure sets *err and returns false.
static bool narra_int_parse(lstr* s, int64_t base, int64_t* out, lstr** err) {
  if (base < 2 || base > 36) { *err = lstr_from_cstr("base must be between 2 and 36"); return false; }
  int64_t i = 0;
  bool neg = false;
  if (s->len > 0 && s->data[0] == '-') { neg = true; i = 1; }
  if (i >= s->len) { *err = lstr_from_cstr(s->len == 0 ? "empty string" : "no digits after '-'"); return false; }
  uint64_t limit = neg ? (uint64_t)INT64_MAX + 1 : (uint64_t)INT64_MAX;
  uint64_t acc = 0;
  for (; i < s->len; i++) {
    char ch = s->data[i];
    int64_t d;
    if (ch >= '0' && ch <= '9') d = ch - '0';
    else if (ch >= 'a' && ch <= 'z') d = ch - 'a' + 10;
    else if (ch >= 'A' && ch <= 'Z') d = ch - 'A' + 10;
    else d = 99;
    if (d >= base) { *err = lstr_from_cstr("invalid digit for the base"); return false; }
    if (acc > (limit - (uint64_t)d) / (uint64_t)base) { *err = lstr_from_cstr("number does not fit in 64 bits"); return false; }
    acc = acc * (uint64_t)base + (uint64_t)d;
  }
  *out = neg ? (int64_t)(0 - acc) : (int64_t)acc;
  return true;
}

static lstr* narra_int_to_hex(int64_t n) {
  char buf[24];
  uint64_t m = n < 0 ? 0 - (uint64_t)n : (uint64_t)n;
  int pos = (int)sizeof buf;
  buf[--pos] = 0;
  do { buf[--pos] = "0123456789abcdef"[m & 15]; m >>= 4; } while (m != 0);
  if (n < 0) buf[--pos] = '-';
  return lstr_from_cstr(buf + pos);
}
`

func init() {
	registerModule(&Module{
		Path:    "std/int",
		Runtime: runtimeInt,
		Funcs: map[string]*Builtin{
			"toString": {
				Name: "Int.toString", Params: []Type{TInt}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_int(%s)", e.argCode(x, 0))), true
				},
			},
			// parse(s, base): bases 2 to 36, optional leading `-`, no prefix.
			"parse": {
				Name: "Int.parse", Effects: []string{"alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr, TInt}, resultOf(c, TInt))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					s := e.argCode(x, 0)
					base := e.argCode(x, 1)
					n := e.temp(TInt, "0")
					bad := stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_int_parse(%s, %s, &%s, &%s)", s, base, n, bad))
					return ioResult(e, x.Type().(*EnumType), ok, n, bad), true
				},
			},
			// toHex(n): lowercase digits, no prefix, `-` for negatives.
			"toHex": {
				Name: "Int.toHex", Params: []Type{TInt}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.ownedTemp(TStr, fmt.Sprintf("narra_int_to_hex(%s)", e.argCode(x, 0))), true
				},
			},
			"abs": {
				Name: "Int.abs", Params: []Type{TInt}, Result: TInt,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.temp(TInt, fmt.Sprintf("narra_abs(%s, %s)", e.argCode(x, 0), w)), false
				},
			},
		},
	})
}
