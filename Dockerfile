FROM golang:1.25.6-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /collector-fe-instrumentation ./cmd/collector

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /collector-fe-instrumentation .

EXPOSE 3000

CMD ["./collector-fe-instrumentation"]
