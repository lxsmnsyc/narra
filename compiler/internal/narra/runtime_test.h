// std/test runtime. Emitted only when a module imports std/test.
//
// Test.panics runs the closure in a child process (POSIX fork). The child
// sets narra_panic_fd, so a panic writes its message to a pipe and exits
// 101. The parent reads the pipe until it closes, waits for the child and
// checks how it ended. The child's heap is a copy, so nothing it allocates
// counts in the parent's leak check.
//
// The child gets a timeout (alarm), 10 seconds or NARRA_TEST_TIMEOUT, so a
// closure that never returns fails instead of hanging the run. This also
// bounds a nested child that would keep the pipe open.
//
// The failure paths free what they built, except the final message, which
// narra_panic prints before the process ends.
#include <errno.h>
#include <signal.h>
#include <sys/types.h>
#include <sys/wait.h>

typedef struct narra_test_proc {
  pid_t pid;
  int fd;
  unsigned timeout;
} narra_test_proc;

static unsigned narra_test_timeout(void) {
  const char* s = getenv("NARRA_TEST_TIMEOUT");
  if (s && *s) {
    char* end;
    long v = strtol(s, &end, 10);
    if (*end == 0 && v > 0 && v < 100000) return (unsigned)v;
  }
  return 10;
}

// narra_test_fork returns true in the child, which must call the closure
// and then narra_test_returned. It returns false in the parent.
static bool narra_test_fork(narra_test_proc* t, const char* where) {
  int fds[2];
  fflush(stdout);
  fflush(stderr);
  t->timeout = narra_test_timeout();
  if (pipe(fds) != 0) narra_panic(where, "Test.panics: cannot create a pipe");
  pid_t pid = fork();
  if (pid < 0) narra_panic(where, "Test.panics: cannot fork");
  if (pid == 0) {
    close(fds[0]);
    narra_panic_fd = fds[1];
    narra_panic_pos = 0;
    alarm(t->timeout);
    return true;
  }
  close(fds[1]);
  t->pid = pid;
  t->fd = fds[0];
  return false;
}

// narra_test_returned ends the child when the closure returned.
static void narra_test_returned(void) {
  fflush(stdout);
  _exit(0);
}

// narra_test_expect waits for the child and panics unless it panicked with
// a message that contains want (any message when want is NULL). An empty
// want matches any message, and the match stops at a NUL byte in either.
static void narra_test_expect(narra_test_proc* t, lstr* want, const char* where) {
  size_t cap = 256, len = 0;
  char* msg = malloc(cap);
  if (!msg) narra_panic("runtime", "out of memory");
  for (;;) {
    if (len + 1 == cap) {
      cap *= 2;
      char* grown = realloc(msg, cap);
      if (!grown) narra_panic("runtime", "out of memory");
      msg = grown;
    }
    ssize_t n = read(t->fd, msg + len, cap - 1 - len);
    if (n < 0 && errno == EINTR) continue;
    if (n <= 0) break;
    len += (size_t)n;
  }
  msg[len] = 0;
  close(t->fd);
  int status = 0;
  while (waitpid(t->pid, &status, 0) < 0) {
    if (errno != EINTR) narra_panic(where, "Test.panics: cannot wait for the child process");
  }
  char buf[160];
  const char* fail = NULL;
  if (WIFEXITED(status) && WEXITSTATUS(status) == 0) {
    fail = "expected a panic, but the function returned";
  } else if (!WIFEXITED(status) || WEXITSTATUS(status) != 101) {
    if (WIFSIGNALED(status) && WTERMSIG(status) == SIGALRM) {
      snprintf(buf, sizeof buf, "expected a panic, but the function ran past the %u s timeout", t->timeout);
    } else if (WIFSIGNALED(status)) {
      snprintf(buf, sizeof buf, "expected a panic, but the function was killed by signal %d", WTERMSIG(status));
    } else {
      snprintf(buf, sizeof buf, "expected a panic, but the function exited with code %d", WEXITSTATUS(status));
    }
    fail = buf;
  }
  if (fail) {
    free(msg);
    narra_panic(where, fail);
  }
  if (want && !strstr(msg, want->data)) {
    lstr* w = narra_show_string(want);
    lstr* g = narra_quote(msg, (int64_t)len, '"');
    size_t n = (size_t)w->len + (size_t)g->len + 64;
    char* out = malloc(n);
    if (!out) narra_panic("runtime", "out of memory");
    snprintf(out, n, "expected a panic containing %s, got %s", w->data, g->data);
    lstr_release(w);
    lstr_release(g);
    free(msg);
    narra_panic(where, out);
  }
  free(msg);
}

