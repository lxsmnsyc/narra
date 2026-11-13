// A C-level test of the descriptor waits of the loop (narra_loop_wait_io),
// which no std module calls yet. Run it on macOS (kqueue) and Linux (epoll):
//   sh roadmaps/selfhost/nodes/event-loop/io_wait_test.sh
#include "../../../../selfhost/runtime/runtime.h"
#include "../../../../selfhost/runtime/std_async.c"
#include "../../../../selfhost/runtime/std_async_loop.c"
#include <sys/socket.h>

static int failures = 0;

static void expect(bool condition, const char* what) {
  if (!condition) {
    failures++;
    fprintf(stderr, "FAIL: %s\n", what);
  }
}

static int64_t value(narra_task* handle) {
  return *(int64_t*)handle->value;
}

int main(void) {
  int pair[2];
  if (socketpair(AF_UNIX, SOCK_STREAM, 0, pair) != 0) return 2;
  char byte = 'x';

  // A reader waits until the other end writes.
  narra_task* reader = narra_loop_wait_io(pair[0], false);
  expect(!reader->done, "a reader with nothing to read waits");
  // A second reader on the same descriptor is EEXIST at once on every platform.
  narra_task* second = narra_loop_wait_io(pair[0], false);
  expect(second->done && value(second) == EEXIST, "a second reader gives EEXIST");
  // A writer on the same descriptor waits beside the reader and is ready at once.
  narra_task* writer = narra_loop_wait_io(pair[0], true);
  expect(!writer->done, "a writer is registered beside the reader");
  expect(narra_loop_poll(), "the poll has waits");
  expect(writer->done && value(writer) == 0, "the writer wakes with 0");
  expect(!reader->done, "the reader still waits after the writer woke");
  expect(write(pair[1], &byte, 1) == 1, "write a byte");
  expect(narra_loop_poll(), "the poll has the reader");
  expect(reader->done && value(reader) == 0, "the reader wakes with 0");
  expect(read(pair[0], &byte, 1) == 1, "read the byte");
  expect(!narra_loop_poll(), "nothing is waited for");

  // A cancelled wait finishes with -1 and leaves the poller: a later byte
  // wakes nothing.
  narra_task* cancelled = narra_loop_wait_io(pair[0], false);
  narra_task waiting;
  memset(&waiting, 0, sizeof(waiting));
  waiting.awaited = cancelled;
  narra_loop_cancel_wait(&waiting);
  expect(cancelled->done && value(cancelled) == -1, "a cancelled wait gives -1");
  expect(!narra_loop_poll(), "a cancelled wait is out of the poller");
  // The same direction can wait again after a cancel and wakes on its byte.
  narra_task* again = narra_loop_wait_io(pair[0], false);
  expect(!again->done, "a reader waits again after a cancel");
  expect(write(pair[1], &byte, 1) == 1, "write another byte");
  expect(narra_loop_poll(), "the poll has the new reader");
  expect(again->done && value(again) == 0, "the new reader wakes with 0");

  // A sleep past the clock's range saturates instead of overflowing.
  narra_task* forever = narra_async_sleep(INT64_MAX);
  expect(((narra_wait*)forever)->deadline == INT64_MAX, "the deadline saturates");
  expect(!forever->done, "a saturated sleep waits");

  narra_task* handles[] = {reader, second, writer, cancelled, again, forever};
  narra_async_exit();
  expect(forever->done && value(forever) == -1, "exit cancels the long sleep");
  for (size_t i = 0; i < sizeof(handles) / sizeof(handles[0]); i++) {
    narra_task_release(handles[i]);
  }
  close(pair[0]);
  close(pair[1]);
  if (failures == 0) printf("io_wait_test: PASS\n");
  return failures == 0 ? 0 : 1;
}
