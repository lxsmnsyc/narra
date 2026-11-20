#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {
    volatile int64_t count = 100000;
    char *s = calloc(1, 1);
    size_t length = 0;
    for (int64_t i = 0; i < count; i++) {
        // A new string each round, like the format calls in the other versions.
        char *next = malloc(length + 21);
        length = (size_t)snprintf(next, length + 21, "%s%lld", s, (long long)(i % 10));
        free(s);
        s = next;
    }
    printf("%zu\n", strlen(s));
    free(s);
    return 0;
}
