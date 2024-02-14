FROM golang:1.21-alpine3.18 as builder

WORKDIR /app

COPY ./go.mod ./
RUN go mod download
ADD . /app/

RUN GOOS=linux go build ./cmd/go-http

FROM alpine:3.18 as prod
WORKDIR /app
COPY --from=builder /app/go-http /app/go-http

ENTRYPOINT /app/go-http