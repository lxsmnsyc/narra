// std/internal/async. Tasks and their handles (D109 items 1 and 2).
//
// A task frame starts with this header; the value slot and the body's
// variables follow it (emitasync.nra). The frame is the handle: Async<T> is a
// narra_task*, and copies of a handle share the frame.
//
// Who holds a reference: each copy of the handle, each task in a waiter list,
// and the loop until the task finishes. release releases what the frame holds
// at its state, or the value once the task is done; the frame itself is freed
// when the count reaches zero.
typedef struct narra_task narra_task;
typedef struct narra_task_callback narra_task_callback;
struct narra_task {
  int64_t rc;
  // The resume point: 0 before the first run, then the await it waits at.
  int64_t state;
  bool done;
  bool cancelled;
  // Set at exit once release has run for an unfinished frame.
  bool dropped;
  // Set when the task that started this one was cancelled then: a suspending
  // std call of this task returns Err(AsyncError.Cancelled) at once.
  bool started_cancelled;
  // Set on the task of a suspending std function (std/async/time) once it
  // waits in the loop: a cancel of the task awaiting it reaches its wait.
  bool forwards_cancel;
  // Set on the task of Async.all: a cancel of it stops there and does not
  // reach the handles it awaits.
  bool keeps_inputs;
  // The unfinished Async.all tasks that were given this handle. Like a then
  // callback, each observes the handle, so a cancel walk does not cancel it.
  int64_t observers;
  void (*resume)(narra_task* task);
  void (*release)(narra_task* task);
  // The value slot of the frame, or NULL for an Async<void>.
  void* value;
  // The handle this task waits on, the next task waiting on that handle,
  // and the first task waiting on this one.
  narra_task* awaited;
  narra_task* next_waiter;
  narra_task* waiters;
  // The ready queue and the loop's list of unfinished tasks.
  narra_task* next_ready;
  bool queued;
  narra_task* next_live;
  narra_task* previous_live;
  // The Async.then callbacks, in the order they were registered.
  narra_task_callback* callbacks;
  narra_task_callback* last_callback;
};

// An Async.then callback: a closure that takes the finished handle. Each
// callback owns a reference to its environment, which release drops. The
// closure runtime may come later in the C file, so only the struct tag is named.
struct narra_env;
struct narra_task_callback {
  void (*fn)(struct narra_env* env, narra_task* task);
  struct narra_env* env;
  void (*release)(struct narra_env* env);
  narra_task_callback* next;
};

static void narra_loop_add(narra_task* task);
static void narra_loop_remove(narra_task* task);
static void narra_loop_queue(narra_task* task);
static narra_task* narra_loop_enter(narra_task* task);
static void narra_loop_leave(narra_task* previous);
static void narra_task_finish(narra_task* task);
static void narra_loop_cancel_wait(narra_task* task);

static void narra_task_free(void* pointer) {
  narra_task* task = pointer;
  if (!task->dropped) task->release(task);
  narra_free(task);
}

static void narra_task_retain(narra_task* task) {
  task->rc++;
}

static void narra_task_release(narra_task* task) {
  if (--task->rc == 0) narra_drop(task, narra_task_free);
}

// Fills the header of a new frame: one reference for the handle the creation
// returns and one for the loop until the task finishes.
static void narra_task_init(
  narra_task* task,
  void (*resume)(narra_task*),
  void (*release)(narra_task*),
  void* value
) {
  task->rc = 2;
  task->state = 0;
  task->resume = resume;
  task->release = release;
  task->value = value;
  narra_loop_add(task);
}

// Runs the body of a new task until it first waits or returns, as a
// TypeScript async function does, and gives the handle.
static narra_task* narra_task_start(narra_task* task) {
  narra_task* previous = narra_loop_enter(task);
  task->resume(task);
  narra_loop_leave(previous);
  return task;
}

// The running task waits until handle finishes. The waiter list holds a
// reference to it. A task cancelled before it waits passes the cancel down
// the chain now, as Async.cancel does for a task already waiting.
static void narra_task_wait(narra_task* task, narra_task* handle) {
  narra_task_retain(task);
  task->awaited = handle;
  task->next_waiter = handle->waiters;
  handle->waiters = task;
  if (task->cancelled) narra_loop_cancel_wait(task);
}

// Takes the callbacks off the task and gives them, so a callback that
// registers or cancels cannot change the list being walked.
static narra_task_callback* narra_task_take_callbacks(narra_task* task) {
  narra_task_callback* list = task->callbacks;
  task->callbacks = NULL;
  task->last_callback = NULL;
  return list;
}

// Runs each callback with the task when run is true, then frees it and
// releases its environment.
static void narra_task_run_callbacks(narra_task* task, narra_task_callback* list, bool run) {
  while (list) {
    narra_task_callback* callback = list;
    list = callback->next;
    if (run) callback->fn(callback->env, task);
    callback->release(callback->env);
    narra_free(callback);
  }
}

