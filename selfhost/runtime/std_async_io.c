// std/internal/asyncio. Sockets for std/net.
//
// A Handle is a counted object that holds one descriptor; the last release
// closes it. A handle whose call failed holds -1. Every call stores its
// errno value in narra_io_error (0 on success), which Handles.error reads
// before any other call can run. Nothing here blocks: the waits are the
// loop's (narra_loop_wait_io in std_async_loop.c), and the emitter calls them
// itself, so this text needs nothing from the task runtime.
#include <errno.h>
#include <fcntl.h>
#include <limits.h>
#ifdef _WIN32
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#else
#include <arpa/inet.h>
#include <netdb.h>
#include <netinet/in.h>
#include <pthread.h>
#include <signal.h>
#include <sys/socket.h>
#endif

// Name lookup runs getaddrinfo on a helper thread, which needs POSIX threads
// and a pipe. Windows and WASI have no lookup: a name there gives a failure.
#if !defined(_WIN32) && !defined(__wasi__)
#define NARRA_LOOKUP 1
#endif
#ifdef __linux__
// Hidden under _POSIX_C_SOURCE on glibc.
extern int pipe2(int descriptors[2], int flags);
#endif

// Sockets on Windows are Winsock handles that the C runtime cannot read,
// write or close, and their errors come from WSAGetLastError. The calls below
// go through these names; elsewhere they are the plain calls. A Winsock handle
// is a small number, so it fits the int that holds a descriptor, and
// INVALID_SOCKET becomes -1.
#ifdef _WIN32
static int narra_net_errno(void) {
  switch (WSAGetLastError()) {
    case WSAEWOULDBLOCK: return EWOULDBLOCK;
    case WSAEINPROGRESS: return EINPROGRESS;
    case WSAEINTR: return EINTR;
    case WSAEBADF: return EBADF;
    case WSAEACCES: return EACCES;
    case WSAEINVAL: return EINVAL;
    case WSAEMFILE: return EMFILE;
    case WSAENOTSOCK: return ENOTSOCK;
    case WSAEAFNOSUPPORT: return EAFNOSUPPORT;
    case WSAEADDRINUSE: return EADDRINUSE;
    case WSAEADDRNOTAVAIL: return EADDRNOTAVAIL;
    case WSAENETDOWN: return ENETDOWN;
    case WSAENETUNREACH: return ENETUNREACH;
    case WSAECONNABORTED: return ECONNABORTED;
    case WSAECONNRESET: return ECONNRESET;
    case WSAESHUTDOWN: return EPIPE;
    case WSAENOTCONN: return ENOTCONN;
    case WSAETIMEDOUT: return ETIMEDOUT;
    case WSAECONNREFUSED: return ECONNREFUSED;
    case WSAEHOSTUNREACH: return EHOSTUNREACH;
    case WSAEMSGSIZE: return EMSGSIZE;
    case WSAEISCONN: return EISCONN;
    case WSAEALREADY: return EALREADY;
    default: return EIO;
  }
}

static void narra_net_start(void) {
  static bool started = false;
  if (started) return;
  WSADATA data;
  WSAStartup(MAKEWORD(2, 2), &data);
  started = true;
}

static int narra_net_prepare(int descriptor) {
  u_long nonblocking = 1;
  if (ioctlsocket((SOCKET)descriptor, FIONBIO, &nonblocking) != 0) return narra_net_errno();
  SetHandleInformation((HANDLE)(uintptr_t)(SOCKET)descriptor, HANDLE_FLAG_INHERIT, 0);
  return 0;
}

#define narra_net_close(descriptor) closesocket((SOCKET)(descriptor))
#define narra_net_read(descriptor, buffer, length) recv((SOCKET)(descriptor), (buffer), (int)(length), 0)
#define narra_net_send(descriptor, buffer, length) send((SOCKET)(descriptor), (buffer), (length) > INT_MAX ? INT_MAX : (int)(length), 0)
#else
#define narra_net_errno() errno
#define narra_net_start()
#define narra_net_close(descriptor) close(descriptor)
#define narra_net_read(descriptor, buffer, length) read((descriptor), (buffer), (length))
#define narra_net_send(descriptor, buffer, length) send((descriptor), (buffer), (length), MSG_NOSIGNAL)
#endif

