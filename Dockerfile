FROM golang:alpine AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /root/zoneinfo.zip
ENV ZONEINFO=/zoneinfo.zip

CMD ["./server"]