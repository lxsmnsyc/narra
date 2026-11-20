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
#include <netinet/in.h>
#include <sys/socket.h>
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

typedef struct narra_handle {
  int64_t rc;
  int descriptor;
} narra_handle;

// Not an errno value: the host is not a numeric IPv4 or IPv6 address.
#define NARRA_IO_BAD_ADDRESS (-2)

static _Thread_local int narra_io_error = 0;

static narra_handle* narra_handle_new(int descriptor) {
  narra_handle* handle = narra_alloc(sizeof(narra_handle));
  handle->rc = 1;
  handle->descriptor = descriptor;
  return handle;
}

static void narra_handle_retain(narra_handle* handle) {
  handle->rc++;
}

static void narra_handle_release(narra_handle* handle) {
  if (--handle->rc == 0) {
    if (handle->descriptor >= 0) narra_net_close(handle->descriptor);
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
