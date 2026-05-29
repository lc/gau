# Build image: golang:1.21.0-alpine3.17
FROM golang:1.21.0-alpine3.17 as build

WORKDIR /app

COPY . .
RUN go mod download && go build -o ./build/gau ./cmd/gau

ENTRYPOINT ["/app/gau/build/gau"]

# Release image: alpine:3.17
FROM alpine:3.17

RUN apk -U upgrade --no-cache
COPY --from=build /app/build/gau /usr/local/bin/gau

RUN adduser \
    --gecos "" \
    --disabled-password \
    gau

USER gau
ENTRYPOINT ["gau"]
