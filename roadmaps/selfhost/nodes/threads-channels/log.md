# Log: threads-channels

## Proof first

- base: cb39b8b6 (feat/batch-13), compiler built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/threads-channels/base` (-O2 -DNARRA_RELEASE).
- `base/narrac run testdata/channel_basic.nra`: exit 1, `channel_basic.nra:3:1: unknown module "std/channel"`.

## Work

- Chose a non-generic builtin `Sync` (lock, two conditions, two end counts) plus boxes in Narra over a generic `SyncSlot<T>`. The end objects need a release hook, which a value slot does not give.
- The SPEC sample's void thread function does not infer `T` (also on base, for any `call<T>(f: () => T)`), so the sample's thread function now returns a count.
- Scripts in /private/tmp/claude-501/threads-channels/: rebuild.sh, try.sh (samples, -release and -sanitize with the leak and copy checks), repeat.sh (50 runs each), compare.sh (emit comparison).
