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

// The absolute working folder, as the system spells it.
static bool narra_proc_cwd(lstr** out, lstr** err) {
  size_t capacity = 256;
  for (;;) {
    char* buffer = malloc(capacity);
    if (!buffer) {
      *err = lstr_from_cstr("out of memory");
      return false;
    }
    if (getcwd(buffer, capacity) != NULL) {
      *out = narra_io_lossy(buffer, strlen(buffer));
      free(buffer);
      return true;
    }
    int reason = errno;
    free(buffer);
    if (reason != ERANGE) {
      *err = lstr_from_cstr(strerror(reason));
      return false;
    }
    capacity *= 2;
  }
}

#ifdef __wasi__
// WASI preview 1 starts no processes. The compiler refuses Process.run on this
// target (E0991); the stub keeps a build that is not checked linking.
static bool narra_proc_run(lstr* cmd, lstr** argv, int64_t argc, lstr* input, bool has_input, lstr** added, int64_t added_count, lstr** removed, int64_t removed_count, lstr* folder, bool inherit, int64_t* code, lstr** out, lstr** errout, lstr** err) {
  (void)cmd; (void)argv; (void)argc; (void)input; (void)has_input; (void)added; (void)added_count;
  (void)removed; (void)removed_count; (void)folder; (void)inherit; (void)code; (void)out; (void)errout;
  *err = lstr_from_cstr("processes are not available on WASI");
  return false;
}
#elif defined(_WIN32)
#include <winsock2.h>
#include <windows.h>

// Processes on Windows: CreateProcessW with a command line built from the
// arguments, one pipe per output stream with a thread that drains it, and a
// job object that ends the child when this process ends. There is no fork and
// no signal forwarding; a console sends Ctrl-C to the child itself.
typedef struct narra_proc_buf {
  char* p;
  size_t len, cap;
} narra_proc_buf;

typedef struct narra_proc_reader {
  HANDLE pipe;
  narra_proc_buf buffer;
  bool failed;
} narra_proc_reader;

// False when the buffer cannot grow; the bytes are then not stored.
static bool narra_proc_append(narra_proc_buf* b, const char* chunk, size_t n) {
  if (n > SIZE_MAX - b->len) return false;
  if (b->len + n > b->cap) {
    size_t cap = b->cap ? b->cap : 4096;
    while (cap < b->len + n) {
      if (cap > SIZE_MAX / 2) return false;
      cap *= 2;
    }
    char* bigger = realloc(b->p, cap);
    if (!bigger) return false;
    b->p = bigger;
    b->cap = cap;
  }
  memcpy(b->p + b->len, chunk, n);
  b->len += n;
  return true;
}

// Reads the pipe to its end. After an allocation failure it keeps reading and
// drops the bytes, so the child never blocks on a full pipe; the call then fails.
static DWORD WINAPI narra_proc_drain(LPVOID argument) {
  narra_proc_reader* reader = (narra_proc_reader*)argument;
  char chunk[4096];
  DWORD count;
  while (ReadFile(reader->pipe, chunk, sizeof chunk, &count, NULL) && count > 0) {
    if (!reader->failed && !narra_proc_append(&reader->buffer, chunk, count)) reader->failed = true;
  }
  return 0;
}

// The text of a Windows error code, with the POSIX wording for the two codes
// a missing program gives so the message matches the other systems.
static const char* narra_proc_message(DWORD code, char* buffer, size_t size) {
  if (code == ERROR_FILE_NOT_FOUND || code == ERROR_PATH_NOT_FOUND) return "No such file or directory";
  if (code == ERROR_ACCESS_DENIED) return "Permission denied";
  DWORD n = FormatMessageA(FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS, NULL, code, 0, buffer, (DWORD)size, NULL);
  while (n > 0 && (buffer[n - 1] == '\r' || buffer[n - 1] == '\n' || buffer[n - 1] == ' ' || buffer[n - 1] == '.')) buffer[--n] = 0;
  if (n == 0) snprintf(buffer, size, "error %lu", (unsigned long)code);
  return buffer;
}

// Appends one argument to the command line in the form CommandLineToArgvW
// reads back as the same argument. False when the line cannot grow.
static bool narra_proc_quote(narra_proc_buf* line, const char* text, size_t length) {
  bool plain = length > 0;
  for (size_t i = 0; i < length; i++) {
    char c = text[i];
    if (c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '"') plain = false;
  }
  if (plain) return narra_proc_append(line, text, length);
  if (!narra_proc_append(line, "\"", 1)) return false;
  size_t i = 0;
  while (i < length) {
    size_t slashes = 0;
    while (i < length && text[i] == '\\') { slashes++; i++; }
    if (i == length) {
      // Backslashes before the closing quote are doubled.
      for (size_t k = 0; k < slashes * 2; k++) {
        if (!narra_proc_append(line, "\\", 1)) return false;
      }
    } else if (text[i] == '"') {
      for (size_t k = 0; k < slashes * 2 + 1; k++) {
        if (!narra_proc_append(line, "\\", 1)) return false;
      }
      if (!narra_proc_append(line, "\"", 1)) return false;
      i++;
    } else {
      for (size_t k = 0; k < slashes; k++) {
        if (!narra_proc_append(line, "\\", 1)) return false;
      }
      if (!narra_proc_append(line, text + i, 1)) return false;
      i++;
    }
  }
  return narra_proc_append(line, "\"", 1);
}

// The UTF-16 form of the command line, or NULL when it is not valid UTF-8.
static wchar_t* narra_proc_wide(const char* text, size_t length) {
  int count = MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, text, (int)length, NULL, 0);
  if (count <= 0 && length > 0) return NULL;
  wchar_t* wide = malloc(((size_t)count + 1) * sizeof(wchar_t));
  if (!wide) return NULL;
  MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, text, (int)length, wide, count);
  wide[count] = 0;
  return wide;
}

static void narra_proc_close_handle(HANDLE* handle) {
  if (*handle != NULL && *handle != INVALID_HANDLE_VALUE) CloseHandle(*handle);
  *handle = NULL;
}

