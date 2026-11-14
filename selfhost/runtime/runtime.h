// Runtime for programs compiled by narra. Included at the top of every
// generated C file.
// -std=c11 hides POSIX names (clock_gettime, kill, ...) on glibc.
#ifndef _POSIX_C_SOURCE
#define _POSIX_C_SOURCE 200809L
#endif
#include <inttypes.h>
#include <math.h>
#include <setjmp.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// Thread mode (D111): a program that imports std/thread is emitted with
// NARRA_THREADS defined. Shared objects then carry negative counts that change
// atomically, each thread allocates from a heap of its own, and the state
// below that a thread changes for itself is per thread. Without it the runtime
// is the single-threaded one.
#ifdef NARRA_THREADS
#include <pthread.h>
#define NARRA_PER_THREAD _Thread_local
#else
#define NARRA_PER_THREAD
#endif

// Test builds only (std/test): in a child process that runs a Test.panics
// closure, a panic writes its message to this pipe and exits 101 at once.
static int narra_panic_fd = -1;
// Set by the test runner's child: the pipe then gets "<where>\n<message>".
static int narra_panic_pos = 0;
// Where Test.golden reads and writes files; narra test sets it with --golden-dir.
static const char* narra_golden_dir = "testdata";

// Test builds: while Co.resume runs a coroutine it sets this, so a panic in
// the coroutine returns to Co.resume as Panicked (std/coroutine).
static NARRA_PER_THREAD jmp_buf* narra_panic_jmp = NULL;
static NARRA_PER_THREAD char narra_panic_text[256];

// Failure and copy paths stay out of line so the hot paths inline. The
// attributes are GNU extensions; other compilers get plain C11.
#if defined(__GNUC__)
#define NARRA_COLD __attribute__((cold, noinline))
#define NARRA_NOINLINE __attribute__((noinline))
#define NARRA_UNLIKELY(condition) __builtin_expect((condition) != 0, 0)
#define NARRA_LIKELY(condition) __builtin_expect((condition) != 0, 1)
#else
#define NARRA_COLD
#define NARRA_NOINLINE
#define NARRA_UNLIKELY(condition) (condition)
#define NARRA_LIKELY(condition) (condition)
#endif

#ifdef NARRA_THREADS
// The threads started and not yet ended. While it is 0 only the main thread
// runs, which then uses the static heap without reading a thread-local.
static int64_t narra_threads_running = 0;
// The exit of a thread's event loop (D109 item 6). std_async_loop.c sets it
// when a task starts, so a thread that ends runs it.
static void (*narra_loop_exit_hook)(void) = NULL;
// A panic holds this lock until the process ends, so one message is written.
static pthread_mutex_t narra_panic_lock = PTHREAD_MUTEX_INITIALIZER;

// The first read of a constant builds it under this lock (D61). Building one
// constant may read another, so the thread that holds the lock may take it again.
static pthread_mutex_t narra_const_mutex = PTHREAD_MUTEX_INITIALIZER;
static NARRA_PER_THREAD int narra_const_depth = 0;
static void narra_const_lock(void) {
  if (narra_const_depth++ == 0) pthread_mutex_lock(&narra_const_mutex);
}
static void narra_const_unlock(void) {
  if (--narra_const_depth == 0) pthread_mutex_unlock(&narra_const_mutex);
}
#endif

_Noreturn NARRA_COLD static void narra_panic(const char* where, const char* msg) {
#ifdef NARRA_THREADS
  if (!narra_panic_jmp) pthread_mutex_lock(&narra_panic_lock);
#endif
  fflush(stdout);
  if (narra_panic_jmp) {
    strncpy(narra_panic_text, msg, sizeof narra_panic_text - 1);
    narra_panic_text[sizeof narra_panic_text - 1] = 0;
    longjmp(*narra_panic_jmp, 1);
  }
  if (narra_panic_fd >= 0) {
    if (narra_panic_pos) {
      if (write(narra_panic_fd, where, strlen(where)) < 0 || write(narra_panic_fd, "\n", 1) < 0) _exit(101);
    }
    size_t n = strlen(msg);
    while (n > 0) {
      ssize_t w = write(narra_panic_fd, msg, n);
      if (w <= 0) break;
      msg += w;
      n -= (size_t)w;
    }
    _exit(101);
  }
  fprintf(stderr, "panic at %s: %s\n", where, msg);
#ifdef NARRA_THREADS
  // No exit handler may run beside threads that still run.
  if (__atomic_load_n(&narra_threads_running, __ATOMIC_ACQUIRE) > 0) {
    fflush(stderr);
    _exit(101);
  }
#endif
  exit(101);
}

// AddressSanitizer builds keep plain malloc and free, so the sanitizer sees
// every block. Defining NARRA_PLAIN_MALLOC does the same in any build.
#if defined(__SANITIZE_ADDRESS__)
#define NARRA_PLAIN_MALLOC 1
#elif defined(__has_feature)
#if __has_feature(address_sanitizer)
#define NARRA_PLAIN_MALLOC 1
#endif
#endif

#ifndef NARRA_THREADS
// Live allocation count. With NARRA_CHECK_LEAKS set, the program reports
// allocations still live at exit, which shows reference counting bugs.
static int64_t narra_live = 0;

static int64_t narra_live_total(void) { return narra_live; }
#endif

#if !defined(NARRA_PLAIN_MALLOC) || defined(NARRA_THREADS)
// Blocks of up to 512 bytes, header included, come from size-class free
// lists in steps of 16 bytes, carved from 1 MiB chunks. Larger blocks use
// malloc. An 8-byte header in front of each block holds its class, 0 for a
// malloc block. Chunks are never returned to the system.
//
// In thread mode each thread allocates from a heap of its own (D111 item 3).
// The header keeps the class in its low byte and the number of the owning
// heap above it; the static heap, which the main thread uses, is number 0. A
// free from another thread pushes the block onto the owner's remote stack
// with one compare-and-swap, and the owner takes the whole stack with one
// exchange when a size class runs empty. A heap whose thread has ended goes
// to the next thread that starts. live counts the allocations minus the frees
// done on the heap's thread; the leak check adds them up.
#define NARRA_SMALL_MAX 512
#define NARRA_CHUNK_SIZE ((size_t)1 << 20)
typedef struct narra_heap {
#ifndef NARRA_PLAIN_MALLOC
  uint64_t* free_lists[NARRA_SMALL_MAX / 16 + 1];
  char* at;
  char* end;
  void* chunks; // linked through each chunk's first word, so they stay reachable
#endif
#ifdef NARRA_THREADS
  int64_t live;
  uint64_t number;
  uint64_t* remote;
  struct narra_heap* next_idle;
#endif
} narra_heap;
static narra_heap narra_small;
#endif

