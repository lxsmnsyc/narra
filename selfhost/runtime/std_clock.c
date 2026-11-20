// std/internal/clock. The monotonic clock and the wall clock.
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
