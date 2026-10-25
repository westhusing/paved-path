# Shared service image: ONE Dockerfile for every service in the repo.
# Build with: docker build --build-arg SERVICE=gateway -t img .
#
# Pinned base, non-root user, distroless runtime. The build arg is the only
# thing that differs between services — which is exactly what you want when
# 40 services share an image policy: one file to review, one place to change.
# Pinned by digest, not tag: a tag can be re-pushed upstream and silently
# change your build inputs. The digest is the actual pin.
# Go version is declared in go.mod; this pin must match it (see AGENTS.md).
# 1.27, not just "new enough": the stdlib CVEs that older toolchains embed in
# static binaries are what image scanners flag, and 1.27's stdlib is clean.
FROM golang:1.27-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS build
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
