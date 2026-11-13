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
};

static void narra_loop_add(narra_task* task);
static void narra_loop_remove(narra_task* task);
static void narra_loop_queue(narra_task* task);
static narra_task* narra_loop_enter(narra_task* task);
static void narra_loop_leave(narra_task* previous);

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
// reference to it.
static void narra_task_wait(narra_task* task, narra_task* handle) {
  narra_task_retain(task);
  task->awaited = handle;
  task->next_waiter = handle->waiters;
  handle->waiters = task;
}

// The body has returned: it already released its locals, and the value is in
// its slot. The waiters become ready, and the loop drops its reference, which
// may free the frame, so nothing reads the task after this call.
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
  narra_loop_remove(task);
  narra_task_release(task);
}
