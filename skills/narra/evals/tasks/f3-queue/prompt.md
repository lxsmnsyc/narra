`narrac check .` reports errors in `src/queue.nra`. Fix them. `addAll` must add the extra numbers to the caller's array, and `firstOr` must keep returning the first number or the fallback.
