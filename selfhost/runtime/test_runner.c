
// Test runner
#include <sys/types.h>
#include <time.h>
#include <errno.h>
#ifdef _WIN32
#include <winsock2.h>
#include <windows.h>
#else
#include <sys/wait.h>
#include <signal.h>
#include <pthread.h>
#include <sched.h>
#include <stdatomic.h>
#endif

typedef struct narra_test_case {
  const char* name;
  void (*fn)(void);
} narra_test_case;

static int64_t narra_test_now_ms(void) {
  struct timespec ts;
  clock_gettime(CLOCK_MONOTONIC, &ts);
  return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

// The time limit of one test: NARRA_TEST_TIMEOUT seconds, default 10.
static unsigned narra_test_limit(void) {
  const char* s = getenv("NARRA_TEST_TIMEOUT");
  if (s && *s) {
    char* end;
    long v = strtol(s, &end, 10);
    if (*end == 0 && v > 0 && v < 100000) return (unsigned)v;
  }
  return 10;
}

// narra_test_child runs one test in the child process.
static void narra_test_child(const narra_test_case* c, int wfd) {
  narra_panic_fd = wfd;
  narra_panic_pos = 1;
  // Line buffering keeps the output of a test that crashes.
  setvbuf(stdout, NULL, _IOLBF, 0);
#ifndef _WIN32
  alarm(narra_test_limit());
#endif
  NARRA_COPY_START();
  c->fn();
  narra_consts_release();
  // A test-only hook, so the leak report can be tested.
  if (getenv("NARRA_TEST_FORCE_LEAK")) narra_live++;
  if (getenv("NARRA_CHECK_LEAKS") && narra_live != 0) {
    char buf[64];
    snprintf(buf, sizeof buf, "leaked %lld allocations", (long long)narra_live);
    narra_panic("leak check", buf);
  }
  NARRA_COPY_REPORT();
  fflush(stdout);
  _exit(0);
}

// narra_test_arg returns the option name of an argument ("--time" and
// "-time" both give "time"), or NULL for an argument that is not an option.
static const char* narra_test_opt(const char* a) {
  if (a[0] != '-') return NULL;
  return a[1] == '-' ? a + 2 : a + 1;
}

#ifndef _WIN32
// While a call runs, SIGTERM, SIGINT and SIGHUP are passed on to its child, so
// killing this process does not orphan the child. Several threads may call at
// once: a count of active calls (under a mutex) installs the handlers for the
// first call and restores them after the last, and each call owns a slot in a
// fixed array of child pids that the handler walks with plain atomic loads.
// A slot holds -1 from the claim until the pid is stored; a signal that
// arrives in between is remembered in the pending array and sent on attach.
#define narra_runner_SLOTS 64
static pthread_mutex_t narra_runner_lock = PTHREAD_MUTEX_INITIALIZER;
static int narra_runner_active = 0;
static struct sigaction narra_runner_saved[3];
static _Atomic int narra_runner_pids[narra_runner_SLOTS];
static _Atomic int narra_runner_pending[narra_runner_SLOTS];
static volatile sig_atomic_t narra_runner_caught = 0;
static const int narra_runner_signals[3] = {SIGTERM, SIGINT, SIGHUP};

typedef struct narra_runner_guard {
  sigset_t mask;
  int slot;
} narra_runner_guard;

static void narra_runner_forward(int signal_number) {
  narra_runner_caught = signal_number;
  for (int i = 0; i < narra_runner_SLOTS; i++) {
    if (atomic_load(&narra_runner_pids[i]) == 0) continue;
    atomic_store(&narra_runner_pending[i], signal_number);
    int child = atomic_load(&narra_runner_pids[i]);
    if (child > 0) kill((pid_t)child, signal_number);
  }
}

static void narra_runner_begin(narra_runner_guard* guard) {
  sigset_t set;
  sigemptyset(&set);
  for (int i = 0; i < 3; i++) sigaddset(&set, narra_runner_signals[i]);
  pthread_sigmask(SIG_BLOCK, &set, &guard->mask);
  // Claim a slot; wait for one when all are taken.
  guard->slot = -1;
  while (guard->slot < 0) {
    for (int i = 0; i < narra_runner_SLOTS; i++) {
      int expected = 0;
      if (atomic_compare_exchange_strong(&narra_runner_pids[i], &expected, -1)) {
        atomic_store(&narra_runner_pending[i], 0);
        guard->slot = i;
        break;
      }
    }
    if (guard->slot < 0) sched_yield();
  }
  pthread_mutex_lock(&narra_runner_lock);
  if (narra_runner_active++ == 0) {
    struct sigaction action;
    memset(&action, 0, sizeof action);
    action.sa_handler = narra_runner_forward;
    sigemptyset(&action.sa_mask);
    for (int i = 0; i < 3; i++) {
      // A signal that was ignored on entry (a background job, nohup) stays ignored.
      sigaction(narra_runner_signals[i], NULL, &narra_runner_saved[i]);
      if (narra_runner_saved[i].sa_handler != SIG_IGN) sigaction(narra_runner_signals[i], &action, NULL);
    }
  }
  pthread_mutex_unlock(&narra_runner_lock);
}

// The child takes no lock: the saved handlers do not change while a call is active.
static void narra_runner_restore_child(narra_runner_guard* guard) {
  for (int i = 0; i < 3; i++) sigaction(narra_runner_signals[i], &narra_runner_saved[i], NULL);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
}

static void narra_runner_attach(narra_runner_guard* guard, pid_t pid) {
  atomic_store(&narra_runner_pids[guard->slot], (int)pid);
  int missed = atomic_load(&narra_runner_pending[guard->slot]);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
  if (missed) kill(pid, missed);
}

// Frees the slot with the signals still blocked, restores the old handlers
// after the last call, then restores the old mask, so no signal reaches a
// reaped pid.
static void narra_runner_detach(narra_runner_guard* guard) {
  sigset_t set;
  sigemptyset(&set);
  for (int i = 0; i < 3; i++) sigaddset(&set, narra_runner_signals[i]);
  pthread_sigmask(SIG_BLOCK, &set, NULL);
  atomic_store(&narra_runner_pids[guard->slot], 0);
  pthread_mutex_lock(&narra_runner_lock);
  if (--narra_runner_active == 0) {
    for (int i = 0; i < 3; i++) sigaction(narra_runner_signals[i], &narra_runner_saved[i], NULL);
  }
  pthread_mutex_unlock(&narra_runner_lock);
  pthread_sigmask(SIG_SETMASK, &guard->mask, NULL);
}
#endif
// narra_test_parse reads the options; it returns -1 when the run goes on and
// the exit code otherwise.
static int narra_test_parse(int argc, char** argv, const char** filter_out, int* timing_out) {
  const char* filter = NULL;
  int timing = 0;
  for (int i = 1; i < argc; i++) {
    const char* o = narra_test_opt(argv[i]);
    const char* eq = o ? strchr(o, '=') : NULL;
    size_t n = o ? (eq ? (size_t)(eq - o) : strlen(o)) : 0;
    if (o && n == 4 && strncmp(o, "time", 4) == 0 && !eq) {
      timing = 1;
    } else if (o && ((n == 6 && strncmp(o, "filter", 6) == 0) || (n == 10 && strncmp(o, "golden-dir", 10) == 0))) {
      const char* v = eq ? eq + 1 : (i + 1 < argc ? argv[++i] : NULL);
      if (!v) {
        fprintf(stderr, "%s needs a value\n", argv[i]);
        return 2;
      }
      if (n == 6) filter = v; else narra_golden_dir = v;
    } else {
      fprintf(stderr, "unknown argument %s; options: --filter <text>, --time, --golden-dir <dir>\n", argv[i]);
      return 2;
    }
  }
  *filter_out = filter;
  *timing_out = timing;
  return -1;
}

#ifdef _WIN32
// There is no fork on Windows. The runner starts the binary again for each
// test, with NARRA_TEST_CHILD set to the number of the case and
// NARRA_TEST_PIPE to an inheritable handle that takes the panic message. A
// job object ends the child, and what it started, on a timeout or when this
// process ends.
typedef struct narra_test_reader {
  HANDLE pipe;
  char* text;
  size_t length, capacity;
} narra_test_reader;

static DWORD WINAPI narra_test_drain(LPVOID argument) {
  narra_test_reader* reader = (narra_test_reader*)argument;
  char chunk[1024];
  DWORD count;
  while (ReadFile(reader->pipe, chunk, sizeof chunk, &count, NULL) && count > 0) {
    if (reader->length + count + 1 > reader->capacity) {
      size_t capacity = reader->capacity ? reader->capacity * 2 : 1024;
      while (capacity < reader->length + count + 1) capacity *= 2;
      char* bigger = realloc(reader->text, capacity);
      if (!bigger) break;
      reader->text = bigger;
      reader->capacity = capacity;
    }
    memcpy(reader->text + reader->length, chunk, count);
    reader->length += count;
  }
  return 0;
}

// Runs the case named by NARRA_TEST_CHILD, or returns -1 in the parent.
static int narra_test_child_main(const narra_test_case* cases) {
  const char* number = getenv("NARRA_TEST_CHILD");
  if (!number || !*number) return -1;
  const char* handle_text = getenv("NARRA_TEST_PIPE");
  if (!handle_text) return 2;
  long index = strtol(number, NULL, 10);
  HANDLE handle = (HANDLE)(uintptr_t)_strtoui64(handle_text, NULL, 10);
  // A test that starts this binary again must not run as a child itself.
  _putenv("NARRA_TEST_CHILD=");
  _putenv("NARRA_TEST_PIPE=");
  int descriptor = _open_osfhandle((intptr_t)handle, _O_BINARY);
  if (descriptor < 0) return 2;
  narra_test_child(&cases[index], descriptor);
  return 0;
}

// Starts the child for case index. Returns false when it cannot start.
static bool narra_test_start(int index, HANDLE* process, HANDLE* job, HANDLE* read_end) {
  SECURITY_ATTRIBUTES attributes;
  memset(&attributes, 0, sizeof attributes);
  attributes.nLength = sizeof attributes;
  attributes.bInheritHandle = TRUE;
  HANDLE write_end = NULL;
  if (!CreatePipe(read_end, &write_end, &attributes, 0)) return false;
  SetHandleInformation(*read_end, HANDLE_FLAG_INHERIT, 0);
  char number[32], handle_text[32];
  snprintf(number, sizeof number, "%d", index);
  snprintf(handle_text, sizeof handle_text, "%llu", (unsigned long long)(uintptr_t)write_end);
  wchar_t path[MAX_PATH * 4];
  DWORD length = GetModuleFileNameW(NULL, path, (DWORD)(sizeof path / sizeof path[0]));
  const wchar_t* original = GetCommandLineW();
  size_t line_length = wcslen(original);
  wchar_t* line = malloc((line_length + 1) * sizeof(wchar_t));
  if (!line || length == 0 || length >= sizeof path / sizeof path[0]) {
    CloseHandle(*read_end);
    CloseHandle(write_end);
    free(line);
    return false;
  }
  memcpy(line, original, (line_length + 1) * sizeof(wchar_t));
  SetEnvironmentVariableA("NARRA_TEST_CHILD", number);
  SetEnvironmentVariableA("NARRA_TEST_PIPE", handle_text);
  STARTUPINFOW startup;
  memset(&startup, 0, sizeof startup);
  startup.cb = sizeof startup;
  PROCESS_INFORMATION information;
  memset(&information, 0, sizeof information);
  BOOL started = CreateProcessW(path, line, NULL, NULL, TRUE, CREATE_SUSPENDED, NULL, NULL, &startup, &information);
  free(line);
  SetEnvironmentVariableA("NARRA_TEST_CHILD", NULL);
  SetEnvironmentVariableA("NARRA_TEST_PIPE", NULL);
  CloseHandle(write_end);
  if (!started) {
    CloseHandle(*read_end);
    return false;
  }
  *job = CreateJobObjectW(NULL, NULL);
  if (*job) {
    JOBOBJECT_EXTENDED_LIMIT_INFORMATION limits;
    memset(&limits, 0, sizeof limits);
    limits.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
    SetInformationJobObject(*job, JobObjectExtendedLimitInformation, &limits, sizeof limits);
    AssignProcessToJobObject(*job, information.hProcess);
  }
  ResumeThread(information.hThread);
  CloseHandle(information.hThread);
  *process = information.hProcess;
  return true;
}

// narra_test_run returns the exit code of the run.
static int narra_test_run(const narra_test_case* cases, int argc, char** argv) {
  const char* filter = NULL;
  int timing = 0;
  int parsed = narra_test_parse(argc, argv, &filter, &timing);
  if (parsed >= 0) return parsed;
  int child = narra_test_child_main(cases);
  if (child >= 0) return child;
  int passed = 0, failed = 0;
  unsigned limit = narra_test_limit();
  int index = 0;
  for (const narra_test_case* c = cases; c->name; c++, index++) {
    if (filter && !strstr(c->name, filter)) continue;
    fflush(stdout);
    fflush(stderr);
    int64_t start = narra_test_now_ms();
    HANDLE process = NULL, job = NULL, read_end = NULL;
    if (!narra_test_start(index, &process, &job, &read_end)) {
      fprintf(stderr, "cannot start a child process\n");
      return 1;
    }
    narra_test_reader reader;
    memset(&reader, 0, sizeof reader);
    reader.pipe = read_end;
    HANDLE thread = CreateThread(NULL, 0, narra_test_drain, &reader, 0, NULL);
    bool timed_out = WaitForSingleObject(process, (DWORD)limit * 1000) == WAIT_TIMEOUT;
    if (timed_out) {
      if (job) TerminateJobObject(job, 3);
      else TerminateProcess(process, 3);
      WaitForSingleObject(process, INFINITE);
    }
    if (thread) {
      WaitForSingleObject(thread, INFINITE);
      CloseHandle(thread);
    }
    DWORD status = 0;
    GetExitCodeProcess(process, &status);
    CloseHandle(process);
    if (job) CloseHandle(job);
    CloseHandle(read_end);
    char* msg = reader.text;
    if (!msg) {
      msg = malloc(1);
      reader.length = 0;
    }
    msg[reader.length] = 0;
    int64_t ms = narra_test_now_ms() - start;
    char tbuf[48] = "";
    if (timing) snprintf(tbuf, sizeof tbuf, " (%lld ms)", (long long)ms);
    if (timed_out) {
      printf("FAIL - %s: timed out after %u s%s\n", c->name, limit, tbuf);
      failed++;
    } else if (status == 0) {
      printf("ok - %s%s\n", c->name, tbuf);
      passed++;
    } else if (status == 101) {
      char* nl = strchr(msg, '\n');
      const char* where = "";
      const char* text = msg;
      if (nl) { *nl = 0; where = msg; text = nl + 1; }
      printf("FAIL - %s: %s (%s)%s\n", c->name, text, where, tbuf);
      failed++;
    } else if (status >= 0xC0000000u) {
      printf("FAIL - %s: crashed with status 0x%08lX%s\n", c->name, (unsigned long)status, tbuf);
      failed++;
    } else {
      printf("FAIL - %s: exited with code %lu%s\n", c->name, (unsigned long)status, tbuf);
      failed++;
    }
    free(msg);
  }
  if (filter && passed + failed == 0) {
    printf("no tests match \"%s\"\n", filter);
    fflush(stdout);
    return 1;
  }
  printf("%d passed, %d failed\n", passed, failed);
  fflush(stdout);
  return failed ? 1 : 0;
}
#else
// narra_test_run returns the exit code of the run.
static int narra_test_run(const narra_test_case* cases, int argc, char** argv) {
  const char* filter = NULL;
  int timing = 0;
  int parsed = narra_test_parse(argc, argv, &filter, &timing);
  if (parsed >= 0) return parsed;
  int passed = 0, failed = 0;
  unsigned limit = narra_test_limit();
  for (const narra_test_case* c = cases; c->name; c++) {
    if (filter && !strstr(c->name, filter)) continue;
    int fds[2];
    fflush(stdout);
    fflush(stderr);
    if (pipe(fds) != 0) { fprintf(stderr, "cannot create a pipe\n"); return 1; }
    int64_t start = narra_test_now_ms();
    narra_runner_guard guard;
    narra_runner_begin(&guard);
    pid_t pid = fork();
    if (pid < 0) { narra_runner_detach(&guard); fprintf(stderr, "cannot fork\n"); return 1; }
    if (pid == 0) {
      narra_runner_restore_child(&guard);
      close(fds[0]);
      narra_test_child(c, fds[1]);
    }
    narra_runner_attach(&guard, pid);
    close(fds[1]);
    size_t cap = 256, len = 0;
    char* msg = malloc(cap);
    for (;;) {
      if (len + 1 == cap) { cap *= 2; msg = realloc(msg, cap); }
      ssize_t n = read(fds[0], msg + len, cap - 1 - len);
      if (n < 0 && errno == EINTR) continue;
      if (n <= 0) break;
      len += (size_t)n;
    }
    msg[len] = 0;
    close(fds[0]);
    int status = 0;
    while (waitpid(pid, &status, 0) < 0 && errno == EINTR) {}
    narra_runner_detach(&guard);
    if (narra_runner_caught) {
      free(msg);
      fflush(stdout);
      return 128 + narra_runner_caught;
    }
    int64_t ms = narra_test_now_ms() - start;
    char tbuf[48] = "";
    if (timing) snprintf(tbuf, sizeof tbuf, " (%lld ms)", (long long)ms);
    if (WIFEXITED(status) && WEXITSTATUS(status) == 0) {
      printf("ok - %s%s\n", c->name, tbuf);
      passed++;
    } else if (WIFEXITED(status) && WEXITSTATUS(status) == 101) {
      char* nl = strchr(msg, '\n');
      const char* where = "";
      const char* text = msg;
      if (nl) { *nl = 0; where = msg; text = nl + 1; }
      printf("FAIL - %s: %s (%s)%s\n", c->name, text, where, tbuf);
      failed++;
    } else if (WIFSIGNALED(status) && WTERMSIG(status) == SIGALRM) {
      printf("FAIL - %s: timed out after %u s%s\n", c->name, limit, tbuf);
      failed++;
    } else if (WIFSIGNALED(status)) {
      printf("FAIL - %s: killed by signal %d (%s)%s\n", c->name, WTERMSIG(status), strsignal(WTERMSIG(status)), tbuf);
      failed++;
    } else {
      printf("FAIL - %s: exited with code %d%s\n", c->name, WEXITSTATUS(status), tbuf);
      failed++;
    }
    free(msg);
  }
  if (filter && passed + failed == 0) {
    printf("no tests match \"%s\"\n", filter);
    fflush(stdout);
    return 1;
  }
  printf("%d passed, %d failed\n", passed, failed);
  fflush(stdout);
  return failed ? 1 : 0;
}
#endif
