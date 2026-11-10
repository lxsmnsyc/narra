# Log

- Base: feat/batch-4 at ffdea9c, `emit selfhost`, 3 runs: 9.39e9, 9.36e9, 9.35e9 instructions retired; peak 121-134 MB.
- Branch: 8.37e9, 8.29e9, 8.32e9 instructions retired; peak 115-123 MB.
- Keys packed into ints; valuesAt maps a key to a slot in valueLists so a push no longer copies the array.
