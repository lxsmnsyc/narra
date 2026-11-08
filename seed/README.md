# The seed

`seed/narrac.c` is the Narra compiler, written out as C. The compiler is written in Narra, so something has to build it the first time. This file does that. Any machine with a C compiler (`cc`) can build it. Nothing else is needed.

Never edit `narrac.c` by hand. It is generated. Its first lines say how it is made.

## How to build the compiler

Run `sh selfhost/bootstrap.sh`. It does three things:

1. It builds the seed with `cc`. This gives a first compiler.
2. The first compiler compiles the source in `selfhost/`. This gives a second compiler.
3. The second compiler compiles `selfhost/` again. The C from step 2 and from step 3 must be the same. If they differ, the script fails.

Set `CC` to pick another C compiler. CC must name a single program: it is not split into words, so `CC="ccache cc"` does not work. Use a wrapper script instead.

## The seed rule

The compiler's source may only use features that the seed can compile.

Say you add a new feature to the language, such as a new kind of loop. The seed does not know it yet. So:

1. Add the feature to the compiler. Write that code without using the feature itself.
2. Regenerate the seed (see below). Now the seed knows the feature.
3. Only now may the compiler's own source use the feature.

If you skip step 2 and use the feature early, the bootstrap fails. The seed cannot read the new code.

## How to regenerate the seed

Run `sh scripts/seed-update.sh`. It runs the bootstrap from the current seed. If that passes, it replaces `seed/narrac.c` with the new C. If the source did not change, the file stays the same, byte for byte.

Then commit the new `seed/narrac.c`.

## When to regenerate

- When the compiler source wants to use a feature the seed does not have yet.
- At each release.

Do not regenerate on every commit. The file is large, and each change adds a big diff to the history.

## Diffs

`.gitattributes` marks the seed as generated. Git does not show its diff, and GitHub leaves it out of language stats.
