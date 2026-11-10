# log
- Branch renamed, reset to feat/batch-6. None of the functions existed.
- Design: builtin modules cannot hold Narra source, so a Narra part std/ext/<name> (file std/src/ext__<name>.nra) is loaded beside std/<name> and found after the builtins.
- divmod and floorDiv return Option (no divisor proof needed).
