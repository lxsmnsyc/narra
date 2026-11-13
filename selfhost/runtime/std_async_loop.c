// std/internal/async. The loop of a thread (D109 items 6, 7 and 8): a ready
// queue of tasks to resume, the unfinished tasks, the tasks waiting in
// Async.idle, a timer heap and a poller (kqueue on macOS, epoll on Linux).
//
// A turn runs the ready queue in order until it is empty, then waits in the
// poller until the nearest timer deadline or until a registered descriptor is
// ready. When nothing is ready and nothing is waited for, the oldest idle
// handle finishes.
//
// The loop waits through wait handles: bare tasks with no body (narra_wait)
// whose value is an int64_t. A timer or descriptor wait finishes its handle
// with 0 or more (the nanoseconds slept, or 0 when a descriptor is ready),
// -1 when the wait was cancelled, and an errno value above 0 when the poller
// refused the descriptor. A std function awaits the handle and turns the value
// into its Result.
#if defined(__APPLE__)
#include <sys/event.h>
#elif defined(__linux__)
#include <sys/epoll.h>
#else
#error "async needs a poller: kqueue (macOS) or epoll (Linux)"
#endif
#include <errno.h>
#include <time.h>

typedef struct narra_wait narra_wait;
struct narra_wait {
  narra_task task;
  int64_t value;
  // The monotonic time the wait started, its deadline and its creation
  // number, which orders equal deadlines.
  int64_t started;
  int64_t deadline;
  int64_t sequence;
  // The index in the timer heap, or -1.
  int64_t slot;
  // The descriptor it waits on, or -1, and the list of descriptor waits.
  int descriptor;
  bool writable;
  narra_wait* next_io;
  narra_wait* previous_io;
};

typedef struct narra_loop {
  narra_task* ready_head;
  narra_task* ready_tail;
  narra_task* live;
  narra_task* idle_head;
  narra_task* idle_tail;
  // The task whose body runs now, or NULL in a plain function.
  narra_task* current;
  // Set once the entry has finished: every task is cancelled from then on.
  bool exiting;
  // The timers: a binary min-heap by deadline, then by creation number.
  narra_wait** timers;
  int64_t timer_count;
  int64_t timer_capacity;
  int64_t timer_sequence;
  // The descriptor waits, and the poller, opened on the first wait.
  narra_wait* io;
  int64_t io_count;
  int poller;
  bool poller_open;
} narra_loop;

static _Thread_local narra_loop narra_async_loop;

static void narra_loop_add(narra_task* task) {
  narra_task* current = narra_async_loop.current;
  task->cancelled = narra_async_loop.exiting;
  task->started_cancelled = current != NULL && current->cancelled;
  task->next_live = narra_async_loop.live;
  task->previous_live = NULL;
  if (narra_async_loop.live) narra_async_loop.live->previous_live = task;
  narra_async_loop.live = task;
}

static void narra_loop_remove(narra_task* task) {
  if (task->previous_live) {
    task->previous_live->next_live = task->next_live;
  } else {
    narra_async_loop.live = task->next_live;
  }
  if (task->next_live) task->next_live->previous_live = task->previous_live;
  task->next_live = NULL;
  task->previous_live = NULL;
}

static void narra_loop_queue(narra_task* task) {
  if (task->queued) return;
  task->queued = true;
  task->next_ready = NULL;
  if (narra_async_loop.ready_tail) {
    narra_async_loop.ready_tail->next_ready = task;
  } else {
    narra_async_loop.ready_head = task;
  }
  narra_async_loop.ready_tail = task;
}

static narra_task* narra_loop_enter(narra_task* task) {
  narra_task* previous = narra_async_loop.current;
  narra_async_loop.current = task;
  return previous;
}

static void narra_loop_leave(narra_task* previous) {
  narra_async_loop.current = previous;
}