#ifdef NARRA_THREADS
#define NARRA_HEAP_MAX 4096
static NARRA_PER_THREAD narra_heap* narra_heap_mine = &narra_small;
static narra_heap* narra_heaps[NARRA_HEAP_MAX] = {&narra_small};
static uint64_t narra_heap_count = 1;
static narra_heap* narra_heap_idle = NULL;
static pthread_mutex_t narra_heap_lock = PTHREAD_MUTEX_INITIALIZER;

// The heap of the calling thread. A plain load decides while no thread runs.
static inline narra_heap* narra_heap_here(void) {
  if (NARRA_LIKELY(__atomic_load_n(&narra_threads_running, __ATOMIC_RELAXED) == 0)) return &narra_small;
  return narra_heap_mine;
}

#define narra_live (narra_heap_here()->live)

static int64_t narra_live_total(void) {
  int64_t total = 0;
  pthread_mutex_lock(&narra_heap_lock);
  for (uint64_t i = 0; i < narra_heap_count; i++) total += narra_heaps[i]->live;
  pthread_mutex_unlock(&narra_heap_lock);
  return total;
}

// Gives a started thread a heap: one whose thread has ended, or a new one.
static void narra_heap_enter(void) {
  pthread_mutex_lock(&narra_heap_lock);
  narra_heap* heap = narra_heap_idle;
  if (heap) {
    narra_heap_idle = heap->next_idle;
  } else {
    if (narra_heap_count == NARRA_HEAP_MAX) {
      pthread_mutex_unlock(&narra_heap_lock);
      narra_panic("runtime", "too many threads at once");
    }
    // Plain calloc: a heap is never freed and stays reachable through narra_heaps.
    heap = calloc(1, sizeof(narra_heap));
    if (!heap) {
      pthread_mutex_unlock(&narra_heap_lock);
      narra_panic("runtime", "out of memory");
    }
    heap->number = narra_heap_count;
    narra_heaps[narra_heap_count++] = heap;
  }
  pthread_mutex_unlock(&narra_heap_lock);
  narra_heap_mine = heap;
}

// Hands the heap of an ending thread to the next thread that starts.
static void narra_heap_leave(void) {
  narra_heap* heap = narra_heap_mine;
  pthread_mutex_lock(&narra_heap_lock);
  heap->next_idle = narra_heap_idle;
  narra_heap_idle = heap;
  pthread_mutex_unlock(&narra_heap_lock);
}
#endif

static void narra_report_leaks(void) {
  if (getenv("NARRA_CHECK_LEAKS") && narra_live_total() != 0) {
    fprintf(stderr, "leak check: %" PRId64 " allocations still live\n", narra_live_total());
    _exit(102);
  }
}

#ifndef NARRA_PLAIN_MALLOC
// In thread mode the block functions take the heap, which the caller read once.
#ifdef NARRA_THREADS
#define NARRA_HEAP_ARG narra_heap* heap,
#define NARRA_HEAP_PASS heap,
#define NARRA_SIZE_CLASS(header) ((header) & 0xff)
#else
#define NARRA_HEAP_ARG
#define NARRA_HEAP_PASS
#define NARRA_SIZE_CLASS(header) (header)
#define heap (&narra_small)
#endif

static uint64_t* narra_carve(NARRA_HEAP_ARG size_t size) {
  if (!heap->at || (size_t)(heap->end - heap->at) < size) {
    char* chunk = malloc(NARRA_CHUNK_SIZE);
    if (!chunk) narra_panic("runtime", "out of memory");
    *(void**)chunk = heap->chunks;
    heap->chunks = chunk;
    heap->at = chunk + 16;
    heap->end = chunk + NARRA_CHUNK_SIZE;
  }
  uint64_t* block = (uint64_t*)heap->at;
  heap->at += size;
  return block;
}

#ifdef NARRA_THREADS
// Moves the blocks other threads freed to the free lists of their classes.
NARRA_COLD static bool narra_heap_collect(narra_heap* heap) {
  uint64_t* block = __atomic_exchange_n(&heap->remote, NULL, __ATOMIC_ACQUIRE);
  if (!block) return false;
  while (block) {
    uint64_t* next = *(uint64_t**)(block + 1);
    uint64_t size_class = NARRA_SIZE_CLASS(*block);
    *(uint64_t**)(block + 1) = heap->free_lists[size_class];
    heap->free_lists[size_class] = block;
    block = next;
  }
  return true;
}
#endif

static void* narra_block_new(NARRA_HEAP_ARG size_t n) {
  uint64_t* block;
  if (n <= NARRA_SMALL_MAX - 8) {
    // The header keeps its class while the block sits in a free list.
    size_t size_class = (n + 8 + 15) >> 4;
    block = heap->free_lists[size_class];
#ifdef NARRA_THREADS
    if (!block && __atomic_load_n(&heap->remote, __ATOMIC_RELAXED) && narra_heap_collect(heap)) {
      block = heap->free_lists[size_class];
    }
#endif
    if (block) {
      heap->free_lists[size_class] = *(uint64_t**)(block + 1);
    } else {
      block = narra_carve(NARRA_HEAP_PASS size_class << 4);
#ifdef NARRA_THREADS
      *block = size_class | (heap->number << 8);
#else
      *block = size_class;
#endif
    }
  } else {
    if (n > SIZE_MAX - 8) narra_panic("runtime", "out of memory");
    block = malloc(n + 8);
    if (!block) narra_panic("runtime", "out of memory");
    *block = 0;
  }
  return block + 1;
}

static void narra_block_free(NARRA_HEAP_ARG void* p) {
  uint64_t* block = (uint64_t*)p - 1;
  uint64_t header = *block;
  if (header == 0) {
    free(block);
    return;
  }
#ifdef NARRA_THREADS
  // A block of another heap goes onto that heap's remote stack.
  uint64_t owner = header >> 8;
  if (NARRA_UNLIKELY(owner != heap->number)) {
    narra_heap* home = narra_heaps[owner];
    uint64_t* head = __atomic_load_n(&home->remote, __ATOMIC_RELAXED);
    do {
      *(uint64_t**)p = head;
    } while (!__atomic_compare_exchange_n(&home->remote, &head, block, true, __ATOMIC_RELEASE, __ATOMIC_RELAXED));
    return;
  }
#endif
  uint64_t size_class = NARRA_SIZE_CLASS(header);
  *(uint64_t**)p = heap->free_lists[size_class];
  heap->free_lists[size_class] = block;
}
#ifndef NARRA_THREADS
#undef heap
#endif
#endif