struct narra_lookup;

typedef struct narra_handle {
  int64_t rc;
  int descriptor;
  // The lookup a lookup handle waits for, or NULL. The descriptor of such a
  // handle is the read end of the lookup's pipe, which the lookup closes.
  struct narra_lookup* lookup;
} narra_handle;

// Not an errno value: the host is not a numeric IPv4 or IPv6 address.
#define NARRA_IO_BAD_ADDRESS (-2)
// Not an errno value: a lookup that found no address. Its text is the result.
#define NARRA_IO_LOOKUP_FAILED (-3)
// Not an errno value: this platform has no name lookup.
#define NARRA_IO_NO_LOOKUP (-4)

static _Thread_local int narra_io_error = 0;

static narra_handle* narra_handle_new(int descriptor) {
  narra_handle* handle = narra_alloc(sizeof(narra_handle));
  handle->rc = 1;
  handle->descriptor = descriptor;
  handle->lookup = NULL;
  return handle;
}

static void narra_lookup_release(struct narra_lookup* lookup);

static void narra_handle_retain(narra_handle* handle) {
  handle->rc++;
}

static void narra_handle_release(narra_handle* handle) {
  if (--handle->rc == 0) {
    if (handle->lookup) {
      narra_lookup_release(handle->lookup);
    } else if (handle->descriptor >= 0) {
      narra_net_close(handle->descriptor);
    }
    narra_free(handle);
  }
}

// A failed call: keeps errno and gives a handle with no descriptor.
static narra_handle* narra_handle_failed(int descriptor, int error) {
  if (descriptor >= 0) narra_net_close(descriptor);
  narra_io_error = error;
  return narra_handle_new(-1);
}

// Makes the descriptor nonblocking and closed on exec.
static int narra_handle_prepare(int descriptor) {
#ifdef _WIN32
  return narra_net_prepare(descriptor);
#else
  int flags = fcntl(descriptor, F_GETFL, 0);
  if (flags < 0 || fcntl(descriptor, F_SETFL, flags | O_NONBLOCK) < 0) return errno;
  if (fcntl(descriptor, F_SETFD, FD_CLOEXEC) < 0) return errno;
  return 0;
#endif
}

// Fills address from a numeric host and a port; false when the host is not
// a numeric address or the port is out of range.
static bool narra_handle_address(lstr* host, int64_t port, struct sockaddr_storage* address, socklen_t* size) {
  memset(address, 0, sizeof(*address));
  if (port < 0 || port > 65535) return false;
  struct sockaddr_in* ipv4 = (struct sockaddr_in*)address;
  if (inet_pton(AF_INET, host->data, &ipv4->sin_addr) == 1) {
    ipv4->sin_family = AF_INET;
    ipv4->sin_port = htons((uint16_t)port);
    *size = sizeof(struct sockaddr_in);
    return true;
  }
  struct sockaddr_in6* ipv6 = (struct sockaddr_in6*)address;
  if (inet_pton(AF_INET6, host->data, &ipv6->sin6_addr) == 1) {
    ipv6->sin6_family = AF_INET6;
    ipv6->sin6_port = htons((uint16_t)port);
    *size = sizeof(struct sockaddr_in6);
    return true;
  }
  return false;
}

// A new nonblocking stream socket for address, or -1 with narra_io_error set.
static int narra_handle_socket(struct sockaddr_storage* address) {
  narra_net_start();
  int descriptor = (int)socket(address->ss_family, SOCK_STREAM, 0);
  if (descriptor < 0) {
    narra_io_error = narra_net_errno();
    return -1;
  }
  int failure = narra_handle_prepare(descriptor);
  if (failure != 0) {
    narra_net_close(descriptor);
    narra_io_error = failure;
    return -1;
  }
  return descriptor;
}

