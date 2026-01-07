FROM golang:1.24 AS base

# Set by buildx
ARG TARGETPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT

# Download and install kepubify
RUN set -e; \
    if [ "$TARGETOS" != "linux" ]; then \
    printf 'Unsupported TARGETOS=%s for kepubify\n' "$TARGETOS" >&2; \
    exit 1; \
    fi; \
    case "$TARGETARCH/$TARGETVARIANT" in \
    amd64/) KEPUBIFY_ARCH=64bit ;; \
    386/) KEPUBIFY_ARCH=32bit ;; \
    arm64/v8|arm64/) KEPUBIFY_ARCH=arm64 ;; \
    arm/v6) KEPUBIFY_ARCH=armv6 ;; \
    arm/v7|arm/) KEPUBIFY_ARCH=arm ;; \
    *) printf 'Unsupported TARGETARCH=%s TARGETVARIANT=%s for kepubify\n' "$TARGETARCH" "$TARGETVARIANT" >&2; exit 1 ;; \
    esac; \
    KEPUBIFY_NAME="kepubify-${TARGETOS}-${KEPUBIFY_ARCH}"; \
    wget "https://github.com/pgaskin/kepubify/releases/download/v4.0.4/${KEPUBIFY_NAME}" && \
    mv "${KEPUBIFY_NAME}" /usr/local/bin/kepubify && \
    chmod +x /usr/local/bin/kepubify

# Download and install kindlegen
RUN set -e; \
    target_platform="${TARGETPLATFORM:-${TARGETOS}/${TARGETARCH}}"; \
    case "$target_platform" in \
      linux/amd64|linux/386) \
        wget https://web.archive.org/web/20150803131026if_/https://kindlegen.s3.amazonaws.com/kindlegen_linux_2.6_i386_v2_9.tar.gz && \
        mkdir kindlegen && \
        tar xvf kindlegen_linux_2.6_i386_v2_9.tar.gz --directory kindlegen && \
        cp kindlegen/kindlegen /usr/local/bin/kindlegen && \
        chmod +x /usr/local/bin/kindlegen ;; \
      *) \
        printf 'Skipping kindlegen for %s\n' "$target_platform" ;; \
    esac

WORKDIR /src/opds-proxy/app/

COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go mod verify

COPY . .

ARG VERSION=dev
ARG REVISION=unknown
ARG BUILDTIME=unknown

RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${REVISION} -X main.date=${BUILDTIME}" -o opds-proxy

RUN mkdir -p /out/usr/local/bin && \
    cp /usr/local/bin/kepubify /out/usr/local/bin/ && \
    if [ -f /usr/local/bin/kindlegen ]; then cp /usr/local/bin/kindlegen /out/usr/local/bin/; fi && \
    cp /src/opds-proxy/app/opds-proxy /out/opds-proxy

FROM gcr.io/distroless/static

COPY --from=base /out/ /

ENTRYPOINT ["./opds-proxy"]
