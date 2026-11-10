"""Hand-written emitter variants on top of variant 1 (the files given are v1 C).

v3: the array is made unique once before the loop that writes it; the write in the
    loop is a bounds check and a store (no rc check, no call on any path).
v4: v3 plus header fields typed `long`, which TBAA keeps apart from int64_t (long long
    on macOS) elements, so clang may keep len and items in registers.
v5: v3 plus len and items read once into C locals before the loop (what the emitter
    would emit when the loop body cannot resize the array).
Usage: python3 variant3.py sieve-v1.c particles-v1.c"""
import re
import sys

sieve_path, particles_path = sys.argv[1], sys.argv[2]


def long_header(text):
    return re.sub(
        r"^struct (A_Array_[0-9a-f]+) \{ int64_t rc; int64_t len; int64_t cap;",
        r"struct \1 { long rc; long len; long cap;",
        text,
        flags=re.M,
    )


def replace_once(text, old, new):
    assert text.count(old) == 1, old
    return text.replace(old, new)


sieve = open(sieve_path).read()
sieve_old_loop = """      while (1) {
        bool t11 = (v_j_4 < v_n_1);
        if (!t11) break;
        int64_t t12 = v_j_4;
        int64_t* t13 = A_Array_e7fe1e37_at(&v_marks_2, t12, "sieve.nra:14:14");
        *t13 = INT64_C(0);"""
sieve_v3 = replace_once(
    sieve,
    sieve_old_loop,
    """      v_marks_2 = A_Array_e7fe1e37_unique(v_marks_2);
      while (1) {
        bool t11 = (v_j_4 < v_n_1);
        if (!t11) break;
        int64_t t12 = v_j_4;
        narra_bounds(t12, v_marks_2->len, "sieve.nra:14:14");
        int64_t* t13 = &v_marks_2->items[t12];
        *t13 = INT64_C(0);""",
)
sieve_v5 = replace_once(
    sieve,
    sieve_old_loop,
    """      v_marks_2 = A_Array_e7fe1e37_unique(v_marks_2);
      int64_t* items_2 = v_marks_2->items;
      int64_t len_2 = v_marks_2->len;
      while (1) {
        bool t11 = (v_j_4 < v_n_1);
        if (!t11) break;
        int64_t t12 = v_j_4;
        narra_bounds(t12, len_2, "sieve.nra:14:14");
        int64_t* t13 = &items_2[t12];
        *t13 = INT64_C(0);""",
)

particles = open(particles_path).read()
particles_old_loop = """    for (int64_t v_i_10 = 0; v_i_10 < t23; v_i_10++) {
      S_6d08b998_Particle t24 = A_Array_186399d5_get(v_particles_7, v_i_10, "particles.nra:33:36");
      S_6d08b998_Particle t25 = f_6d08b998_move(&t24);
      int64_t t26 = v_i_10;
      S_6d08b998_Particle* t27 = A_Array_186399d5_at(&v_particles_7, t26, "particles.nra:33:16");"""
particles_v3 = replace_once(
    particles,
    particles_old_loop,
    """    v_particles_7 = A_Array_186399d5_unique(v_particles_7);
    for (int64_t v_i_10 = 0; v_i_10 < t23; v_i_10++) {
      S_6d08b998_Particle t24 = A_Array_186399d5_get(v_particles_7, v_i_10, "particles.nra:33:36");
      S_6d08b998_Particle t25 = f_6d08b998_move(&t24);
      int64_t t26 = v_i_10;
      narra_bounds(t26, v_particles_7->len, "particles.nra:33:16");
      S_6d08b998_Particle* t27 = &v_particles_7->items[t26];""",
)
particles_v5 = replace_once(
    particles,
    particles_old_loop,
    """    v_particles_7 = A_Array_186399d5_unique(v_particles_7);
    S_6d08b998_Particle* items_7 = v_particles_7->items;
    int64_t len_7 = v_particles_7->len;
    for (int64_t v_i_10 = 0; v_i_10 < t23; v_i_10++) {
      narra_bounds(v_i_10, len_7, "particles.nra:33:36");
      S_6d08b998_Particle t24 = items_7[v_i_10];
      S_6d08b998_Particle t25 = f_6d08b998_move(&t24);
      int64_t t26 = v_i_10;
      narra_bounds(t26, len_7, "particles.nra:33:16");
      S_6d08b998_Particle* t27 = &items_7[t26];""",
)

outputs = {
    "sieve-v3.c": sieve_v3,
    "sieve-v4.c": long_header(sieve_v3),
    "sieve-v5.c": sieve_v5,
    "particles-v3.c": particles_v3,
    "particles-v4.c": long_header(particles_v3),
    "particles-v5.c": particles_v5,
}
for name, text in outputs.items():
    open(name, "w").write(text)
print("wrote", ", ".join(outputs))
