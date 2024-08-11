FROM golang:1.22 AS base

RUN
# Download and install kepubify
RUN wget https://github.com/pgaskin/kepubify/releases/download/v4.0.4/kepubify-linux-64bit && \
    mv kepubify-linux-64bit /usr/local/bin/kepubify && \
    chmod +x /usr/local/bin/kepubify

# Download and install kindlegen
RUN wget https://web.archive.org/web/20150803131026if_/https://kindlegen.s3.amazonaws.com/kindlegen_linux_2.6_i386_v2_9.tar.gz && \
    mkdir kindlegen && \
    tar xvf kindlegen_linux_2.6_i386_v2_9.tar.gz --directory kindlegen && \
    cp kindlegen/kindlegen /usr/local/bin/kindlegen && \
    chmod +x /usr/local/bin/kindlegen

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

FROM gcr.io/distroless/static

COPY --from=base /usr/local/bin/kepubify /usr/local/bin/kepubify
COPY --from=base /usr/local/bin/kindlegen /usr/local/bin/kindlegen
COPY --from=base /src/opds-proxy/app/opds-proxy .

ENTRYPOINT ["./opds-proxy"]