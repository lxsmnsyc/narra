# Premortem: batch-12

It is two weeks from now and batch-12 is late or reverted. Why?

1. **The async chain ran long.** async-design (gate), async-core (L), then event-loop or async-api, then std-async-io is a hard chain of four steps, and the gate waits for the maintainer. Change: async-design runs in wave 1 beside the fixes, so the maintainer sees D109 early. If async-core grows past L, async-design splits it (syntax and checks first, emit and runtime second). If the gate is not answered, the fixes and review-nits ship as the batch and async moves to batch-13.
2. **Tasks leaked or crashed under the sanitizers.** A pending task holds a frame, captured values and callbacks; cancel at exit and `then` callbacks dropped after a cancel are where counts go wrong. Change: every async sample runs under AddressSanitizer and the leak check, including tasks still running at exit and cancelled handles; D109 states how the leak check sees pending tasks.
3. **epoll behaved differently from kqueue.** Linux is checked only in a container, late. Change: event-loop and std-async-io run `sh scripts/linux-check.sh` before landing, not only at the end of the batch.
4. **Network samples were flaky in parallel shards.** Fixed ports collide across shards. Change: samples bind port 0 on loopback and read the port back.
5. **The generator transform broke.** async-core shares `emitgenerators.nra`; a change there can break generators and coroutines. Change: async-core keeps generator behavior byte-identical (`narrac emit selfhost` output unchanged for code with no async) and runs the generator and coroutine suites.

Accepted risk: `diagnostics.nra` and the async runtime module are edited by several nodes; the hard chain orders the async ones, and the table rows sit in reserved code ranges, so rebases stay mechanical.