static narra_handle* narra_handle_listen(lstr* host, int64_t port) {
  narra_io_error = 0;
  struct sockaddr_storage address;
  socklen_t size = 0;
  if (!narra_handle_address(host, port, &address, &size)) {
    return narra_handle_failed(-1, NARRA_IO_BAD_ADDRESS);
  }
  int descriptor = narra_handle_socket(&address);
  if (descriptor < 0) return narra_handle_new(-1);
  // On Windows SO_REUSEADDR would let a second socket take a port in use.
#ifndef _WIN32
  int reuse = 1;
  setsockopt(descriptor, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));
#endif
  if (bind(descriptor, (struct sockaddr*)&address, size) < 0) {
    return narra_handle_failed(descriptor, narra_net_errno());
  }
  if (listen(descriptor, SOMAXCONN) < 0) return narra_handle_failed(descriptor, narra_net_errno());
  return narra_handle_new(descriptor);
}

// Starts a connect. narra_io_error is EINPROGRESS while it is under way:
// wait until the descriptor is writable, then read Handles.connected.
static narra_handle* narra_handle_connect(lstr* host, int64_t port) {
  narra_io_error = 0;
  struct sockaddr_storage address;
  socklen_t size = 0;
  if (!narra_handle_address(host, port, &address, &size)) {
    return narra_handle_failed(-1, NARRA_IO_BAD_ADDRESS);
  }
  int descriptor = narra_handle_socket(&address);
  if (descriptor < 0) return narra_handle_new(-1);
  if (connect(descriptor, (struct sockaddr*)&address, size) < 0) {
    int failure = narra_net_errno();
    // A nonblocking connect on Windows reports that it would block.
    if (failure != EINPROGRESS && failure != EINTR && failure != EWOULDBLOCK) return narra_handle_failed(descriptor, failure);
    narra_io_error = EINPROGRESS;
  }
  return narra_handle_new(descriptor);
}

// The result of a connect that was under way: 0 or its errno value.
static int64_t narra_handle_connected(narra_handle* handle) {
  int failure = 0;
  socklen_t size = sizeof(failure);
  if (getsockopt(handle->descriptor, SOL_SOCKET, SO_ERROR, (char*)&failure, &size) < 0) {
    failure = narra_net_errno();
  }
#ifdef _WIN32
  else if (failure != 0) {
    // SO_ERROR holds a Winsock code.
    WSASetLastError(failure);
    failure = narra_net_errno();
  }
#endif
  narra_io_error = failure;
  return failure;
}

static narra_handle* narra_handle_accept(narra_handle* listener) {
  narra_io_error = 0;
  if (listener->descriptor < 0) return narra_handle_failed(-1, EBADF);
  int descriptor = (int)accept(listener->descriptor, NULL, NULL);
  if (descriptor < 0) return narra_handle_failed(-1, narra_net_errno());
  int failure = narra_handle_prepare(descriptor);
  if (failure != 0) return narra_handle_failed(descriptor, failure);
  return narra_handle_new(descriptor);
}

// The port the socket is bound to, or -1.
static int64_t narra_handle_local_port(narra_handle* handle) {
  struct sockaddr_storage address;
  socklen_t size = sizeof(address);
  if (handle->descriptor < 0) return -1;
  if (getsockname(handle->descriptor, (struct sockaddr*)&address, &size) < 0) return -1;
  if (address.ss_family == AF_INET) return ntohs(((struct sockaddr_in*)&address)->sin_port);
  if (address.ss_family == AF_INET6) return ntohs(((struct sockaddr_in6*)&address)->sin6_port);
  return -1;
}

// One read takes at most this many bytes, whatever the caller asks for.
#define NARRA_IO_READ_LIMIT 65536

static _Thread_local char narra_io_buffer[NARRA_IO_READ_LIMIT];

