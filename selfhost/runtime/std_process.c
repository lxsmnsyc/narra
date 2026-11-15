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

#ifdef __wasi__
// WASI preview 1 starts no processes. The compiler refuses Process.run on this
// target (E0991); the stub keeps a build that is not checked linking.
static bool narra_proc_run(lstr* cmd, lstr** argv, int64_t argc, int64_t* code, lstr** out, lstr** errout, lstr** err) {
  (void)cmd; (void)argv; (void)argc; (void)code; (void)out; (void)errout;
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

// Tries name as written when it has an extension, then name with .exe, then
// with .com, inside one directory. Only these two extensions run directly;
// .bat and .cmd need the shell and are not guessed.
static wchar_t* narra_proc_probe(const wchar_t* directory, size_t directory_length, const wchar_t* name, bool* memory) {
  const wchar_t* base = name;
  for (const wchar_t* c = name; *c; c++) {
    if (*c == L'\\' || *c == L'/' || *c == L':') base = c + 1;
  }
  bool has_extension = wcschr(base, L'.') != NULL;
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
// NULL when nothing is found; *memory is set when an allocation failed.
static wchar_t* narra_proc_find(const wchar_t* name, bool* memory) {
  bool bare = true;
  for (const wchar_t* c = name; *c; c++) {
    if (*c == L'\\' || *c == L'/' || *c == L':') bare = false;
  }
  if (!bare) return narra_proc_probe(L"", 0, name, memory);
  DWORD size = GetEnvironmentVariableW(L"PATH", NULL, 0);
  if (size == 0) return NULL;
  wchar_t* search = malloc(size * sizeof(wchar_t));
  wchar_t* directory = malloc(size * sizeof(wchar_t));
  if (!search || !directory) {
    free(search);
    free(directory);
    *memory = true;
    return NULL;
  }
  DWORD got = GetEnvironmentVariableW(L"PATH", search, size);
  wchar_t* found = NULL;
  if (got > 0 && got < size) {
    const wchar_t* entry = search;
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
  free(search);
  free(directory);
  return found;
}

// Runs cmd (found through PATH) with argv[0..argc) as its arguments and its
// standard input closed. Returns false with *err set when the process cannot
// start. The exit code is the one the process ended with.
static bool narra_proc_run(lstr* cmd, lstr** argv, int64_t argc, int64_t* code, lstr** out, lstr** errout, lstr** err) {
  if (!narra_io_no_nul(cmd, err)) return false;
  for (int64_t i = 0; i < argc; i++) {
    if (!narra_io_no_nul(argv[i], err)) return false;
  }
  narra_proc_buf line = {0};
  bool built = narra_proc_quote(&line, cmd->data, (size_t)cmd->len);
  for (int64_t i = 0; built && i < argc; i++) {
    built = narra_proc_append(&line, " ", 1) && narra_proc_quote(&line, argv[i]->data, (size_t)argv[i]->len);
  }
  if (!built) {
    free(line.p);
    *err = narra_io_fail(cmd->data, cmd->len, "out of memory");
    return false;
  }
  wchar_t* wide = line.p ? narra_proc_wide(line.p, line.len) : NULL;
  free(line.p);
  if (!wide) {
    *err = narra_io_fail(cmd->data, cmd->len, "arguments are not valid UTF-8");
    return false;
  }
  // The program is resolved here, so CreateProcess never searches the current
  // directory or the folder of this program.
  wchar_t* program_name = narra_proc_wide(cmd->data, (size_t)cmd->len);
  bool memory = program_name == NULL;
  wchar_t* program = program_name ? narra_proc_find(program_name, &memory) : NULL;
  free(program_name);
  if (!program) {
    *err = narra_io_fail(cmd->data, cmd->len, memory ? "out of memory" : "No such file or directory");
    free(wide);
    return false;
  }
  HANDLE out_read = NULL, out_write = NULL, err_read = NULL, err_write = NULL, input = NULL, job = NULL;
  SECURITY_ATTRIBUTES inherit;
  memset(&inherit, 0, sizeof inherit);
  inherit.nLength = sizeof inherit;
  inherit.bInheritHandle = TRUE;
  input = CreateFileA("NUL", GENERIC_READ, FILE_SHARE_READ | FILE_SHARE_WRITE, &inherit, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
  if (input == INVALID_HANDLE_VALUE) input = NULL;
  if (!input || !narra_proc_pipe(&out_read, &out_write) || !narra_proc_pipe(&err_read, &err_write)) {
    char text[256];
    *err = narra_io_fail(cmd->data, cmd->len, narra_proc_message(GetLastError(), text, sizeof text));
    narra_proc_close_handle(&input);
    narra_proc_close_handle(&out_read); narra_proc_close_handle(&out_write);
    narra_proc_close_handle(&err_read); narra_proc_close_handle(&err_write);
    free(wide);
    free(program);
    return false;
  }
  fflush(NULL);
  // Only the three standard handles pass to the child, so a process started
  // by another thread at the same time does not hold these pipes open.
  HANDLE inherited[3] = {input, out_write, err_write};
  STARTUPINFOEXW startup;
  memset(&startup, 0, sizeof startup);
  startup.StartupInfo.cb = sizeof startup;
  startup.StartupInfo.dwFlags = STARTF_USESTDHANDLES;
  startup.StartupInfo.hStdInput = input;
  startup.StartupInfo.hStdOutput = out_write;
  startup.StartupInfo.hStdError = err_write;
  SIZE_T list_size = 0;
  InitializeProcThreadAttributeList(NULL, 1, 0, &list_size);
  startup.lpAttributeList = (LPPROC_THREAD_ATTRIBUTE_LIST)malloc(list_size);
  bool listed = startup.lpAttributeList != NULL &&
    InitializeProcThreadAttributeList(startup.lpAttributeList, 1, 0, &list_size) &&
    UpdateProcThreadAttribute(startup.lpAttributeList, 0, PROC_THREAD_ATTRIBUTE_HANDLE_LIST, inherited, sizeof inherited, NULL, NULL);
  PROCESS_INFORMATION process;
  memset(&process, 0, sizeof process);
  BOOL started = FALSE;
  DWORD failure = ERROR_OUTOFMEMORY;
  if (listed) {
    started = CreateProcessW(program, wide, NULL, NULL, TRUE, EXTENDED_STARTUPINFO_PRESENT | CREATE_SUSPENDED | CREATE_NO_WINDOW, NULL, NULL, &startup.StartupInfo, &process);
    if (!started) failure = GetLastError();
  }
  if (startup.lpAttributeList) {
    if (listed) DeleteProcThreadAttributeList(startup.lpAttributeList);
    free(startup.lpAttributeList);
  }
  free(wide);
  free(program);
  narra_proc_close_handle(&input);
  narra_proc_close_handle(&out_write);
  narra_proc_close_handle(&err_write);
  if (!started) {
    char text[256];
    *err = narra_io_fail(cmd->data, cmd->len, narra_proc_message(failure, text, sizeof text));
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
  // Both streams need a thread: draining one pipe after the other can
  // deadlock when the child fills the second. The threads start before the
  // child runs, and the call fails, with the child ended, when one cannot.
  HANDLE threads[2] = {NULL, NULL};
  for (int i = 0; i < 2; i++) threads[i] = CreateThread(NULL, 0, narra_proc_drain, &readers[i], 0, NULL);
  if (!threads[0] || !threads[1]) {
    TerminateProcess(process.hProcess, 1);
    for (int i = 0; i < 2; i++) {
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
    free(readers[0].buffer.p);
    free(readers[1].buffer.p);
    return false;
  }
  ResumeThread(process.hThread);
  narra_proc_close_handle(&process.hThread);
  for (int i = 0; i < 2; i++) {
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
    for (int i = 0; i < 3; i++) sigaction(narra_proc_signals[i], &narra_proc_saved[i], NULL);
  }
  pthread_mutex_unlock(&narra_proc_lock);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
}
// Runs cmd (found through PATH) with argv[0..argc) as its arguments and its
// standard input closed. Returns false with *err set when the process cannot
// start. The exit code of a process killed by a signal is 128 plus the signal.
static bool narra_proc_run(lstr* cmd, lstr** argv, int64_t argc, int64_t* code, lstr** out, lstr** errout, lstr** err) {
  if (!narra_io_no_nul(cmd, err)) return false;
  for (int64_t i = 0; i < argc; i++) {
    if (!narra_io_no_nul(argv[i], err)) return false;
  }
  char** vec = malloc(sizeof(char*) * (size_t)(argc + 2));
  if (!vec) { *err = narra_io_fail(cmd->data, cmd->len, "out of memory"); return false; }
  vec[0] = cmd->data;
  for (int64_t i = 0; i < argc; i++) vec[i + 1] = argv[i]->data;
  vec[argc + 1] = NULL;
  int po[2], pe[2], px[2];
  po[0] = po[1] = pe[0] = pe[1] = px[0] = px[1] = -1;
  if (narra_proc_pipe(po) != 0 || narra_proc_pipe(pe) != 0 || narra_proc_pipe(px) != 0) {
    *err = narra_io_fail(cmd->data, cmd->len, strerror(errno));
    narra_proc_close(po); narra_proc_close(pe); narra_proc_close(px);
    free(vec);
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
    narra_proc_close(po); narra_proc_close(pe); narra_proc_close(px);
    free(vec);
    return false;
  }
  if (pid == 0) {
    int e;
    narra_proc_restore_child(&guard);
    int in = open("/dev/null", O_RDONLY);
    if (in < 0) goto child_fail;
    if (in != 0) {
      if (dup2(in, 0) < 0) goto child_fail;
      close(in);
    }
    if (dup2(po[1], 1) < 0 || dup2(pe[1], 2) < 0) goto child_fail;
    execvp(vec[0], vec);
  child_fail:
    e = errno;
    ssize_t w = write(px[1], &e, sizeof e);
    (void)w;
    _exit(127);
  }
  free(vec);
  close(po[1]); close(pe[1]); close(px[1]);
  // The close-on-exec pipe is empty when the program started, and holds
  // the child's errno when it did not.
  int child_errno = 0;
  ssize_t got;
  do { got = read(px[0], &child_errno, sizeof child_errno); } while (got < 0 && errno == EINTR);
  close(px[0]);
  int status = 0;
  if (got == (ssize_t)sizeof child_errno) {
    close(po[0]); close(pe[0]);
    while (waitpid(pid, &status, 0) < 0 && errno == EINTR) {}
    narra_proc_detach(&guard);
    *err = narra_io_fail(cmd->data, cmd->len, strerror(child_errno));
    return false;
  }
  narra_proc_buf bo = {0}, be = {0};
  struct pollfd fds[2] = {{po[0], POLLIN, 0}, {pe[0], POLLIN, 0}};
  int open_fds = 2;
  while (open_fds > 0) {
    if (poll(fds, 2, -1) < 0) {
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
  }
  if (fds[0].fd >= 0) close(fds[0].fd);
  if (fds[1].fd >= 0) close(fds[1].fd);
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
