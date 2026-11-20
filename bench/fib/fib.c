#include <stdint.h>
#include <stdio.h>

static int64_t fib(int64_t n) {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

int main(void) {
    volatile int64_t n = 38;
    printf("%lld\n", (long long)fib(n));
    return 0;
}
