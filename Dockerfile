# Build image: golang:1.14-alpine3.13
FROM golang@sha256:ef409ff24dd3d79ec313efe88153d703fee8b80a522d294bb7908216dc7aa168 as build

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o ./build/gau

ENTRYPOINT ["/app/gau/build/gau"]

# Image: alpine:3.14.1
FROM alpine@sha256:be9bdc0ef8e96dbc428dc189b31e2e3b05523d96d12ed627c37aa2936653258c

RUN apk -U upgrade --no-cache
COPY --from=build /app/build/gau /usr/local/bin/gau

RUN adduser \
    --gecos "" \
    --disabled-password \
    gau

USER gau
ENTRYPOINT ["gau"]
