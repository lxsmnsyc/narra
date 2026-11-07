# Toolchain for the Linux fast check. Go version matches compiler/go.mod.
FROM golang:1.23-bookworm

# gcc with AddressSanitizer, UBSan and LeakSanitizer; the tests call `cc`.
RUN apt-get update \
 && apt-get install -y --no-install-recommends gcc libc6-dev \
 && rm -rf /var/lib/apt/lists/*

ENV CC=cc \
    GOCACHE=/cache/go-build \
    GOMODCACHE=/cache/go-mod \
    GOFLAGS=-buildvcs=false

WORKDIR /repo/compiler