// Resumes the ready tasks in order until none is ready.
static void narra_loop_run_ready(void) {
  while (narra_async_loop.ready_head) {
    narra_task* task = narra_async_loop.ready_head;
    narra_async_loop.ready_head = task->next_ready;
    if (!narra_async_loop.ready_head) narra_async_loop.ready_tail = NULL;
    task->next_ready = NULL;
    task->queued = false;
    narra_task* previous = narra_loop_enter(task);
    task->resume(task);
    narra_loop_leave(previous);
  }
}

static void narra_task_release_nothing(narra_task* task) {
  (void)task;
}

static int64_t narra_loop_now(void) {
  struct timespec now;
  clock_gettime(CLOCK_MONOTONIC, &now);
  return (int64_t)now.tv_sec * 1000000000 + (int64_t)now.tv_nsec;
}

// The timer heap.

static bool narra_timer_before(narra_wait* first, narra_wait* second) {
  if (first->deadline != second->deadline) return first->deadline < second->deadline;
  return first->sequence < second->sequence;
}

static void narra_timer_place(int64_t slot, narra_wait* wait) {
  narra_async_loop.timers[slot] = wait;
  wait->slot = slot;
}

static void narra_timer_up(int64_t slot) {
  narra_wait* wait = narra_async_loop.timers[slot];
  while (slot > 0) {
    int64_t parent = (slot - 1) / 2;
    if (!narra_timer_before(wait, narra_async_loop.timers[parent])) break;
    narra_timer_place(slot, narra_async_loop.timers[parent]);
    slot = parent;
  }
  narra_timer_place(slot, wait);
}

static void narra_timer_down(int64_t slot) {
  narra_wait* wait = narra_async_loop.timers[slot];
  int64_t count = narra_async_loop.timer_count;
  while (true) {
    int64_t child = slot * 2 + 1;
    if (child >= count) break;
    if (child + 1 < count && narra_timer_before(narra_async_loop.timers[child + 1], narra_async_loop.timers[child])) {
      child++;
    }
    if (!narra_timer_before(narra_async_loop.timers[child], wait)) break;
    narra_timer_place(slot, narra_async_loop.timers[child]);
    slot = child;
  }
  narra_timer_place(slot, wait);
}

static void narra_timer_push(narra_wait* wait) {
  if (narra_async_loop.timer_count == narra_async_loop.timer_capacity) {
    int64_t capacity = narra_async_loop.timer_capacity ? narra_async_loop.timer_capacity * 2 : 16;
    narra_wait** timers = realloc(narra_async_loop.timers, (size_t)capacity * sizeof(narra_wait*));
    if (!timers) {
      fflush(stdout);
      fprintf(stderr, "async: out of memory for timers\n");
      exit(101);
    }
    narra_async_loop.timers = timers;
    narra_async_loop.timer_capacity = capacity;
  }
  int64_t slot = narra_async_loop.timer_count++;
  narra_timer_place(slot, wait);
  narra_timer_up(slot);
}

static void narra_timer_remove(narra_wait* wait) {
  int64_t slot = wait->slot;
  int64_t last = --narra_async_loop.timer_count;
  wait->slot = -1;
  if (slot == last) return;
  // The last timer fills the hole and moves up or down to its place.
  narra_wait* moved = narra_async_loop.timers[last];
  narra_timer_place(slot, moved);
  narra_timer_up(slot);
  narra_timer_down(moved->slot);
}

// The poller.

static void narra_poller_open(void) {
  if (narra_async_loop.poller_open) return;
#if defined(__APPLE__)
  narra_async_loop.poller = kqueue();
#else
  narra_async_loop.poller = epoll_create1(0);
#endif
  if (narra_async_loop.poller < 0) {
    fflush(stdout);
    fprintf(stderr, "async: cannot open the poller: %s\n", strerror(errno));
    exit(101);
  }
  narra_async_loop.poller_open = true;
}

static void narra_poller_close(void) {
  if (!narra_async_loop.poller_open) return;
  close(narra_async_loop.poller);
  narra_async_loop.poller_open = false;
}