// Up to limit bytes, at most NARRA_IO_READ_LIMIT; empty bytes for a limit of
// zero or less, at the end of the stream and on a failure, which
// narra_io_error tells apart. The result holds only the bytes that came.
static lstr* narra_handle_read(narra_handle* handle, int64_t limit) {
  narra_io_error = 0;
  if (limit < 1) return lstr_new(0);
  if (limit > NARRA_IO_READ_LIMIT) limit = NARRA_IO_READ_LIMIT;
  if (handle->descriptor < 0) {
    narra_io_error = EBADF;
    return lstr_new(0);
  }
  ssize_t count = narra_net_read(handle->descriptor, narra_io_buffer, (size_t)limit);
  if (count < 0) {
    narra_io_error = narra_net_errno();
    return lstr_new(0);
  }
  lstr* data = lstr_new(count);
  memcpy(data->data, narra_io_buffer, (size_t)count);
  return data;
}

// Writes the bytes of data from offset on and gives how many went, or -1.
// MSG_NOSIGNAL: a closed peer is an EPIPE error, not a SIGPIPE (Windows has no SIGPIPE).
static int64_t narra_handle_write(narra_handle* handle, lstr* data, int64_t offset) {
  narra_io_error = 0;
  if (handle->descriptor < 0) {
    narra_io_error = EBADF;
    return -1;
  }
  if (offset < 0 || offset > data->len) offset = data->len;
  ssize_t count = narra_net_send(handle->descriptor, data->data + offset, (size_t)(data->len - offset));
  if (count < 0) {
    narra_io_error = narra_net_errno();
    return -1;
  }
  return count;
}

// Whether accept failed only for the connection it took, which the peer
// dropped or the network lost (accept(2)), so the next one may work.
static bool narra_handle_accept_again(int64_t error) {
  if (error == ECONNABORTED) return true;
#ifdef __linux__
  // Linux passes on network errors of the new connection; the others are POSIX.
  if (error == ENETDOWN || error == EPROTO || error == ENOPROTOOPT) return true;
  if (error == EHOSTUNREACH || error == EOPNOTSUPP || error == ENETUNREACH) return true;
#ifdef EHOSTDOWN
  if (error == EHOSTDOWN) return true;
#endif
#ifdef ENONET
  if (error == ENONET) return true;
#endif
#endif
  return false;
}

// Whether a wait failed because another task waits on the socket in the same
// direction: one reader and one writer at a time.
static bool narra_handle_busy(int64_t error) {
  return error == EEXIST;
}

// Whether a call failed only because it would have to wait.
static bool narra_handle_retry(int64_t error) {
  return error == EAGAIN || error == EWOULDBLOCK || error == EINTR;
}

static lstr* narra_handle_message(int64_t error) {
  if (error == NARRA_IO_BAD_ADDRESS) return lstr_from_cstr("not a numeric IPv4 or IPv6 address and port");
  if (error == NARRA_IO_NO_LOOKUP) return lstr_from_cstr("name lookup is not supported on this platform");
#ifdef _WIN32
  if (error == ECONNREFUSED) return lstr_from_cstr("Connection refused");
  if (error == ECONNRESET) return lstr_from_cstr("Connection reset by peer");
  if (error == EADDRINUSE) return lstr_from_cstr("Address already in use");
  if (error == ETIMEDOUT) return lstr_from_cstr("Connection timed out");
  if (error == ENOTCONN) return lstr_from_cstr("Socket is not connected");
  if (error == EPIPE) return lstr_from_cstr("Broken pipe");
#endif
  return lstr_from_cstr(strerror((int)error));
}

// The errno value of the last call, 0 when it succeeded.
static int64_t narra_handle_error(void) {
  return narra_io_error;
}

// Whether a connect is under way.
static bool narra_handle_pending(int64_t error) {
  return error == EINPROGRESS;
}

