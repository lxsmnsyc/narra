#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

// Eight scalar accumulators, summed in the order of Simd.sum in dot.nra.
static float dot(const float *left, const float *right, size_t count, size_t passes) {
    float s[8] = {0};
    for (size_t pass = 0; pass < passes; pass++) {
        for (size_t i = 0; i + 8 <= count; i += 8) {
            for (int k = 0; k < 8; k++) {
                s[k] += left[i + k] * right[i + k];
            }
        }
    }
    return ((s[0] + s[4]) + (s[2] + s[6])) + ((s[1] + s[5]) + (s[3] + s[7]));
}

int main(void) {
    size_t count = 1048576;
    float *left = malloc(count * sizeof *left);
    float *right = malloc(count * sizeof *right);
    for (int64_t i = 0; i < (int64_t)count; i++) {
        left[i] = (float)(i % 16 - 8) * 0.125f;
        right[i] = (float)(i % 7 - 3) * 0.25f;
    }
    volatile size_t passes = 200;
    float result = dot(left, right, count, passes);
    printf("%lld\n", (long long)(result * 32.0f));
    free(left);
    free(right);
    return 0;
}
