# Log: shared-results-walk

- Base cost measured with `checkSharedResults` returning at once: +0.080G (+3.8%) on `emit selfhost`.
- Disabled parts one at a time to locate the cost: `noteMutableArguments` first. Its extra cost came from `addGiven` copying a growing array per call, held-type set copies and `typeKey` strings for scalar arguments.
- Gates: skip bodies with no arrow function and no way to share a parameter into the result.
- A Narra function cannot go unreported when the walk is skipped: reports need a stop local that is a parameter (functions) or an arrow.
- Scratch: /private/tmp/claude-501/shared-results-walk/.
- sh scripts/test.sh: PASS in 146s (8 shards plus bootstrap, 0 failed). Log: /private/tmp/claude-501/shared-results-walk/test.log
