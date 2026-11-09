// std/internal/clock. The monotonic clock, the wall clock and sleep.
#include <errno.h>
#include <time.h>

static int64_t narra_clock_monotonic_nanos(void) {
  struct timespec now;
  clock_gettime(CLOCK_MONOTONIC, &now);
  return (int64_t)now.tv_sec * 1000000000 + (int64_t)now.tv_nsec;
}

static int64_t narra_clock_wall_millis(void) {
  struct timespec now;
  clock_gettime(CLOCK_REALTIME, &now);
  return (int64_t)now.tv_sec * 1000 + (int64_t)now.tv_nsec / 1000000;
}

// Sleeps for the given nanoseconds, again after a signal; a value of zero or less returns at once.
static void narra_clock_sleep_nanos(int64_t nanos) {
  if (nanos <= 0) return;
  struct timespec remaining;
  remaining.tv_sec = (time_t)(nanos / 1000000000);
  remaining.tv_nsec = (long)(nanos % 1000000000);
  while (nanosleep(&remaining, &remaining) != 0 && errno == EINTR) {
  }
}
