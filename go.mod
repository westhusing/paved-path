module example.com/paved-path

// Pinned to the exact patch, not just "1.27": setup-go (GitHub) reads this
// line and installs that exact toolchain, so CI builds with the same Go as
// the Dockerfile's digest-pinned image (1.27.2). A bare "go 1.27" would let
// setup-go float to the newest 1.27 patch and drift from the image.
go 1.27.2
