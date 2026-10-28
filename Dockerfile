# Shared service image: ONE Dockerfile for every service in the repo.
# Build with: docker build --build-arg SERVICE=gateway -t img .
#
# Pinned base, non-root user, distroless runtime. The build arg is the only
# thing that differs between services — which is exactly what you want when
# 40 services share an image policy: one file to review, one place to change.
# Pinned by digest, not tag: a tag can be re-pushed upstream and silently
# change your build inputs. The digest is the actual pin.
# Go version is declared in go.mod; this pin must match it (see AGENTS.md).
# The image-scan gate fails on stdlib CVEs that a toolchain embeds in static
# binaries; when one lands, the fix is to bump this pin (and go.mod), not to
# ignore it. Keep the toolchain current so that gate stays green.
FROM golang:1.27-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS build
ARG SERVICE
WORKDIR /src

# Copy go.mod first so dependency layers cache across commits.
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/app ./services/${SERVICE}

# Pinned by digest, not tag: a tag can be re-pushed upstream and silently
# change your build inputs. The digest is the actual pin.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/app /app
# Every service listens on 8080 inside the container; host port mapping is
# a deployment decision (see docker-compose.yml).
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
