#include <stdio.h>

/* The defines of the tables package must not reach this file. */
#ifdef EXTRA
#error "a define of another package leaked into wrapper"
#endif

static char buffer[64];

const char *wrap_describe(int value, int scale) {
  snprintf(buffer, sizeof buffer, "value=%d scale=%d", value, scale);
  return buffer;
}