// The failure of Test.eq. The shown values are owned and released here.
static void narra_test_fail_eq(lstr* expected, lstr* actual, const char* where) {
  size_t n = (size_t)expected->len + (size_t)actual->len + 32;
  char* out = malloc(n);
  if (!out) narra_panic("runtime", "out of memory");
  snprintf(out, n, "expected %s, got %s", expected->data, actual->data);
  lstr_release(expected);
  lstr_release(actual);
  narra_panic(where, out);
}

static void narra_test_fail_ne(lstr* value, const char* where) {
  size_t n = (size_t)value->len + 48;
  char* out = malloc(n);
  if (!out) narra_panic("runtime", "out of memory");
  snprintf(out, n, "expected a value other than %s", value->data);
  lstr_release(value);
  narra_panic(where, out);
}

// Test.golden: compare actual with <dir>/<name>.golden. With
// NARRA_UPDATE_GOLDEN=1 the file is written instead.
#include <sys/stat.h>

static void narra_test_golden(lstr* name, lstr* actual, const char* where) {
  const char* dir = narra_golden_dir;
  const char* up = getenv("NARRA_UPDATE_GOLDEN");
  int update = up && strcmp(up, "1") == 0;
  if (name->len == 0) narra_panic(where, "Test.golden: the name is empty");
  int bad = memchr(name->data, 0, (size_t)name->len) != NULL || name->data[0] == '/';
  // No path segment may be "..".
  for (const char* s = name->data; !bad && *s;) {
    const char* e = strchr(s, '/');
    size_t k = e ? (size_t)(e - s) : strlen(s);
    if (k == 2 && s[0] == '.' && s[1] == '.') bad = 1;
    s += k + (e ? 1 : 0);
    if (!e) break;
  }
  if (bad) narra_panic(where, "Test.golden: the name must be a relative path without a `..` segment");
  size_t cap = strlen(dir) + (size_t)name->len + 160;
  char* path = malloc(cap);
  if (!path) narra_panic("runtime", "out of memory");
  snprintf(path, cap, "%s/%s.golden", dir, name->data);
  if (update) {
    // Create the directories on the way.
    mkdir(dir, 0777);
    for (char* p = path + strlen(dir) + 1; *p; p++) {
      if (*p == '/') {
        *p = 0;
        mkdir(path, 0777);
        *p = '/';
      }
    }
    FILE* f = fopen(path, "wb");
    if (!f || fwrite(actual->data, 1, (size_t)actual->len, f) != (size_t)actual->len) {
      char* out = malloc(cap + 48);
      if (!out) narra_panic("runtime", "out of memory");
      snprintf(out, cap + 48, "Test.golden: cannot write %s.golden", name->data);
      if (f) fclose(f);
      free(path);
      narra_panic(where, out);
    }
    fclose(f);
    free(path);
    return;
  }
  FILE* f = fopen(path, "rb");
  if (!f) {
    // The message names the golden file, not the directory, so test output
    // does not depend on where the package is.
    char* out = malloc(cap + 96);
    if (!out) narra_panic("runtime", "out of memory");
    snprintf(out, cap + 96, "golden file %s.golden does not exist; run with NARRA_UPDATE_GOLDEN=1 to create it", name->data);
    free(path);
    narra_panic(where, out);
  }
  size_t wcap = 4096, wlen = 0;
  char* want = malloc(wcap);
  if (!want) narra_panic("runtime", "out of memory");
  for (;;) {
    if (wlen == wcap) {
      wcap *= 2;
      want = realloc(want, wcap);
      if (!want) narra_panic("runtime", "out of memory");
    }
    size_t n = fread(want + wlen, 1, wcap - wlen, f);
    if (n == 0) break;
    wlen += n;
  }
  fclose(f);
  const char* got = actual->data;
  size_t glen = (size_t)actual->len;
  size_t i = 0;
  while (i < wlen && i < glen && want[i] == got[i]) i++;
  if (i == wlen && i == glen) {
    free(want);
    free(path);
    return;
  }
  // The first differing line: both lines start after the last newline before i.
  size_t start = i;
  int64_t line = 1;
  for (size_t k = 0; k < i; k++) if (want[k] == '\n') line++;
  while (start > 0 && want[start - 1] != '\n') start--;
  size_t we = start, ge = start;
  while (we < wlen && want[we] != '\n') we++;
  while (ge < glen && got[ge] != '\n') ge++;
  lstr* w = narra_quote(want + start, (int64_t)(we - start), '"');
  lstr* g = narra_quote(got + start, (int64_t)(ge - start), '"');
  size_t n = (size_t)w->len + (size_t)g->len + (size_t)name->len + 96;
  char* out = malloc(n);
  if (!out) narra_panic("runtime", "out of memory");
  snprintf(out, n, "golden %s differs at line %lld: expected %s, got %s", name->data, (long long)line, w->data, g->data);
  lstr_release(w);
  lstr_release(g);
  free(want);
  free(path);
  narra_panic(where, out);
}
