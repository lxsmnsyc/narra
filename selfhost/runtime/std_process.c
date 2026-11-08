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


#include <fcntl.h>
#include <poll.h>
#include <signal.h>
#include <sys/wait.h>

typedef struct narra_proc_buf {
  char* p;
  size_t len, cap;
} narra_proc_buf;

// Reads one chunk from fd. False at end of input or on a hard error.
static bool narra_proc_take(narra_proc_buf* b, int fd) {
  char chunk[4096];
  ssize_t n = read(fd, chunk, sizeof chunk);
  if (n < 0) return errno == EINTR || errno == EAGAIN;
  if (n == 0) return false;
  if (b->len + (size_t)n > b->cap) {
    size_t cap = b->cap ? b->cap * 2 : 4096;
    while (cap < b->len + (size_t)n) cap *= 2;
    char* bigger = realloc(b->p, cap);
    if (!bigger) return false;
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
  pid_t pid = fork();
  if (pid < 0) {
    *err = narra_io_fail(cmd->data, cmd->len, strerror(errno));
    narra_proc_close(po); narra_proc_close(pe); narra_proc_close(px);
    free(vec);
    return false;
  }
  if (pid == 0) {
    int e;
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
  *code = WIFEXITED(status) ? WEXITSTATUS(status) : 128 + WTERMSIG(status);
  *out = narra_io_lossy(bo.p ? bo.p : "", bo.len);
  *errout = narra_io_lossy(be.p ? be.p : "", be.len);
  free(bo.p);
  free(be.p);
  return true;
}