// A pipe whose read end stays with this process; only the write end is inherited.
static bool narra_proc_pipe(HANDLE* read_end, HANDLE* write_end) {
  SECURITY_ATTRIBUTES attributes;
  memset(&attributes, 0, sizeof attributes);
  attributes.nLength = sizeof attributes;
  attributes.bInheritHandle = TRUE;
  if (!CreatePipe(read_end, write_end, &attributes, 0)) return false;
  return SetHandleInformation(*read_end, HANDLE_FLAG_INHERIT, 0) != 0;
}

// True when path names a file that is not a directory.
static bool narra_proc_is_file(const wchar_t* path) {
  DWORD attributes = GetFileAttributesW(path);
  return attributes != INVALID_FILE_ATTRIBUTES && (attributes & FILE_ATTRIBUTE_DIRECTORY) == 0;
}

// "<directory>\<name><suffix>" when that names a file, else NULL. A zero
// length directory means name is used as written. *memory is set when an
// allocation failed.
static wchar_t* narra_proc_candidate(const wchar_t* directory, size_t directory_length, const wchar_t* name, const wchar_t* suffix, bool* memory) {
  size_t name_length = wcslen(name);
  size_t suffix_length = wcslen(suffix);
  wchar_t* path = malloc((directory_length + name_length + suffix_length + 2) * sizeof(wchar_t));
  if (!path) {
    *memory = true;
    return NULL;
  }
  size_t length = 0;
  if (directory_length > 0) {
    memcpy(path, directory, directory_length * sizeof(wchar_t));
    length = directory_length;
    if (path[length - 1] != L'\\' && path[length - 1] != L'/') path[length++] = L'\\';
  }
  memcpy(path + length, name, name_length * sizeof(wchar_t));
  length += name_length;
  memcpy(path + length, suffix, (suffix_length + 1) * sizeof(wchar_t));
  if (narra_proc_is_file(path)) return path;
  free(path);
  return NULL;
}

// Tries name as written when it ends in .exe, .com, .bat or .cmd, then name with .exe, then
// with .com, inside one directory. Only these two extensions run directly;
// .bat and .cmd need the shell and are not guessed.
static wchar_t* narra_proc_probe(const wchar_t* directory, size_t directory_length, const wchar_t* name, bool* memory) {
  const wchar_t* base = name;
  for (const wchar_t* c = name; *c; c++) {
    if (*c == L'\\' || *c == L'/' || *c == L':') base = c + 1;
  }
  // Only a program extension counts: a name such as python3.11 is not tried as written, so a
  // file of that name in an earlier directory does not hide python3.11.exe.
  const wchar_t* dot = wcsrchr(base, L'.');
  bool has_extension = dot != NULL && (_wcsicmp(dot, L".exe") == 0 || _wcsicmp(dot, L".com") == 0 || _wcsicmp(dot, L".bat") == 0 || _wcsicmp(dot, L".cmd") == 0);
  wchar_t* found = NULL;
  if (has_extension) found = narra_proc_candidate(directory, directory_length, name, L"", memory);
  if (!found && !*memory) found = narra_proc_candidate(directory, directory_length, name, L".exe", memory);
  if (!found && !*memory) found = narra_proc_candidate(directory, directory_length, name, L".com", memory);
  return found;
}

// The program a command names, as a path for lpApplicationName. A name with a
// directory part is taken as written. A bare name is looked for in each
// directory of PATH, in order; the current directory and the folder of this
// program are never searched, which CreateProcess would do for a bare name.
// path_list is the PATH of the child, or NULL when it has none, and then a bare
// name is not found here. NULL when nothing is found; *memory is set when an
// allocation failed.
static wchar_t* narra_proc_find(const wchar_t* name, const wchar_t* path_list, bool* memory) {
  bool bare = true;
  for (const wchar_t* c = name; *c; c++) {
    if (*c == L'\\' || *c == L'/' || *c == L':') bare = false;
  }
  if (!bare) return narra_proc_probe(L"", 0, name, memory);
  if (!path_list || !*path_list) return NULL;
  size_t size = wcslen(path_list) + 1;
  wchar_t* directory = malloc(size * sizeof(wchar_t));
  if (!directory) {
    *memory = true;
    return NULL;
  }
  wchar_t* found = NULL;
  {
    const wchar_t* entry = path_list;
    while (!found && !*memory) {
      const wchar_t* end = wcschr(entry, L';');
      size_t entry_length = end ? (size_t)(end - entry) : wcslen(entry);
      // An entry may be quoted; the quotes are not part of the path.
      size_t length = 0;
      for (size_t i = 0; i < entry_length; i++) {
        if (entry[i] != L'"') directory[length++] = entry[i];
      }
      directory[length] = 0;
      if (length > 0) found = narra_proc_probe(directory, length, name, memory);
      if (!end) break;
      entry = end + 1;
    }
  }
  free(directory);
  return found;
}

// Writes the whole text to the pipe, then closes it so the child sees the end.
// A child that stops reading ends the write with an error, and the rest is dropped.
typedef struct narra_proc_writer {
  HANDLE pipe;
  const char* data;
  size_t len;
} narra_proc_writer;

static DWORD WINAPI narra_proc_feed(LPVOID argument) {
  narra_proc_writer* writer = (narra_proc_writer*)argument;
  size_t done = 0;
  while (done < writer->len) {
    size_t rest = writer->len - done;
    DWORD chunk = rest > 65536 ? 65536 : (DWORD)rest;
    DWORD count = 0;
    if (!WriteFile(writer->pipe, writer->data + done, chunk, &count, NULL) || count == 0) break;
    done += count;
  }
  CloseHandle(writer->pipe);
  writer->pipe = NULL;
  return 0;
}

// A pipe whose write end stays with this process; only the read end is inherited.
static bool narra_proc_input_pipe(HANDLE* read_end, HANDLE* write_end) {
  SECURITY_ATTRIBUTES attributes;
  memset(&attributes, 0, sizeof attributes);
  attributes.nLength = sizeof attributes;
  attributes.bInheritHandle = TRUE;
  if (!CreatePipe(read_end, write_end, &attributes, 0)) return false;
  return SetHandleInformation(*write_end, HANDLE_FLAG_INHERIT, 0) != 0;
}

