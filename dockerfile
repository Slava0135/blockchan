# syntax=docker/dockerfile:1

FROM golang:1.20 AS builder
COPY . /go/src/github.com/slava0135/blockchan
WORKDIR /go/src/github.com/slava0135/blockchan
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 go build -o app .

FROM alpine:3.17.2
COPY --from=builder /go/src/github.com/slava0135/blockchan/app /
ENTRYPOINT ["/app"]