static void* narra_alloc(size_t n) {
#ifdef NARRA_THREADS
  narra_heap* heap = narra_heap_here();
  heap->live++;
#ifdef NARRA_PLAIN_MALLOC
  void* p = malloc(n ? n : 1);
  if (!p) narra_panic("runtime", "out of memory");
  return p;
#else
  return narra_block_new(heap, n);
#endif
#else
#ifdef NARRA_PLAIN_MALLOC
  void* p = malloc(n ? n : 1);
  if (!p) narra_panic("runtime", "out of memory");
#else
  void* p = narra_block_new(n);
#endif
  narra_live++;
  return p;
#endif
}

static void narra_free(void* p) {
#ifdef NARRA_THREADS
  narra_heap* heap = narra_heap_here();
  heap->live--;
#ifdef NARRA_PLAIN_MALLOC
  free(p);
#else
  narra_block_free(heap, p);
#endif
#else
  narra_live--;
#ifdef NARRA_PLAIN_MALLOC
  free(p);
#else
  narra_block_free(p);
#endif
#endif
}

// The debug copy check. With NARRA_CHECK_COPIES set, a debug build counts each
// copy on write of a shared array, box, map, string builder or generator, and
// at exit reports the ones that no copy(...) explains, with exit code 103.
// copy(x) retains x with granting on, so each object the retain reaches gets a
// credit. A copy on write uses one credit of the object it copies, and the
// retains of its children during the copy grant each child one. The credits of
// an array, box, map, builder or generator end when its count falls back to 2
// or less. A heap enum value also carries a copy_held id that copy(x) gives to
// the holder it makes. Retains of that holder share the id, and the credit ends
// when the last of them is released, so releasing the original holder ends no
// credit; the rest end when the box is freed. The credits live in an
// open-addressing table of plain malloc memory, outside narra_live. Release
// builds (NARRA_RELEASE) have none of it: the hooks expand to nothing.
#ifndef NARRA_RELEASE
typedef struct narra_copy_slot { const void* p; int64_t credits; } narra_copy_slot;
static bool narra_copy_on = false;
static NARRA_PER_THREAD bool narra_copy_granting = false;
// In thread mode the table is behind one lock, which a thread may take again
// while it holds it (debug builds only).
#ifdef NARRA_THREADS
static pthread_mutex_t narra_copy_mutex = PTHREAD_MUTEX_INITIALIZER;
static NARRA_PER_THREAD int narra_copy_depth = 0;
static void narra_copy_lock(void) {
  if (narra_copy_depth++ == 0) pthread_mutex_lock(&narra_copy_mutex);
}
static void narra_copy_unlock(void) {
  if (--narra_copy_depth == 0) pthread_mutex_unlock(&narra_copy_mutex);
}
#else
#define narra_copy_lock() ((void)0)
#define narra_copy_unlock() ((void)0)
#endif
static narra_copy_slot* narra_copy_slots = NULL;
static size_t narra_copy_cap = 0, narra_copy_used = 0;
static int64_t narra_copy_unexplained[5];
static const char* const narra_copy_kinds[5] = {"array", "box", "map", "string builder", "generator"};

static size_t narra_copy_home(const void* p) {
  uint64_t x = (uint64_t)(uintptr_t)p * UINT64_C(0x9E3779B97F4A7C15);
  return (size_t)(x >> 32) & (narra_copy_cap - 1);
}

static narra_copy_slot* narra_copy_find(const void* p) {
  if (narra_copy_used == 0) return NULL;
  for (size_t i = narra_copy_home(p);; i = (i + 1) & (narra_copy_cap - 1)) {
    if (narra_copy_slots[i].p == p) return &narra_copy_slots[i];
    if (!narra_copy_slots[i].p) return NULL;
  }
}

static void narra_copy_put(const void* p, int64_t credits) {
  size_t i = narra_copy_home(p);
  while (narra_copy_slots[i].p) i = (i + 1) & (narra_copy_cap - 1);
  narra_copy_slots[i].p = p;
  narra_copy_slots[i].credits = credits;
  narra_copy_used++;
}

static void narra_copy_credit(const void* p) {
  narra_copy_slot* slot = narra_copy_find(p);
  if (slot) {
    slot->credits++;
    return;
  }
  // The table doubles when it would be more than half full.
  if ((narra_copy_used + 1) * 2 > narra_copy_cap) {
    narra_copy_slot* old = narra_copy_slots;
    size_t old_cap = narra_copy_cap;
    narra_copy_cap = old_cap ? old_cap * 2 : 64;
    narra_copy_slots = calloc(narra_copy_cap, sizeof(narra_copy_slot));
    if (!narra_copy_slots) narra_panic("runtime", "out of memory");
    narra_copy_used = 0;
    for (size_t i = 0; i < old_cap; i++) {
      if (old[i].p) narra_copy_put(old[i].p, old[i].credits);
    }
    free(old);
  }
  narra_copy_put(p, 1);
}

static void narra_copy_forget(const void* p) {
  narra_copy_slot* slot = narra_copy_find(p);
  if (!slot) return;
  // Later entries of the probe chain move back into the hole when their home
  // allows it, so every chain stays unbroken.
  size_t mask = narra_copy_cap - 1;
  size_t hole = (size_t)(slot - narra_copy_slots);
  for (size_t i = (hole + 1) & mask; narra_copy_slots[i].p; i = (i + 1) & mask) {
    size_t home = narra_copy_home(narra_copy_slots[i].p);
    if (((i - home) & mask) >= ((i - hole) & mask)) {
      narra_copy_slots[hole] = narra_copy_slots[i];
      hole = i;
    }
  }
  narra_copy_slots[hole].p = NULL;
  narra_copy_slots[hole].credits = 0;
  narra_copy_used--;
}

// Each heap enum holder that copy(...) makes gets an id, and the count of the
// holders sharing that id ends its credit when it falls to zero. Ids are not
// reused, so the table grows 4 bytes per copy(...) and a run is limited to
// 2^31 - 2 copies with the check on.
static int32_t* narra_copy_holders = NULL;
static int32_t narra_copy_ids = 0;
static size_t narra_copy_ids_cap = 0;

