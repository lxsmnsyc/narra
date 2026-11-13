// std/internal/async. The loop of a thread (D109 items 6 and 7): a ready queue
// of tasks to resume, the unfinished tasks, and the tasks waiting in
// Async.idle. There is no I/O and no timer yet.
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
} narra_loop;

static _Thread_local narra_loop narra_async_loop;

static void narra_loop_add(narra_task* task) {
  task->cancelled = narra_async_loop.exiting;
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

// Async.idle: a handle that finishes once no task is ready. In a cancelled
// task, and at exit, it finishes at once.
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

// Finishes the idle handles in the order they were made; false when there is none.
static bool narra_loop_wake_idle(void) {
  narra_task* list = narra_async_loop.idle_head;
  narra_async_loop.idle_head = NULL;
  narra_async_loop.idle_tail = NULL;
  if (!list) return false;
  while (list) {
    narra_task* next = list->next_ready;
    list->next_ready = NULL;
    narra_task_finish(list);
    list = next;
  }
  return true;
}

// Runs the loop until the entry's handle is done and gives its value, the
// exit code. Nothing ready and nothing to wait for while the entry waits
// means the tasks wait on each other: the program stops as a panic does.
static int64_t narra_async_run_entry(narra_task* entry) {
  while (!entry->done) {
    narra_loop_run_ready();
    if (entry->done) break;
    if (narra_loop_wake_idle()) continue;
    fflush(stdout);
    fprintf(stderr, "async: every task waits on another task\n");
    exit(101);
  }
  int64_t code = *(int64_t*)entry->value;
  narra_task_release(entry);
  return code;
}

// The exit of D95: every unfinished task is cancelled and the loop runs until
// nothing is ready. A task still waiting then waits on a handle that cannot
// finish; its frame releases what it holds, and the loop drops it. Releasing
// every frame before dropping any breaks the cycles through those frames.
static void narra_async_exit(void) {
  narra_async_loop.exiting = true;
  for (narra_task* task = narra_async_loop.live; task; task = task->next_live) {
    task->cancelled = true;
  }
  do {
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
}
