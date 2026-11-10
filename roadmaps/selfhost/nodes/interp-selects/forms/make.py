#!/usr/bin/env python3
# Writes the hand-edited forms of bench/interp into OUT from SRC/inline.c and SRC/heap.c,
# the C that narrac emits after and before D99. gen.sh writes SRC.
# Usage: make.py SRC OUT
import os
import re
import sys

source_dir = sys.argv[1]
out = sys.argv[2]
os.makedirs(out, exist_ok=True)

inline = open(os.path.join(source_dir, "inline.c")).read()
heap = open(os.path.join(source_dir, "heap.c")).read()

step_pattern = re.compile(
    r"static int64_t f_45cc5e7f_step\([^)]*\) \{.*?\n\}\n",
    re.S,
)

# A switch with each payload read inside its case.
switch_step = """static int64_t f_45cc5e7f_step(E_45cc5e7f_Op v_op_1, int64_t v_acc_2) {
  switch (v_op_1.tag) {
  case 0: return narra_wrap_add(v_acc_2, v_op_1.u.v_Add.f_n);
  case 1: return narra_wrap_sub(v_acc_2, v_op_1.u.v_Sub.f_n);
  case 2: return narra_wrap_mul(v_acc_2, v_op_1.u.v_Mul.f_n);
  case 3: return narra_rem(v_acc_2, v_op_1.u.v_Mod.f_n, "interp.nra:18:22");
  default: return v_acc_2;
  }
}
"""


def barrier(source, keyword):
    # An empty asm on each payload binding, right after the read.
    def add(match):
        return match.group(0) + f" __asm__ {keyword}(\"\" : \"+r\"({match.group(1)}));"

    body = step_pattern.search(source).group(0)
    edited = re.sub(r"int64_t (v_n_\d+) = t2\.u\.v_\w+\.f_n;", add, body)
    return source.replace(body, edited)


# No copy between the array and the arm: the loop passes the element's address, and the
# emitted if-chain reads the tag and payload through it (what Rust's `step(op: &Op)` does).
def by_pointer(source):
    source = source.replace(
        "static int64_t f_45cc5e7f_step(E_45cc5e7f_Op v_op_1, int64_t v_acc_2);",
        "static int64_t f_45cc5e7f_step(const E_45cc5e7f_Op* v_op_1, int64_t v_acc_2);",
    )
    body = step_pattern.search(source).group(0)
    edited = body.replace("E_45cc5e7f_Op v_op_1,", "const E_45cc5e7f_Op* v_op_1,")
    edited = edited.replace("E_45cc5e7f_Op t2 = v_op_1;", "const E_45cc5e7f_Op* t2 = v_op_1;")
    edited = edited.replace("t2.", "t2->")
    source = source.replace(body, edited)
    return source.replace(
        "f_45cc5e7f_step(v_op_13, v_acc_11)", "f_45cc5e7f_step(&t26->items[i27], v_acc_11)"
    )


# Only the match reads through a pointer; the parameter and the loop item stay copies.
def match_pointer(source):
    body = step_pattern.search(source).group(0)
    edited = body.replace("E_45cc5e7f_Op t2 = v_op_1;", "const E_45cc5e7f_Op* t2 = &v_op_1;")
    edited = edited.replace("t2.", "t2->")
    return source.replace(body, edited)


# Only the parameter is a pointer; the match copies the value into its subject temp.
def parameter_pointer(source):
    source = by_pointer(source)
    body = step_pattern.search(source).group(0)
    edited = body.replace("const E_45cc5e7f_Op* t2 = v_op_1;", "E_45cc5e7f_Op t2 = *v_op_1;")
    edited = edited.replace("t2->", "t2.")
    return source.replace(body, edited)


# Program order. "cycle" is the bench: a five-op cycle. "random" picks each op from a
# splitmix64 hash of its index (1000 ops, 150000 rounds); the predictor learns it.
# "wide" uses the same hash over 1000000 ops and 150 rounds, too long to learn.
# "addsub" is defined below.
mix = """static uint64_t mix(uint64_t x) {
  x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9ull;
  x = (x ^ (x >> 27)) * 0x94d049bb133111ebull;
  return x ^ (x >> 31);
}

static int64_t f_45cc5e7f_run(A_Array_5fce1d7c* v_args_7) {"""


def add_mix(source):
    return source.replace("static int64_t f_45cc5e7f_run(A_Array_5fce1d7c* v_args_7) {", mix, 1)


def randomize(source):
    return add_mix(source).replace(
        'narra_rem(v_i_9, INT64_C(5), "interp.nra:26:17")',
        'narra_rem((int64_t)(mix((uint64_t)v_i_9) >> 1), INT64_C(5), "interp.nra:26:17")',
    )


forms = {
    "inline": inline,
    "heap": heap,
    "switch": step_pattern.sub(lambda m: switch_step, inline, count=1),
    "asm": barrier(inline, ""),
    "asmvolatile": barrier(inline, "volatile"),
    "pointer": by_pointer(inline),
    "matchpointer": match_pointer(inline),
    "parampointer": parameter_pointer(inline),
}


# Only Add and Sub, picked at random over 1000000 ops: the case where selects beat
# branches, because a branch here mispredicts half the time.
def add_sub(source):
    source = add_mix(source).replace(
        'narra_rem(v_i_9, INT64_C(5), "interp.nra:26:17")',
        '2 * (int64_t)(mix((uint64_t)v_i_9) & 1)',
    )
    source = source.replace("int64_t t15 = INT64_C(1000);", "int64_t t15 = INT64_C(1000000);")
    return source.replace("= INT64_C(150000);", "= INT64_C(150);")


def widen(source):
    source = randomize(source)
    source = source.replace("int64_t t15 = INT64_C(1000);", "int64_t t15 = INT64_C(1000000);")
    return source.replace("= INT64_C(150000);", "= INT64_C(150);")


for name, source in forms.items():
    orders = (
        ("cycle", source),
        ("random", randomize(source)),
        ("wide", widen(source)),
        ("addsub", add_sub(source)),
    )
    for order, text in orders:
        with open(os.path.join(out, f"{name}-{order}.c"), "w") as file:
            file.write(text)
