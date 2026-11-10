"""Variant 1, runtime only: the panic is noreturn and cold, the bounds failure and
the copy-on-write copy are out of line, so Array_at and Array_get inline as a few
instructions. Usage: python3 variant1.py in.c out.c"""
import re
import sys

text = open(sys.argv[1]).read()

text = text.replace(
    "static void narra_panic(const char* where, const char* msg) {",
    "__attribute__((noreturn, cold, noinline)) static void narra_panic(const char* where, const char* msg) {",
    1,
)
text = text.replace(
    """static void narra_bounds(int64_t i, int64_t len, const char* where) {
  if (i < 0 || i >= len) {
    char msg[96];""",
    """__attribute__((noreturn, cold, noinline)) static void narra_bounds_fail(int64_t i, int64_t len, const char* where) {
  {
    char msg[96];""",
)
text = text.replace(
    """    narra_panic(where, msg);
  }
}
""",
    """    narra_panic(where, msg);
  }
}
static inline void narra_bounds(int64_t i, int64_t len, const char* where) {
  if (__builtin_expect((uint64_t)i >= (uint64_t)len, 0)) narra_bounds_fail(i, len, where);
}
""",
    1,
)

# Split each Array_unique into a fast check and an out-of-line copy.
pattern = re.compile(
    r"static (A_Array_\w+)\* (A_Array_\w+)_unique\(\1\* a\) \{\n  if \(a->rc == 1\) return a;\n"
)


def split(match):
    t, name = match.group(1), match.group(2)
    return (
        f"static {t}* {name}_copy({t}* a);\n"
        f"static inline {t}* {name}_unique({t}* a) {{\n"
        f"  if (__builtin_expect(a->rc == 1, 1)) return a;\n"
        f"  return {name}_copy(a);\n"
        f"}}\n"
        f"__attribute__((noinline)) static {t}* {name}_copy({t}* a) {{\n"
    )


text, count = pattern.subn(split, text)
assert count > 0, "no unique function found"
open(sys.argv[2], "w").write(text)
print("split", count, "unique functions")
