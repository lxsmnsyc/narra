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
#else
#include <fcntl.h>
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

// "<from> -> <to>: <message>", for a failure the call cannot pin on one path.
static lstr* narra_fs_fail_pair(lstr* from, lstr* to, const char* message) {
  size_t message_length = strlen(message);
  lstr* s = lstr_new(from->len + 4 + to->len + 2 + (int64_t)message_length);
  char* out = s->data;
  memcpy(out, from->data, (size_t)from->len);
  out += from->len;
  memcpy(out, " -> ", 4);
  out += 4;
  memcpy(out, to->data, (size_t)to->len);
  out += to->len;
  memcpy(out, ": ", 2);
  memcpy(out + 2, message, message_length);
  return s;
}

#ifdef _WIN32
// The system's message for a Windows error code, without the line end.
static void narra_fs_windows_message(DWORD code, char* message, DWORD size) {
  DWORD length = FormatMessageA(
      FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS, NULL, code, 0, message, size - 1, NULL);
  while (length > 0 && (message[length - 1] == '\n' || message[length - 1] == '\r' || message[length - 1] == ' ' || message[length - 1] == '.')) length--;
  message[length] = 0;
  if (length == 0) snprintf(message, size, "Windows error %lu", (unsigned long)code);
}

static lstr* narra_fs_windows_fail(lstr* path, DWORD code) {
  char message[256];
  narra_fs_windows_message(code, message, (DWORD)sizeof(message));
  return narra_io_fail(path->data, path->len, message);
}
#endif

// Moves a file or folder. On POSIX systems rename replaces an existing file
// (a folder only when the target is an empty folder); Windows replaces an
// existing file and refuses an existing folder. Neither call says which path
// failed, so the message names both.
static bool narra_fs_rename(lstr* from, lstr* to, lstr** err) {
  if (!narra_io_no_nul(from, err) || !narra_io_no_nul(to, err)) return false;
#ifdef _WIN32
  if (!MoveFileExA(from->data, to->data, MOVEFILE_REPLACE_EXISTING | MOVEFILE_COPY_ALLOWED)) {
    char message[256];
    narra_fs_windows_message(GetLastError(), message, (DWORD)sizeof(message));
    *err = narra_fs_fail_pair(from, to, message);
    return false;
  }
#else
  if (rename(from->data, to->data) != 0) {
    *err = narra_fs_fail_pair(from, to, strerror(errno));
    return false;
  }
#endif
  return true;
}

// Copies the bytes of a file to the target, which is created or emptied first.
// A copy keeps the read, write and run bits of the source where the system has
// them. A copy that fails part way leaves a partial target.
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
    char message[256];
    narra_fs_windows_message(GetLastError(), message, (DWORD)sizeof(message));
    *err = narra_fs_fail_pair(from, to, message);
    return false;
  }
  *copied = (int64_t)source_info.st_size;
  return true;
#else
  struct stat target_info;
  if (stat(to->data, &target_info) == 0 && target_info.st_dev == source_info.st_dev && target_info.st_ino == source_info.st_ino) {
    *err = narra_fs_fail_pair(from, to, "source and target are the same file");
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
  lstr* failed = to;
  for (;;) {
    size_t got = fread(buffer, 1, sizeof(buffer), input);
    if (got > 0) {
      errno = 0;
      if (fwrite(buffer, 1, got, output) != got) { reason = errno ? errno : EIO; break; }
      total += (int64_t)got;
    }
    if (got < sizeof(buffer)) {
      if (ferror(input)) {
        reason = errno ? errno : EIO;
        failed = from;
      }
      break;
    }
  }
  fclose(input);
  errno = 0;
  if (fflush(output) != 0 && reason == 0) reason = errno ? errno : EIO;
#ifndef __wasi__
  // Only the permission bits: setuid, setgid and sticky stay off the copy.
  if (reason == 0) (void)fchmod(fileno(output), source_info.st_mode & 0777);
#endif
  if (fclose(output) != 0 && reason == 0) reason = errno ? errno : EIO;
  if (reason != 0) {
    *err = narra_io_fail(failed->data, failed->len, strerror(reason));
    return false;
  }
  *copied = total;
  return true;
#endif
}

static bool narra_fs_is_separator(char c) {
  return c == '/' || (NARRA_PATH_WINDOWS && c == '\\');
}