static int32_t narra_copy_brand(void) {
  if (narra_copy_ids == INT32_MAX - 1) narra_panic("runtime", "too many copy(...) calls for the copy check");
  if ((size_t)narra_copy_ids + 1 >= narra_copy_ids_cap) {
    narra_copy_ids_cap = narra_copy_ids_cap ? narra_copy_ids_cap * 2 : 64;
    narra_copy_holders = realloc(narra_copy_holders, narra_copy_ids_cap * sizeof(int32_t));
    if (!narra_copy_holders) narra_panic("runtime", "out of memory");
  }
  narra_copy_holders[++narra_copy_ids] = 1;
  return narra_copy_ids;
}

static void narra_copy_start(void) { narra_copy_on = getenv("NARRA_CHECK_COPIES") != NULL; }

static void narra_copy_grant(bool on) { narra_copy_granting = on && narra_copy_on; }

static void narra_copy_write(const void* p, int kind) {
  if (!narra_copy_on) return;
  narra_copy_slot* slot = narra_copy_find(p);
  if (slot && slot->credits > 0) {
    slot->credits--;
    narra_copy_granting = true;
    return;
  }
  narra_copy_unexplained[kind]++;
}

static void narra_report_copies(void) {
  if (!narra_copy_on) return;
  int64_t total = 0;
  for (int kind = 0; kind < 5; kind++) total += narra_copy_unexplained[kind];
  if (total == 0) return;
  fflush(stdout);
  fprintf(stderr, "copy check: copies on write that no copy(...) explains:");
  const char* separator = " ";
  for (int kind = 0; kind < 5; kind++) {
    if (narra_copy_unexplained[kind] > 0) {
      fprintf(stderr, "%s%s %" PRId64, separator, narra_copy_kinds[kind], narra_copy_unexplained[kind]);
      separator = ", ";
    }
  }
  fprintf(stderr, "\n");
  _exit(103);
}

// SHARED follows a retain, DROP comes before a decrement of p->rc, and WRITE and
// WRITTEN surround the copy that a copy on write makes.
// ONLY holds code that exists for the check, CREDIT gives credits at a copy(...),
// GRANTED runs a retain with granting on, and ENDED takes back the credits of
// the copy whose last holder was released, HELD counts a retain of a holder
// made by copy(...) and BRAND marks the holder it makes. All expand to nothing, or
// to the plain retain, under NARRA_RELEASE.
#define NARRA_COPY_SHARED(p) do { if (narra_copy_granting) { narra_copy_lock(); narra_copy_credit(p); narra_copy_unlock(); } } while (0)
// In thread mode a shared count is negative, so the tests compare the number
// of holders, and the table is read only under its lock.
#ifdef NARRA_THREADS
#define NARRA_COPY_DROP(p) do { if (narra_copy_on) { narra_copy_lock(); if (narra_copy_used && NARRA_RC_HOLDERS((p)->rc) <= 2) narra_copy_forget(p); narra_copy_unlock(); } } while (0)
#else
#define NARRA_COPY_DROP(p) do { if (narra_copy_used && (p)->rc <= 2) narra_copy_forget(p); } while (0)
#endif
#define NARRA_COPY_WRITE(p, kind) do { narra_copy_lock(); narra_copy_write(p, kind); narra_copy_unlock(); } while (0)
#define NARRA_COPY_WRITTEN() (narra_copy_granting = false)
#define NARRA_COPY_GRANT(on) narra_copy_grant(on)
#define NARRA_COPY_ONLY(...) __VA_ARGS__
#define NARRA_COPY_CREDIT(...) if (!narra_copy_on) {} else { narra_copy_lock(); __VA_ARGS__ narra_copy_unlock(); }
#define NARRA_COPY_GRANTED(...) do { narra_copy_grant(true); __VA_ARGS__ narra_copy_grant(false); } while (0)
#define NARRA_COPY_ENDED(v, uncredit) do { narra_copy_lock(); bool retired = (v).copy_held && --narra_copy_holders[(v).copy_held] == 0; if ((v).u && narra_copy_used) { narra_copy_slot* mark = narra_copy_find((v).u); if (mark) { bool last = NARRA_RC_HOLDERS((v).u->rc) <= 1; int64_t ended = last ? mark->credits : (retired && mark->credits > 0 ? 1 : 0); if (ended > 0 && ended < mark->credits) mark->credits -= ended; else if (ended > 0 || last) narra_copy_forget((v).u); while (ended-- > 0) uncredit(v); } } narra_copy_unlock(); } while (0)
#define NARRA_COPY_HELD(v) do { if ((v).copy_held) { narra_copy_lock(); narra_copy_holders[(v).copy_held]++; narra_copy_unlock(); } } while (0)
#define NARRA_COPY_BRAND(v) do { narra_copy_lock(); if ((v).copy_held) narra_copy_holders[(v).copy_held]--; (v).copy_held = narra_copy_brand(); narra_copy_unlock(); } while (0)
#define NARRA_COPY_START() narra_copy_start()
#define NARRA_COPY_REPORT() narra_report_copies()
#else
#define NARRA_COPY_SHARED(p) ((void)0)
#define NARRA_COPY_DROP(p) ((void)0)
#define NARRA_COPY_WRITE(p, kind) ((void)0)
#define NARRA_COPY_WRITTEN() ((void)0)
#define NARRA_COPY_GRANT(on) ((void)0)
#define NARRA_COPY_ONLY(...)
#define NARRA_COPY_CREDIT(...)
#define NARRA_COPY_GRANTED(...) do { __VA_ARGS__ } while (0)
#define NARRA_COPY_ENDED(v, uncredit) ((void)0)
#define NARRA_COPY_HELD(v) ((void)0)
#define NARRA_COPY_BRAND(v) ((void)0)
#define NARRA_COPY_START() ((void)0)
#define NARRA_COPY_REPORT() ((void)0)
#endif

