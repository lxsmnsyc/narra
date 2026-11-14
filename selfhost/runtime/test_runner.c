
// Test runner
#include <sys/types.h>
#include <sys/wait.h>
#include <time.h>
#include <errno.h>
#include <signal.h>
#include <pthread.h>
#include <sched.h>
#include <stdatomic.h>

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
  alarm(narra_test_limit());
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
// narra_test_run returns the exit code of the run.
static int narra_test_run(const narra_test_case* cases, int argc, char** argv) {
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
