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



// On Windows both "/" and "\\" separate segments, and a leading drive such as
// "C:" is kept in front of the result. Results always use "/".
static bool narra_path_is_separator(char c) {
  return c == '/' || (NARRA_PATH_WINDOWS && c == '\\');
}

static int64_t narra_path_drive(const char* p, int64_t n) {
  if (!NARRA_PATH_WINDOWS || n < 2 || p[1] != ':') return 0;
  bool letter = (p[0] >= 'a' && p[0] <= 'z') || (p[0] >= 'A' && p[0] <= 'Z');
  return letter ? 2 : 0;
}

static lstr* narra_path_with_drive(const char* p, int64_t drive, lstr* rest) {
  if (drive == 0) return rest;
  lstr* s = lstr_new(drive + rest->len);
  memcpy(s->data, p, (size_t)drive);
  memcpy(s->data + drive, rest->data, (size_t)rest->len);
  lstr_release(rest);
  return s;
}

static lstr* narra_path_slice(const char* p, int64_t start, int64_t end) {
  lstr* s = lstr_new(end - start);
  memcpy(s->data, p + start, (size_t)(end - start));
  return s;
}

// Node's path.posix.normalize: drops "." and empty segments, resolves "..".
static lstr* narra_path_normalize_rest(const char* p, int64_t n) {
  bool absolute = n > 0 && narra_path_is_separator(p[0]);
  bool trailing = n > 0 && narra_path_is_separator(p[n - 1]);
  char* out = malloc((size_t)n + 4);
  if (!out) narra_panic("runtime", "out of memory");
  int64_t o = 0;
  int64_t i = 0;
  while (i < n) {
    while (i < n && narra_path_is_separator(p[i])) i++;
    int64_t start = i;
    while (i < n && !narra_path_is_separator(p[i])) i++;
    int64_t len = i - start;
    if (len == 0 || (len == 1 && p[start] == '.')) continue;
    if (len == 2 && p[start] == '.' && p[start + 1] == '.') {
      bool last_up = o >= 2 && out[o - 1] == '.' && out[o - 2] == '.' && (o == 2 || out[o - 3] == '/');
      if (o > 0 && !last_up) {
        while (o > 0 && out[o - 1] != '/') o--;
        if (o > 0) o--;
      } else if (!absolute) {
        if (o > 0) out[o++] = '/';
        out[o++] = '.';
        out[o++] = '.';
      }
      continue;
    }
    if (o > 0) out[o++] = '/';
    memcpy(out + o, p + start, (size_t)len);
    o += len;
  }
  lstr* s;
  if (o == 0) {
    s = lstr_from_cstr(absolute ? "/" : (trailing ? "./" : "."));
  } else {
    int64_t total = (absolute ? 1 : 0) + o + (trailing ? 1 : 0);
    s = lstr_new(total);
    int64_t at = 0;
    if (absolute) s->data[at++] = '/';
    memcpy(s->data + at, out, (size_t)o);
    at += o;
    if (trailing) s->data[at] = '/';
  }
  free(out);
  return s;
}

static lstr* narra_path_normalize(const char* p, int64_t n) {
  int64_t drive = narra_path_drive(p, n);
  return narra_path_with_drive(p, drive, narra_path_normalize_rest(p + drive, n - drive));
}

static lstr* narra_path_join(lstr* left, lstr* right) {
  if (left->len == 0 && right->len == 0) return lstr_from_cstr(".");
  if (left->len == 0) return narra_path_normalize(right->data, right->len);
  if (right->len == 0) return narra_path_normalize(left->data, left->len);
  lstr* both = lstr_new(left->len + 1 + right->len);
  memcpy(both->data, left->data, (size_t)left->len);
  both->data[left->len] = '/';
  memcpy(both->data + left->len + 1, right->data, (size_t)right->len);
  lstr* s = narra_path_normalize(both->data, both->len);
  lstr_release(both);
  return s;
}

static lstr* narra_path_dirname_rest(const char* p, int64_t n) {
  if (n == 0) return lstr_from_cstr(".");
  bool root = narra_path_is_separator(p[0]);
  int64_t end = -1;
  bool matched = true;
  for (int64_t i = n - 1; i >= 1; i--) {
    if (narra_path_is_separator(p[i])) {
      if (!matched) { end = i; break; }
    } else {
      matched = false;
    }
  }
  if (end == -1) return lstr_from_cstr(root ? "/" : ".");
  if (root && end == 1) return lstr_from_cstr("//");
  return narra_path_slice(p, 0, end);
}

static lstr* narra_path_dirname(lstr* path) {
  int64_t drive = narra_path_drive(path->data, path->len);
  lstr* rest = narra_path_dirname_rest(path->data + drive, path->len - drive);
  if (drive > 0 && rest->len == 1 && rest->data[0] == '.') {
    // "C:" and "C:a" have the drive itself as their directory.
    lstr_release(rest);
    return narra_path_slice(path->data, 0, drive);
  }
  return narra_path_with_drive(path->data, drive, rest);
}

static lstr* narra_path_basename(lstr* path) {
  int64_t drive = narra_path_drive(path->data, path->len);
  const char* p = path->data + drive;
  int64_t start = 0, end = -1;
  bool matched = true;
  for (int64_t i = path->len - drive - 1; i >= 0; i--) {
    if (narra_path_is_separator(p[i])) {
      if (!matched) { start = i + 1; break; }
    } else if (end == -1) {
      matched = false;
      end = i + 1;
    }
  }
  if (end == -1) return lstr_from_cstr("");
  return narra_path_slice(p, start, end);
}

static lstr* narra_path_extname(lstr* path) {
  int64_t drive = narra_path_drive(path->data, path->len);
  const char* p = path->data + drive;
  int64_t start_dot = -1, start_part = 0, end = -1;
  int pre_dot = 0;
  bool matched = true;
  for (int64_t i = path->len - drive - 1; i >= 0; i--) {
    char c = p[i];
    if (narra_path_is_separator(c)) {
      if (!matched) { start_part = i + 1; break; }
      continue;
    }
    if (end == -1) { matched = false; end = i + 1; }
    if (c == '.') {
      if (start_dot == -1) start_dot = i;
      else if (pre_dot != 1) pre_dot = 1;
    } else if (start_dot != -1) {
      pre_dot = -1;
    }
  }
  if (start_dot == -1 || end == -1 || pre_dot == 0 || (pre_dot == 1 && start_dot == end - 1 && start_dot == start_part + 1)) {
    return lstr_from_cstr("");
  }
  return narra_path_slice(p, start_dot, end);
}