// Reference counts (D111 item 2). A count above 0 is plain and owned by one
// thread, a count below 0 marks a shared object whose count changes
// atomically (minus one per holder), and 0 marks an immortal object (string
// literals, closure environments on the stack). RETAIN adds a holder, RELEASE
// removes one and is true for the last, and UNSHARE is true when a write must
// copy first. Only thread mode makes counts negative, so without it the macros
// are the plain operations. SHARE marks an object shared and is true when it
// was not, so the caller goes on to the values it holds.
#ifdef NARRA_THREADS
static inline void narra_rc_retain(int64_t* rc) {
  int64_t count = __atomic_load_n(rc, __ATOMIC_RELAXED);
  if (NARRA_LIKELY(count > 0)) {
    __atomic_store_n(rc, count + 1, __ATOMIC_RELAXED);
  } else if (count < 0) {
    __atomic_fetch_sub(rc, 1, __ATOMIC_RELAXED);
  }
}
static inline bool narra_rc_release(int64_t* rc) {
  int64_t count = __atomic_load_n(rc, __ATOMIC_RELAXED);
  if (NARRA_LIKELY(count > 0)) {
    __atomic_store_n(rc, count - 1, __ATOMIC_RELAXED);
    return count == 1;
  }
  return count < 0 && __atomic_add_fetch(rc, 1, __ATOMIC_ACQ_REL) == 0;
}
static inline bool narra_rc_share(int64_t* rc) {
  // Only the owning thread sees a positive count, so a plain store would do;
  // the loads are atomic because a shared count changes on other threads.
  int64_t count = __atomic_load_n(rc, __ATOMIC_RELAXED);
  if (count <= 0) return false;
  __atomic_store_n(rc, -count, __ATOMIC_RELAXED);
  return true;
}
static inline int64_t narra_rc_holders(int64_t* rc) {
  int64_t count = __atomic_load_n(rc, __ATOMIC_RELAXED);
  return count < 0 ? -count : count;
}
#define NARRA_RC_RETAIN(rc) narra_rc_retain(&(rc))
#define NARRA_RC_RELEASE(rc) narra_rc_release(&(rc))
#define NARRA_RC_UNSHARE(rc) (__atomic_load_n(&(rc), __ATOMIC_RELAXED) != 1)
#define NARRA_RC_HOLDERS(rc) narra_rc_holders(&(rc))
#else
#define NARRA_RC_RETAIN(rc) ((void)((rc)++))
#define NARRA_RC_RELEASE(rc) (--(rc) == 0)
#define NARRA_RC_UNSHARE(rc) ((rc) != 1)
#define NARRA_RC_HOLDERS(rc) (rc)
#endif

// Dropping a heap value queues the values it holds instead of recursing, so
// a long list or a deep tree is freed with a worklist, not the C stack. The
// worklist grows and is reused by later drops; it is never freed. It is not
// counted by narra_live (plain realloc), and it stays reachable through a
// global, so neither the leak check nor LeakSanitizer reports it.
// A Range value is its two bounds and its step (1 counts up, a negative step counts down).
typedef struct narra_range { int64_t low; int64_t high; int64_t step; } narra_range;

typedef struct narra_pending { void* p; void (*drop)(void*); } narra_pending;
static _Thread_local narra_pending* narra_pend = NULL;
static _Thread_local size_t narra_pend_n = 0, narra_pend_cap = 0;
static _Thread_local bool narra_draining = false;

static void narra_drop(void* p, void (*drop)(void*)) {
  if (narra_draining) {
    if (narra_pend_n == narra_pend_cap) {
      narra_pend_cap = narra_pend_cap ? narra_pend_cap * 2 : 64;
      narra_pending* grown = realloc(narra_pend, narra_pend_cap * sizeof(narra_pending));
      if (!grown) narra_panic("runtime", "out of memory");
      narra_pend = grown;
    }
    narra_pend[narra_pend_n].p = p;
    narra_pend[narra_pend_n].drop = drop;
    narra_pend_n++;
    return;
  }
  narra_draining = true;
  drop(p);
  while (narra_pend_n > 0) {
    narra_pending d = narra_pend[--narra_pend_n];
    d.drop(d.p);
  }
  narra_draining = false;
}

#ifdef NARRA_THREADS
// Marking a value shared (D111 item 2) queues the parts of each object it
// marks, as narra_drop does, so a long list is walked without the C stack.
static NARRA_PER_THREAD narra_pending* narra_shares = NULL;
static NARRA_PER_THREAD size_t narra_shares_n = 0, narra_shares_cap = 0;
static NARRA_PER_THREAD bool narra_sharing = false;

static void narra_share(void* p, void (*walk)(void*)) {
  if (narra_sharing) {
    if (narra_shares_n == narra_shares_cap) {
      narra_shares_cap = narra_shares_cap ? narra_shares_cap * 2 : 64;
      narra_pending* grown = realloc(narra_shares, narra_shares_cap * sizeof(narra_pending));
      if (!grown) narra_panic("runtime", "out of memory");
      narra_shares = grown;
    }
    narra_shares[narra_shares_n].p = p;
    narra_shares[narra_shares_n].drop = walk;
    narra_shares_n++;
    return;
  }
  narra_sharing = true;
  walk(p);
  while (narra_shares_n > 0) {
    narra_pending next = narra_shares[--narra_shares_n];
    next.drop(next.p);
  }
  narra_sharing = false;
}
#endif

static void* narra_realloc(void* p, size_t n) {
#ifdef NARRA_PLAIN_MALLOC
  void* q = realloc(p, n ? n : 1);
  if (!q) narra_panic("runtime", "out of memory");
  return q;
#else
#ifdef NARRA_THREADS
  narra_heap* heap = narra_heap_here();
#endif
  if (!p) return narra_block_new(NARRA_HEAP_PASS n);
  uint64_t* block = (uint64_t*)p - 1;
  uint64_t size_class = NARRA_SIZE_CLASS(*block);
  // A malloc block stays one, even when it shrinks.
  if (size_class == 0) {
    if (n > SIZE_MAX - 8) narra_panic("runtime", "out of memory");
    block = realloc(block, n + 8);
    if (!block) narra_panic("runtime", "out of memory");
    return block + 1;
  }
  // A small block that still fits stays; otherwise it moves to another class or to malloc.
  size_t room = (size_t)(size_class << 4) - 8;
  if (n <= room) return p;
  void* q = narra_block_new(NARRA_HEAP_PASS n);
  memcpy(q, p, room);
  narra_block_free(NARRA_HEAP_PASS p);
  return q;
#endif
}

// Generators (emitgenerators.nra). A Generator<T> value points to the header
// of a counted frame. resume writes the next value through out and returns
// true, or returns false once the generator has finished. state is 0 before
// the first resume, -1 after the last, and otherwise the yield it waits at.
typedef struct narra_gen {
  int64_t rc;
  int64_t state;
  bool (*resume)(struct narra_gen* g, void* out);
  struct narra_gen* (*clone)(struct narra_gen* g);
  void (*drop)(void* g);
  const void* in; // the input of the resume in progress (coroutines only)
} narra_gen;
static void narra_gen_retain(narra_gen* g) {
  NARRA_RC_RETAIN(g->rc);
  NARRA_COPY_SHARED(g);
}
static void narra_gen_release(narra_gen* g) {
  NARRA_COPY_DROP(g);
  if (NARRA_RC_RELEASE(g->rc)) narra_drop(g, g->drop);
}
// narra_gen_next resumes *p. A shared generator is copied first, so the
// other holders of the value keep their state.
static bool narra_gen_next(narra_gen** p, void* out) {
  if (NARRA_RC_UNSHARE((*p)->rc)) {
    NARRA_COPY_WRITE(*p, 4);
    narra_gen* copy = (*p)->clone(*p);
    NARRA_COPY_WRITTEN();
    narra_gen_release(*p);
    *p = copy;
  }
  return (*p)->resume(*p, out);
}

