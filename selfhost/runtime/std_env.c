#ifndef NARRA_IO_COMMON
#define NARRA_IO_COMMON
#include <errno.h>

// True when p[0..n) is valid UTF-8: no overlong forms, no surrogates and
// nothing above U+10FFFF.
static bool narra_io_utf8(const unsigned char* p, int64_t n) {
  int64_t i = 0;
  while (i < n) {
    unsigned char c = p[i];
    int64_t k;
    uint32_t cp;
    if (c < 0x80) { i++; continue; }
    if (c >= 0xC2 && c <= 0xDF) { k = 1; cp = c & 0x1F; }
    else if (c >= 0xE0 && c <= 0xEF) { k = 2; cp = c & 0x0F; }
    else if (c >= 0xF0 && c <= 0xF4) { k = 3; cp = c & 0x07; }
    else return false;
    if (i + k >= n) return false;
    for (int64_t j = 1; j <= k; j++) {
      if ((p[i + j] & 0xC0) != 0x80) return false;
      cp = (cp << 6) | (p[i + j] & 0x3F);
    }
    if (k == 2 && (cp < 0x800 || (cp >= 0xD800 && cp <= 0xDFFF))) return false;
    if (k == 3 && (cp < 0x10000 || cp > 0x10FFFF)) return false;
    i += k + 1;
  }
  return true;
}

// A string from bytes; each byte that is not valid UTF-8 becomes U+FFFD.
static lstr* narra_io_lossy(const char* p, size_t n) {
  if (narra_io_utf8((const unsigned char*)p, (int64_t)n)) {
    lstr* s = lstr_new((int64_t)n);
    memcpy(s->data, p, n);
    return s;
  }
  lstr* s = lstr_new((int64_t)n * 3);
  int64_t o = 0;
  size_t i = 0;
  while (i < n) {
    size_t k = 1;
    unsigned char c = (unsigned char)p[i];
    if (c >= 0xC2 && c <= 0xDF) k = 2;
    else if (c >= 0xE0 && c <= 0xEF) k = 3;
    else if (c >= 0xF0 && c <= 0xF4) k = 4;
    if (c < 0x80 || (k > 1 && i + k <= n && narra_io_utf8((const unsigned char*)p + i, (int64_t)k))) {
      memcpy(s->data + o, p + i, k);
      o += (int64_t)k;
      i += k;
    } else {
      memcpy(s->data + o, "\xEF\xBF\xBD", 3);
      o += 3;
      i++;
    }
  }
  s->len = o;
  s->data[o] = 0;
  return s;
}

// "<what>: <message>"
static lstr* narra_io_fail(const char* what, int64_t whatlen, const char* msg) {
  size_t ml = strlen(msg);
  lstr* s = lstr_new(whatlen + 2 + (int64_t)ml);
  memcpy(s->data, what, (size_t)whatlen);
  memcpy(s->data + whatlen, ": ", 2);
  memcpy(s->data + whatlen + 2, msg, ml);
  return s;
}

// C strings end at the first zero byte, so a string with one inside cannot
// name a file, a command or a variable.
static bool narra_io_no_nul(lstr* s, lstr** err) {
  if (memchr(s->data, 0, (size_t)s->len)) {
    *err = lstr_from_cstr("argument contains a NUL byte");
    return false;
  }
  return true;
}
#endif


#ifdef __APPLE__
#include <crt_externs.h>
#endif

// NULL when the variable is not set.
static lstr* narra_env_get(lstr* name) {
  if (memchr(name->data, 0, (size_t)name->len)) return NULL;
  const char* v = getenv(name->data);
  return v ? narra_io_lossy(v, strlen(v)) : NULL;
}

#ifdef __APPLE__
static int64_t narra_env_argc(void) { return *_NSGetArgc(); }
static lstr* narra_env_arg(int64_t i) {
  const char* a = (*_NSGetArgv())[i];
  return narra_io_lossy(a, strlen(a));
}
#else
// Elsewhere the arguments are read once from /proc/self/cmdline.
static char* narra_env_buf;
static size_t narra_env_len;
static int64_t narra_env_n = -1;

static void narra_env_load(void) {
  if (narra_env_n >= 0) return;
  narra_env_n = 0;
  FILE* f = fopen("/proc/self/cmdline", "rb");
  if (!f) return;
  size_t cap = 4096;
  narra_env_buf = malloc(cap);
  while (narra_env_buf) {
    if (narra_env_len == cap) {
      char* bigger = realloc(narra_env_buf, cap * 2);
      if (!bigger) break;
      narra_env_buf = bigger;
      cap *= 2;
    }
    size_t got = fread(narra_env_buf + narra_env_len, 1, cap - narra_env_len, f);
    if (got == 0) break;
    narra_env_len += got;
  }
  fclose(f);
  for (size_t i = 0; i < narra_env_len; i++) {
    if (narra_env_buf[i] == 0) narra_env_n++;
  }
}
static int64_t narra_env_argc(void) { narra_env_load(); return narra_env_n; }
static lstr* narra_env_arg(int64_t i) {
  narra_env_load();
  const char* p = narra_env_buf;
  while (i-- > 0) p += strlen(p) + 1;
  return narra_io_lossy(p, strlen(p));
}
#endif
