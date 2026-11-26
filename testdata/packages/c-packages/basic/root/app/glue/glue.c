#include <mini.h>

int app_glue_sum(int first, int second) {
  return mini_add(first, second) + mini_twice(first);
}

int app_glue_has_secret(void) {
#ifdef MINI_SECRET
  return 1;
#else
  return 0;
#endif
}
