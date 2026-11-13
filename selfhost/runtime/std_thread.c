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
