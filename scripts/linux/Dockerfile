# Toolchain for the Linux check: a C compiler with AddressSanitizer, UBSan and
# LeakSanitizer. The compiler is built from seed/narrac.c, so nothing else is needed.
FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends gcc libc6-dev \
 && rm -rf /var/lib/apt/lists/*

ENV CC=cc

WORKDIR /repo
