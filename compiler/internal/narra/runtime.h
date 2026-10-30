// Runtime for programs compiled by narra. Included at the top of every
// generated C file.
#include <inttypes.h>
#include <math.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static void narra_panic(const char* where, const char* msg) {
  fflush(stdout);
  fprintf(stderr, "panic at %s: %s\n", where, msg);
  exit(101);
}

// Live allocation count. With NARRA_CHECK_LEAKS set, the program reports
// allocations still live at exit, which shows reference counting bugs.
static int64_t narra_live = 0;

static void narra_report_leaks(void) {
  if (getenv("NARRA_CHECK_LEAKS") && narra_live != 0) {
    fprintf(stderr, "leak check: %" PRId64 " allocations still live\n", narra_live);
    _exit(102);
  }
}

static void* narra_alloc(size_t n) {
  void* p = malloc(n ? n : 1);
  if (!p) narra_panic("runtime", "out of memory");
  narra_live++;
  return p;
}

static void narra_free(void* p) {
  narra_live--;
  free(p);
}

// Dropping a heap value queues the values it holds instead of recursing, so
// a long list or a deep tree is freed with a worklist, not the C stack. The
// worklist grows and is reused by later drops; it is never freed. It is not
// counted by narra_live (plain realloc), and it stays reachable through a
// global, so neither the leak check nor LeakSanitizer reports it.
typedef struct narra_pending { void* p; void (*drop)(void*); } narra_pending;
static _Thread_local narra_pending* narra_pend = NULL;
static _Thread_local size_t narra_pend_n = 0, narra_pend_cap = 0;
static _Thread_local bool narra_draining = false;

static void narra_drop(void* p, void (*drop)(void*)) {
  if (narra_draining) {
    if (narra_pend_n == narra_pend_cap) {
      narra_pend_cap = narra_pend_cap ? narra_pend_cap * 2 : 64;
      narra_pending* grown = realloc(narra_pend, narra_pend_cap * sizeof(narra_pending));
      if (!grown) narra_panic("runtime", "out of memory");
      narra_pend = grown;
    }
    narra_pend[narra_pend_n].p = p;
    narra_pend[narra_pend_n].drop = drop;
    narra_pend_n++;
    return;
  }
  narra_draining = true;
  drop(p);
  while (narra_pend_n > 0) {
    narra_pending d = narra_pend[--narra_pend_n];
    d.drop(d.p);
  }
  narra_draining = false;
}

static void* narra_realloc(void* p, size_t n) {
  void* q = realloc(p, n ? n : 1);
  if (!q) narra_panic("runtime", "out of memory");
  return q;
}

// Strings. A negative count marks a literal that is never freed.
typedef struct lstr {
  int64_t rc;
  int64_t len;
  char* data;
} lstr;

static lstr* lstr_new(int64_t len) {
  lstr* s = narra_alloc(sizeof(lstr));
  s->rc = 1;
  s->len = len;
  s->data = narra_alloc((size_t)len + 1);
  s->data[len] = 0;
  return s;
}

static void lstr_retain(lstr* s) {
  if (s->rc >= 0) s->rc++;
}

static void lstr_release(lstr* s) {
  if (s->rc >= 0 && --s->rc == 0) {
    narra_free(s->data);
    narra_free(s);
  }
}

static lstr* lstr_from_cstr(const char* c) {
  int64_t n = (int64_t)strlen(c);
  lstr* s = lstr_new(n);
  memcpy(s->data, c, (size_t)n);
  return s;
}

static lstr* lstr_concat(lstr* a, lstr* b) {
  lstr* s = lstr_new(a->len + b->len);
  memcpy(s->data, a->data, (size_t)a->len);
  memcpy(s->data + a->len, b->data, (size_t)b->len);
  return s;
}

static lstr* lstr_repeat(lstr* a, int64_t n, const char* where) {
  if (n < 0) narra_panic(where, "Str.repeat with a negative count");
  lstr* s = lstr_new(a->len * n);
  for (int64_t i = 0; i < n; i++) memcpy(s->data + i * a->len, a->data, (size_t)a->len);
  return s;
}

static lstr* lstr_from_int(int64_t v) {
  char buf[32];
  snprintf(buf, sizeof buf, "%" PRId64, v);
  return lstr_from_cstr(buf);
}

// Shortest text that reads back as the same double.
static lstr* lstr_from_f64(double v) {
  char buf[40];
  for (int prec = 1; prec <= 17; prec++) {
    snprintf(buf, sizeof buf, "%.*g", prec, v);
    if (strtod(buf, NULL) == v) break;
  }
  if (!strchr(buf, '.') && !strchr(buf, 'e') && !strchr(buf, 'n') && !strchr(buf, 'i')) {
    strcat(buf, ".0");
  }
  return lstr_from_cstr(buf);
}

