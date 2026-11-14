#!/usr/bin/env python3
# Rewrites a C file emitted by narrac into a measurement variant.
# Copy of regions-design/prototype/variant.py for D113: it also matches the
# free(p) that fast-alloc moved under #ifdef NARRA_PLAIN_MALLOC.
# Usage: variant.py <mode> <in.c> <out.c>
#   count    counts allocations, frees, reallocs, retains, releases, copies on write
#   bump     bump allocation, frees do nothing (counts unchanged)
#   nofree   bump, and a release that reaches zero frees nothing (no drop walk)
#   nocount  bump, no retains, no releases (every count stays 1)
#   freelist size-class free lists for blocks up to 512 bytes, malloc above (frees reuse blocks)
#   sign     adds a sign check (rc >= 0) to every array, map, box and enum retain and release
import re
import sys

mode, source, target = sys.argv[1], sys.argv[2], sys.argv[3]
lines = open(source).read().split("\n")

COUNTERS = r"""
static long long narra_n_alloc, narra_n_bytes, narra_n_free, narra_n_realloc;
static long long narra_n_retain, narra_n_release, narra_n_cow;
static void narra_print_counters(void) {
  fprintf(stderr, "alloc %lld bytes %lld free %lld realloc %lld retain %lld release %lld cow %lld\n",
    narra_n_alloc, narra_n_bytes, narra_n_free, narra_n_realloc, narra_n_retain, narra_n_release, narra_n_cow);
}
__attribute__((constructor)) static void narra_counters_init(void) { atexit(narra_print_counters); }
"""

BUMP = r"""
// Bump allocation in 1 MiB chunks; each block keeps its size in front.
static char* narra_bump_at = NULL;
static char* narra_bump_end = NULL;
static void* narra_bump(size_t n) {
  size_t need = ((n ? n : 1) + 8 + 15) & ~(size_t)15;
  if ((size_t)(narra_bump_end - narra_bump_at) < need) {
    size_t chunk = need > (1u << 20) ? need : (1u << 20);
    narra_bump_at = malloc(chunk);
    if (!narra_bump_at) abort();
    narra_bump_end = narra_bump_at + chunk;
  }
  char* p = narra_bump_at;
  narra_bump_at += need;
  *(size_t*)p = n;
  return p + 8;
}
"""

FREELIST = r"""
// Size classes of 16 bytes up to 512; each block keeps its size in front.
static void* narra_fl_heads[33];
static char* narra_fl_at = NULL;
static char* narra_fl_end = NULL;
static void* narra_fl_alloc(size_t n) {
  size_t need = ((n ? n : 1) + 8 + 15) & ~(size_t)15;
  if (need > 512) {
    char* big = malloc(need);
    if (!big) abort();
    *(size_t*)big = n;
    return big + 8;
  }
  size_t k = need >> 4;
  char* p = narra_fl_heads[k];
  if (p) {
    narra_fl_heads[k] = *(void**)(p + 8);
  } else {
    if ((size_t)(narra_fl_end - narra_fl_at) < need) {
      narra_fl_at = malloc(1u << 20);
      if (!narra_fl_at) abort();
      narra_fl_end = narra_fl_at + (1u << 20);
    }
    p = narra_fl_at;
    narra_fl_at += need;
  }
  *(size_t*)p = n;
  return p + 8;
}
static void narra_fl_free(void* q) {
  char* p = (char*)q - 8;
  size_t need = ((*(size_t*)p ? *(size_t*)p : 1) + 8 + 15) & ~(size_t)15;
  if (need > 512) { free(p); return; }
  *(void**)(p + 8) = narra_fl_heads[need >> 4];
  narra_fl_heads[need >> 4] = p;
}
"""

retain_re = re.compile(r"((?:\(\*p\)|[a-z]+(?:\.u)?)->rc)\+\+")
release_re = re.compile(r"--((?:[a-z]+(?:\.u)?)->rc) == 0")