// A callback registered on a finished handle runs on the next loop turn, as
// in TypeScript: a small task in the ready queue holds it and the handle.
typedef struct narra_task_later {
  narra_task h;
  narra_task* handle;
  narra_task_callback* callback;
} narra_task_later;

static void narra_task_later_resume(narra_task* task) {
  narra_task_later* later = (narra_task_later*)task;
  narra_task_callback* callback = later->callback;
  later->callback = NULL;
  narra_task_run_callbacks(later->handle, callback, true);
  narra_task_finish(task);
}

static void narra_task_later_release(narra_task* task) {
  narra_task_later* later = (narra_task_later*)task;
  narra_task_run_callbacks(later->handle, later->callback, false);
  later->callback = NULL;
  narra_task_release(later->handle);
}

// Async.then: a cancelled handle drops the callback, a finished one queues it
// for the next loop turn, and an unfinished one keeps it until it finishes.
static void narra_task_then(
  narra_task* task,
  void (*fn)(struct narra_env*, narra_task*),
  struct narra_env* env,
  void (*release)(struct narra_env*)
) {
  narra_task_callback* callback = narra_alloc(sizeof(narra_task_callback));
  callback->fn = fn;
  callback->env = env;
  callback->release = release;
  callback->next = NULL;
  if (task->cancelled) {
    narra_task_run_callbacks(task, callback, false);
    return;
  }
  if (task->done) {
    narra_task_later* later = narra_alloc(sizeof(narra_task_later));
    memset(later, 0, sizeof(narra_task_later));
    narra_task_init(&later->h, narra_task_later_resume, narra_task_later_release, NULL);
    narra_task_retain(task);
    later->handle = task;
    later->callback = callback;
    narra_loop_queue(&later->h);
    // Nobody holds the handle of this task: the loop does until it runs.
    narra_task_release(&later->h);
    return;
  }
  if (task->last_callback) {
    task->last_callback->next = callback;
  } else {
    task->callbacks = callback;
  }
  task->last_callback = callback;
}

// Marks an unfinished task cancelled and drops its callbacks. Exit calls this
// for every task, then finishes the waits itself in timer order.
static bool narra_task_mark_cancelled(narra_task* task) {
  if (task->done || task->cancelled) return false;
  task->cancelled = true;
  narra_task_run_callbacks(task, narra_task_take_callbacks(task), false);
  return true;
}

// Async.cancel: marks the task and wakes its timer, descriptor or idle wait,
// which then gives Err(AsyncError.Cancelled). A wait on another task's handle
// goes on (D109 item 8).
static void narra_task_cancel(narra_task* task) {
  if (narra_task_mark_cancelled(task)) narra_loop_cancel_wait(task);
}

// Async.all takes each handle it was given as observed until it has awaited it.
static void narra_task_observe(narra_task* task) {
  task->observers++;
}

static void narra_task_unobserve(narra_task* task) {
  task->observers--;
}

// Whether Async.peek gives the value: the task finished and was not cancelled.
static bool narra_task_finished(narra_task* task) {
  return task->done && !task->cancelled;
}

// The body has returned: it already released its locals, and the value is in
// its slot. The waiters become ready, the callbacks run unless the task was
// cancelled, and the loop drops its reference, which may free the frame, so
// nothing reads the task after this call.
static void narra_task_finish(narra_task* task) {
  task->done = true;
  task->state = -1;
  // The list is built by pushing at the front, so it is reversed to wake the
  // waiters in the order they started to wait.
  narra_task* reversed = NULL;
  while (task->waiters) {
    narra_task* waiter = task->waiters;
    task->waiters = waiter->next_waiter;
    waiter->next_waiter = reversed;
    reversed = waiter;
  }
  while (reversed) {
    narra_task* waiter = reversed;
    reversed = waiter->next_waiter;
    waiter->next_waiter = NULL;
    waiter->awaited = NULL;
    narra_loop_queue(waiter);
    narra_task_release(waiter);
  }
  narra_task_run_callbacks(task, narra_task_take_callbacks(task), !task->cancelled);
  narra_loop_remove(task);
  narra_task_release(task);
}

// Async.race: a bare task whose value is the index of the first handle to
// finish. A watcher task waits on each unfinished handle and sets the index
// if no other handle came first. Once the index is set, the other watchers
// leave their handles' waiter lists and are freed, so a race decided early
// leaves nothing behind on a handle that stays pending.
typedef struct narra_task_watcher narra_task_watcher;

typedef struct narra_task_winner {
  narra_task h;
  int64_t value;
  // The watchers still in the race; this list holds no reference.
  narra_task_watcher* watchers;
} narra_task_winner;

