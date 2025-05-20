FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o news-svc-test-task ./

FROM alpine:3.17
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/news-svc-test-task .

EXPOSE 8080

ENTRYPOINT ["./news-svc-test-task"]
