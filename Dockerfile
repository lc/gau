# Release image: alpine:3.17
FROM alpine:3.17

RUN apk -U upgrade --no-cache
COPY gau /usr/local/bin/gau

RUN adduser \
    --gecos "" \
    --disabled-password \
    gau

USER gau
ENTRYPOINT ["gau"]