// Whether host is a numeric IPv4 or IPv6 address, which needs no lookup.
static bool narra_handle_is_numeric(lstr* host) {
  struct in_addr ipv4;
  struct in6_addr ipv6;
  return inet_pton(AF_INET, host->data, &ipv4) == 1 || inet_pton(AF_INET6, host->data, &ipv6) == 1;
}

#ifdef NARRA_LOOKUP
// A name lookup. getaddrinfo blocks, so a helper thread runs it and then
// writes one byte to a pipe, which the loop waits on like any descriptor. The
// thread and the handle each hold one reference, and the last release frees
// the lookup and closes both ends of the pipe. A cancelled wait therefore
// leaves the thread to finish and free the result, and the pipe is never
// closed under a writer. The lookup uses plain malloc: the thread has no heap.
typedef struct narra_lookup {
  int references;
  int reader;
  int writer;
  char* host;
  // The addresses one per line, or the failure text when failed is set.
  char* result;
  bool failed;
} narra_lookup;

static void narra_lookup_release(narra_lookup* lookup) {
  if (__atomic_fetch_sub(&lookup->references, 1, __ATOMIC_ACQ_REL) != 1) return;
  close(lookup->reader);
  close(lookup->writer);
  free(lookup->host);
  free(lookup->result);
  free(lookup);
}

static char* narra_lookup_text(const char* text, size_t length) {
  char* copy = malloc(length + 1);
  if (!copy) return NULL;
  memcpy(copy, text, length);
  copy[length] = '\0';
  return copy;
}

static void narra_lookup_fail(narra_lookup* lookup, const char* message) {
  lookup->failed = true;
  lookup->result = narra_lookup_text(message, strlen(message));
}

// Whether text has a line equal to the address.
static bool narra_lookup_has_line(const char* text, const char* address) {
  size_t count = strlen(address);
  for (const char* line = text; *line;) {
    const char* end = strchr(line, '\n');
    size_t size = end ? (size_t)(end - line) : strlen(line);
    if (size == count && memcmp(line, address, count) == 0) return true;
    line += size + (end ? 1 : 0);
  }
  return false;
}

// Runs getaddrinfo and keeps every distinct address in the system's order.
static void narra_lookup_resolve(narra_lookup* lookup) {
  struct addrinfo hints;
  memset(&hints, 0, sizeof(hints));
  hints.ai_family = AF_UNSPEC;
  hints.ai_socktype = SOCK_STREAM;
  struct addrinfo* found = NULL;
  int status = getaddrinfo(lookup->host, NULL, &hints, &found);
  if (status != 0) {
    narra_lookup_fail(lookup, gai_strerror(status));
    return;
  }
  size_t capacity = 64;
  size_t length = 0;
  char* text = malloc(capacity);
  if (!text) {
    freeaddrinfo(found);
    narra_lookup_fail(lookup, "out of memory");
    return;
  }
  text[0] = '\0';
  for (struct addrinfo* entry = found; entry; entry = entry->ai_next) {
    char address[INET6_ADDRSTRLEN];
    const void* source = NULL;
    if (entry->ai_family == AF_INET) source = &((struct sockaddr_in*)entry->ai_addr)->sin_addr;
    if (entry->ai_family == AF_INET6) source = &((struct sockaddr_in6*)entry->ai_addr)->sin6_addr;
    if (!source || !inet_ntop(entry->ai_family, source, address, sizeof(address))) continue;
    if (narra_lookup_has_line(text, address)) continue;
    size_t count = strlen(address);
    if (length + count + 2 > capacity) {
      while (length + count + 2 > capacity) capacity *= 2;
      char* grown = realloc(text, capacity);
      if (!grown) {
        free(text);
        freeaddrinfo(found);
        narra_lookup_fail(lookup, "out of memory");
        return;
      }
      text = grown;
    }
    if (length > 0) text[length++] = '\n';
    memcpy(text + length, address, count);
    length += count;
    text[length] = '\0';
  }
  freeaddrinfo(found);
  if (length == 0) {
    free(text);
    narra_lookup_fail(lookup, "no address found");
    return;
  }
  lookup->result = text;
}

