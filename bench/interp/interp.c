#include <stdint.h>
#include <stdio.h>

typedef enum { OP_ADD, OP_SUB, OP_MUL, OP_MOD, OP_NOP } OpKind;

typedef struct {
    OpKind kind;
    int64_t n;
} Op;

static int64_t step(Op op, int64_t acc) {
    switch (op.kind) {
    case OP_ADD:
        return acc + op.n;
    case OP_SUB:
        return acc - op.n;
    case OP_MUL:
        return acc * op.n;
    case OP_MOD:
        return acc % op.n;
    case OP_NOP:
        return acc;
    }
    return acc;
}

int main(void) {
    Op program[1000];
    for (int64_t i = 0; i < 1000; i++) {
        switch (i % 5) {
        case 0:
            program[i] = (Op){OP_ADD, i};
            break;
        case 1:
            program[i] = (Op){OP_MUL, 3};
            break;
        case 2:
            program[i] = (Op){OP_SUB, 7};
            break;
        case 3:
            program[i] = (Op){OP_MOD, 1000003};
            break;
        default:
            program[i] = (Op){OP_NOP, 0};
            break;
        }
    }
    volatile int64_t rounds = 150000;
    int64_t acc = 1;
    for (int64_t round = 0; round < rounds; round++) {
        for (int k = 0; k < 1000; k++) {
            acc = step(program[k], acc);
        }
    }
    printf("%lld\n", (long long)acc);
    return 0;
}
