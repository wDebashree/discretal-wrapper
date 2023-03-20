## Build
FROM golang:1.19-alpine AS builder

WORKDIR /

COPY . .

RUN go build

## Deploy
FROM alpine

COPY --from=builder /pubsubapi /pubsubapi

EXPOSE 5000

ENTRYPOINT ["./pubsubapi"]