static void* narra_lookup_main(void* argument) {
  narra_lookup* lookup = argument;
  narra_lookup_resolve(lookup);
  // Wakes the loop: one byte goes into an empty pipe, so the write never blocks.
  char byte = 1;
  ssize_t written = write(lookup->writer, &byte, 1);
  (void)written;
  narra_lookup_release(lookup);
  return NULL;
}

// A handle that becomes readable once the lookup of host is done; read the
// result with narra_handle_lookup_result. A failed start sets narra_io_error.
static narra_handle* narra_handle_lookup(lstr* host) {
  narra_io_error = 0;
  int ends[2];
#ifdef __linux__
  // Both ends close on exec from the start, so a concurrent fork and exec never inherits them.
  if (pipe2(ends, O_CLOEXEC) < 0) return narra_handle_failed(-1, errno);
#else
  if (pipe(ends) < 0) return narra_handle_failed(-1, errno);
  fcntl(ends[0], F_SETFD, FD_CLOEXEC);
  fcntl(ends[1], F_SETFD, FD_CLOEXEC);
#endif
  narra_lookup* lookup = malloc(sizeof(narra_lookup));
  if (!lookup) {
    close(ends[0]);
    close(ends[1]);
    return narra_handle_failed(-1, ENOMEM);
  }
  memset(lookup, 0, sizeof(*lookup));
  lookup->references = 2;
  lookup->reader = ends[0];
  lookup->writer = ends[1];
  lookup->host = narra_lookup_text(host->data, (size_t)host->len);
  if (!lookup->host) {
    lookup->references = 1;
    narra_lookup_release(lookup);
    return narra_handle_failed(-1, ENOMEM);
  }
  pthread_attr_t attributes;
  pthread_attr_init(&attributes);
  pthread_attr_setdetachstate(&attributes, PTHREAD_CREATE_DETACHED);
  pthread_t thread;
  // The helper runs with every signal blocked, so a signal for the process
  // lands on a thread that can handle it and getaddrinfo never sees EINTR.
  sigset_t everything;
  sigset_t previous;
  sigfillset(&everything);
  pthread_sigmask(SIG_SETMASK, &everything, &previous);
  int failure = pthread_create(&thread, &attributes, narra_lookup_main, lookup);
  pthread_sigmask(SIG_SETMASK, &previous, NULL);
  pthread_attr_destroy(&attributes);
  if (failure != 0) {
    lookup->references = 1;
    narra_lookup_release(lookup);
    return narra_handle_failed(-1, failure);
  }
  narra_handle* handle = narra_handle_new(ends[0]);
  handle->lookup = lookup;
  return handle;
}

// The addresses of a finished lookup, one per line. When it failed the result
// is the failure text and narra_io_error is NARRA_IO_LOOKUP_FAILED.
static lstr* narra_handle_lookup_result(narra_handle* handle) {
  narra_io_error = 0;
  narra_lookup* lookup = handle->lookup;
  if (!lookup || !lookup->result) {
    narra_io_error = NARRA_IO_LOOKUP_FAILED;
    return lstr_from_cstr("lookup did not finish");
  }
  if (lookup->failed) narra_io_error = NARRA_IO_LOOKUP_FAILED;
  return lstr_from_cstr(lookup->result);
}
#else
static void narra_lookup_release(struct narra_lookup* lookup) {
  (void)lookup;
}

static narra_handle* narra_handle_lookup(lstr* host) {
  (void)host;
  return narra_handle_failed(-1, NARRA_IO_NO_LOOKUP);
}

static lstr* narra_handle_lookup_result(narra_handle* handle) {
  (void)handle;
  narra_io_error = NARRA_IO_LOOKUP_FAILED;
  return lstr_from_cstr("name lookup is not supported on this platform");
}
#endif