// The standard handle of this process, made inheritable, or NULL when it has none.
// *flags is the inherit flag it had, for narra_proc_restore_standard.
static HANDLE narra_proc_standard(DWORD which, DWORD* flags) {
  HANDLE handle = GetStdHandle(which);
  if (handle == NULL || handle == INVALID_HANDLE_VALUE) return NULL;
  DWORD before = 0;
  if (!GetHandleInformation(handle, &before)) return NULL;
  *flags = before & HANDLE_FLAG_INHERIT;
  if (!SetHandleInformation(handle, HANDLE_FLAG_INHERIT, HANDLE_FLAG_INHERIT)) return NULL;
  return handle;
}

// Puts the inherit flag of a standard handle back as narra_proc_standard found it.
static void narra_proc_restore_standard(HANDLE handle, DWORD flags) {
  if (handle != NULL) SetHandleInformation(handle, HANDLE_FLAG_INHERIT, flags);
}

// The UTF-16 environment block of the child: the variables of this process
// without the removed names and the names set again (Windows compares names
// without case), then the added entries. Entries that start with '=' are the
// per-drive folders and stay. NULL with *memory or *invalid set on failure.
static wchar_t* narra_proc_environment(lstr** added, int64_t added_count, lstr** removed, int64_t removed_count, bool* memory, bool* invalid) {
  wchar_t** added_wide = calloc((size_t)added_count + 1, sizeof(wchar_t*));
  wchar_t** removed_wide = calloc((size_t)removed_count + 1, sizeof(wchar_t*));
  wchar_t* block = NULL;
  wchar_t* current = GetEnvironmentStringsW();
  bool failed = !added_wide || !removed_wide || !current;
  for (int64_t i = 0; !failed && i < added_count; i++) {
    added_wide[i] = narra_proc_wide(added[i]->data, (size_t)added[i]->len);
    if (!added_wide[i]) { failed = true; *invalid = true; }
  }
  for (int64_t i = 0; !failed && i < removed_count; i++) {
    removed_wide[i] = narra_proc_wide(removed[i]->data, (size_t)removed[i]->len);
    if (!removed_wide[i]) { failed = true; *invalid = true; }
  }
  size_t total = 1;
  if (!failed) {
    for (const wchar_t* entry = current; *entry; entry += wcslen(entry) + 1) total += wcslen(entry) + 1;
    for (int64_t i = 0; i < added_count; i++) total += wcslen(added_wide[i]) + 1;
    block = malloc(total * sizeof(wchar_t));
    if (!block) failed = true;
  }
  if (!failed) {
    size_t used = 0;
    for (const wchar_t* entry = current; *entry; entry += wcslen(entry) + 1) {
      const wchar_t* equals = wcschr(entry + 1, L'=');
      size_t name_length = equals ? (size_t)(equals - entry) : wcslen(entry);
      bool dropped = false;
      for (int64_t i = 0; i < removed_count && !dropped; i++) {
        dropped = wcslen(removed_wide[i]) == name_length && _wcsnicmp(entry, removed_wide[i], name_length) == 0;
      }
      for (int64_t i = 0; i < added_count && !dropped; i++) {
        const wchar_t* added_equals = wcschr(added_wide[i], L'=');
        dropped = (size_t)(added_equals - added_wide[i]) == name_length && _wcsnicmp(entry, added_wide[i], name_length) == 0;
      }
      if (dropped) continue;
      size_t length = wcslen(entry) + 1;
      memcpy(block + used, entry, length * sizeof(wchar_t));
      used += length;
    }
    // A name set twice, in any case, keeps its last value.
    for (int64_t i = 0; i < added_count; i++) {
      size_t name_length = (size_t)(wcschr(added_wide[i], L'=') - added_wide[i]);
      bool replaced = false;
      for (int64_t j = i + 1; j < added_count && !replaced; j++) {
        const wchar_t* later = wcschr(added_wide[j], L'=');
        replaced = (size_t)(later - added_wide[j]) == name_length && _wcsnicmp(added_wide[i], added_wide[j], name_length) == 0;
      }
      if (replaced) continue;
      size_t length = wcslen(added_wide[i]) + 1;
      memcpy(block + used, added_wide[i], length * sizeof(wchar_t));
      used += length;
    }
    block[used] = 0;
  }
  if (current) FreeEnvironmentStringsW(current);
  for (int64_t i = 0; added_wide && i < added_count; i++) free(added_wide[i]);
  for (int64_t i = 0; removed_wide && i < removed_count; i++) free(removed_wide[i]);
  free(added_wide);
  free(removed_wide);
  if (failed) {
    free(block);
    if (!*invalid) *memory = true;
    return NULL;
  }
  return block;
}