static lstr* lstr_from_char(uint32_t c) {
  char buf[5] = {0};
  int64_t n;
  if (c < 0x80) {
    buf[0] = (char)c;
    n = 1;
  } else if (c < 0x800) {
    buf[0] = (char)(0xC0 | (c >> 6));
    buf[1] = (char)(0x80 | (c & 0x3F));
    n = 2;
  } else if (c < 0x10000) {
    buf[0] = (char)(0xE0 | (c >> 12));
    buf[1] = (char)(0x80 | ((c >> 6) & 0x3F));
    buf[2] = (char)(0x80 | (c & 0x3F));
    n = 3;
  } else {
    buf[0] = (char)(0xF0 | (c >> 18));
    buf[1] = (char)(0x80 | ((c >> 12) & 0x3F));
    buf[2] = (char)(0x80 | ((c >> 6) & 0x3F));
    buf[3] = (char)(0x80 | (c & 0x3F));
    n = 4;
  }
  // Use the encoded length: a NUL character has no strlen.
  lstr* r = lstr_new(n);
  memcpy(r->data, buf, (size_t)n);
  return r;
}

static bool lstr_eq(lstr* a, lstr* b) {
  return a->len == b->len && memcmp(a->data, b->data, (size_t)a->len) == 0;
}

static int lstr_cmp(lstr* a, lstr* b) {
  int64_t n = a->len < b->len ? a->len : b->len;
  int c = memcmp(a->data, b->data, (size_t)n);
  if (c != 0) return c;
  return a->len < b->len ? -1 : (a->len > b->len ? 1 : 0);
}

// Standard trait functions for primitive types (trait-model contract).
static bool narra_eq_int(int64_t a, int64_t b) { return a == b; }
static bool narra_eq_f64(double a, double b) { return a == b; }
static bool narra_eq_bool(bool a, bool b) { return a == b; }
static bool narra_eq_char(uint32_t a, uint32_t b) { return a == b; }

static uint64_t narra_hash_mix(uint64_t h, uint64_t x) {
  h ^= x;
  h *= UINT64_C(1099511628211);
  return h ^ (h >> 29);
}

static uint64_t narra_hash_int(int64_t v) {
  uint64_t x = (uint64_t)v + UINT64_C(0x9E3779B97F4A7C15);
  x = (x ^ (x >> 30)) * UINT64_C(0xBF58476D1CE4E5B9);
  x = (x ^ (x >> 27)) * UINT64_C(0x94D049BB133111EB);
  return x ^ (x >> 31);
}
static uint64_t narra_hash_bool(bool v) { return narra_hash_int(v ? 1 : 0); }
static uint64_t narra_hash_char(uint32_t v) { return narra_hash_int((int64_t)v); }

static uint64_t lstr_hash(lstr* s) {
  uint64_t h = UINT64_C(14695981039346656037);
  for (int64_t i = 0; i < s->len; i++) {
    h ^= (unsigned char)s->data[i];
    h *= UINT64_C(1099511628211);
  }
  return h;
}

static int narra_cmp_int(int64_t a, int64_t b) { return (a > b) - (a < b); }
static int narra_cmp_f64(double a, double b) { return (a > b) - (a < b); }
static int narra_cmp_bool(bool a, bool b) { return (a > b) - (a < b); }
static int narra_cmp_char(uint32_t a, uint32_t b) { return (a > b) - (a < b); }

// Ordering's tag for a comparison result: Less = 0, Equal = 1, Greater = 2.
static int32_t narra_ordering(int c) { return c < 0 ? 0 : (c > 0 ? 2 : 1); }

// Appends c to *acc, replacing *acc with the result.
static void lstr_append_c(lstr** acc, const char* c) {
  int64_t n = (int64_t)strlen(c);
  lstr* s = lstr_new((*acc)->len + n);
  memcpy(s->data, (*acc)->data, (size_t)(*acc)->len);
  memcpy(s->data + (*acc)->len, c, (size_t)n);
  lstr_release(*acc);
  *acc = s;
}

// Appends piece to *acc and releases piece.
static void lstr_append_owned(lstr** acc, lstr* piece) {
  lstr* s = lstr_concat(*acc, piece);
  lstr_release(*acc);
  lstr_release(piece);
  *acc = s;
}

static lstr* narra_show_bool(bool v) { return lstr_from_cstr(v ? "true" : "false"); }

