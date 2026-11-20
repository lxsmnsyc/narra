#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static int64_t count_primes(size_t n) {
    int64_t *marks = malloc(n * sizeof *marks);
    for (size_t k = 0; k < n; k++) {
        marks[k] = 1;
    }
    marks[0] = 0;
    marks[1] = 0;
    for (size_t i = 2; i * i < n; i++) {
        if (marks[i] == 1) {
            for (size_t j = i * i; j < n; j += i) {
                marks[j] = 0;
            }
        }
    }
    int64_t count = 0;
    for (size_t k = 0; k < n; k++) {
        count += marks[k];
    }
    free(marks);
    return count;
}

int main(void) {
    volatile size_t n = 20000000;
    printf("%lld\n", (long long)count_primes(n));
    return 0;
}
