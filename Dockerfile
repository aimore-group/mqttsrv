FROM golang:1.21.0-alpine3.18 AS builder

RUN apk update
RUN apk add git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /app/mqttsrv ./cmd


FROM alpine

WORKDIR /
COPY --from=builder /app/mqttsrv .

# tcp
EXPOSE 1883

# websockets
EXPOSE 1882

# dashboard
EXPOSE 8080

ENTRYPOINT [ "/mqttsrv" ]
