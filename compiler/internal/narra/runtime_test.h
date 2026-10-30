// std/test runtime. Emitted only when a module imports std/test.
//
// Test.panics runs the closure in a child process (POSIX fork). The child
// sets narra_panic_fd, so a panic writes its message to a pipe and exits
// 101. The parent reads the pipe until it closes, waits for the child and
// checks how it ended. The child's heap is a copy, so nothing it allocates
// counts in the parent's leak check.
#include <errno.h>
#include <sys/types.h>
#include <sys/wait.h>

typedef struct narra_test_proc {
  pid_t pid;
  int fd;
} narra_test_proc;

// narra_test_fork returns true in the child, which must call the closure
// and then narra_test_returned. It returns false in the parent.
static bool narra_test_fork(narra_test_proc* t, const char* where) {
  int fds[2];
  fflush(stdout);
  fflush(stderr);
  if (pipe(fds) != 0) narra_panic(where, "Test.panics: cannot create a pipe");
  pid_t pid = fork();
  if (pid < 0) narra_panic(where, "Test.panics: cannot fork");
  if (pid == 0) {
    close(fds[0]);
    narra_panic_fd = fds[1];
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
// a message that contains want (any message when want is NULL).
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
    if (WIFSIGNALED(status)) {
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
    narra_panic(where, out);
  }
  free(msg);
}

// The failure of Test.eq: shown values are owned and the program ends.
static void narra_test_fail_eq(lstr* expected, lstr* actual, const char* where) {
  size_t n = (size_t)expected->len + (size_t)actual->len + 32;
  char* out = malloc(n);
  if (!out) narra_panic("runtime", "out of memory");
  snprintf(out, n, "expected %s, got %s", expected->data, actual->data);
  narra_panic(where, out);
}

static void narra_test_fail_ne(lstr* value, const char* where) {
  size_t n = (size_t)value->len + 48;
  char* out = malloc(n);
  if (!out) narra_panic("runtime", "out of memory");
  snprintf(out, n, "expected a value other than %s", value->data);
  narra_panic(where, out);
}
