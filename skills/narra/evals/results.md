# Skill eval results

Ran 2026-11-16 with `claude plugin eval` 2.1.325 and model claude-opus-5-5 in both arms. The command used `--ablation with-without -j 2`. There are 10 tasks (`tasks/`), 2 runs per task and 2 arms, 40 runs in all, on a machine with nothing else running.

A run passes when `narrac check` and the task's test (`checks.nra`) both pass, run by `grade.sh` on the run's working folder.

| Arm | Passed | Share | Median `narrac check` calls |
| --- | --- | --- | --- |
| With the skill | 19 of 20 | 95% | 2 |
| Without the skill | 15 of 20 | 75% | 2.5 |

The threshold fixed before the run was a pass share up by at least 15 points. It was met: the share rose by 20 points.

## Method

- Both arms got the same system prompt, which says `narrac` is on `PATH` and names the working folder.
- The harness grants `Bash`, `Write` and `Edit`. `narrac test` could not run inside its sandbox, so no agent ran the task tests. They ran only in grading.
- `prepare-eval.sh` builds a plugin copy without the `solution/` folders and `checks.nra`, so no agent could read them.
- Cost: 585 s wall time and $6.40 reported ($3.42 with the skill, $2.98 without).

## Limits

- With 20 runs per arm, a gap of 4 passes is not far from noise.
- 5 of the 6 failed runs were in the arm without the skill: syntax from other languages, such as `+` to join strings. The one failed run with the skill called a `std/test` function that does not exist.
