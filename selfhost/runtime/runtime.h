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

// Test builds only (std/test): in a child process that runs a Test.panics
// closure, a panic writes its message to this pipe and exits 101 at once.
static int narra_panic_fd = -1;
// Set by the test runner's child: the pipe then gets "<where>\n<message>".
static int narra_panic_pos = 0;
// Where Test.golden reads and writes files; narra test sets it with --golden-dir.
static const char* narra_golden_dir = "testdata";

// Test builds: while Co.resume runs a coroutine it sets this, so a panic in
// the coroutine returns to Co.resume as Panicked (std/coroutine).
static jmp_buf* narra_panic_jmp = NULL;
static char narra_panic_text[256];

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

_Noreturn NARRA_COLD static void narra_panic(const char* where, const char* msg) {
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
  exit(101);
}

// Live allocation count. With NARRA_CHECK_LEAKS set, the program reports
// allocations still live at exit, which shows reference counting bugs.
static int64_t narra_live = 0;

static void narra_report_leaks(void) {
  if (getenv("NARRA_CHECK_LEAKS") && narra_live != 0) {
    fprintf(stderr, "leak check: %" PRId64 " allocations still live\n", narra_live);
    _exit(102);
  }
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

#ifndef NARRA_PLAIN_MALLOC
// Blocks of up to 512 bytes, header included, come from size-class free
// lists in steps of 16 bytes, carved from 1 MiB chunks. Larger blocks use
// malloc. An 8-byte header in front of each block holds its class, 0 for a
// malloc block. Chunks are never returned to the system.
#define NARRA_SMALL_MAX 512
#define NARRA_CHUNK_SIZE ((size_t)1 << 20)
// Make this _Thread_local when threads exist: one heap per thread.
#define NARRA_HEAP_LOCAL
typedef struct narra_heap {
  uint64_t* free_lists[NARRA_SMALL_MAX / 16 + 1];
  char* at;
  char* end;
  void* chunks; // linked through each chunk's first word, so they stay reachable
} narra_heap;
static NARRA_HEAP_LOCAL narra_heap narra_small;

static uint64_t* narra_carve(size_t size) {
  if (!narra_small.at || (size_t)(narra_small.end - narra_small.at) < size) {
    char* chunk = malloc(NARRA_CHUNK_SIZE);
    if (!chunk) narra_panic("runtime", "out of memory");
    *(void**)chunk = narra_small.chunks;
    narra_small.chunks = chunk;
    narra_small.at = chunk + 16;
    narra_small.end = chunk + NARRA_CHUNK_SIZE;
  }
  uint64_t* block = (uint64_t*)narra_small.at;
  narra_small.at += size;
  return block;
}

static void* narra_block_new(size_t n) {
  uint64_t* block;
  if (n <= NARRA_SMALL_MAX - 8) {
    // The header keeps its class while the block sits in a free list.
    size_t size_class = (n + 8 + 15) >> 4;
    block = narra_small.free_lists[size_class];
    if (block) {
      narra_small.free_lists[size_class] = *(uint64_t**)(block + 1);
    } else {
      block = narra_carve(size_class << 4);
      *block = size_class;
    }
  } else {
    if (n > SIZE_MAX - 8) narra_panic("runtime", "out of memory");
    block = malloc(n + 8);
    if (!block) narra_panic("runtime", "out of memory");
    *block = 0;
  }
  return block + 1;
}

static void narra_block_free(void* p) {
  uint64_t* block = (uint64_t*)p - 1;
  uint64_t size_class = *block;
  if (size_class == 0) {
    free(block);
    return;
  }
  *(uint64_t**)p = narra_small.free_lists[size_class];
  narra_small.free_lists[size_class] = block;
}
#endif

static void* narra_alloc(size_t n) {
#ifdef NARRA_PLAIN_MALLOC
  void* p = malloc(n ? n : 1);
  if (!p) narra_panic("runtime", "out of memory");
#else
  void* p = narra_block_new(n);
#endif
  narra_live++;
  return p;
}

static void narra_free(void* p) {
  narra_live--;
#ifdef NARRA_PLAIN_MALLOC
  free(p);
#else
  narra_block_free(p);
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
static bool narra_copy_granting = false;
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
#define NARRA_COPY_SHARED(p) do { if (narra_copy_granting) narra_copy_credit(p); } while (0)
#define NARRA_COPY_DROP(p) do { if (narra_copy_used && (p)->rc <= 2) narra_copy_forget(p); } while (0)
#define NARRA_COPY_WRITE(p, kind) narra_copy_write(p, kind)
#define NARRA_COPY_WRITTEN() (narra_copy_granting = false)
#define NARRA_COPY_GRANT(on) narra_copy_grant(on)
#define NARRA_COPY_ONLY(...) __VA_ARGS__
#define NARRA_COPY_CREDIT(...) if (!narra_copy_on) {} else { __VA_ARGS__ }
#define NARRA_COPY_GRANTED(...) do { narra_copy_grant(true); __VA_ARGS__ narra_copy_grant(false); } while (0)
#define NARRA_COPY_ENDED(v, uncredit) do { bool retired = (v).copy_held && --narra_copy_holders[(v).copy_held] == 0; if ((v).u && narra_copy_used) { narra_copy_slot* mark = narra_copy_find((v).u); if (mark) { bool last = (v).u->rc <= 1; int64_t ended = last ? mark->credits : (retired && mark->credits > 0 ? 1 : 0); if (ended > 0 && ended < mark->credits) mark->credits -= ended; else if (ended > 0 || last) narra_copy_forget((v).u); while (ended-- > 0) uncredit(v); } } } while (0)
#define NARRA_COPY_HELD(v) do { if ((v).copy_held) narra_copy_holders[(v).copy_held]++; } while (0)
#define NARRA_COPY_BRAND(v) do { if ((v).copy_held) narra_copy_holders[(v).copy_held]--; (v).copy_held = narra_copy_brand(); } while (0)
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

static void* narra_realloc(void* p, size_t n) {
#ifdef NARRA_PLAIN_MALLOC
  void* q = realloc(p, n ? n : 1);
  if (!q) narra_panic("runtime", "out of memory");
  return q;
#else
  if (!p) return narra_block_new(n);
  uint64_t* block = (uint64_t*)p - 1;
  uint64_t size_class = *block;
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
  void* q = narra_block_new(n);
  memcpy(q, p, room);
  narra_block_free(p);
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
  g->rc++;
  NARRA_COPY_SHARED(g);
}
static void narra_gen_release(narra_gen* g) {
  NARRA_COPY_DROP(g);
  if (--g->rc == 0) narra_drop(g, g->drop);
}
// narra_gen_next resumes *p. A shared generator is copied first, so the
// other holders of the value keep their state.
static bool narra_gen_next(narra_gen** p, void* out) {
  if ((*p)->rc > 1) {
    NARRA_COPY_WRITE(*p, 4);
    narra_gen* copy = (*p)->clone(*p);
    NARRA_COPY_WRITTEN();
    NARRA_COPY_DROP(*p);
    (*p)->rc--;
    *p = copy;
  }
  return (*p)->resume(*p, out);
}

// narra_co_next is Co.resume's step: it resumes *p with the input at in (NULL
// for a generator) and returns 1 with a value at out, 0 when the generator is
// done, or -1 when it panicked. Only a guarded call (test builds) catches a
// panic: the frame is then finished, and what it held is not freed.
static bool narra_co_panicked = false;
static int narra_co_next(narra_gen** p, void* out, const void* in, bool guard) {
  if ((*p)->rc > 1) {
    NARRA_COPY_WRITE(*p, 4);
    narra_gen* copy = (*p)->clone(*p);
    NARRA_COPY_WRITTEN();
    NARRA_COPY_DROP(*p);
    (*p)->rc--;
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

// Strings. A negative count marks a literal that is never freed.
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
  if (s->rc >= 0) s->rc++;
}

static void lstr_release(lstr* s) {
  if (s->rc >= 0 && --s->rc == 0) {
    narra_free(s->data);
    narra_free(s);
  }
}

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

// Shortest text that reads back as the same double.
static lstr* lstr_from_f64(double v) {
  char buf[40];
  for (int prec = 1; prec <= 17; prec++) {
    snprintf(buf, sizeof buf, "%.*g", prec, v);
    if (strtod(buf, NULL) == v) break;
  }
  if (!strchr(buf, '.') && !strchr(buf, 'e') && !strchr(buf, 'n') && !strchr(buf, 'i')) {
    strcat(buf, ".0");
  }
  return lstr_from_cstr(buf);
}

// Shortest text that reads back as the same float.
static lstr* lstr_from_f32(float v) {
  char buf[40];
  for (int prec = 1; prec <= 9; prec++) {
    snprintf(buf, sizeof buf, "%.*g", prec, (double)v);
    if (strtof(buf, NULL) == v) break;
  }
  if (!strchr(buf, '.') && !strchr(buf, 'e') && !strchr(buf, 'n') && !strchr(buf, 'i')) {
    strcat(buf, ".0");
  }
  return lstr_from_cstr(buf);
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

static bool narra_range_stepped(int64_t low, int64_t high, int64_t step, narra_range* out) {
  if (step == 0) return false;
  *out = (narra_range){low, high, step};
  return true;
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
