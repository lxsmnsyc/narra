#!/bin/sh
# D113 measurement variants against the current runtime.
# Usage: WORK=<dir> sh build.sh   (WORK holds out/stage2.c from `sh selfhost/bootstrap.sh $WORK/out`)
# base and signfl use the size-class allocator; the other variants hook the
# plain malloc path, so they are built with -DNARRA_PLAIN_MALLOC.
set -eu
proto=$(cd "$(dirname "$0")" && pwd)
cd "$WORK"
cc -O2 -DNARRA_RELEASE -std=c11 -w -o v-base out/stage2.c -lm
cc -O2 -DNARRA_RELEASE -DNARRA_PLAIN_MALLOC -std=c11 -w -o v-plain out/stage2.c -lm
for m in count bump nofree nocount sign; do
  python3 "$proto/variant.py" "$m" out/stage2.c "v-$m.c"
  cc -O2 -DNARRA_RELEASE -DNARRA_PLAIN_MALLOC -std=c11 -w -o "v-$m" "v-$m.c" -lm
done
python3 "$proto/variant.py" sign out/stage2.c v-signfl.c
cc -O2 -DNARRA_RELEASE -std=c11 -w -o v-signfl v-signfl.c -lm
