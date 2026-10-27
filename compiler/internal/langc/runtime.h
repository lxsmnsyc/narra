// Runtime for programs compiled by langc. Included at the top of every
// generated C file.
#include <inttypes.h>
#include <math.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static void lang_panic(const char* where, const char* msg) {
  fflush(stdout);
  fprintf(stderr, "panic at %s: %s\n", where, msg);
  exit(101);
}

// Live allocation count. With LANG_CHECK_LEAKS set, the program reports
// allocations still live at exit, which shows reference counting bugs.
static int64_t lang_live = 0;

static void lang_report_leaks(void) {
  if (getenv("LANG_CHECK_LEAKS") && lang_live != 0) {
    fprintf(stderr, "leak check: %" PRId64 " allocations still live\n", lang_live);
    _exit(102);
  }
}

static void* lang_alloc(size_t n) {
  void* p = malloc(n ? n : 1);
  if (!p) lang_panic("runtime", "out of memory");
  lang_live++;
  return p;
}

static void lang_free(void* p) {
  lang_live--;
  free(p);
}

static void* lang_realloc(void* p, size_t n) {
  void* q = realloc(p, n ? n : 1);
  if (!q) lang_panic("runtime", "out of memory");
  return q;
}

// Strings. A negative count marks a literal that is never freed.
typedef struct lstr {
  int64_t rc;
  int64_t len;
  char* data;
} lstr;

static lstr* lstr_new(int64_t len) {
  lstr* s = lang_alloc(sizeof(lstr));
  s->rc = 1;
  s->len = len;
  s->data = lang_alloc((size_t)len + 1);
  s->data[len] = 0;
  return s;
}

static void lstr_retain(lstr* s) {
  if (s->rc >= 0) s->rc++;
}

static void lstr_release(lstr* s) {
  if (s->rc >= 0 && --s->rc == 0) {
    lang_free(s->data);
    lang_free(s);
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
  if (n < 0) lang_panic(where, "Str.repeat with a negative count");
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
  if (c < 0x80) {
    buf[0] = (char)c;
  } else if (c < 0x800) {
    buf[0] = (char)(0xC0 | (c >> 6));
    buf[1] = (char)(0x80 | (c & 0x3F));
  } else if (c < 0x10000) {
    buf[0] = (char)(0xE0 | (c >> 12));
    buf[1] = (char)(0x80 | ((c >> 6) & 0x3F));
    buf[2] = (char)(0x80 | (c & 0x3F));
  } else {
    buf[0] = (char)(0xF0 | (c >> 18));
    buf[1] = (char)(0x80 | ((c >> 12) & 0x3F));
    buf[2] = (char)(0x80 | ((c >> 6) & 0x3F));
    buf[3] = (char)(0x80 | (c & 0x3F));
  }
  return lstr_from_cstr(buf);
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

static void lang_print(lstr* s) {
  fwrite(s->data, 1, (size_t)s->len, stdout);
  fputc('\n', stdout);
}

static void lang_assert(bool ok, const char* where) {
  if (!ok) lang_panic(where, "assertion failed");
}

static void lang_panic_str(lstr* msg, const char* where) {
  lang_panic(where, msg->data);
}

// Integer arithmetic. Overflow panics in debug builds and wraps in release
// builds. Division by zero panics in every build.
#ifdef LANG_RELEASE
#define LANG_CHECKED 0
#else
#define LANG_CHECKED 1
#endif

static int64_t lang_add(int64_t a, int64_t b, const char* where) {
  int64_t r;
  if (__builtin_add_overflow(a, b, &r) && LANG_CHECKED) lang_panic(where, "integer overflow");
  return r;
}

static int64_t lang_sub(int64_t a, int64_t b, const char* where) {
  int64_t r;
  if (__builtin_sub_overflow(a, b, &r) && LANG_CHECKED) lang_panic(where, "integer overflow");
  return r;
}

static int64_t lang_mul(int64_t a, int64_t b, const char* where) {
  int64_t r;
  if (__builtin_mul_overflow(a, b, &r) && LANG_CHECKED) lang_panic(where, "integer overflow");
  return r;
}

static int64_t lang_neg(int64_t a, const char* where) {
  if (a == INT64_MIN) {
    if (LANG_CHECKED) lang_panic(where, "integer overflow");
    return a;
  }
  return -a;
}

static int64_t lang_div(int64_t a, int64_t b, const char* where) {
  if (b == 0) lang_panic(where, "division by zero");
  if (a == INT64_MIN && b == -1) {
    if (LANG_CHECKED) lang_panic(where, "integer overflow");
    return a;
  }
  return a / b;
}

static int64_t lang_rem(int64_t a, int64_t b, const char* where) {
  if (b == 0) lang_panic(where, "division by zero");
  if (b == -1) return 0;
  return a % b;
}

static int64_t lang_wrap_add(int64_t a, int64_t b) { return (int64_t)((uint64_t)a + (uint64_t)b); }
static int64_t lang_wrap_sub(int64_t a, int64_t b) { return (int64_t)((uint64_t)a - (uint64_t)b); }
static int64_t lang_wrap_mul(int64_t a, int64_t b) { return (int64_t)((uint64_t)a * (uint64_t)b); }

static int64_t lang_shl(int64_t a, int64_t b, const char* where) {
  if (b < 0 || b > 63) lang_panic(where, "shift count out of range");
  return (int64_t)((uint64_t)a << b);
}

static int64_t lang_shr(int64_t a, int64_t b, const char* where) {
  if (b < 0 || b > 63) lang_panic(where, "shift count out of range");
  return a >> b;
}

static int64_t lang_abs(int64_t a, const char* where) { return a < 0 ? lang_neg(a, where) : a; }

static int64_t lang_truncate(double v, const char* where) {
  if (!(v >= -9223372036854775808.0 && v < 9223372036854775808.0)) {
    lang_panic(where, "Float.truncate: value out of range");
  }
  return (int64_t)v;
}

static void lang_bounds(int64_t i, int64_t len, const char* where) {
  if (i < 0 || i >= len) {
    char msg[96];
    snprintf(msg, sizeof msg, "index %" PRId64 " out of range for length %" PRId64, i, len);
    lang_panic(where, msg);
  }
}