// Why removing the tree at path is refused, or NULL when it is allowed. An
// empty path, a root folder (a drive root on Windows) and a path whose last
// part is "." or ".." would remove more than the caller named.
static const char* narra_fs_remove_all_refusal(lstr* path) {
  if (path->len == 0) return "refusing to remove an empty path";
  int64_t end = path->len;
  while (end > 0 && narra_fs_is_separator(path->data[end - 1])) end--;
  if (end == 0) return "refusing to remove the root folder";
  bool has_drive = NARRA_PATH_WINDOWS && end >= 2 && path->data[1] == ':';
  if (has_drive && end == 2) return "refusing to remove the root of a drive";
  int64_t start = end;
  while (start > 0 && !narra_fs_is_separator(path->data[start - 1])) start--;
  if (has_drive && start == 0) start = 2;
  int64_t length = end - start;
  const char* last = path->data + start;
  if ((length == 1 && last[0] == '.') || (length == 2 && last[0] == '.' && last[1] == '.')) {
    return "refusing to remove a path that ends in . or ..";
  }
  return NULL;
}

#ifdef _WIN32
typedef struct narra_fs_tree_entry {
  char* name;
  DWORD attributes;
} narra_fs_tree_entry;

// Clears the read-only attribute, which makes DeleteFileA and RemoveDirectoryA fail.
static void narra_fs_windows_writable(const char* path, DWORD attributes) {
  if ((attributes & FILE_ATTRIBUTE_READONLY) != 0) SetFileAttributesA(path, attributes & ~FILE_ATTRIBUTE_READONLY);
}