// narra_co_next is Co.resume's step: it resumes *p with the input at in (NULL
// for a generator) and returns 1 with a value at out, 0 when the generator is
// done, or -1 when it panicked. Only a guarded call (test builds) catches a
// panic: the frame is then finished, and what it held is not freed.
static NARRA_PER_THREAD bool narra_co_panicked = false;
static int narra_co_next(narra_gen** p, void* out, const void* in, bool guard) {
  if (NARRA_RC_UNSHARE((*p)->rc)) {
    NARRA_COPY_WRITE(*p, 4);
    narra_gen* copy = (*p)->clone(*p);
    NARRA_COPY_WRITTEN();
    narra_gen_release(*p);
    *p = copy;
  }
  (*p)->in = in;
  narra_co_panicked = false;
  if (!guard) return (*p)->resume(*p, out) ? 1 : 0;
  jmp_buf here;
  jmp_buf* outer = narra_panic_jmp;
  int64_t live = narra_live;
  narra_gen* g = *p;
  narra_panic_jmp = &here;
  if (setjmp(here)) {
    narra_panic_jmp = outer;
    g->state = -1;
    // The count goes back to its value before the resume. This assumes the
    // coroutine only allocated since then: a free of memory from before the
    // resume would be miscounted (the leak check would then see one extra).
    narra_live = live;
    narra_co_panicked = true;
    return -1;
  }
  bool more = g->resume(g, out);
  narra_panic_jmp = outer;
  return more ? 1 : 0;
}

// Strings. A count of 0 marks a literal that is never freed.
typedef struct lstr {
  int64_t rc;
  int64_t len;
  char* data;
} lstr;

static lstr* lstr_new(int64_t len) {
  lstr* s = narra_alloc(sizeof(lstr));
  s->rc = 1;
  s->len = len;
  s->data = narra_alloc((size_t)len + 1);
  s->data[len] = 0;
  return s;
}

static void lstr_retain(lstr* s) {
#ifdef NARRA_THREADS
  NARRA_RC_RETAIN(s->rc);
#else
  if (s->rc != 0) s->rc++;
#endif
}

