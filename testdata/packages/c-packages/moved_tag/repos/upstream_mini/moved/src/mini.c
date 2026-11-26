#include "mini.h"

int mini_add(int first, int second) {
  return first + second + 100;
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
