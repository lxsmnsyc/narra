#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

static int64_t *multiply(const int64_t *a, const int64_t *b, size_t n) {
    int64_t *c = calloc(n * n, sizeof *c);
    for (size_t i = 0; i < n; i++) {
        for (size_t j = 0; j < n; j++) {
            int64_t sum = 0;
            for (size_t k = 0; k < n; k++) {
                sum += a[i * n + k] * b[k * n + j];
            }
            c[i * n + j] = sum;
        }
    }
    return c;
}

int main(void) {
    volatile size_t size = 600;
    size_t n = size;
    int64_t *a = calloc(n * n, sizeof *a);
    int64_t *b = calloc(n * n, sizeof *b);
    for (size_t i = 0; i < n; i++) {
        for (size_t j = 0; j < n; j++) {
            a[i * n + j] = (int64_t)((i + j) % 10);
            b[i * n + j] = (int64_t)((i * j) % 10);
        }
    }
    int64_t *c = multiply(a, b, n);
    int64_t total = 0;
    for (size_t k = 0; k < n * n; k++) {
        total += c[k];
    }
    printf("%lld\n", (long long)total);
    free(a);
    free(b);
    free(c);
    return 0;
}
