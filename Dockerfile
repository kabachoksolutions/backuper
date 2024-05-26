FROM golang:1.21.4-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG PG_VERSION='16'
RUN apk add --update --no-cache postgresql${PG_VERSION}-client --repository=https://dl-cdn.alpinelinux.org/alpine/edge/main

RUN go build -o worker ./*.go

CMD ["./worker"]
