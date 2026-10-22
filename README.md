# Storage-network starter

This is your independent exercise workspace for **Building a Storage Network**.
It contains tests and unfinished function declarations, not completed solutions.

## Set up

1. Install Go 1.24 or later.
2. Extract this folder somewhere outside the book's repository.
3. Open a terminal here and initialise your own Git repository:

    git init
    git add .
    git commit -m "Start storage-network exercises"

The starter has no Git history, remote, completed app or reference implementations.
You can rename this folder and add your own remote. Never commit private recovery keys or node data.

## Work one test group at a time

    go test -run '^TestFrames$' -v
    go test -run '^TestChunks$' -v
    go test -run '^TestErasure$' -v
    go test -run '^TestEncryption$' -v
    go test -run '^TestStore$' -v
    go test -race -run '^TestWorkers$' -v
    go test -run '^TestHealth$' -v
    go test -run '^TestLedger$' -v

Tests fail until the relevant functions in workshop.go are implemented. The encryption tests also require the erasure functions. Store tests require Digest, a lowercase SHA-256 helper. The framing kernel is a four-byte big-endian payload length, capped at 8 MiB, followed by the exact payload.

The Store exercise assumes one Store instance owns a trusted directory; configure its Stage failure hook before concurrent use. Worker callbacks must cooperate with context cancellation. These are focused mechanisms; design the integrated network in your own packages as you progress through the book.

For hints and optional reference solutions, return to the relevant chapter or Appendix B in the book.
