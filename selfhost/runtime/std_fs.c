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


#include <dirent.h>
#include <sys/stat.h>

typedef struct narra_fs_names {
  lstr** items;
  int64_t len;
} narra_fs_names;

static bool narra_fs_read(lstr* path, lstr** out, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  FILE* f = fopen(path->data, "rb");
  if (!f) {
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  size_t cap = 4096, len = 0;
  char* buf = malloc(cap);
  const char* fail = buf ? NULL : "out of memory";
  while (!fail) {
    if (len == cap) {
      char* bigger = realloc(buf, cap * 2);
      if (!bigger) { fail = "out of memory"; break; }
      buf = bigger;
      cap *= 2;
    }
    size_t got = fread(buf + len, 1, cap - len, f);
    len += got;
    if (got == 0) {
      if (ferror(f)) fail = strerror(errno);
      break;
    }
  }
  fclose(f);
  if (!fail && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "file is not valid UTF-8";
  if (fail) {
    free(buf);
    *err = narra_io_fail(path->data, path->len, fail);
    return false;
  }
  *out = lstr_new((int64_t)len);
  memcpy((*out)->data, buf, len);
  free(buf);
  return true;
}

static bool narra_fs_write(lstr* path, lstr* text, int64_t* written, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  FILE* f = fopen(path->data, "wb");
  if (!f) {
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  errno = 0;
  size_t put = fwrite(text->data, 1, (size_t)text->len, f);
  bool ok = put == (size_t)text->len;
  if (fclose(f) != 0) ok = false;
  if (!ok) {
    *err = narra_io_fail(path->data, path->len, strerror(errno ? errno : EIO));
    return false;
  }
  *written = text->len;
  return true;
}

static bool narra_fs_exists(lstr* path) {
  if (memchr(path->data, 0, (size_t)path->len)) return false;
  struct stat st;
  return stat(path->data, &st) == 0;
}

static bool narra_fs_is_dir(lstr* path) {
  if (memchr(path->data, 0, (size_t)path->len)) return false;
  struct stat st;
  return stat(path->data, &st) == 0 && S_ISDIR(st.st_mode);
}

// Creates one directory (mode 0777 before the umask).
static bool narra_fs_mkdir(lstr* path, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  if (mkdir(path->data, 0777) != 0) {
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  return true;
}

// Creates the directory and its missing parents. *created is false when the
// directory was already there.
static bool narra_fs_mkdir_all(lstr* path, bool* created, lstr** err) {
  *created = false;
  if (!narra_io_no_nul(path, err)) return false;
  char* buf = malloc((size_t)path->len + 1);
  if (!buf) {
    *err = narra_io_fail(path->data, path->len, "out of memory");
    return false;
  }
  memcpy(buf, path->data, (size_t)path->len + 1);
  int failed = 0;
  for (int64_t i = 1; i < path->len && !failed; i++) {
    if (buf[i] != '/') continue;
    buf[i] = 0;
    if (mkdir(buf, 0777) == 0) *created = true;
    else if (errno != EEXIST) failed = errno;
    buf[i] = '/';
  }
  if (!failed) {
    if (mkdir(buf, 0777) == 0) *created = true;
    else if (errno != EEXIST) failed = errno;
    else {
      struct stat st;
      if (stat(buf, &st) != 0 || !S_ISDIR(st.st_mode)) failed = EEXIST;
    }
  }
  free(buf);
  if (failed) {
    *err = narra_io_fail(path->data, path->len, strerror(failed));
    return false;
  }
  return true;
}

static bool narra_fs_unlink(lstr* path, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  if (unlink(path->data) != 0) {
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  return true;
}

static bool narra_fs_rmdir(lstr* path, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  if (rmdir(path->data) != 0) {
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  return true;
}

static int narra_fs_name_cmp(const void* a, const void* b) {
  return lstr_cmp(*(lstr* const*)a, *(lstr* const*)b);
}

// The names in a directory, sorted by bytes, without "." and "..".
static bool narra_fs_list(lstr* path, narra_fs_names* out, lstr** err) {
  out->items = NULL;
  out->len = 0;
  if (!narra_io_no_nul(path, err)) return false;
  DIR* d = opendir(path->data);
  if (!d) {
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  int64_t cap = 0;
  const char* fail = NULL;
  struct dirent* ent;
  while (!fail && (ent = readdir(d)) != NULL) {
    if (strcmp(ent->d_name, ".") == 0 || strcmp(ent->d_name, "..") == 0) continue;
    size_t n = strlen(ent->d_name);
    if (!narra_io_utf8((const unsigned char*)ent->d_name, (int64_t)n)) { fail = "file name is not valid UTF-8"; break; }
    if (out->len == cap) {
      cap = cap ? cap * 2 : 16;
      lstr** bigger = realloc(out->items, sizeof(lstr*) * (size_t)cap);
      if (!bigger) { fail = "out of memory"; break; }
      out->items = bigger;
    }
    out->items[out->len++] = lstr_from_cstr(ent->d_name);
  }
  closedir(d);
  if (fail) {
    for (int64_t i = 0; i < out->len; i++) lstr_release(out->items[i]);
    free(out->items);
    out->items = NULL;
    out->len = 0;
    *err = narra_io_fail(path->data, path->len, fail);
    return false;
  }
  if (out->len > 1) qsort(out->items, (size_t)out->len, sizeof(lstr*), narra_fs_name_cmp);
  return true;
}