// The wait of the io list on descriptor in one direction, or NULL.
static narra_wait* narra_io_find(int descriptor, bool writable) {
  for (narra_wait* wait = narra_async_loop.io; wait; wait = wait->next_io) {
    if (wait->descriptor == descriptor && wait->writable == writable) return wait;
  }
  return NULL;
}

static void narra_io_unlink(narra_wait* wait) {
  if (wait->previous_io) {
    wait->previous_io->next_io = wait->next_io;
  } else {
    narra_async_loop.io = wait->next_io;
  }
  if (wait->next_io) wait->next_io->previous_io = wait->previous_io;
  wait->next_io = NULL;
  wait->previous_io = NULL;
  narra_async_loop.io_count--;
}

#if defined(__linux__)
// epoll keeps one registration per descriptor: it watches the directions the
// io list still waits for, and is deleted once there is none. registered is
// true when the descriptor was in epoll before this change. 0, or the errno
// value.
static int narra_epoll_sync(int descriptor, bool registered) {
  uint32_t events = 0;
  if (narra_io_find(descriptor, false)) events |= EPOLLIN;
  if (narra_io_find(descriptor, true)) events |= EPOLLOUT;
  if (events == 0) {
    if (registered) epoll_ctl(narra_async_loop.poller, EPOLL_CTL_DEL, descriptor, NULL);
    return 0;
  }
  struct epoll_event event;
  memset(&event, 0, sizeof(event));
  event.events = events;
  event.data.fd = descriptor;
  int operation = registered ? EPOLL_CTL_MOD : EPOLL_CTL_ADD;
  if (epoll_ctl(narra_async_loop.poller, operation, descriptor, &event) < 0) return errno;
  return 0;
}
#endif

// Adds a wait, already in the io list, to the poller; 0, or the errno value.
static int narra_poller_add(narra_wait* wait) {
  narra_poller_open();
#if defined(__APPLE__)
  struct kevent change;
  EV_SET(&change, wait->descriptor, wait->writable ? EVFILT_WRITE : EVFILT_READ, EV_ADD | EV_ONESHOT, 0, 0, wait);
  if (kevent(narra_async_loop.poller, &change, 1, NULL, 0, NULL) < 0) return errno;
  return 0;
#else
  return narra_epoll_sync(wait->descriptor, narra_io_find(wait->descriptor, !wait->writable) != NULL);
#endif
}

// Takes a wait, already out of the io list, out of the poller. woke is true
// when the poller reported it, so kqueue has deleted its one-shot event.
static void narra_poller_remove(int descriptor, bool writable, bool woke) {
#if defined(__APPLE__)
  if (woke) return;
  struct kevent change;
  EV_SET(&change, descriptor, writable ? EVFILT_WRITE : EVFILT_READ, EV_DELETE, 0, 0, NULL);
  kevent(narra_async_loop.poller, &change, 1, NULL, 0, NULL);
#else
  (void)writable;
  (void)woke;
  narra_epoll_sync(descriptor, true);
#endif
}

// Wait handles.

static void narra_wait_release(narra_task* task) {
  (void)task;
}

static narra_wait* narra_wait_new(void) {
  narra_wait* wait = narra_alloc(sizeof(narra_wait));
  memset(wait, 0, sizeof(narra_wait));
  narra_task_init(&wait->task, NULL, narra_wait_release, &wait->value);
  wait->started = narra_loop_now();
  wait->slot = -1;
  wait->descriptor = -1;
  return wait;
}

static bool narra_wait_is(narra_task* task) {
  return task->release == narra_wait_release;
}

// Marks the running task (an Async.all block) so a cancel of it stops there.
static void narra_loop_keep_inputs(void) {
  if (narra_async_loop.current) narra_async_loop.current->keeps_inputs = true;
}

// True when a suspending std call of the running task must not wait: the task,
// or the task that started it, is cancelled, or the program exits.
static bool narra_loop_cancelled_now(void) {
  narra_task* current = narra_async_loop.current;
  if (narra_async_loop.exiting) return true;
  return current != NULL && (current->cancelled || current->started_cancelled);
}