// The escape for byte c inside quotes q, or 0 when c stands for itself.
static char narra_escape(char c, char q) {
  switch (c) {
  case '\n': return 'n';
  case '\t': return 't';
  case '\r': return 'r';
  case '\\': return '\\';
  default: return c == q ? q : 0;
  }
}

// Bytes in quotes q, with q, `\`, newline, tab and carriage return escaped.
static lstr* narra_quote(const char* data, int64_t len, char q) {
  int64_t n = 2;
  for (int64_t i = 0; i < len; i++) n += narra_escape(data[i], q) ? 2 : 1;
  lstr* s = lstr_new(n);
  int64_t j = 0;
  s->data[j++] = q;
  for (int64_t i = 0; i < len; i++) {
    char e = narra_escape(data[i], q);
    if (e) {
      s->data[j++] = '\\';
      s->data[j++] = e;
    } else {
      s->data[j++] = data[i];
    }
  }
  s->data[j++] = q;
  return s;
}

static lstr* narra_show_char(uint32_t c) {
  lstr* u = lstr_from_char(c);
  lstr* s = narra_quote(u->data, u->len, '\'');
  lstr_release(u);
  return s;
}

static lstr* narra_show_string(lstr* v) { return narra_quote(v->data, v->len, '"'); }

static void narra_print(lstr* s) {
  fwrite(s->data, 1, (size_t)s->len, stdout);
  fputc('\n', stdout);
}

static void narra_assert(bool ok, const char* where) {
  if (!ok) narra_panic(where, "assertion failed");
}

static void narra_panic_str(lstr* msg, const char* where) {
  narra_panic(where, msg->data);
}

// Integer arithmetic. Overflow panics in debug builds and wraps in release
// builds. Division by zero panics in every build.
#ifdef NARRA_RELEASE
#define NARRA_CHECKED 0
#else
#define NARRA_CHECKED 1
#endif

static int64_t narra_add(int64_t a, int64_t b, const char* where) {
  int64_t r;
  if (__builtin_add_overflow(a, b, &r) && NARRA_CHECKED) narra_panic(where, "integer overflow");
  return r;
}

static int64_t narra_sub(int64_t a, int64_t b, const char* where) {
  int64_t r;
  if (__builtin_sub_overflow(a, b, &r) && NARRA_CHECKED) narra_panic(where, "integer overflow");
  return r;
}

static int64_t narra_mul(int64_t a, int64_t b, const char* where) {
  int64_t r;
  if (__builtin_mul_overflow(a, b, &r) && NARRA_CHECKED) narra_panic(where, "integer overflow");
  return r;
}

static int64_t narra_neg(int64_t a, const char* where) {
  if (a == INT64_MIN) {
    if (NARRA_CHECKED) narra_panic(where, "integer overflow");
    return a;
  }
  return -a;
}

static int64_t narra_div(int64_t a, int64_t b, const char* where) {
  if (b == 0) narra_panic(where, "division by zero");
  if (a == INT64_MIN && b == -1) {
    if (NARRA_CHECKED) narra_panic(where, "integer overflow");
    return a;
  }
  return a / b;
}

static int64_t narra_rem(int64_t a, int64_t b, const char* where) {
  if (b == 0) narra_panic(where, "division by zero");
  if (b == -1) return 0;
  return a % b;
}

static int64_t narra_wrap_add(int64_t a, int64_t b) { return (int64_t)((uint64_t)a + (uint64_t)b); }
static int64_t narra_wrap_sub(int64_t a, int64_t b) { return (int64_t)((uint64_t)a - (uint64_t)b); }
static int64_t narra_wrap_mul(int64_t a, int64_t b) { return (int64_t)((uint64_t)a * (uint64_t)b); }

static int64_t narra_shl(int64_t a, int64_t b, const char* where) {
  if (b < 0 || b > 63) narra_panic(where, "shift count out of range");
  return (int64_t)((uint64_t)a << b);
}

static int64_t narra_shr(int64_t a, int64_t b, const char* where) {
  if (b < 0 || b > 63) narra_panic(where, "shift count out of range");
  return a >> b;
}

static int64_t narra_abs(int64_t a, const char* where) { return a < 0 ? narra_neg(a, where) : a; }

static int64_t narra_truncate(double v, const char* where) {
  if (!(v >= -9223372036854775808.0 && v < 9223372036854775808.0)) {
    narra_panic(where, "Float.truncate: value out of range");
  }
  return (int64_t)v;
}

static void narra_bounds(int64_t i, int64_t len, const char* where) {
  if (i < 0 || i >= len) {
    char msg[96];
    snprintf(msg, sizeof msg, "index %" PRId64 " out of range for length %" PRId64, i, len);
    narra_panic(where, msg);
  }
}
