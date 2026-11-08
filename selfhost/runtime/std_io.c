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



// Reads up to limit bytes of stdin (all of it when limit is negative) into a
// malloc'ed buffer. *fail is set on a read error.
static char* narra_io_slurp(int64_t limit, size_t* len, const char** fail) {
  size_t cap = 4096;
  char* buf = malloc(cap);
  *len = 0;
  *fail = buf ? NULL : "out of memory";
  while (!*fail && (limit < 0 || (int64_t)*len < limit)) {
    if (*len == cap) {
      char* bigger = realloc(buf, cap * 2);
      if (!bigger) { *fail = "out of memory"; break; }
      buf = bigger;
      cap *= 2;
    }
    size_t want = cap - *len;
    if (limit >= 0 && (int64_t)want > limit - (int64_t)*len) want = (size_t)(limit - (int64_t)*len);
    size_t got = fread(buf + *len, 1, want, stdin);
    *len += got;
    if (got == 0) {
      if (ferror(stdin)) *fail = strerror(errno);
      break;
    }
  }
  return buf;
}

// The next line without its "\n" (a "\r" stays). *out is NULL at the end of
// the input.
static bool narra_io_read_line(lstr** out, lstr** err) {
  *out = NULL;
  size_t cap = 128, n = 0;
  char* line = malloc(cap);
  if (!line) {
    *err = narra_io_fail("stdin", 5, "out of memory");
    return false;
  }
  int c;
  errno = 0;
  while ((c = fgetc(stdin)) != EOF && c != '\n') {
    if (n == cap) {
      char* bigger = realloc(line, cap * 2);
      if (!bigger) {
        free(line);
        *err = narra_io_fail("stdin", 5, "out of memory");
        return false;
      }
      line = bigger;
      cap *= 2;
    }
    line[n++] = (char)c;
  }
  if (c == EOF && (ferror(stdin) || n == 0)) {
    bool bad = ferror(stdin) != 0;
    free(line);
    if (bad) {
      *err = narra_io_fail("stdin", 5, strerror(errno ? errno : EIO));
      return false;
    }
    return true;
  }
  if (!narra_io_utf8((const unsigned char*)line, (int64_t)n)) {
    free(line);
    *err = narra_io_fail("stdin", 5, "line is not valid UTF-8");
    return false;
  }
  *out = lstr_new((int64_t)n);
  memcpy((*out)->data, line, (size_t)n);
  free(line);
  return true;
}

// Exactly count bytes.
static bool narra_io_read_exact(int64_t count, lstr** out, lstr** err) {
  if (count < 0) {
    *err = narra_io_fail("stdin", 5, "negative byte count");
    return false;
  }
  size_t len;
  const char* fail;
  char* buf = narra_io_slurp(count, &len, &fail);
  if (!fail && (int64_t)len < count) fail = "unexpected end of input";
  if (!fail && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "input is not valid UTF-8";
  if (fail) {
    free(buf);
    *err = narra_io_fail("stdin", 5, fail);
    return false;
  }
  *out = lstr_new((int64_t)len);
  memcpy((*out)->data, buf, len);
  free(buf);
  return true;
}

// The rest of the input.
static bool narra_io_read_all(lstr** out, lstr** err) {
  size_t len;
  const char* fail;
  char* buf = narra_io_slurp(-1, &len, &fail);
  if (!fail && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "input is not valid UTF-8";
  if (fail) {
    free(buf);
    *err = narra_io_fail("stdin", 5, fail);
    return false;
  }
  *out = lstr_new((int64_t)len);
  memcpy((*out)->data, buf, len);
  free(buf);
  return true;
}

static void narra_io_write(lstr* text) { fwrite(text->data, 1, (size_t)text->len, stdout); }

static void narra_io_write_error(lstr* text) { fwrite(text->data, 1, (size_t)text->len, stderr); }

static void narra_io_flush(void) { fflush(stdout); }