// Takes a wait out of the heap or the poller and finishes its handle with
// value. The handle's waiters become ready. woke is true when the poller
// reported the descriptor.
static void narra_wait_end(narra_wait* wait, int64_t value, bool woke) {
  if (wait->slot >= 0) narra_timer_remove(wait);
  if (wait->descriptor >= 0) {
    int descriptor = wait->descriptor;
    narra_io_unlink(wait);
    wait->descriptor = -1;
    narra_poller_remove(descriptor, wait->writable, woke);
  }
  wait->value = value;
  narra_task_finish(&wait->task);
}

static void narra_wait_finish(narra_wait* wait, int64_t value) {
  narra_wait_end(wait, value, false);
}

// The handle of a timer that fires after nanos (at once for zero or less,
// after the ready tasks have run). Its value is the time slept, or -1 when
// the running task is cancelled or is cancelled while it waits. A deadline
// past the clock's range is the end of the range: it waits until a cancel.
static narra_task* narra_async_sleep(int64_t nanos) {
  narra_task* current = narra_async_loop.current;
  if (current) current->forwards_cancel = true;
  narra_wait* wait = narra_wait_new();
  narra_task* handle = &wait->task;
  if (narra_loop_cancelled_now()) {
    narra_wait_finish(wait, -1);
    return handle;
  }
  if (nanos <= 0) {
    wait->deadline = wait->started;
  } else if (nanos > INT64_MAX - wait->started) {
    wait->deadline = INT64_MAX;
  } else {
    wait->deadline = wait->started + nanos;
  }
  wait->sequence = narra_async_loop.timer_sequence++;
  narra_timer_push(wait);
  return handle;
}

// The handle of a wait until descriptor is ready to read, or to write when
// writable is set. Its value is 0 when the descriptor is ready, -1 on a
// cancel, or the errno value when the poller refuses the descriptor. One
// reader and one writer may wait on a descriptor at once; a second wait in
// the same direction gives EEXIST on every platform.
static narra_task* narra_loop_wait_io(int descriptor, bool writable) {
  narra_task* current = narra_async_loop.current;
  if (current) current->forwards_cancel = true;
  narra_wait* wait = narra_wait_new();
  narra_task* handle = &wait->task;
  if (narra_loop_cancelled_now()) {
    narra_wait_finish(wait, -1);
    return handle;
  }
  if (narra_io_find(descriptor, writable)) {
    narra_wait_finish(wait, EEXIST);
    return handle;
  }
  wait->writable = writable;
  wait->descriptor = descriptor;
  wait->next_io = narra_async_loop.io;
  if (narra_async_loop.io) narra_async_loop.io->previous_io = wait;
  narra_async_loop.io = wait;
  narra_async_loop.io_count++;
  int failure = narra_poller_add(wait);
  if (failure != 0) {
    narra_io_unlink(wait);
    wait->descriptor = -1;
    narra_wait_finish(wait, failure);
  }
  return handle;
}

static void narra_loop_cancel_idle(narra_task* idle);

// A cancel of task passes down the await chain (D109 item 8, amended
// 2026-11-13) to the wait at its end: an Async.idle handle, which finishes,
// or a timer or descriptor wait, which finishes with -1. The task of a
// suspending std function passes it on. A user task it awaits is cancelled
// too when the waiter is the only task waiting on it and it has no then
// callbacks; a handle that other tasks await or that has callbacks keeps
// running, and so do the tasks a cancelled task started but does not await.
// The task of Async.all stops the walk, as the winner of Async.race does, so
// neither cancels the handles it was given. Async.cancel calls this after it
// sets the cancelled flag, and so does an await in a task that is already
// cancelled.
static void narra_loop_cancel_wait(narra_task* task) {
  narra_task* waiter = task;
  narra_task* awaited = task->awaited;
  while (awaited && !awaited->done && !waiter->keeps_inputs) {
    if (awaited->release == narra_task_release_nothing) {
      narra_loop_cancel_idle(awaited);
      return;
    }
    if (narra_wait_is(awaited)) {
      narra_wait_finish((narra_wait*)awaited, -1);
      return;
    }
    if (!awaited->forwards_cancel) {
      // A bare task (a race winner) has no wait of its own to end.
      if (awaited->resume == NULL) return;
      bool alone = awaited->waiters == waiter && waiter->next_waiter == NULL;
      if (!alone || awaited->callbacks != NULL) return;
      if (!narra_task_mark_cancelled(awaited)) return;
    }
    waiter = awaited;
    awaited = awaited->awaited;
  }
}