struct narra_task_watcher {
  narra_task h;
  narra_task* winner;
  int64_t index;
  narra_task_watcher* next_watcher;
  narra_task_watcher* previous_watcher;
};

// Not narra_task_release_nothing: the loop takes a task with that release for
// an idle handle, and a cancel must not finish a winner.
static void narra_task_winner_release(narra_task* task) {
  (void)task;
}

static narra_task* narra_task_winner_new(void) {
  narra_task_winner* winner = narra_alloc(sizeof(narra_task_winner));
  memset(winner, 0, sizeof(narra_task_winner));
  narra_task_init(&winner->h, NULL, narra_task_winner_release, &winner->value);
  return &winner->h;
}

// Takes a watcher off its winner's list.
static void narra_task_unlist_watcher(narra_task_watcher* watcher) {
  narra_task_winner* winner = (narra_task_winner*)watcher->winner;
  if (watcher->previous_watcher) {
    watcher->previous_watcher->next_watcher = watcher->next_watcher;
  } else if (winner->watchers == watcher) {
    winner->watchers = watcher->next_watcher;
  }
  if (watcher->next_watcher) watcher->next_watcher->previous_watcher = watcher->previous_watcher;
  watcher->next_watcher = NULL;
  watcher->previous_watcher = NULL;
}

// Ends a watcher that lost: it leaves the waiter list of its handle and the
// loop, and both drop their references, which frees it.
static void narra_task_drop_watcher(narra_task_watcher* watcher) {
  narra_task* task = &watcher->h;
  narra_task* handle = task->awaited;
  narra_task** link = &handle->waiters;
  while (*link && *link != task) link = &(*link)->next_waiter;
  if (*link) *link = task->next_waiter;
  task->awaited = NULL;
  task->next_waiter = NULL;
  task->done = true;
  task->state = -1;
  narra_loop_remove(task);
  narra_task_release(task);
  narra_task_release(task);
}

static void narra_task_settle(narra_task* winner, int64_t index) {
  if (winner->done) return;
  *(int64_t*)winner->value = index;
  // The winner stays alive while the watchers that hold it go.
  narra_task_retain(winner);
  narra_task_winner* frame = (narra_task_winner*)winner;
  while (frame->watchers) {
    narra_task_watcher* watcher = frame->watchers;
    narra_task_unlist_watcher(watcher);
    // A watcher already queued ends when it resumes.
    if (watcher->h.awaited) narra_task_drop_watcher(watcher);
  }
  narra_task_finish(winner);
  narra_task_release(winner);
}

static void narra_task_watcher_resume(narra_task* task) {
  narra_task_watcher* watcher = (narra_task_watcher*)task;
  narra_task_unlist_watcher(watcher);
  narra_task_settle(watcher->winner, watcher->index);
  narra_task_finish(task);
}

static void narra_task_watcher_release(narra_task* task) {
  narra_task_watcher* watcher = (narra_task_watcher*)task;
  narra_task_unlist_watcher(watcher);
  narra_task_release(watcher->winner);
}

static void narra_task_watch(narra_task* winner, narra_task* handle, int64_t index) {
  if (winner->done) return;
  if (handle->done) {
    narra_task_settle(winner, index);
    return;
  }
  narra_task_watcher* watcher = narra_alloc(sizeof(narra_task_watcher));
  memset(watcher, 0, sizeof(narra_task_watcher));
  narra_task_init(&watcher->h, narra_task_watcher_resume, narra_task_watcher_release, NULL);
  narra_task_retain(winner);
  watcher->winner = winner;
  watcher->index = index;
  narra_task_winner* frame = (narra_task_winner*)winner;
  watcher->next_watcher = frame->watchers;
  if (frame->watchers) frame->watchers->previous_watcher = watcher;
  frame->watchers = watcher;
  narra_task_wait(&watcher->h, handle);
  // Nobody holds the watcher's handle: the loop and the waiter list do.
  narra_task_release(&watcher->h);
}

// Part of the leak check, called at exit with the unfinished tasks: a watcher
// whose race is decided must have left the loop. Exit would free it, so only
// this check sees one that stayed.
static void narra_task_check_watchers(narra_task* live) {
  if (!getenv("NARRA_CHECK_LEAKS")) return;
  int64_t stale = 0;
  for (narra_task* task = live; task; task = task->next_live) {
    if (task->resume == narra_task_watcher_resume) {
      narra_task_watcher* watcher = (narra_task_watcher*)task;
      if (watcher->winner->done) stale++;
    }
  }
  if (stale != 0) {
    fflush(stdout);
    fprintf(stderr, "leak check: %" PRId64 " race watchers outlived their race\n", stale);
    exit(102);
  }
}
