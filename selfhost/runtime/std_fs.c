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
#ifdef _WIN32
#include <winsock2.h>
#include <windows.h>
#endif

typedef struct narra_fs_names {
  lstr** items;
  int64_t len;
} narra_fs_names;

// Reads a whole file. With utf8 set, a file that is not valid UTF-8 fails.
static bool narra_fs_read_all(lstr* path, bool utf8, lstr** out, lstr** err) {
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
  if (!fail && utf8 && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "file is not valid UTF-8";
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

static bool narra_fs_read(lstr* path, lstr** out, lstr** err) {
  return narra_fs_read_all(path, true, out, err);
}

static bool narra_fs_read_bytes(lstr* path, lstr** out, lstr** err) {
  return narra_fs_read_all(path, false, out, err);
}

static bool narra_fs_write(lstr* path, lstr* text, int64_t* written, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
#if defined(__wasi__) || defined(_WIN32)
  // WASI and Windows have no /dev: the two standard output paths write descriptors 1 and 2.
  FILE* standard = NULL;
  if (strcmp(path->data, "/dev/stdout") == 0) standard = stdout;
  if (strcmp(path->data, "/dev/stderr") == 0) standard = stderr;
  if (standard) {
    errno = 0;
    size_t sent = fwrite(text->data, 1, (size_t)text->len, standard);
    if (sent != (size_t)text->len || fflush(standard) != 0) {
      *err = narra_io_fail(path->data, path->len, strerror(errno ? errno : EIO));
      return false;
    }
    *written = text->len;
    return true;
  }
#endif
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

// Creates one directory (mode 0777 before the umask; Windows has no mode).
static bool narra_fs_mkdir(lstr* path, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  if (narra_mkdir(path->data) != 0) {
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
    char separator = buf[i];
    if (separator != '/' && !(NARRA_PATH_WINDOWS && separator == '\\')) continue;
    // A drive such as "C:" is not a directory to create.
    if (NARRA_PATH_WINDOWS && buf[i - 1] == ':') continue;
    buf[i] = 0;
    if (narra_mkdir(buf) == 0) *created = true;
    else if (errno != EEXIST) failed = errno;
    buf[i] = separator;
  }
  if (!failed) {
    if (narra_mkdir(buf) == 0) *created = true;
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

// Adds text to the end of a file and creates the file when it is missing.
static bool narra_fs_append(lstr* path, lstr* text, int64_t* written, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  FILE* f = fopen(path->data, "ab");
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

#ifdef _WIN32
// The system's message for a Windows error code, without the line end.
static lstr* narra_fs_windows_fail(lstr* path, DWORD code) {
  char message[256];
  DWORD length = FormatMessageA(
      FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS, NULL, code, 0, message,
      (DWORD)sizeof(message) - 1, NULL);
  while (length > 0 && (message[length - 1] == '\n' || message[length - 1] == '\r' || message[length - 1] == ' ' || message[length - 1] == '.')) length--;
  message[length] = 0;
  if (length == 0) snprintf(message, sizeof(message), "Windows error %lu", (unsigned long)code);
  return narra_io_fail(path->data, path->len, message);
}
#endif

// Moves a file or folder. On POSIX systems rename replaces an existing file
// (a folder only when the target is an empty folder); Windows replaces an
// existing file and refuses an existing folder.
static bool narra_fs_rename(lstr* from, lstr* to, lstr** err) {
  if (!narra_io_no_nul(from, err) || !narra_io_no_nul(to, err)) return false;
#ifdef _WIN32
  if (!MoveFileExA(from->data, to->data, MOVEFILE_REPLACE_EXISTING | MOVEFILE_COPY_ALLOWED)) {
    *err = narra_fs_windows_fail(from, GetLastError());
    return false;
  }
#else
  if (rename(from->data, to->data) != 0) {
    *err = narra_io_fail(from->data, from->len, strerror(errno));
    return false;
  }
#endif
  return true;
}

// Copies the bytes of a file to the target, which is created or emptied first.
// A copy keeps the permission bits of the source where the system has them.
static bool narra_fs_copy(lstr* from, lstr* to, int64_t* copied, lstr** err) {
  if (!narra_io_no_nul(from, err) || !narra_io_no_nul(to, err)) return false;
  struct stat source_info;
  if (stat(from->data, &source_info) != 0) {
    *err = narra_io_fail(from->data, from->len, strerror(errno));
    return false;
  }
  if (S_ISDIR(source_info.st_mode)) {
    *err = narra_io_fail(from->data, from->len, "Is a directory");
    return false;
  }
#ifdef _WIN32
  if (!CopyFileA(from->data, to->data, FALSE)) {
    *err = narra_fs_windows_fail(to, GetLastError());
    return false;
  }
  *copied = (int64_t)source_info.st_size;
  return true;
#else
  struct stat target_info;
  if (stat(to->data, &target_info) == 0 && target_info.st_dev == source_info.st_dev && target_info.st_ino == source_info.st_ino) {
    *err = narra_io_fail(to->data, to->len, "source and target are the same file");
    return false;
  }
  FILE* input = fopen(from->data, "rb");
  if (!input) {
    *err = narra_io_fail(from->data, from->len, strerror(errno));
    return false;
  }
  FILE* output = fopen(to->data, "wb");
  if (!output) {
    int open_reason = errno;
    fclose(input);
    *err = narra_io_fail(to->data, to->len, strerror(open_reason));
    return false;
  }
  char buffer[65536];
  int64_t total = 0;
  int reason = 0;
  for (;;) {
    size_t got = fread(buffer, 1, sizeof(buffer), input);
    if (got > 0) {
      errno = 0;
      if (fwrite(buffer, 1, got, output) != got) { reason = errno ? errno : EIO; break; }
      total += (int64_t)got;
    }
    if (got < sizeof(buffer)) {
      if (ferror(input)) reason = errno ? errno : EIO;
      break;
    }
  }
  fclose(input);
  errno = 0;
  if (fflush(output) != 0 && reason == 0) reason = errno ? errno : EIO;
#ifndef __wasi__
  if (reason == 0) (void)fchmod(fileno(output), source_info.st_mode & 07777);
#endif
  if (fclose(output) != 0 && reason == 0) reason = errno ? errno : EIO;
  if (reason != 0) {
    *err = narra_io_fail(to->data, to->len, strerror(reason));
    return false;
  }
  *copied = total;
  return true;
#endif
}

// 0 when the path does not exist, 1 for a folder, 2 for anything else. A
// symbolic link (a junction on Windows) is never a folder here, so the
// removal below never walks through one.
static int narra_fs_entry_kind(const char* path) {
#ifdef _WIN32
  DWORD attributes = GetFileAttributesA(path);
  if (attributes == INVALID_FILE_ATTRIBUTES) return 0;
  if ((attributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0) return 2;
  return (attributes & FILE_ATTRIBUTE_DIRECTORY) != 0 ? 1 : 2;
#else
  struct stat info;
  if (lstat(path, &info) != 0) return 0;
  return S_ISDIR(info.st_mode) ? 1 : 2;
#endif
}

static bool narra_fs_remove_tree(const char* path, int* reason) {
  DIR* directory = opendir(path);
  if (!directory) {
    *reason = errno;
    return false;
  }
  size_t base_length = strlen(path);
  struct dirent* entry;
  while ((entry = readdir(directory)) != NULL) {
    if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) continue;
    size_t name_length = strlen(entry->d_name);
    char* child = malloc(base_length + name_length + 2);
    if (!child) {
      *reason = ENOMEM;
      closedir(directory);
      return false;
    }
    memcpy(child, path, base_length);
    child[base_length] = '/';
    memcpy(child + base_length + 1, entry->d_name, name_length + 1);
    bool removed;
    if (narra_fs_entry_kind(child) == 1) {
      removed = narra_fs_remove_tree(child, reason);
    } else {
      removed = unlink(child) == 0;
      if (!removed) *reason = errno;
    }
    free(child);
    if (!removed) {
      closedir(directory);
      return false;
    }
  }
  closedir(directory);
  if (rmdir(path) != 0) {
    *reason = errno;
    return false;
  }
  return true;
}

// Removes a folder with everything in it. *removed is false when the path was
// not there. A symbolic link inside the tree is removed as a link and never
// followed; a path that names a file or a link fails.
static bool narra_fs_remove_all(lstr* path, bool* removed, lstr** err) {
  *removed = false;
  if (!narra_io_no_nul(path, err)) return false;
  int kind = narra_fs_entry_kind(path->data);
  if (kind == 0) return true;
  if (kind != 1) {
    *err = narra_io_fail(path->data, path->len, "Not a directory");
    return false;
  }
  int reason = 0;
  if (!narra_fs_remove_tree(path->data, &reason)) {
    *err = narra_io_fail(path->data, path->len, strerror(reason));
    return false;
  }
  *removed = true;
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

