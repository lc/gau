FROM golang:1.14

WORKDIR /app/gau

COPY . .
RUN go mod download
RUN go build -o ./build/gau

ENTRYPOINT ["/app/gau/build/gau"]
