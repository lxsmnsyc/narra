package narra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The test runner (SPEC.md section 13, D66). A test build's main runs every
// test block in its own child process, so a failed test does not stop the
// others and the leak check covers each test alone.
//
// Options come from argv of the test binary: `--filter <text>`, `--time` and
// `--golden-dir <dir>` (one or two dashes). `narra test` turns its own flags
// into these arguments (TestArgs). Nothing about the machine is in the C.

// TestArgs returns the test binary arguments for the `narra test` flags. src
// is the file or package directory under test: golden files are read from
// its `testdata/` directory, or from the file's directory for a single file.
func TestArgs(filter string, timing bool, src string) []string {
	var args []string
	if filter != "" {
		args = append(args, "--filter", filter)
	}
	if timing {
		args = append(args, "--time")
	}
	return append(args, "--golden-dir", goldenDir(src))
}

// goldenDir is where Test.golden looks for `<name>.golden` under narra test.
func goldenDir(src string) string {
	dir := filepath.Dir(src)
	if info, err := os.Stat(src); err == nil && info.IsDir() {
		dir = filepath.Join(src, "testdata")
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return dir
}

// emitTestMain writes the C main of a test build.
func emitTestMain(b *strings.Builder, prog *Program) {
	b.WriteString(testRunnerC)
	b.WriteString("\nint main(int argc, char** argv) {\n  static const narra_test_case cases[] = {\n")
	for i, t := range prog.Tests {
		fmt.Fprintf(b, "    {%s, test_%d},\n", cString(t.Name), i)
	}
	b.WriteString("    {NULL, NULL}\n  };\n  return narra_test_run(cases, argc, argv);\n}\n")
}

// testRunnerC runs after the program's functions, so it can call
// narra_consts_release.
const testRunnerC = `
// Test runner
#include <sys/types.h>
#include <sys/wait.h>
#include <time.h>
#include <errno.h>
#include <signal.h>

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
  c->fn();
  narra_consts_release();
  // A test-only hook, so the leak report can be tested.
  if (getenv("NARRA_TEST_FORCE_LEAK")) narra_live++;
  if (getenv("NARRA_CHECK_LEAKS") && narra_live != 0) {
    char buf[64];
    snprintf(buf, sizeof buf, "leaked %lld allocations", (long long)narra_live);
    narra_panic("leak check", buf);
  }
  fflush(stdout);
  _exit(0);
}

// narra_test_arg returns the option name of an argument ("--time" and
// "-time" both give "time"), or NULL for an argument that is not an option.
static const char* narra_test_opt(const char* a) {
  if (a[0] != '-') return NULL;
  return a[1] == '-' ? a + 2 : a + 1;
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
    pid_t pid = fork();
    if (pid < 0) { fprintf(stderr, "cannot fork\n"); return 1; }
    if (pid == 0) {
      close(fds[0]);
      narra_test_child(c, fds[1]);
    }
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
`