static void lstr_release(lstr* s) {
#ifdef NARRA_THREADS
  if (NARRA_RC_RELEASE(s->rc)) {
#else
  if (s->rc != 0 && --s->rc == 0) {
#endif
    narra_free(s->data);
    narra_free(s);
  }
}

#ifdef NARRA_THREADS
static void lstr_share(lstr* s) { narra_rc_share(&s->rc); }
#endif

static lstr* lstr_from_cstr(const char* c) {
  int64_t n = (int64_t)strlen(c);
  lstr* s = lstr_new(n);
  memcpy(s->data, c, (size_t)n);
  return s;
}

static lstr* lstr_concat(lstr* a, lstr* b) {
  lstr* s = lstr_new(a->len + b->len);
  memcpy(s->data, a->data, (size_t)a->len);
  memcpy(s->data + a->len, b->data, (size_t)b->len);
  return s;
}

static lstr* lstr_concat_n(int64_t count, lstr** parts) {
  // Sum the lengths, then copy each part once
  int64_t total = 0;
  for (int64_t i = 0; i < count; i++) total += parts[i]->len;
  lstr* s = lstr_new(total);
  char* cursor = s->data;
  for (int64_t i = 0; i < count; i++) {
    memcpy(cursor, parts[i]->data, (size_t)parts[i]->len);
    cursor += parts[i]->len;
  }
  return s;
}

static bool lstr_repeat(lstr* a, int64_t n, lstr** out) {
  int64_t size;
  if (n < 0 || __builtin_mul_overflow(a->len, n, &size)) return false;
  if (a->len == 0) {
    *out = lstr_new(0);
    return true;
  }
  lstr* s = lstr_new(size);
  for (int64_t i = 0; i < n; i++) memcpy(s->data + i * a->len, a->data, (size_t)a->len);
  *out = s;
  return true;
}

static lstr* lstr_from_digits(uint64_t magnitude, int negative) {
  char buf[24];
  char* end = buf + sizeof buf;
  char* start = end;
  // Write the digits from the last one backwards
  do {
    *--start = (char)('0' + magnitude % 10);
    magnitude /= 10;
  } while (magnitude != 0);
  if (negative) *--start = '-';
  lstr* s = lstr_new(end - start);
  memcpy(s->data, start, (size_t)(end - start));
  return s;
}

static lstr* lstr_from_int(int64_t v) {
  // Negate in unsigned arithmetic so INT64_MIN does not overflow
  if (v < 0) return lstr_from_digits(0 - (uint64_t)v, 1);
  return lstr_from_digits((uint64_t)v, 0);
}

static lstr* lstr_from_uint(uint64_t v) {
  return lstr_from_digits(v, 0);
}

// Lays out the digits of a "%e" text the way JavaScript does, except that a
// whole number keeps ".0" so the text still reads as a float.
static lstr* lstr_from_scientific(const char* sci) {
  char out[400];
  char digits[40];
  size_t used = 0;
  size_t count = 0;
  if (*sci == '-') out[used++] = *sci++;
  for (; *sci != 'e'; sci++) {
    if (*sci != '.') digits[count++] = *sci;
  }
  int exponent = atoi(sci + 1);
  while (count > 1 && digits[count - 1] == '0') count--;
  int point = exponent + 1;
  if (count == 1 && digits[0] == '0') {
    memcpy(out + used, "0.0", 3);
    used += 3;
  } else if (point > 0 && point <= 21) {
    for (int i = 0; i < point; i++) out[used++] = (size_t)i < count ? digits[i] : '0';
    out[used++] = '.';
    if (count > (size_t)point) {
      for (size_t i = (size_t)point; i < count; i++) out[used++] = digits[i];
    } else {
      out[used++] = '0';
    }
  } else if (point <= 0 && point > -6) {
    out[used++] = '0';
    out[used++] = '.';
    for (int i = point; i < 0; i++) out[used++] = '0';
    for (size_t i = 0; i < count; i++) out[used++] = digits[i];
  } else {
    out[used++] = digits[0];
    if (count > 1) {
      out[used++] = '.';
      for (size_t i = 1; i < count; i++) out[used++] = digits[i];
    }
    out[used++] = 'e';
    out[used++] = exponent < 0 ? '-' : '+';
    used += (size_t)snprintf(out + used, sizeof out - used, "%d", exponent < 0 ? -exponent : exponent);
  }
  out[used] = 0;
  return lstr_from_cstr(out);
}

// Shortest text that reads back as the same double.
static lstr* lstr_from_f64(double v) {
  if (v != v) return lstr_from_cstr("nan");
  if (v - v != 0.0) return lstr_from_cstr(v > 0 ? "inf" : "-inf");
  char buf[40];
  for (int prec = 0; prec <= 16; prec++) {
    snprintf(buf, sizeof buf, "%.*e", prec, v);
    if (strtod(buf, NULL) == v) break;
  }
  return lstr_from_scientific(buf);
}

// Shortest text that reads back as the same float.
static lstr* lstr_from_f32(float v) {
  if (v != v) return lstr_from_cstr("nan");
  if (v - v != 0.0f) return lstr_from_cstr(v > 0 ? "inf" : "-inf");
  char buf[40];
  for (int prec = 0; prec <= 8; prec++) {
    snprintf(buf, sizeof buf, "%.*e", prec, (double)v);
    if (strtof(buf, NULL) == v) break;
  }
  return lstr_from_scientific(buf);
}

static lstr* lstr_from_char(uint32_t c) {
  char buf[5] = {0};
  int64_t n;
  if (c < 0x80) {
    buf[0] = (char)c;
    n = 1;
  } else if (c < 0x800) {
    buf[0] = (char)(0xC0 | (c >> 6));
    buf[1] = (char)(0x80 | (c & 0x3F));
    n = 2;
  } else if (c < 0x10000) {
    buf[0] = (char)(0xE0 | (c >> 12));
    buf[1] = (char)(0x80 | ((c >> 6) & 0x3F));
    buf[2] = (char)(0x80 | (c & 0x3F));
    n = 3;
  } else {
    buf[0] = (char)(0xF0 | (c >> 18));
    buf[1] = (char)(0x80 | ((c >> 12) & 0x3F));
    buf[2] = (char)(0x80 | ((c >> 6) & 0x3F));
    buf[3] = (char)(0x80 | (c & 0x3F));
    n = 4;
  }
  // Use the encoded length: a NUL character has no strlen.
  lstr* r = lstr_new(n);
  memcpy(r->data, buf, (size_t)n);
  return r;
}

static bool lstr_eq(lstr* a, lstr* b) {
  return a->len == b->len && memcmp(a->data, b->data, (size_t)a->len) == 0;
}

static int lstr_cmp(lstr* a, lstr* b) {
  int64_t n = a->len < b->len ? a->len : b->len;
  int c = memcmp(a->data, b->data, (size_t)n);
  if (c != 0) return c;
  return a->len < b->len ? -1 : (a->len > b->len ? 1 : 0);
}

// Standard trait functions for primitive types (trait-model contract).
static bool narra_eq_int(int64_t a, int64_t b) { return a == b; }
static bool narra_eq_f64(double a, double b) { return a == b; }
static bool narra_eq_f32(float a, float b) { return a == b; }
static bool narra_eq_bool(bool a, bool b) { return a == b; }
static bool narra_eq_char(uint32_t a, uint32_t b) { return a == b; }

static uint64_t narra_hash_mix(uint64_t h, uint64_t x) {
  h ^= x;
  h *= UINT64_C(1099511628211);
  return h ^ (h >> 29);
}

static uint64_t narra_hash_int(int64_t v) {
  uint64_t x = (uint64_t)v + UINT64_C(0x9E3779B97F4A7C15);
  x = (x ^ (x >> 30)) * UINT64_C(0xBF58476D1CE4E5B9);
  x = (x ^ (x >> 27)) * UINT64_C(0x94D049BB133111EB);
  return x ^ (x >> 31);
}
static uint64_t narra_hash_bool(bool v) { return narra_hash_int(v ? 1 : 0); }
static uint64_t narra_hash_char(uint32_t v) { return narra_hash_int((int64_t)v); }

static uint64_t lstr_hash(lstr* s) {
  uint64_t h = UINT64_C(14695981039346656037);
  for (int64_t i = 0; i < s->len; i++) {
    h ^= (unsigned char)s->data[i];
    h *= UINT64_C(1099511628211);
  }
  return h;
}

static int narra_cmp_int(int64_t a, int64_t b) { return (a > b) - (a < b); }
static int narra_cmp_f64(double a, double b) { return (a > b) - (a < b); }
static int narra_cmp_f32(float a, float b) { return (a > b) - (a < b); }
static int narra_cmp_bool(bool a, bool b) { return (a > b) - (a < b); }
static int narra_cmp_char(uint32_t a, uint32_t b) { return (a > b) - (a < b); }

// Ordering's tag for a comparison result: Less = 0, Equal = 1, Greater = 2.
static int32_t narra_ordering(int c) { return c < 0 ? 0 : (c > 0 ? 2 : 1); }

// Appends c to *acc, replacing *acc with the result.
static void lstr_append_c(lstr** acc, const char* c) {
  int64_t n = (int64_t)strlen(c);
  lstr* s = lstr_new((*acc)->len + n);
  memcpy(s->data, (*acc)->data, (size_t)(*acc)->len);
  memcpy(s->data + (*acc)->len, c, (size_t)n);
  lstr_release(*acc);
  *acc = s;
}

// Appends piece to *acc and releases piece.
static void lstr_append_owned(lstr** acc, lstr* piece) {
  lstr* s = lstr_concat(*acc, piece);
  lstr_release(*acc);
  lstr_release(piece);
  *acc = s;
}

static lstr* narra_show_bool(bool v) { return lstr_from_cstr(v ? "true" : "false"); }

// The escape for byte c inside quotes q, or 0 when c stands for itself.
static char narra_escape(char c, char q) {
  switch (c) {
  case '\n': return 'n';
  case '\t': return 't';
  case '\r': return 'r';
  case '\\': return '\\';
  default: return c == q ? q : 0;
  }
}

// Bytes in quotes q, with q, `\`, newline, tab and carriage return escaped.
static lstr* narra_quote(const char* data, int64_t len, char q) {
  int64_t n = 2;
  for (int64_t i = 0; i < len; i++) n += narra_escape(data[i], q) ? 2 : 1;
  lstr* s = lstr_new(n);
  int64_t j = 0;
  s->data[j++] = q;
  for (int64_t i = 0; i < len; i++) {
    char e = narra_escape(data[i], q);
    if (e) {
      s->data[j++] = '\\';
      s->data[j++] = e;
    } else {
      s->data[j++] = data[i];
    }
  }
  s->data[j++] = q;
  return s;
}

static lstr* narra_show_char(uint32_t c) {
  lstr* u = lstr_from_char(c);
  lstr* s = narra_quote(u->data, u->len, '\'');
  lstr_release(u);
  return s;
}

static lstr* narra_show_string(lstr* v) { return narra_quote(v->data, v->len, '"'); }

// bytes as `bytes[68 69 0a]`: two lowercase hex digits per byte.
static lstr* narra_show_bytes(lstr* v) {
  static const char digits[] = "0123456789abcdef";
  int64_t n = 7 + (v->len > 0 ? v->len * 3 - 1 : 0);
  lstr* s = lstr_new(n);
  memcpy(s->data, "bytes[", 6);
  int64_t j = 6;
  for (int64_t i = 0; i < v->len; i++) {
    unsigned char c = (unsigned char)v->data[i];
    if (i > 0) s->data[j++] = ' ';
    s->data[j++] = digits[c >> 4];
    s->data[j++] = digits[c & 15];
  }
  s->data[j] = ']';
  return s;
}

static void narra_print(lstr* s) {
  fwrite(s->data, 1, (size_t)s->len, stdout);
  fputc('\n', stdout);
}

static void narra_assert(bool ok, const char* where) {
  if (!ok) narra_panic(where, "assertion failed");
}

_Noreturn static void narra_panic_str(lstr* msg, const char* where) {
  narra_panic(where, msg->data);
}

// Integer arithmetic. + - * and unary - wrap at the type's width in every
// build. The checker proves every divisor is not zero and every shift count is
// inside the bit width, so the functions below do not test them. MIN / -1 gives
// MIN and MIN % -1 gives 0.
static int64_t narra_wrap_add(int64_t a, int64_t b) { return (int64_t)((uint64_t)a + (uint64_t)b); }
static int64_t narra_wrap_sub(int64_t a, int64_t b) { return (int64_t)((uint64_t)a - (uint64_t)b); }
static int64_t narra_wrap_mul(int64_t a, int64_t b) { return (int64_t)((uint64_t)a * (uint64_t)b); }
static int64_t narra_wrap_neg(int64_t a) { return (int64_t)(0 - (uint64_t)a); }

static int64_t narra_div(int64_t a, int64_t b) {
  if (a == INT64_MIN && b == -1) return a;
  return a / b;
}

static int64_t narra_rem(int64_t a, int64_t b) {
  if (b == -1) return 0;
  return a % b;
}

static int64_t narra_shl(int64_t a, int64_t b) {
  return (int64_t)((uint64_t)a << b);
}

static int64_t narra_shr(int64_t a, int64_t b) {
  return a >> b;
}

// The sized integers (D82): one set of functions per type, named with the
// type's suffix (narra_wrap_add_u8). They wrap at the type's own width.
#define NARRA_SIZED(N, T, MIN, SHOW) \
  static T narra_wrap_add_##N(T a, T b) { return (T)((uint64_t)a + (uint64_t)b); } \
  static T narra_wrap_sub_##N(T a, T b) { return (T)((uint64_t)a - (uint64_t)b); } \
  static T narra_wrap_mul_##N(T a, T b) { return (T)((uint64_t)a * (uint64_t)b); } \
  static T narra_wrap_neg_##N(T a) { return (T)(0 - (uint64_t)a); } \
  static T narra_div_##N(T a, T b) { \
    if ((MIN) != 0 && a == (T)(MIN) && b == (T)-1) return a; \
    return (T)(a / b); \
  } \
  static T narra_rem_##N(T a, T b) { \
    if ((MIN) != 0 && b == (T)-1) return 0; \
    return (T)(a % b); \
  } \
  static T narra_shl_##N(T a, T b) { \
    return (T)((uint64_t)a << b); \
  } \
  static T narra_shr_##N(T a, T b) { \
    return (T)(a >> b); \
  } \
  static bool narra_eq_##N(T a, T b) { return a == b; } \
  static int narra_cmp_##N(T a, T b) { return (a > b) - (a < b); } \
  static uint64_t narra_hash_##N(T v) { return narra_hash_int((int64_t)v); } \
  static lstr* narra_show_##N(T v) { return SHOW; }

