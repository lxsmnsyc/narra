// std/internal/entropy. A seed from the operating system's random source.
#include <time.h>

static int64_t narra_entropy_seed(void) {
  uint64_t seed = 0;
  FILE* source = fopen("/dev/urandom", "rb");
  if (source != NULL) {
    if (fread(&seed, sizeof seed, 1, source) != 1) seed = 0;
    fclose(source);
  }
  if (seed == 0) {
    seed = (uint64_t)time(NULL) * 6364136223846793005ULL + (uint64_t)(uintptr_t)&seed;
  }
  return (int64_t)seed;
}