// Finishes every timer and descriptor wait with -1 (at exit).
static void narra_loop_cancel_all_waits(void) {
  while (narra_async_loop.timer_count > 0) {
    narra_wait_finish(narra_async_loop.timers[0], -1);
  }
  while (narra_async_loop.io) {
    narra_wait_finish(narra_async_loop.io, -1);
  }
}

// The longest single wait in the poller, one day; the loop waits again after it.
#define NARRA_POLL_LIMIT ((int64_t)86400 * 1000000000)

// A poller error other than a signal is a broken loop: stop as a panic does
// rather than spin.
static void narra_poller_failed(void) {
  fflush(stdout);
  fprintf(stderr, "async: the poller failed: %s\n", strerror(errno));
  exit(101);
}

// Waits in the poller until the nearest deadline (at most a day) or a ready
// descriptor, then finishes the waits that are due. False when there is
// nothing to wait for.
static bool narra_loop_poll(void) {
  if (narra_async_loop.timer_count == 0 && narra_async_loop.io_count == 0) return false;
  narra_poller_open();
  int64_t timeout = NARRA_POLL_LIMIT;
  if (narra_async_loop.timer_count > 0) {
    int64_t deadline = narra_async_loop.timers[0]->deadline;
    int64_t now = narra_loop_now();
    int64_t left = deadline <= now ? 0 : deadline - now;
    if (left < timeout) timeout = left;
  }
  enum { NARRA_POLL_EVENTS = 64 };
  int count;
#if defined(__APPLE__)
  struct kevent events[NARRA_POLL_EVENTS];
  struct timespec limit;
  limit.tv_sec = (time_t)(timeout / 1000000000);
  limit.tv_nsec = (long)(timeout % 1000000000);
  count = kevent(narra_async_loop.poller, NULL, 0, events, NARRA_POLL_EVENTS, &limit);
#else
  struct epoll_event events[NARRA_POLL_EVENTS];
  // epoll waits in milliseconds: round up, so a timer is never woken early.
  int milliseconds = (int)((timeout + 999999) / 1000000);
  count = epoll_wait(narra_async_loop.poller, events, NARRA_POLL_EVENTS, milliseconds);
#endif
  if (count < 0) {
    if (errno != EINTR) narra_poller_failed();
    count = 0;
  }
  for (int i = 0; i < count; i++) {
#if defined(__APPLE__)
    narra_wait_end(events[i].udata, 0, true);
#else
    // One event per descriptor: wake the reader and the writer it reports.
    int descriptor = events[i].data.fd;
    uint32_t ready = events[i].events;
    uint32_t broken = EPOLLERR | EPOLLHUP;
    narra_wait* reader = narra_io_find(descriptor, false);
    narra_wait* writer = narra_io_find(descriptor, true);
    if (reader && (ready & (EPOLLIN | broken))) narra_wait_end(reader, 0, true);
    if (writer && (ready & (EPOLLOUT | broken))) narra_wait_end(writer, 0, true);
#endif
  }
  int64_t now = narra_loop_now();
  while (narra_async_loop.timer_count > 0 && narra_async_loop.timers[0]->deadline <= now) {
    narra_wait* wait = narra_async_loop.timers[0];
    narra_wait_finish(wait, now - wait->started);
  }
  return true;
}

