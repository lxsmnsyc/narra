// std/internal/terminal. Whether a stream is a terminal, and its width in columns.
#ifndef __wasi__
#include <sys/ioctl.h>
#endif
#include <unistd.h>

static bool narra_terminal_is_terminal(int64_t descriptor) {
  return isatty((int)descriptor) != 0;
}

// The columns of the terminal on the descriptor; 0 when it is not a terminal or reports no width.
static int64_t narra_terminal_columns(int64_t descriptor) {
#ifdef __wasi__
  // WASI has no ioctl, so a terminal reports no width.
  (void)descriptor;
  return 0;
#else
  struct winsize size;
  if (ioctl((int)descriptor, TIOCGWINSZ, &size) != 0) return 0;
  return (int64_t)size.ws_col;
#endif
}