// Removes the folder at path and everything in it. A link or junction inside
// it is removed as a link and never followed. Every name in a folder is read
// before the walk goes deeper, so one search handle is open at a time.
static bool narra_fs_remove_tree(const char* path, DWORD attributes, DWORD* code) {
  size_t base_length = strlen(path);
  char* pattern = malloc(base_length + 3);
  if (!pattern) {
    *code = ERROR_NOT_ENOUGH_MEMORY;
    return false;
  }
  memcpy(pattern, path, base_length);
  memcpy(pattern + base_length, "\\*", 3);
  WIN32_FIND_DATAA found;
  HANDLE search = FindFirstFileA(pattern, &found);
  free(pattern);
  if (search == INVALID_HANDLE_VALUE) {
    *code = GetLastError();
    return false;
  }
  narra_fs_tree_entry* entries = NULL;
  size_t count = 0, capacity = 0;
  DWORD failure = 0;
  do {
    if (strcmp(found.cFileName, ".") == 0 || strcmp(found.cFileName, "..") == 0) continue;
    if (count == capacity) {
      capacity = capacity ? capacity * 2 : 16;
      narra_fs_tree_entry* bigger = realloc(entries, sizeof(narra_fs_tree_entry) * capacity);
      if (!bigger) { failure = ERROR_NOT_ENOUGH_MEMORY; break; }
      entries = bigger;
    }
    entries[count].name = _strdup(found.cFileName);
    if (!entries[count].name) { failure = ERROR_NOT_ENOUGH_MEMORY; break; }
    entries[count].attributes = found.dwFileAttributes;
    count++;
  } while (FindNextFileA(search, &found));
  if (failure == 0 && GetLastError() != ERROR_NO_MORE_FILES) failure = GetLastError();
  FindClose(search);
  for (size_t i = 0; i < count && failure == 0; i++) {
    size_t name_length = strlen(entries[i].name);
    char* child = malloc(base_length + name_length + 2);
    if (!child) { failure = ERROR_NOT_ENOUGH_MEMORY; break; }
    memcpy(child, path, base_length);
    child[base_length] = '\\';
    memcpy(child + base_length + 1, entries[i].name, name_length + 1);
    DWORD child_attributes = entries[i].attributes;
    bool is_folder = (child_attributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
    bool is_link = (child_attributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0;
    if (is_folder && !is_link) {
      DWORD child_code = ERROR_GEN_FAILURE;
      if (!narra_fs_remove_tree(child, child_attributes, &child_code)) failure = child_code;
    } else {
      // A link to a folder or a junction is removed as a folder, its target untouched.
      narra_fs_windows_writable(child, child_attributes);
      BOOL removed = is_folder ? RemoveDirectoryA(child) : DeleteFileA(child);
      if (!removed && GetLastError() != ERROR_FILE_NOT_FOUND) failure = GetLastError();
    }
    free(child);
  }
  for (size_t i = 0; i < count; i++) free(entries[i].name);
  free(entries);
  if (failure != 0) {
    *code = failure;
    return false;
  }
  narra_fs_windows_writable(path, attributes);
  if (!RemoveDirectoryA(path)) {
    *code = GetLastError();
    return false;
  }
  return true;
}

// Removes a folder with everything in it. *removed is false when the path was
// not there. A link or junction inside the tree is removed as a link and never
// followed; a path that names a file or a link fails.
static bool narra_fs_remove_all(lstr* path, bool* removed, lstr** err) {
  *removed = false;
  if (!narra_io_no_nul(path, err)) return false;
  const char* refusal = narra_fs_remove_all_refusal(path);
  if (refusal) {
    *err = path->len == 0 ? lstr_from_cstr(refusal) : narra_io_fail(path->data, path->len, refusal);
    return false;
  }
  DWORD attributes = GetFileAttributesA(path->data);
  if (attributes == INVALID_FILE_ATTRIBUTES) {
    DWORD code = GetLastError();
    if (code == ERROR_FILE_NOT_FOUND || code == ERROR_PATH_NOT_FOUND) return true;
    *err = narra_fs_windows_fail(path, code);
    return false;
  }
  if ((attributes & FILE_ATTRIBUTE_DIRECTORY) == 0 || (attributes & FILE_ATTRIBUTE_REPARSE_POINT) != 0) {
    *err = narra_io_fail(path->data, path->len, "Not a directory");
    return false;
  }
  DWORD code = 0;
  if (!narra_fs_remove_tree(path->data, attributes, &code)) {
    *err = narra_fs_windows_fail(path, code);
    return false;
  }
  *removed = true;
  return true;
}
#else
// macOS hides O_NOFOLLOW under _POSIX_C_SOURCE. The device and inode check in
// narra_fs_open_folder still refuses a folder that was swapped for a link.
#ifndef O_NOFOLLOW
#define O_NOFOLLOW 0
#endif
#ifndef O_CLOEXEC
#define O_CLOEXEC 0
#endif

// One folder of the walk: its names, the next one to remove, its descriptor
// (-1 while closed) and the device and inode that identify it again.
typedef struct narra_fs_tree_level {
  char** names;
  size_t count;
  size_t next;
  int descriptor;
  dev_t device;
  ino_t inode;
} narra_fs_tree_level;

// Opens the folder name inside parent without following a link, and checks it
// is the folder with the given device and inode.
static int narra_fs_open_folder(int parent, const char* name, dev_t device, ino_t inode, const char** message) {
  int descriptor = openat(parent, name, O_RDONLY | O_DIRECTORY | O_NOFOLLOW | O_CLOEXEC);
  if (descriptor < 0) {
    *message = strerror(errno);
    return -1;
  }
  struct stat info;
  if (fstat(descriptor, &info) != 0) {
    *message = strerror(errno);
    close(descriptor);
    return -1;
  }
  if (info.st_dev != device || info.st_ino != inode) {
    *message = "a folder in the tree changed during the removal";
    close(descriptor);
    return -1;
  }
  return descriptor;
}

// Reads every name in the folder open at level->descriptor, which stays open.
static bool narra_fs_read_names(narra_fs_tree_level* level, const char** message) {
  level->names = NULL;
  level->count = 0;
  level->next = 0;
  int listing = openat(level->descriptor, ".", O_RDONLY | O_DIRECTORY | O_CLOEXEC);
  if (listing < 0) {
    *message = strerror(errno);
    return false;
  }
  DIR* directory = fdopendir(listing);
  if (!directory) {
    *message = strerror(errno);
    close(listing);
    return false;
  }
  size_t capacity = 0;
  const char* failure = NULL;
  struct dirent* entry;
  for (;;) {
    errno = 0;
    entry = readdir(directory);
    if (!entry) {
      if (errno != 0) failure = strerror(errno);
      break;
    }
    if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) continue;
    if (level->count == capacity) {
      capacity = capacity ? capacity * 2 : 16;
      char** bigger = realloc(level->names, sizeof(char*) * capacity);
      if (!bigger) { failure = "out of memory"; break; }
      level->names = bigger;
    }
    level->names[level->count] = strdup(entry->d_name);
    if (!level->names[level->count]) { failure = "out of memory"; break; }
    level->count++;
  }
  closedir(directory);
  if (failure) {
    for (size_t i = 0; i < level->count; i++) free(level->names[i]);
    free(level->names);
    level->names = NULL;
    level->count = 0;
    *message = failure;
    return false;
  }
  return true;
}

// Removes the folder at path and everything in it, relative to open folder
// descriptors so a link is never followed and no full path is built. Every
// name in a folder is read before the walk goes deeper, and outside WASI the
// parent is closed then and opened again through ".." on the way back, so the
// walk holds at most three descriptors however deep the tree is. WASI keeps
// one descriptor per level, because it refuses ".." through a descriptor.
static bool narra_fs_remove_tree(const char* path, const struct stat* top, const char** message) {
  int descriptor = narra_fs_open_folder(AT_FDCWD, path, top->st_dev, top->st_ino, message);
  if (descriptor < 0) return false;
  size_t capacity = 16, depth = 1;
  narra_fs_tree_level* levels = malloc(sizeof(narra_fs_tree_level) * capacity);
  if (!levels) {
    close(descriptor);
    *message = "out of memory";
    return false;
  }
  levels[0].descriptor = descriptor;
  levels[0].device = top->st_dev;
  levels[0].inode = top->st_ino;
  bool ok = narra_fs_read_names(&levels[0], message);
  while (ok && depth > 0) {
    narra_fs_tree_level* level = &levels[depth - 1];
    if (level->next < level->count) {
      const char* name = level->names[level->next++];
      struct stat info;
      if (fstatat(level->descriptor, name, &info, AT_SYMLINK_NOFOLLOW) != 0) {
        if (errno == ENOENT) continue;
        *message = strerror(errno);
        ok = false;
        break;
      }
      if (!S_ISDIR(info.st_mode)) {
        if (unlinkat(level->descriptor, name, 0) != 0 && errno != ENOENT) {
          *message = strerror(errno);
          ok = false;
        }
        continue;
      }
      int child = narra_fs_open_folder(level->descriptor, name, info.st_dev, info.st_ino, message);
      if (child < 0) {
        ok = false;
        break;
      }
      if (depth == capacity) {
        narra_fs_tree_level* bigger = realloc(levels, sizeof(narra_fs_tree_level) * capacity * 2);
        if (!bigger) {
          close(child);
          *message = "out of memory";
          ok = false;
          break;
        }
        levels = bigger;
        capacity *= 2;
        level = &levels[depth - 1];
      }
      narra_fs_tree_level* next = &levels[depth];
      next->descriptor = child;
      next->device = info.st_dev;
      next->inode = info.st_ino;
      if (!narra_fs_read_names(next, message)) {
        close(child);
        ok = false;
        break;
      }
#ifndef __wasi__
      close(level->descriptor);
      level->descriptor = -1;
#endif
      depth++;
      continue;
    }
    // Every entry is gone: free the names, step back to the parent and remove
    // this folder from it.
    for (size_t i = 0; i < level->count; i++) free(level->names[i]);
    free(level->names);
    level->names = NULL;
    level->count = 0;
    if (depth == 1) {
      close(level->descriptor);
      depth = 0;
      break;
    }
    narra_fs_tree_level* parent = &levels[depth - 2];
    if (parent->descriptor < 0) {
      parent->descriptor = narra_fs_open_folder(level->descriptor, "..", parent->device, parent->inode, message);
    }
    close(level->descriptor);
    depth--;
    if (parent->descriptor < 0) {
      ok = false;
      break;
    }
    if (unlinkat(parent->descriptor, parent->names[parent->next - 1], AT_REMOVEDIR) != 0 && errno != ENOENT) {
      *message = strerror(errno);
      ok = false;
    }
  }
  // After a failure, release what the levels still hold.
  for (size_t i = 0; i < depth; i++) {
    for (size_t j = 0; j < levels[i].count; j++) free(levels[i].names[j]);
    free(levels[i].names);
    if (levels[i].descriptor >= 0) close(levels[i].descriptor);
  }
  free(levels);
  if (!ok) return false;
  if (rmdir(path) != 0) {
    *message = strerror(errno);
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
  const char* refusal = narra_fs_remove_all_refusal(path);
  if (refusal) {
    *err = path->len == 0 ? lstr_from_cstr(refusal) : narra_io_fail(path->data, path->len, refusal);
    return false;
  }
  struct stat info;
  if (lstat(path->data, &info) != 0) {
    if (errno == ENOENT) return true;
    *err = narra_io_fail(path->data, path->len, strerror(errno));
    return false;
  }
  if (!S_ISDIR(info.st_mode)) {
    *err = narra_io_fail(path->data, path->len, "Not a directory");
    return false;
  }
  const char* message = "unknown failure";
  if (!narra_fs_remove_tree(path->data, &info, &message)) {
    *err = narra_io_fail(path->data, path->len, message);
    return false;
  }
  *removed = true;
  return true;
}
#endif

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

