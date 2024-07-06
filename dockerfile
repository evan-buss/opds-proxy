# Dockerfile.distroless
FROM golang:1.22 as base

WORKDIR /src/kobo-opds-proxy/app/

COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

FROM gcr.io/distroless/static

COPY --from=base /src/kobo-opds-proxy/app/kobo-opds-proxy .

CMD ["./kobo-opds-proxy"]