out = []
for line in lines:
    literal = line.startswith("static lstr L")
    if not literal and line.startswith("static void* narra_alloc(size_t n) {"):
        if mode == "count":
            out.append(COUNTERS)
        if mode in ("bump", "nofree", "nocount"):
            out.append("#define NARRA_FREEING 0")
            out.append(BUMP)
        if mode == "freelist":
            out.append(FREELIST)
    if literal:
        out.append(line)
        continue
    if mode == "count":
        if line.strip() == "void* p = malloc(n ? n : 1);":
            line += " narra_n_alloc++; narra_n_bytes += (long long)n;"
        if line.strip() == "free(p);" and out and (out[-1].strip() == "narra_live--;" or (len(out) > 1 and out[-2].strip() == "narra_live--;" and out[-1].strip() == "#ifdef NARRA_PLAIN_MALLOC")):
            line = "  narra_n_free++; free(p);"
        if line.strip() == "void* q = realloc(p, n ? n : 1);":
            line += " narra_n_realloc++;"
        line = retain_re.sub(r"(narra_n_retain++, \1++)", line)
        line = release_re.sub(r"(narra_n_release++, --\1) == 0", line)
        if re.search(r"if \((a|m)->rc == 1\) return \1;", line):
            line += " narra_n_cow++;"
    if mode in ("bump", "nofree", "nocount"):
        if line.strip() == "void* p = malloc(n ? n : 1);":
            line = "  void* p = narra_bump(n);"
        if line.strip() == "free(p);" and out and (out[-1].strip() == "narra_live--;" or (len(out) > 1 and out[-2].strip() == "narra_live--;" and out[-1].strip() == "#ifdef NARRA_PLAIN_MALLOC")):
            line = "  (void)p;"
        if line.strip() == "void* q = realloc(p, n ? n : 1);":
            line = "  void* q = narra_bump(n); if (p) memcpy(q, p, *((size_t*)p - 1) < n ? *((size_t*)p - 1) : n);"
    if mode == "freelist":
        if line.strip() == "void* p = malloc(n ? n : 1);":
            line = "  void* p = narra_fl_alloc(n);"
        if line.strip() == "free(p);" and out and (out[-1].strip() == "narra_live--;" or (len(out) > 1 and out[-2].strip() == "narra_live--;" and out[-1].strip() == "#ifdef NARRA_PLAIN_MALLOC")):
            line = "  narra_fl_free(p);"
        if line.strip() == "void* q = realloc(p, n ? n : 1);":
            line = "  void* q = narra_fl_alloc(n); if (p) { size_t old = *((size_t*)p - 1); memcpy(q, p, old < n ? old : n); narra_fl_free(p); }"
    if mode == "nofree":
        line = release_re.sub(r"--\1 == 0 && NARRA_FREEING", line)
    if mode == "nocount":
        line = retain_re.sub(r"(void)0", line)
        line = release_re.sub(r"NARRA_FREEING", line)
        line = line.replace("if (s->rc >= 0 && --s->rc == 0)", "if (NARRA_FREEING)")
    if mode == "sign":
        line = re.sub(r"\{ a->rc\+\+; \}", "{ if (a->rc >= 0) a->rc++; }", line)
        line = re.sub(r"\{ m->rc\+\+; \}", "{ if (m->rc >= 0) m->rc++; }", line)
        line = line.replace("if (v.u) v.u->rc++;", "if (v.u && v.u->rc >= 0) v.u->rc++;")
        line = line.replace("if (--a->rc == 0)", "if (a->rc >= 0 && --a->rc == 0)")
        line = line.replace("if (--m->rc == 0)", "if (m->rc >= 0 && --m->rc == 0)")
        line = line.replace("if (v.u && --v.u->rc == 0)", "if (v.u && v.u->rc >= 0 && --v.u->rc == 0)")
    out.append(line)

open(target, "w").write("\n".join(out))
