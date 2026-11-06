// Runtime C for std/fs, std/io, std/path, std/process and std/env. std_fs.go,
// std_io.go, std_path.go, std_process.go and std_env.go split this file at the "// @section" lines. A module's
// Runtime is the common section plus its own, and the common section guards
// itself, so it is emitted once however many of the modules a program imports.

// @section common
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

// @section fs
#include <dirent.h>
#include <sys/stat.h>

typedef struct narra_fs_names {
  lstr** items;
  int64_t len;
} narra_fs_names;

static bool narra_fs_read(lstr* path, lstr** out, lstr** err) {
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
  if (!fail && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "file is not valid UTF-8";
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

static bool narra_fs_write(lstr* path, lstr* text, int64_t* written, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
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

// Creates one directory (mode 0777 before the umask).
static bool narra_fs_mkdir(lstr* path, lstr** err) {
  if (!narra_io_no_nul(path, err)) return false;
  if (mkdir(path->data, 0777) != 0) {
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
    if (buf[i] != '/') continue;
    buf[i] = 0;
    if (mkdir(buf, 0777) == 0) *created = true;
    else if (errno != EEXIST) failed = errno;
    buf[i] = '/';
  }
  if (!failed) {
    if (mkdir(buf, 0777) == 0) *created = true;
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

// @section io

// Reads up to limit bytes of stdin (all of it when limit is negative) into a
// malloc'ed buffer. *fail is set on a read error.
static char* narra_io_slurp(int64_t limit, size_t* len, const char** fail) {
  size_t cap = 4096;
  char* buf = malloc(cap);
  *len = 0;
  *fail = buf ? NULL : "out of memory";
  while (!*fail && (limit < 0 || (int64_t)*len < limit)) {
    if (*len == cap) {
      char* bigger = realloc(buf, cap * 2);
      if (!bigger) { *fail = "out of memory"; break; }
      buf = bigger;
      cap *= 2;
    }
    size_t want = cap - *len;
    if (limit >= 0 && (int64_t)want > limit - (int64_t)*len) want = (size_t)(limit - (int64_t)*len);
    size_t got = fread(buf + *len, 1, want, stdin);
    *len += got;
    if (got == 0) {
      if (ferror(stdin)) *fail = strerror(errno);
      break;
    }
  }
  return buf;
}

// The next line without its "\n" (a "\r" stays). *out is NULL at the end of
// the input.
static bool narra_io_read_line(lstr** out, lstr** err) {
  *out = NULL;
  char* line = NULL;
  size_t cap = 0;
  errno = 0;
  ssize_t n = getline(&line, &cap, stdin);
  if (n < 0) {
    bool bad = ferror(stdin) != 0;
    free(line);
    if (bad) {
      *err = narra_io_fail("stdin", 5, strerror(errno ? errno : EIO));
      return false;
    }
    return true;
  }
  if (n > 0 && line[n - 1] == '\n') n--;
  if (!narra_io_utf8((const unsigned char*)line, (int64_t)n)) {
    free(line);
    *err = narra_io_fail("stdin", 5, "line is not valid UTF-8");
    return false;
  }
  *out = lstr_new((int64_t)n);
  memcpy((*out)->data, line, (size_t)n);
  free(line);
  return true;
}

// Exactly count bytes.
static bool narra_io_read_exact(int64_t count, lstr** out, lstr** err) {
  if (count < 0) {
    *err = narra_io_fail("stdin", 5, "negative byte count");
    return false;
  }
  size_t len;
  const char* fail;
  char* buf = narra_io_slurp(count, &len, &fail);
  if (!fail && (int64_t)len < count) fail = "unexpected end of input";
  if (!fail && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "input is not valid UTF-8";
  if (fail) {
    free(buf);
    *err = narra_io_fail("stdin", 5, fail);
    return false;
  }
  *out = lstr_new((int64_t)len);
  memcpy((*out)->data, buf, len);
  free(buf);
  return true;
}

// The rest of the input.
static bool narra_io_read_all(lstr** out, lstr** err) {
  size_t len;
  const char* fail;
  char* buf = narra_io_slurp(-1, &len, &fail);
  if (!fail && !narra_io_utf8((const unsigned char*)buf, (int64_t)len)) fail = "input is not valid UTF-8";
  if (fail) {
    free(buf);
    *err = narra_io_fail("stdin", 5, fail);
    return false;
  }
  *out = lstr_new((int64_t)len);
  memcpy((*out)->data, buf, len);
  free(buf);
  return true;
}

static void narra_io_write(lstr* text) { fwrite(text->data, 1, (size_t)text->len, stdout); }

static void narra_io_write_error(lstr* text) { fwrite(text->data, 1, (size_t)text->len, stderr); }

static void narra_io_flush(void) { fflush(stdout); }

// @section path

static lstr* narra_path_slice(const char* p, int64_t start, int64_t end) {
  lstr* s = lstr_new(end - start);
  memcpy(s->data, p + start, (size_t)(end - start));
  return s;
}

// Node's path.posix.normalize: drops "." and empty segments, resolves "..".
static lstr* narra_path_normalize(const char* p, int64_t n) {
  bool absolute = n > 0 && p[0] == '/';
  bool trailing = n > 0 && p[n - 1] == '/';
  char* out = malloc((size_t)n + 4);
  int64_t o = 0;
  int64_t i = 0;
  while (i < n) {
    while (i < n && p[i] == '/') i++;
    int64_t start = i;
    while (i < n && p[i] != '/') i++;
    int64_t len = i - start;
    if (len == 0 || (len == 1 && p[start] == '.')) continue;
    if (len == 2 && p[start] == '.' && p[start + 1] == '.') {
      bool last_up = o >= 2 && out[o - 1] == '.' && out[o - 2] == '.' && (o == 2 || out[o - 3] == '/');
      if (o > 0 && !last_up) {
        while (o > 0 && out[o - 1] != '/') o--;
        if (o > 0) o--;
      } else if (!absolute) {
        if (o > 0) out[o++] = '/';
        out[o++] = '.';
        out[o++] = '.';
      }
      continue;
    }
    if (o > 0) out[o++] = '/';
    memcpy(out + o, p + start, (size_t)len);
    o += len;
  }
  lstr* s;
  if (o == 0) {
    s = lstr_from_cstr(absolute ? "/" : (trailing ? "./" : "."));
  } else {
    int64_t total = (absolute ? 1 : 0) + o + (trailing ? 1 : 0);
    s = lstr_new(total);
    int64_t at = 0;
    if (absolute) s->data[at++] = '/';
    memcpy(s->data + at, out, (size_t)o);
    at += o;
    if (trailing) s->data[at] = '/';
  }
  free(out);
  return s;
}

static lstr* narra_path_join(lstr* left, lstr* right) {
  if (left->len == 0 && right->len == 0) return lstr_from_cstr(".");
  if (left->len == 0) return narra_path_normalize(right->data, right->len);
  if (right->len == 0) return narra_path_normalize(left->data, left->len);
  lstr* both = lstr_new(left->len + 1 + right->len);
  memcpy(both->data, left->data, (size_t)left->len);
  both->data[left->len] = '/';
  memcpy(both->data + left->len + 1, right->data, (size_t)right->len);
  lstr* s = narra_path_normalize(both->data, both->len);
  lstr_release(both);
  return s;
}

static lstr* narra_path_dirname(lstr* path) {
  const char* p = path->data;
  int64_t n = path->len;
  if (n == 0) return lstr_from_cstr(".");
  bool root = p[0] == '/';
  int64_t end = -1;
  bool matched = true;
  for (int64_t i = n - 1; i >= 1; i--) {
    if (p[i] == '/') {
      if (!matched) { end = i; break; }
    } else {
      matched = false;
    }
  }
  if (end == -1) return lstr_from_cstr(root ? "/" : ".");
  if (root && end == 1) return lstr_from_cstr("//");
  return narra_path_slice(p, 0, end);
}

static lstr* narra_path_basename(lstr* path) {
  const char* p = path->data;
  int64_t start = 0, end = -1;
  bool matched = true;
  for (int64_t i = path->len - 1; i >= 0; i--) {
    if (p[i] == '/') {
      if (!matched) { start = i + 1; break; }
    } else if (end == -1) {
      matched = false;
      end = i + 1;
    }
  }
  if (end == -1) return lstr_from_cstr("");
  return narra_path_slice(p, start, end);
}

static lstr* narra_path_extname(lstr* path) {
  const char* p = path->data;
  int64_t start_dot = -1, start_part = 0, end = -1;
  int pre_dot = 0;
  bool matched = true;
  for (int64_t i = path->len - 1; i >= 0; i--) {
    char c = p[i];
    if (c == '/') {
      if (!matched) { start_part = i + 1; break; }
      continue;
    }
    if (end == -1) { matched = false; end = i + 1; }
    if (c == '.') {
      if (start_dot == -1) start_dot = i;
      else if (pre_dot != 1) pre_dot = 1;
    } else if (start_dot != -1) {
      pre_dot = -1;
    }
  }
  if (start_dot == -1 || end == -1 || pre_dot == 0 || (pre_dot == 1 && start_dot == end - 1 && start_dot == start_part + 1)) {
    return lstr_from_cstr("");
  }
  return narra_path_slice(p, start_dot, end);
}

// @section process
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

// @section env
#ifdef __APPLE__
#include <crt_externs.h>
#endif

// NULL when the variable is not set.
static lstr* narra_env_get(lstr* name) {
  if (memchr(name->data, 0, (size_t)name->len)) return NULL;
  const char* v = getenv(name->data);
  return v ? narra_io_lossy(v, strlen(v)) : NULL;
}

#ifdef __APPLE__
static int64_t narra_env_argc(void) { return *_NSGetArgc(); }
static lstr* narra_env_arg(int64_t i) {
  const char* a = (*_NSGetArgv())[i];
  return narra_io_lossy(a, strlen(a));
}
#else
// Elsewhere the arguments are read once from /proc/self/cmdline.
static char* narra_env_buf;
static size_t narra_env_len;
static int64_t narra_env_n = -1;

static void narra_env_load(void) {
  if (narra_env_n >= 0) return;
  narra_env_n = 0;
  FILE* f = fopen("/proc/self/cmdline", "rb");
  if (!f) return;
  size_t cap = 4096;
  narra_env_buf = malloc(cap);
  while (narra_env_buf) {
    if (narra_env_len == cap) {
      char* bigger = realloc(narra_env_buf, cap * 2);
      if (!bigger) break;
      narra_env_buf = bigger;
      cap *= 2;
    }
    size_t got = fread(narra_env_buf + narra_env_len, 1, cap - narra_env_len, f);
    if (got == 0) break;
    narra_env_len += got;
  }
  fclose(f);
  for (size_t i = 0; i < narra_env_len; i++) {
    if (narra_env_buf[i] == 0) narra_env_n++;
  }
}
static int64_t narra_env_argc(void) { narra_env_load(); return narra_env_n; }
static lstr* narra_env_arg(int64_t i) {
  narra_env_load();
  const char* p = narra_env_buf;
  while (i-- > 0) p += strlen(p) + 1;
  return narra_io_lossy(p, strlen(p));
}
#endif
