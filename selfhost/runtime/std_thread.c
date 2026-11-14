// std/internal/thread (D111 item 4). POSIX threads for std/thread; only a
// program in thread mode (NARRA_THREADS) includes this text.
//
// narra_thread_start marks the captures of the thread function shared, then
// runs the function on a new thread with an 8 MiB stack, the size of the main
// thread's. The new thread takes a heap, calls the function, runs the exit of
// its event loop (D109 item 6) when it started tasks, releases the
// environment and gives the heap back. The thread is a number to Narra code:
// its pthread_t, which narra_thread_join waits for.
struct narra_env;

typedef struct narra_thread_work {
  void (*fn)(struct narra_env* env);
  struct narra_env* env;
  void (*release)(struct narra_env* env);
} narra_thread_work;

static void* narra_thread_main(void* argument) {
  narra_thread_work work = *(narra_thread_work*)argument;
  free(argument);
  narra_heap_enter();
  work.fn(work.env);
  void (*loop_exit)(void) = __atomic_load_n(&narra_loop_exit_hook, __ATOMIC_ACQUIRE);
  if (loop_exit) loop_exit();
  work.release(work.env);
  // The drop and share worklists of this thread are plain memory that ends with it.
  free(narra_pend);
  narra_pend = NULL;
  narra_pend_cap = 0;
  free(narra_shares);
  narra_shares = NULL;
  narra_shares_cap = 0;
  narra_heap_leave();
  __atomic_fetch_sub(&narra_threads_running, 1, __ATOMIC_RELEASE);
  return NULL;
}

static int64_t narra_thread_start(
  void (*fn)(struct narra_env*),
  struct narra_env* env,
  void (*share)(struct narra_env*),
  void (*release)(struct narra_env*)
) {
  share(env);
  // Plain malloc: the new thread frees it before it allocates.
  narra_thread_work* work = malloc(sizeof(narra_thread_work));
  if (!work) narra_panic("runtime", "out of memory");
  work->fn = fn;
  work->env = env;
  work->release = release;
  pthread_attr_t attributes;
  pthread_attr_init(&attributes);
  pthread_attr_setstacksize(&attributes, (size_t)8 << 20);
  __atomic_fetch_add(&narra_threads_running, 1, __ATOMIC_ACQ_REL);
  pthread_t thread;
  int failed = pthread_create(&thread, &attributes, narra_thread_main, work);
  pthread_attr_destroy(&attributes);
  if (failed) {
    __atomic_fetch_sub(&narra_threads_running, 1, __ATOMIC_ACQ_REL);
    narra_panic("runtime", "cannot start a thread");
  }
  return (int64_t)(uintptr_t)thread;
}

static void narra_thread_join(int64_t thread) {
  pthread_join((pthread_t)(uintptr_t)thread, NULL);
}

// The cores online, for std/parallel. macOS hides the sysconf name under
// _POSIX_C_SOURCE; its value there is 58.
#if !defined(_SC_NPROCESSORS_ONLN) && defined(__APPLE__)
#define _SC_NPROCESSORS_ONLN 58
#endif
static int64_t narra_thread_cores(void) {
#ifdef _SC_NPROCESSORS_ONLN
  long count = sysconf(_SC_NPROCESSORS_ONLN);
  return count < 1 ? 1 : (int64_t)count;
#else
  return 1;
#endif
}

// Sync (D111 item 4): a lock with two condition variables, for std/channel
// and std/shared. Its count starts shared (-1), so copies on any thread retain
// and release it atomically and the share walk stops at it. A Sync is a core
// (narra_sync_new: the lock, the conditions and two end counts) or an end
// (narra_sync_end: only the header, pointing at a core). An end adds one to
// the core's end count of its kind, and when its last copy is dropped it takes
// that one away again and wakes every waiter of the core, so a channel sees
// its last sender or its last receiver go. The lock functions take a core.
typedef struct narra_sync {
  int64_t rc;
  struct narra_sync_core* core;
  int64_t kind;
} narra_sync;

typedef struct narra_sync_core {
  narra_sync h;
  int64_t ends[2];
  pthread_mutex_t lock;
  pthread_cond_t ready[2];
} narra_sync_core;

static narra_sync* narra_sync_new(void) {
  narra_sync_core* core = narra_alloc(sizeof(narra_sync_core));
  core->h.rc = -1;
  core->h.core = NULL;
  core->h.kind = 0;
  core->ends[0] = 0;
  core->ends[1] = 0;
  pthread_mutex_init(&core->lock, NULL);
  pthread_cond_init(&core->ready[0], NULL);
  pthread_cond_init(&core->ready[1], NULL);
  return &core->h;
}

static void narra_sync_retain(narra_sync* sync) { NARRA_RC_RETAIN(sync->rc); }

static void narra_sync_share(narra_sync* sync) { (void)sync; }

static void narra_sync_lock(narra_sync* sync) {
  pthread_mutex_lock(&((narra_sync_core*)sync)->lock);
}

static void narra_sync_unlock(narra_sync* sync) {
  pthread_mutex_unlock(&((narra_sync_core*)sync)->lock);
}

static void narra_sync_wait(narra_sync* sync, int64_t condition) {
  narra_sync_core* core = (narra_sync_core*)sync;
  pthread_cond_wait(&core->ready[condition & 1], &core->lock);
}

static void narra_sync_wake_one(narra_sync* sync, int64_t condition) {
  pthread_cond_signal(&((narra_sync_core*)sync)->ready[condition & 1]);
}

static void narra_sync_wake_all(narra_sync* sync, int64_t condition) {
  pthread_cond_broadcast(&((narra_sync_core*)sync)->ready[condition & 1]);
}

static int64_t narra_sync_ends(narra_sync* sync, int64_t kind) {
  return ((narra_sync_core*)sync)->ends[kind & 1];
}

static narra_sync* narra_sync_end(narra_sync* sync, int64_t kind) {
  narra_sync_core* core = (narra_sync_core*)sync;
  narra_sync* end = narra_alloc(sizeof(narra_sync));
  end->rc = -1;
  narra_sync_retain(sync);
  end->core = core;
  end->kind = kind & 1;
  pthread_mutex_lock(&core->lock);
  core->ends[end->kind]++;
  pthread_mutex_unlock(&core->lock);
  return end;
}

static void narra_sync_release(narra_sync* sync) {
  if (!NARRA_RC_RELEASE(sync->rc)) return;
  narra_sync_core* core = sync->core;
  if (!core) {
    core = (narra_sync_core*)sync;
    pthread_cond_destroy(&core->ready[0]);
    pthread_cond_destroy(&core->ready[1]);
    pthread_mutex_destroy(&core->lock);
    narra_free(core);
    return;
  }
  pthread_mutex_lock(&core->lock);
  core->ends[sync->kind]--;
  pthread_cond_broadcast(&core->ready[0]);
  pthread_cond_broadcast(&core->ready[1]);
  pthread_mutex_unlock(&core->lock);
  narra_free(sync);
  narra_sync_release(&core->h);
}
