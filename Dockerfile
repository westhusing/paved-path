# Shared service image: ONE Dockerfile for every service in the repo.
# Build with: docker build --build-arg SERVICE=gateway -t img .
#
# Pinned base, non-root user, distroless runtime. The build arg is the only
# thing that differs between services — which is exactly what you want when
# 40 services share an image policy: one file to review, one place to change.
FROM golang:1.24-alpine AS build
ARG SERVICE
WORKDIR /src

# Copy go.mod first so dependency layers cache across commits.
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/app ./services/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