// Async.idle: a handle that finishes once no other task can run: nothing is
// ready, no timer is set and no descriptor is waited on. Idle handles finish
// one at a time in the order they were made, so a task started by an earlier
// one counts for the later ones. In a cancelled task, and at exit, it
// finishes at once.
static narra_task* narra_async_idle(void) {
  narra_task* task = narra_alloc(sizeof(narra_task));
  memset(task, 0, sizeof(narra_task));
  narra_task_init(task, NULL, narra_task_release_nothing, NULL);
  narra_task* current = narra_async_loop.current;
  if (task->cancelled || (current && current->cancelled)) {
    narra_task_finish(task);
    return task;
  }
  if (narra_async_loop.idle_tail) {
    narra_async_loop.idle_tail->next_ready = task;
  } else {
    narra_async_loop.idle_head = task;
  }
  narra_async_loop.idle_tail = task;
  return task;
}

// Takes an idle handle out of the idle list and finishes it (a cancel of the
// task that waits in it).
static void narra_loop_cancel_idle(narra_task* idle) {
  narra_task* previous = NULL;
  for (narra_task* task = narra_async_loop.idle_head; task; task = task->next_ready) {
    if (task != idle) {
      previous = task;
      continue;
    }
    if (previous) {
      previous->next_ready = task->next_ready;
    } else {
      narra_async_loop.idle_head = task->next_ready;
    }
    if (narra_async_loop.idle_tail == task) narra_async_loop.idle_tail = previous;
    task->next_ready = NULL;
    narra_task_finish(task);
    return;
  }
}

// Finishes the oldest idle handle; false when there is none.
static bool narra_loop_wake_idle(void) {
  narra_task* task = narra_async_loop.idle_head;
  if (!task) return false;
  narra_async_loop.idle_head = task->next_ready;
  if (!narra_async_loop.idle_head) narra_async_loop.idle_tail = NULL;
  task->next_ready = NULL;
  narra_task_finish(task);
  return true;
}

// Runs the loop until the entry's handle is done and gives its value, the
// exit code. Nothing ready, no timer, no descriptor and no idle handle while
// the entry waits means the tasks wait on each other: the program stops as a
// panic does.
static int64_t narra_async_run_entry(narra_task* entry) {
  while (!entry->done) {
    narra_loop_run_ready();
    if (entry->done) break;
    if (narra_loop_poll()) continue;
    if (narra_loop_wake_idle()) continue;
    fflush(stdout);
    fprintf(stderr, "async: every task waits on another task\n");
    exit(101);
  }
  int64_t code = *(int64_t*)entry->value;
  narra_task_release(entry);
  return code;
}

// The exit of D95: every unfinished task is cancelled, every timer and
// descriptor wait finishes with -1, and the loop runs until nothing is ready.
// A task still waiting then waits on a handle that cannot finish; its frame
// releases what it holds, and the loop drops it. Releasing every frame before
// dropping any breaks the cycles through those frames. Last, the poller
// closes.
static void narra_async_exit(void) {
  narra_task_check_watchers(narra_async_loop.live);
  narra_async_loop.exiting = true;
  for (narra_task* task = narra_async_loop.live; task; task = task->next_live) {
    narra_task_mark_cancelled(task);
  }
  do {
    narra_loop_cancel_all_waits();
    narra_loop_run_ready();
  } while (narra_loop_wake_idle());
  for (narra_task* task = narra_async_loop.live; task; task = task->next_live) {
    task->release(task);
    task->dropped = true;
  }
  for (narra_task* task = narra_async_loop.live; task; task = task->next_live) {
    while (task->waiters) {
      narra_task* waiter = task->waiters;
      task->waiters = waiter->next_waiter;
      waiter->next_waiter = NULL;
      waiter->awaited = NULL;
      narra_task_release(waiter);
    }
  }
  while (narra_async_loop.live) {
    narra_task* task = narra_async_loop.live;
    narra_loop_remove(task);
    narra_task_release(task);
  }
  free(narra_async_loop.timers);
  narra_async_loop.timers = NULL;
  narra_async_loop.timer_capacity = 0;
  narra_poller_close();
}
