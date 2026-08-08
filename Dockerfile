# The build stage runs natively on the builder and Go cross-compiles to the
# target architecture, so building an arm64 image needs no QEMU emulation.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-arm64} \
    go build -trimpath -ldflags="-s -w" -o /out/blc .

# distroless/static ships CA certificates, which the HTTPS source fetches need,
# and runs as uid 65532 rather than root.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/blc /blc

ENTRYPOINT ["/blc"]
