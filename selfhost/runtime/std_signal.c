// std/internal/signal. SIGINT and SIGTERM as a descriptor the loop can wait on.
//
// The first start installs one handler per signal and opens a pipe (a socket
// pair on Windows). The handler only writes one byte, 1 for SIGINT and 2 for
// SIGTERM, so it stays async-signal-safe and stays installed: every signal
// that comes adds a byte, and each take reads one. A signal that comes while
// Process.run runs goes to the child as well and is recorded the same way
// (std_process.c). The wait on the descriptor is the loop's (narra_loop_wait_io
// in std_async_loop.c). Windows maps Ctrl-C to 1 and Ctrl-Break to 2 through
// SetConsoleCtrlHandler.
#include <errno.h>
#include <fcntl.h>
#ifdef _WIN32
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#else
#include <sched.h>
#include <signal.h>
#include <stdatomic.h>
#include <unistd.h>
#endif

// 0 before the first start, 1 while a thread starts, 2 once it is done.
static int narra_signal_state = 0;
// The read end of the pipe, or the negated errno value of a failed start.
static int64_t narra_signal_descriptor = -1;

#ifdef _WIN32
static SOCKET narra_signal_writer = INVALID_SOCKET;

static BOOL WINAPI narra_signal_console_handler(DWORD event) {
  unsigned char code;
  if (event == CTRL_C_EVENT) {
    code = 1;
  } else if (event == CTRL_BREAK_EVENT) {
    code = 2;
  } else {
    return FALSE;
  }
  send(narra_signal_writer, (const char*)&code, 1, 0);
  return TRUE;
}

static int narra_signal_open(void) {
  WSADATA data;
  WSAStartup(MAKEWORD(2, 2), &data);
  // A connected pair of loopback sockets: the handler thread sends, the loop reads.
  SOCKET listener = socket(AF_INET, SOCK_STREAM, 0);
  if (listener == INVALID_SOCKET) return EIO;
  struct sockaddr_in address;
  memset(&address, 0, sizeof(address));
  address.sin_family = AF_INET;
  address.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
  int size = sizeof(address);
  SOCKET reader = INVALID_SOCKET;
  SOCKET writer = socket(AF_INET, SOCK_STREAM, 0);
  if (writer == INVALID_SOCKET || bind(listener, (struct sockaddr*)&address, size) != 0 ||
      listen(listener, 1) != 0 || getsockname(listener, (struct sockaddr*)&address, &size) != 0 ||
      connect(writer, (struct sockaddr*)&address, size) != 0 ||
      (reader = accept(listener, NULL, NULL)) == INVALID_SOCKET) {
    if (writer != INVALID_SOCKET) closesocket(writer);
    closesocket(listener);
    return EIO;
  }
  closesocket(listener);
  u_long nonblocking = 1;
  ioctlsocket(reader, FIONBIO, &nonblocking);
  ioctlsocket(writer, FIONBIO, &nonblocking);
  SetHandleInformation((HANDLE)(uintptr_t)reader, HANDLE_FLAG_INHERIT, 0);
  SetHandleInformation((HANDLE)(uintptr_t)writer, HANDLE_FLAG_INHERIT, 0);
  narra_signal_writer = writer;
  if (!SetConsoleCtrlHandler(narra_signal_console_handler, TRUE)) {
    narra_signal_writer = INVALID_SOCKET;
    closesocket(reader);
    closesocket(writer);
    return EIO;
  }
  narra_signal_descriptor = (int64_t)reader;
  return 0;
}

static int64_t narra_signal_take(int64_t descriptor) {
  unsigned char code;
  int count = recv((SOCKET)descriptor, (char*)&code, 1, 0);
  if (count == 1) return code;
  return 0;
}
#else
// Shared with std/internal/process, which records the signals of a running child too.
#ifndef NARRA_SIGNAL_COMMON
#define NARRA_SIGNAL_COMMON
typedef void (*narra_signal_function)(int);
static _Atomic int narra_signal_writer = -1;
static _Atomic(narra_signal_function) narra_signal_handler_in_use = NULL;

static void narra_signal_note(int number) {
  int saved = errno;
  int writer = atomic_load(&narra_signal_writer);
  if (writer >= 0 && (number == SIGINT || number == SIGTERM)) {
    unsigned char code = number == SIGINT ? 1 : 2;
    ssize_t ignored = write(writer, &code, 1);
    (void)ignored;
  }
  errno = saved;
}
#endif

static int narra_signal_prepare(int descriptor) {
  int flags = fcntl(descriptor, F_GETFL, 0);
  if (flags < 0 || fcntl(descriptor, F_SETFL, flags | O_NONBLOCK) < 0) return errno;
  if (fcntl(descriptor, F_SETFD, FD_CLOEXEC) < 0) return errno;
  return 0;
}

static int narra_signal_open(void) {
  // pipe2 would set both flags at once, but it is hidden under _POSIX_C_SOURCE.
  int ends[2];
  if (pipe(ends) < 0) return errno;
  int failure = narra_signal_prepare(ends[0]);
  if (failure == 0) failure = narra_signal_prepare(ends[1]);
  if (failure != 0) {
    close(ends[0]);
    close(ends[1]);
    return failure;
  }
  struct sigaction action;
  memset(&action, 0, sizeof(action));
  action.sa_handler = narra_signal_note;
  sigemptyset(&action.sa_mask);
  action.sa_flags = SA_RESTART;
  atomic_store(&narra_signal_writer, ends[1]);
  // Publish the handler before installing it, so a Process.run that ends meanwhile puts it back.
  atomic_store(&narra_signal_handler_in_use, narra_signal_note);
  struct sigaction previous;
  if (sigaction(SIGINT, &action, &previous) < 0) failure = errno;
  if (failure == 0 && sigaction(SIGTERM, &action, NULL) < 0) {
    // Put SIGINT back, so Ctrl-C does not write to a pipe nobody reads.
    failure = errno;
    sigaction(SIGINT, &previous, NULL);
  }
  if (failure != 0) {
    atomic_store(&narra_signal_handler_in_use, NULL);
    atomic_store(&narra_signal_writer, -1);
    close(ends[0]);
    close(ends[1]);
    return failure;
  }
  narra_signal_descriptor = ends[0];
  return 0;
}

static int64_t narra_signal_take(int64_t descriptor) {
  unsigned char code;
  while (true) {
    ssize_t count = read((int)descriptor, &code, 1);
    if (count == 1) return code;
    if (count < 0 && errno == EINTR) continue;
    return 0;
  }
}
#endif

static int64_t narra_signal_start(void) {
  // The first caller installs; the others wait until it is done and share the answer.
  int expected = 0;
  if (__atomic_compare_exchange_n(&narra_signal_state, &expected, 1, false, __ATOMIC_ACQ_REL, __ATOMIC_ACQUIRE)) {
    int failure = narra_signal_open();
    if (failure != 0) narra_signal_descriptor = -(int64_t)failure;
    __atomic_store_n(&narra_signal_state, 2, __ATOMIC_RELEASE);
  } else {
    while (__atomic_load_n(&narra_signal_state, __ATOMIC_ACQUIRE) != 2) {
#ifdef _WIN32
      Sleep(0);
#else
      sched_yield();
#endif
    }
  }
  return narra_signal_descriptor;
}

static lstr* narra_signal_message(int64_t error) {
  return lstr_from_cstr(strerror((int)error));
}