NARRA_SIZED(i8, int8_t, INT8_MIN, lstr_from_int((int64_t)v))
NARRA_SIZED(i16, int16_t, INT16_MIN, lstr_from_int((int64_t)v))
NARRA_SIZED(i32, int32_t, INT32_MIN, lstr_from_int((int64_t)v))
NARRA_SIZED(u8, uint8_t, 0, lstr_from_uint((uint64_t)v))
NARRA_SIZED(u16, uint16_t, 0, lstr_from_uint((uint64_t)v))
NARRA_SIZED(u32, uint32_t, 0, lstr_from_uint((uint64_t)v))
NARRA_SIZED(u64, uint64_t, 0, lstr_from_uint(v))

// A step of 0 gives the empty range, so a loop over it never starts.
static narra_range narra_range_stepped(int64_t low, int64_t high, int64_t step) {
  if (step == 0) return (narra_range){low, low, 1};
  return (narra_range){low, high, step};
}

// Moves a loop variable by the step; a sum past the int range ends the loop.
static void narra_range_advance(int64_t* value, int64_t step, int64_t high) {
  if (__builtin_add_overflow(*value, step, value)) *value = high;
}

static int64_t narra_abs(int64_t a, const char* where) { return a < 0 ? narra_wrap_neg(a) : a; }

static bool narra_float_truncate(double v, int64_t* out) {
  if (!(v >= -9223372036854775808.0 && v < 9223372036854775808.0)) return false;
  *out = (int64_t)v;
  return true;
}


_Noreturn NARRA_COLD static void narra_bounds_fail(int64_t i, int64_t len, const char* where) {
  char msg[96];
  snprintf(msg, sizeof msg, "index %" PRId64 " out of range for length %" PRId64, i, len);
  narra_panic(where, msg);
}

static inline void narra_bounds(int64_t i, int64_t len, const char* where) {
  if (NARRA_UNLIKELY((uint64_t)i >= (uint64_t)len)) narra_bounds_fail(i, len, where);
}
