package narra

import (
	"fmt"
	"path/filepath"
	"strings"
)

// The test runner (SPEC.md section 13, D66). A test build's main runs every
// test block in its own child process, so a failed test does not stop the
// others and the leak check covers each test alone.
//
// Options come from argv of the test binary: `--filter <text>` and `--time`.
// `narra test` turns its own flags into these arguments (TestArgs).

// TestArgs returns the test binary arguments for the `narra test` flags.
func TestArgs(filter string, timing bool) []string {
	var args []string
	if filter != "" {
		args = append(args, "--filter", filter)
	}
	if timing {
		args = append(args, "--time")
	}
	return args
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

typedef struct narra_test_case {
  const char* name;
  void (*fn)(void);
} narra_test_case;

static int64_t narra_test_now_ms(void) {
  struct timespec ts;
  clock_gettime(CLOCK_MONOTONIC, &ts);
  return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

// narra_test_child runs one test in the child process.
static void narra_test_child(const narra_test_case* c, int wfd) {
  narra_panic_fd = wfd;
  narra_panic_pos = 1;
  c->fn();
  narra_consts_release();
  fflush(stdout);
  narra_report_leaks();
  _exit(0);
}

// narra_test_run returns the exit code of the run.
static int narra_test_run(const narra_test_case* cases, int argc, char** argv) {
  const char* filter = NULL;
  int timing = 0;
  for (int i = 1; i < argc; i++) {
    if (strcmp(argv[i], "--filter") == 0 && i + 1 < argc) filter = argv[++i];
    else if (strncmp(argv[i], "--filter=", 9) == 0) filter = argv[i] + 9;
    else if (strcmp(argv[i], "--time") == 0) timing = 1;
  }
  int passed = 0, failed = 0;
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
    } else if (WIFSIGNALED(status)) {
      printf("FAIL - %s: killed by signal %d%s\n", c->name, WTERMSIG(status), tbuf);
      failed++;
    } else {
      printf("FAIL - %s: exited with code %d%s\n", c->name, WEXITSTATUS(status), tbuf);
      failed++;
    }
    free(msg);
  }
  printf("%d passed, %d failed\n", passed, failed);
  fflush(stdout);
  return failed ? 1 : 0;
}
`

// goldenDir is where Test.golden looks for `<name>.golden`: the package's
// testdata directory, or the directory of the file for a single file. The
// path is absolute, so the test binary finds it from any working directory.
func goldenDir(pkg *Package) string {
	dir := pkg.Dir
	if !pkg.single {
		dir = filepath.Join(dir, "testdata")
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return dir
}
