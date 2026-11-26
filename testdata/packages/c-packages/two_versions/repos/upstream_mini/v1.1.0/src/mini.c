#include "mini.h"

/* Version 1.1.0 of the library. */
int mini_add(int first, int second) {
  return first + second;
}

int mini_scale(void) {
  return MINI_SCALE;
}

int mini_secret(void) {
#ifdef MINI_SECRET
  return 1;
#else
  return 0;
#endif
}