// Runs cmd (found through PATH) with argv[0..argc) as its arguments. Its
// standard input is the text input when has_input is set, and closed otherwise.
// The environment is this process's, with the added entries (NAME=VALUE) set and
// the removed names unset; a folder that is not empty is the working folder. When
// inherit is set, the child writes to the standard handles of this process and
// the captured text is empty. Returns false with *err set when the process cannot
// start. The exit code is the one the process ended with.
static bool narra_proc_run(lstr* cmd, lstr** argv, int64_t argc, lstr* input, bool has_input, lstr** added, int64_t added_count, lstr** removed, int64_t removed_count, lstr* folder, bool inherit, int64_t* code, lstr** out, lstr** errout, lstr** err) {
  if (!narra_io_no_nul(cmd, err)) return false;
  for (int64_t i = 0; i < argc; i++) {
    if (!narra_io_no_nul(argv[i], err)) return false;
  }
  if (!narra_io_no_nul(folder, err)) return false;
  for (int64_t i = 0; i < added_count; i++) {
    if (!narra_io_no_nul(added[i], err)) return false;
    const char* equals = strchr(added[i]->data, '=');
    if (!equals || equals == added[i]->data) {
      *err = lstr_from_cstr("environment variable needs a name");
      return false;
    }
  }
  for (int64_t i = 0; i < removed_count; i++) {
    if (!narra_io_no_nul(removed[i], err)) return false;
    if (removed[i]->len == 0 || strchr(removed[i]->data, '=')) {
      *err = lstr_from_cstr("environment variable name is empty or holds '='");
      return false;
    }
  }
  // The environment block and the working folder are built before anything starts.
  wchar_t* environment = NULL;
  if (added_count > 0 || removed_count > 0) {
    bool environment_memory = false, environment_invalid = false;
    environment = narra_proc_environment(added, added_count, removed, removed_count, &environment_memory, &environment_invalid);
    if (!environment) {
      *err = narra_io_fail(cmd->data, cmd->len, environment_invalid ? "environment is not valid UTF-8" : "out of memory");
      return false;
    }
  }
  wchar_t* directory = NULL;
  if (folder->len > 0) {
    directory = narra_proc_wide(folder->data, (size_t)folder->len);
    if (!directory) {
      free(environment);
      *err = narra_io_fail(cmd->data, cmd->len, "folder is not valid UTF-8");
      return false;
    }
  }
  narra_proc_buf line = {0};
  bool built = narra_proc_quote(&line, cmd->data, (size_t)cmd->len);
  for (int64_t i = 0; built && i < argc; i++) {
    built = narra_proc_append(&line, " ", 1) && narra_proc_quote(&line, argv[i]->data, (size_t)argv[i]->len);
  }
  if (!built) {
    free(line.p);
    free(environment);
    free(directory);
    *err = narra_io_fail(cmd->data, cmd->len, "out of memory");
    return false;
  }
  wchar_t* wide = line.p ? narra_proc_wide(line.p, line.len) : NULL;
  free(line.p);
  if (!wide) {
    free(environment);
    free(directory);
    *err = narra_io_fail(cmd->data, cmd->len, "arguments are not valid UTF-8");
    return false;
  }
  // The program is resolved here, in the PATH the child gets, so CreateProcess
  // never searches the current directory or the folder of this program. A child
  // with no PATH leaves the search to CreateProcess.
  wchar_t* program_name = narra_proc_wide(cmd->data, (size_t)cmd->len);
  bool memory = program_name == NULL;
  wchar_t* parent_path = NULL;
  const wchar_t* path_list = NULL;
  if (environment) {
    for (const wchar_t* entry = environment; *entry; entry += wcslen(entry) + 1) {
      if (_wcsnicmp(entry, L"PATH=", 5) == 0) path_list = entry + 5;
    }
  } else {
    DWORD path_size = GetEnvironmentVariableW(L"PATH", NULL, 0);
    parent_path = path_size > 0 ? malloc(path_size * sizeof(wchar_t)) : NULL;
    if (parent_path && GetEnvironmentVariableW(L"PATH", parent_path, path_size) > 0) path_list = parent_path;
  }
  bool bare_name = true;
  for (const wchar_t* c = program_name ? program_name : L""; *c; c++) {
    if (*c == L'\\' || *c == L'/' || *c == L':') bare_name = false;
  }
  bool default_search = program_name != NULL && bare_name && path_list == NULL;
  wchar_t* program = program_name && !default_search ? narra_proc_find(program_name, path_list, &memory) : NULL;
  free(program_name);
  free(parent_path);
  if (!program && !default_search) {
    *err = narra_io_fail(cmd->data, cmd->len, memory ? "out of memory" : "No such file or directory");
    free(wide);
    free(environment);
    free(directory);
    return false;
  }
  HANDLE out_read = NULL, out_write = NULL, err_read = NULL, err_write = NULL, input_read = NULL, input_write = NULL, job = NULL;
  SECURITY_ATTRIBUTES inherit_handle;
  memset(&inherit_handle, 0, sizeof inherit_handle);
  inherit_handle.nLength = sizeof inherit_handle;
  inherit_handle.bInheritHandle = TRUE;
  // The child reads the input pipe, or the null device when there is no input.
  bool ready = true;
  if (has_input) {
    ready = narra_proc_input_pipe(&input_read, &input_write);
  } else {
    input_read = CreateFileA("NUL", GENERIC_READ, FILE_SHARE_READ | FILE_SHARE_WRITE, &inherit_handle, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (input_read == INVALID_HANDLE_VALUE) input_read = NULL;
    ready = input_read != NULL;
  }
  // Passed-through output uses the standard handles of this process, which stay
  // open; captured output uses a pipe per stream.
  bool owns_output = !inherit;
  DWORD out_flags = 0, err_flags = 0;
  HANDLE out_standard = NULL, err_standard = NULL;
  if (ready && inherit) {
    out_write = narra_proc_standard(STD_OUTPUT_HANDLE, &out_flags);
    err_write = narra_proc_standard(STD_ERROR_HANDLE, &err_flags);
    out_standard = out_write;
    err_standard = err_write;
    if (!out_write) {
      out_write = CreateFileA("NUL", GENERIC_WRITE, FILE_SHARE_READ | FILE_SHARE_WRITE, &inherit_handle, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
      if (out_write == INVALID_HANDLE_VALUE) out_write = NULL;
      owns_output = true;
    }
    if (!err_write) {
      err_write = CreateFileA("NUL", GENERIC_WRITE, FILE_SHARE_READ | FILE_SHARE_WRITE, &inherit_handle, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
      if (err_write == INVALID_HANDLE_VALUE) err_write = NULL;
      owns_output = true;
    }
    ready = out_write != NULL && err_write != NULL;
  } else if (ready) {
    ready = narra_proc_pipe(&out_read, &out_write) && narra_proc_pipe(&err_read, &err_write);
  }
  if (!ready) {
    char text[256];
    *err = narra_io_fail(cmd->data, cmd->len, narra_proc_message(GetLastError(), text, sizeof text));
    narra_proc_restore_standard(out_standard, out_flags);
    narra_proc_restore_standard(err_standard, err_flags);
    narra_proc_close_handle(&input_read);
    narra_proc_close_handle(&input_write);
    narra_proc_close_handle(&out_read); narra_proc_close_handle(&err_read);
    if (owns_output) { narra_proc_close_handle(&out_write); narra_proc_close_handle(&err_write); }
    free(wide);
    free(program);
    free(environment);
    free(directory);
    return false;
  }
  fflush(NULL);
  // Only the three standard handles pass to the child, so a process started
  // by another thread at the same time does not hold these pipes open. The
  // list holds each handle once: output and error may be the same handle.
  HANDLE inherited[3] = {input_read, out_write, err_write};
  DWORD inherited_count = 3;
  if (err_write == out_write) inherited_count = 2;
  STARTUPINFOEXW startup;
  memset(&startup, 0, sizeof startup);
  startup.StartupInfo.cb = sizeof startup;
  startup.StartupInfo.dwFlags = STARTF_USESTDHANDLES;
  startup.StartupInfo.hStdInput = input_read;
  startup.StartupInfo.hStdOutput = out_write;
  startup.StartupInfo.hStdError = err_write;
  SIZE_T list_size = 0;
  InitializeProcThreadAttributeList(NULL, 1, 0, &list_size);
  startup.lpAttributeList = (LPPROC_THREAD_ATTRIBUTE_LIST)malloc(list_size);
  bool listed = startup.lpAttributeList != NULL &&
    InitializeProcThreadAttributeList(startup.lpAttributeList, 1, 0, &list_size) &&
    UpdateProcThreadAttribute(startup.lpAttributeList, 0, PROC_THREAD_ATTRIBUTE_HANDLE_LIST, inherited, inherited_count * sizeof(HANDLE), NULL, NULL);
  PROCESS_INFORMATION process;
  memset(&process, 0, sizeof process);
  BOOL started = FALSE;
  DWORD failure = ERROR_OUTOFMEMORY;
  if (listed) {
    // A hidden console would hide the output the parent passes through, so
    // only captured output starts the child without a window.
    DWORD flags = EXTENDED_STARTUPINFO_PRESENT | CREATE_SUSPENDED;
    if (!inherit) flags |= CREATE_NO_WINDOW;
    if (environment) flags |= CREATE_UNICODE_ENVIRONMENT;
    started = CreateProcessW(program, wide, NULL, NULL, TRUE, flags, environment, directory, &startup.StartupInfo, &process);
    if (!started) failure = GetLastError();
  }
  // The parent's own handles stop being inheritable once the child has them.
  narra_proc_restore_standard(out_standard, out_flags);
  narra_proc_restore_standard(err_standard, err_flags);
  if (startup.lpAttributeList) {
    if (listed) DeleteProcThreadAttributeList(startup.lpAttributeList);
    free(startup.lpAttributeList);
  }
  free(wide);
  free(program);
  free(environment);
  bool folder_failure = directory != NULL && (failure == ERROR_DIRECTORY || failure == ERROR_PATH_NOT_FOUND);
  free(directory);
  narra_proc_close_handle(&input_read);
  if (owns_output) {
    narra_proc_close_handle(&out_write);
    narra_proc_close_handle(&err_write);
  }
  if (!started) {
    char text[256];
    if (folder_failure) *err = narra_io_fail(folder->data, folder->len, narra_proc_message(failure, text, sizeof text));
    else *err = narra_io_fail(cmd->data, cmd->len, narra_proc_message(failure, text, sizeof text));
    narra_proc_close_handle(&input_write);
    narra_proc_close_handle(&out_read);
    narra_proc_close_handle(&err_read);
    return false;
  }
  // A job that ends its processes when its handle closes, so the child does
  // not outlive this process. A failure here leaves the child unguarded.
  job = CreateJobObjectW(NULL, NULL);
  if (job) {
    JOBOBJECT_EXTENDED_LIMIT_INFORMATION limits;
    memset(&limits, 0, sizeof limits);
    limits.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
    SetInformationJobObject(job, JobObjectExtendedLimitInformation, &limits, sizeof limits);
    AssignProcessToJobObject(job, process.hProcess);
  }
  narra_proc_reader readers[2];
  memset(readers, 0, sizeof readers);
  readers[0].pipe = out_read;
  readers[1].pipe = err_read;
  narra_proc_writer writer;
  memset(&writer, 0, sizeof writer);
  writer.pipe = input_write;
  writer.data = has_input ? input->data : "";
  writer.len = has_input ? (size_t)input->len : 0;
  // Each stream needs a thread: draining one pipe after the other, or writing
  // the input before reading, can deadlock when the child fills a pipe. The
  // threads start before the child runs, and the call fails, with the child
  // ended, when one cannot.
  HANDLE threads[3] = {NULL, NULL, NULL};
  bool started_all = true;
  if (!inherit) {
    for (int i = 0; i < 2; i++) {
      threads[i] = CreateThread(NULL, 0, narra_proc_drain, &readers[i], 0, NULL);
      if (!threads[i]) started_all = false;
    }
  }
  if (has_input) {
    threads[2] = CreateThread(NULL, 0, narra_proc_feed, &writer, 0, NULL);
    if (!threads[2]) started_all = false;
  }
  if (!started_all) {
    TerminateProcess(process.hProcess, 1);
    for (int i = 0; i < 3; i++) {
      if (threads[i]) {
        WaitForSingleObject(threads[i], INFINITE);
        CloseHandle(threads[i]);
      }
    }
    *err = narra_io_fail(cmd->data, cmd->len, "cannot start a thread to read the output");
    narra_proc_close_handle(&process.hThread);
    narra_proc_close_handle(&process.hProcess);
    narra_proc_close_handle(&job);
    narra_proc_close_handle(&out_read);
    narra_proc_close_handle(&err_read);
    if (!threads[2]) narra_proc_close_handle(&input_write);
    free(readers[0].buffer.p);
    free(readers[1].buffer.p);
    return false;
  }
  ResumeThread(process.hThread);
  narra_proc_close_handle(&process.hThread);
  for (int i = 0; i < 3; i++) {
    if (!threads[i]) continue;
    WaitForSingleObject(threads[i], INFINITE);
    CloseHandle(threads[i]);
  }
  WaitForSingleObject(process.hProcess, INFINITE);
  DWORD exit_code = 0;
  GetExitCodeProcess(process.hProcess, &exit_code);
  narra_proc_close_handle(&process.hProcess);
  narra_proc_close_handle(&job);
  narra_proc_close_handle(&out_read);
  narra_proc_close_handle(&err_read);
  bool complete = !readers[0].failed && !readers[1].failed;
  if (complete) {
    *code = (int64_t)exit_code;
    *out = narra_io_lossy(readers[0].buffer.p ? readers[0].buffer.p : "", readers[0].buffer.len);
    *errout = narra_io_lossy(readers[1].buffer.p ? readers[1].buffer.p : "", readers[1].buffer.len);
  } else {
    *err = narra_io_fail(cmd->data, cmd->len, "out of memory");
  }
  free(readers[0].buffer.p);
  free(readers[1].buffer.p);
  return complete;
}
#else
#include <fcntl.h>
#include <poll.h>
#include <signal.h>
#include <pthread.h>
#include <sched.h>
#include <stdatomic.h>
#include <sys/wait.h>
#include <unistd.h>

// Shared with std/internal/signal: once std/signal is in use, a SIGINT or SIGTERM that comes
// while a child runs is also recorded for it, and the handlers of std/signal come back after the call.
#ifndef NARRA_SIGNAL_COMMON
#define NARRA_SIGNAL_COMMON
typedef void (*narra_signal_function)(int);
static _Atomic int narra_signal_writer = -1;
static _Atomic(narra_signal_function) narra_signal_handler_in_use = NULL;

static void narra_signal_note(int number) {
  int saved = errno;
  int writer = atomic_load(&narra_signal_writer);
  if (writer >= 0 && (number == SIGINT || number == SIGTERM)) {
    unsigned char code = number == SIGINT ? 1 : 2;
    ssize_t ignored = write(writer, &code, 1);
    (void)ignored;
  }
  errno = saved;
}
#endif

typedef struct narra_proc_buf {
  char* p;
  size_t len, cap;
  bool failed;
} narra_proc_buf;

// Reads one chunk from fd. False at end of input or on a hard error. After an
// allocation failure the bytes are dropped but the pipe is still read to its
// end, so the child does not block; the call then fails.
static bool narra_proc_take(narra_proc_buf* b, int fd) {
  char chunk[4096];
  ssize_t n = read(fd, chunk, sizeof chunk);
  if (n < 0) return errno == EINTR || errno == EAGAIN;
  if (n == 0) return false;
  if (b->failed) return true;
  if (b->len + (size_t)n > b->cap) {
    size_t cap = b->cap ? b->cap * 2 : 4096;
    while (cap < b->len + (size_t)n) cap *= 2;
    char* bigger = realloc(b->p, cap);
    if (!bigger) {
      b->failed = true;
      return true;
    }
    b->p = bigger;
    b->cap = cap;
  }
  memcpy(b->p + b->len, chunk, (size_t)n);
  b->len += (size_t)n;
  return true;
}

static void narra_proc_close(int* fds) {
  if (fds[0] >= 0) close(fds[0]);
  if (fds[1] >= 0) close(fds[1]);
}

static int narra_proc_pipe(int* fds) {
  fds[0] = fds[1] = -1;
  if (pipe(fds) != 0) return -1;
  fcntl(fds[0], F_SETFD, FD_CLOEXEC);
  fcntl(fds[1], F_SETFD, FD_CLOEXEC);
  return 0;
}

// While a call runs, SIGTERM, SIGINT and SIGHUP are passed on to its child, so
// killing this process does not orphan the child. Several threads may call at
// once: a count of active calls (under a mutex) installs the handlers for the
// first call and restores them after the last, and each call owns a slot in a
// fixed array of child pids that the handler walks with plain atomic loads.
// A slot holds -1 from the claim until the pid is stored; a signal that
// arrives in between is remembered in the pending array and sent on attach.
#define narra_proc_SLOTS 64
static pthread_mutex_t narra_proc_lock = PTHREAD_MUTEX_INITIALIZER;
static int narra_proc_active = 0;
static struct sigaction narra_proc_saved[3];
static _Atomic int narra_proc_pids[narra_proc_SLOTS];
static _Atomic int narra_proc_pending[narra_proc_SLOTS];
static const int narra_proc_signals[3] = {SIGTERM, SIGINT, SIGHUP};

typedef struct narra_proc_guard {
  sigset_t mask;
  int slot;
} narra_proc_guard;

static void narra_proc_forward(int signal_number) {
  narra_signal_note(signal_number);
  for (int i = 0; i < narra_proc_SLOTS; i++) {
    if (atomic_load(&narra_proc_pids[i]) == 0) continue;
    atomic_store(&narra_proc_pending[i], signal_number);
    int child = atomic_load(&narra_proc_pids[i]);
    if (child > 0) kill((pid_t)child, signal_number);
  }
}

static void narra_proc_begin(narra_proc_guard* guard) {
  sigset_t set;
  sigemptyset(&set);
  for (int i = 0; i < 3; i++) sigaddset(&set, narra_proc_signals[i]);
  pthread_sigmask(SIG_BLOCK, &set, &guard->mask);
  // Claim a slot; wait for one when all are taken.
  guard->slot = -1;
  while (guard->slot < 0) {
    for (int i = 0; i < narra_proc_SLOTS; i++) {
      int expected = 0;
      if (atomic_compare_exchange_strong(&narra_proc_pids[i], &expected, -1)) {
        atomic_store(&narra_proc_pending[i], 0);
        guard->slot = i;
        break;
      }
    }
    if (guard->slot < 0) sched_yield();
  }
  pthread_mutex_lock(&narra_proc_lock);
  if (narra_proc_active++ == 0) {
    struct sigaction action;
    memset(&action, 0, sizeof action);
    action.sa_handler = narra_proc_forward;
    sigemptyset(&action.sa_mask);
    for (int i = 0; i < 3; i++) {
      // A signal that was ignored on entry (a background job, nohup) stays ignored.
      sigaction(narra_proc_signals[i], NULL, &narra_proc_saved[i]);
      if (narra_proc_saved[i].sa_handler != SIG_IGN) sigaction(narra_proc_signals[i], &action, NULL);
    }
  }
  pthread_mutex_unlock(&narra_proc_lock);
}

// The child takes no lock: the saved handlers do not change while a call is active.
static void narra_proc_restore_child(narra_proc_guard* guard) {
  for (int i = 0; i < 3; i++) sigaction(narra_proc_signals[i], &narra_proc_saved[i], NULL);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
}

static void narra_proc_attach(narra_proc_guard* guard, pid_t pid) {
  atomic_store(&narra_proc_pids[guard->slot], (int)pid);
  int missed = atomic_load(&narra_proc_pending[guard->slot]);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
  if (missed) kill(pid, missed);
}

// Frees the slot with the signals still blocked, restores the old handlers
// after the last call, then restores the old mask, so no signal reaches a
// reaped pid.
static void narra_proc_detach(narra_proc_guard* guard) {
  sigset_t set;
  sigemptyset(&set);
  for (int i = 0; i < 3; i++) sigaddset(&set, narra_proc_signals[i]);
  pthread_sigmask(SIG_BLOCK, &set, NULL);
  atomic_store(&narra_proc_pids[guard->slot], 0);
  pthread_mutex_lock(&narra_proc_lock);
  if (--narra_proc_active == 0) {
    narra_signal_function ours = atomic_load(&narra_signal_handler_in_use);
    for (int i = 0; i < 3; i++) {
      int number = narra_proc_signals[i];
      if (ours != NULL && (number == SIGINT || number == SIGTERM)) {
        // std/signal started during the call: its handler stays, whatever was saved.
        struct sigaction action;
        memset(&action, 0, sizeof action);
        action.sa_handler = ours;
        sigemptyset(&action.sa_mask);
        action.sa_flags = SA_RESTART;
        sigaction(number, &action, NULL);
      } else {
        sigaction(number, &narra_proc_saved[i], NULL);
      }
    }
  }
  pthread_mutex_unlock(&narra_proc_lock);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
}
extern char** environ;

// True when the variable entry (NAME=VALUE, or just NAME) names name.
static bool narra_proc_names(const char* entry, const char* name, size_t name_length) {
  return strncmp(entry, name, name_length) == 0 && (entry[name_length] == '=' || entry[name_length] == 0);
}

// The environment of the child: this process's variables without the removed
// names and without the names that are set again, then the added entries
// (NAME=VALUE). NULL when memory runs out.
static char** narra_proc_environment(lstr** added, int64_t added_count, lstr** removed, int64_t removed_count) {
  size_t existing = 0;
  while (environ && environ[existing]) existing++;
  char** vec = malloc(sizeof(char*) * (existing + (size_t)added_count + 1));
  if (!vec) return NULL;
  size_t used = 0;
  for (size_t i = 0; i < existing; i++) {
    bool dropped = false;
    for (int64_t j = 0; j < removed_count && !dropped; j++) {
      dropped = narra_proc_names(environ[i], removed[j]->data, (size_t)removed[j]->len);
    }
    for (int64_t j = 0; j < added_count && !dropped; j++) {
      const char* equals = strchr(added[j]->data, '=');
      dropped = narra_proc_names(environ[i], added[j]->data, (size_t)(equals - added[j]->data));
    }
    if (!dropped) vec[used++] = environ[i];
  }
  // A name set twice keeps its last value.
  for (int64_t i = 0; i < added_count; i++) {
    const char* equals = strchr(added[i]->data, '=');
    size_t name_length = (size_t)(equals - added[i]->data);
    bool replaced = false;
    for (int64_t j = i + 1; j < added_count && !replaced; j++) {
      const char* later = strchr(added[j]->data, '=');
      replaced = (size_t)(later - added[j]->data) == name_length && strncmp(added[j]->data, added[i]->data, name_length) == 0;
    }
    if (!replaced) vec[used++] = added[i]->data;
  }
  vec[used] = NULL;
  return vec;
}

// An added entry needs a name before its first '=', and a removed name needs
// text without '='; no entry may hold a NUL byte.
static bool narra_proc_check_environment(lstr** added, int64_t added_count, lstr** removed, int64_t removed_count, lstr** err) {
  for (int64_t i = 0; i < added_count; i++) {
    if (!narra_io_no_nul(added[i], err)) return false;
    const char* equals = strchr(added[i]->data, '=');
    if (!equals || equals == added[i]->data) {
      *err = lstr_from_cstr("environment variable needs a name");
      return false;
    }
  }
  for (int64_t i = 0; i < removed_count; i++) {
    if (!narra_io_no_nul(removed[i], err)) return false;
    if (removed[i]->len == 0 || strchr(removed[i]->data, '=')) {
      *err = lstr_from_cstr("environment variable name is empty or holds '='");
      return false;
    }
  }
  return true;
}

// Runs cmd (found through PATH) with argv[0..argc) as its arguments. Its
// standard input is the text input when has_input is set, and closed otherwise.
// The environment is this process's, with the added entries (NAME=VALUE) set and
// the removed names unset; a folder that is not empty is the working folder. When
// inherit is set, the child writes to the standard output and standard error of
// this process and the captured text is empty. Writing the input and reading both
// streams happen in one poll loop, so a full pipe never blocks the child. Returns
// false with *err set when the process cannot start. The exit code of a process
// killed by a signal is 128 plus the signal.
static bool narra_proc_run(lstr* cmd, lstr** argv, int64_t argc, lstr* input, bool has_input, lstr** added, int64_t added_count, lstr** removed, int64_t removed_count, lstr* folder, bool inherit, int64_t* code, lstr** out, lstr** errout, lstr** err) {
  if (!narra_io_no_nul(cmd, err)) return false;
  for (int64_t i = 0; i < argc; i++) {
    if (!narra_io_no_nul(argv[i], err)) return false;
  }
  if (!narra_io_no_nul(folder, err)) return false;
  if (!narra_proc_check_environment(added, added_count, removed, removed_count, err)) return false;
  char** vec = malloc(sizeof(char*) * (size_t)(argc + 2));
  if (!vec) { *err = narra_io_fail(cmd->data, cmd->len, "out of memory"); return false; }
  vec[0] = cmd->data;
  for (int64_t i = 0; i < argc; i++) vec[i + 1] = argv[i]->data;
  vec[argc + 1] = NULL;
  char** envp = NULL;
  if (added_count > 0 || removed_count > 0) {
    envp = narra_proc_environment(added, added_count, removed, removed_count);
    if (!envp) {
      free(vec);
      *err = narra_io_fail(cmd->data, cmd->len, "out of memory");
      return false;
    }
  }
  int po[2], pe[2], px[2], pi[2];
  po[0] = po[1] = pe[0] = pe[1] = px[0] = px[1] = pi[0] = pi[1] = -1;
  if (narra_proc_pipe(px) != 0 || (!inherit && (narra_proc_pipe(po) != 0 || narra_proc_pipe(pe) != 0)) || (has_input && narra_proc_pipe(pi) != 0)) {
    *err = narra_io_fail(cmd->data, cmd->len, strerror(errno));
    narra_proc_close(po); narra_proc_close(pe); narra_proc_close(px); narra_proc_close(pi);
    free(vec);
    free(envp);
    return false;
  }
  fflush(NULL);
  // Termination signals that arrive while the child runs are passed on to it,
  // so killing the parent does not orphan the child. The child stays in the
  // foreground process group, so Ctrl-C in a terminal still reaches it. The
  // signals are blocked from before the fork until the child pid is recorded.
  narra_proc_guard guard;
  narra_proc_begin(&guard);
  pid_t pid = fork();
  int fork_errno = errno;
  if (pid > 0) narra_proc_attach(&guard, pid);
  if (pid < 0) {
    narra_proc_detach(&guard);
    *err = narra_io_fail(cmd->data, cmd->len, strerror(fork_errno));
    narra_proc_close(po); narra_proc_close(pe); narra_proc_close(px); narra_proc_close(pi);
    free(vec);
    free(envp);
    return false;
  }
  if (pid == 0) {
    // The report holds the step that failed (1 for the folder) and its errno.
    int report[2] = {0, 0};
    narra_proc_restore_child(&guard);
    int in = has_input ? pi[0] : open("/dev/null", O_RDONLY);
    if (in < 0) goto child_fail;
    if (in != 0) {
      if (dup2(in, 0) < 0) goto child_fail;
      close(in);
    }
    if (!inherit && (dup2(po[1], 1) < 0 || dup2(pe[1], 2) < 0)) goto child_fail;
    if (folder->len > 0 && chdir(folder->data) != 0) {
      report[0] = 1;
      goto child_fail;
    }
    if (envp) environ = envp;
    execvp(vec[0], vec);
  child_fail:
    report[1] = errno;
    ssize_t w = write(px[1], report, sizeof report);
    (void)w;
    _exit(127);
  }
  free(vec);
  free(envp);
  close(po[1]); close(pe[1]); close(px[1]); close(pi[0]);
  po[1] = pe[1] = px[1] = pi[0] = -1;
  // The close-on-exec pipe is empty when the program started, and holds
  // the step and errno of the child when it did not.
  int report[2] = {0, 0};
  ssize_t got;
  do { got = read(px[0], report, sizeof report); } while (got < 0 && errno == EINTR);
  close(px[0]);
  int status = 0;
  if (got == (ssize_t)sizeof report) {
    close(po[0]); close(pe[0]); close(pi[1]);
    while (waitpid(pid, &status, 0) < 0 && errno == EINTR) {}
    narra_proc_detach(&guard);
    if (report[0] == 1) *err = narra_io_fail(folder->data, folder->len, strerror(report[1]));
    else *err = narra_io_fail(cmd->data, cmd->len, strerror(report[1]));
    return false;
  }
  // A child that stops reading its input must not kill this process: SIGPIPE
  // stays blocked while the input is written, and a pending one is dropped.
  sigset_t pipe_set, pipe_previous;
  sigemptyset(&pipe_set);
  sigaddset(&pipe_set, SIGPIPE);
  pthread_sigmask(SIG_BLOCK, &pipe_set, &pipe_previous);
  size_t written = 0;
  if (has_input) {
    if (input->len == 0) {
      close(pi[1]);
      pi[1] = -1;
    } else {
      fcntl(pi[1], F_SETFL, fcntl(pi[1], F_GETFL) | O_NONBLOCK);
#ifdef __APPLE__
      // macOS has no sigtimedwait; this descriptor never raises SIGPIPE. The
      // constant is hidden by strict C11 headers.
#ifndef F_SETNOSIGPIPE
#define F_SETNOSIGPIPE 73
#endif
      fcntl(pi[1], F_SETNOSIGPIPE, 1);
#endif
    }
  }
  narra_proc_buf bo = {0}, be = {0};
  struct pollfd fds[3] = {{po[0], POLLIN, 0}, {pe[0], POLLIN, 0}, {pi[1], POLLOUT, 0}};
  int open_fds = (po[0] >= 0) + (pe[0] >= 0) + (pi[1] >= 0);
  while (open_fds > 0) {
    if (poll(fds, 3, -1) < 0) {
      if (errno == EINTR) continue;
      break;
    }
    for (int i = 0; i < 2; i++) {
      if (fds[i].fd < 0 || !(fds[i].revents & (POLLIN | POLLHUP | POLLERR))) continue;
      if (!narra_proc_take(i == 0 ? &bo : &be, fds[i].fd)) {
        close(fds[i].fd);
        fds[i].fd = -1;
        open_fds--;
      }
    }
    if (fds[2].fd >= 0 && (fds[2].revents & (POLLOUT | POLLHUP | POLLERR))) {
      size_t chunk = (size_t)input->len - written;
      if (chunk > 65536) chunk = 65536;
      ssize_t n = write(fds[2].fd, input->data + written, chunk);
      bool finished = false;
      if (n >= 0) {
        written += (size_t)n;
        finished = written == (size_t)input->len;
      } else if (errno != EINTR && errno != EAGAIN) {
        // The child closed its input; the rest is dropped.
        finished = true;
      }
      if (finished) {
        close(fds[2].fd);
        fds[2].fd = -1;
        open_fds--;
      }
    }
  }
  for (int i = 0; i < 3; i++) {
    if (fds[i].fd >= 0) close(fds[i].fd);
  }
#ifndef __APPLE__
  if (!sigismember(&pipe_previous, SIGPIPE)) {
    // Take a pending SIGPIPE of this thread off the queue, without waiting.
    sigset_t pending;
    sigpending(&pending);
    if (sigismember(&pending, SIGPIPE)) {
      struct timespec none = {0, 0};
      sigtimedwait(&pipe_set, NULL, &none);
    }
  }
#endif
  pthread_sigmask(SIG_SETMASK, &pipe_previous, NULL);
  while (waitpid(pid, &status, 0) < 0 && errno == EINTR) {}
  narra_proc_detach(&guard);
  if (bo.failed || be.failed) {
    free(bo.p);
    free(be.p);
    *err = narra_io_fail(cmd->data, cmd->len, "out of memory");
    return false;
  }
  *code = WIFEXITED(status) ? WEXITSTATUS(status) : 128 + WTERMSIG(status);
  *out = narra_io_lossy(bo.p ? bo.p : "", bo.len);
  *errout = narra_io_lossy(be.p ? be.p : "", be.len);
  free(bo.p);
  free(be.p);
  return true;
}
#endif
