// std/internal/entropy. A seed from the operating system's random source.
#include <time.h>

static int64_t narra_entropy_seed(void) {
  uint64_t seed = 0;
#ifdef _WIN32
  // rand_s needs _CRT_RAND_S before the first stdlib.h, set in runtime.h; it draws from the system's random source and gives 32 bits at a time.
  unsigned int low = 0, high = 0;
  if (rand_s(&low) == 0 && rand_s(&high) == 0) seed = ((uint64_t)high << 32) | low;
#else
  FILE* source = fopen("/dev/urandom", "rb");
  if (source != NULL) {
    if (fread(&seed, sizeof seed, 1, source) != 1) seed = 0;
    fclose(source);
  }
#endif
  if (seed == 0) {
    seed = (uint64_t)time(NULL) * 6364136223846793005ULL + (uint64_t)(uintptr_t)&seed;
  }
  return (int64_t)seed;
}
