FROM golang:1.25-rc-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
RUN apk add --no-cache ca-certificates
EXPOSE 8080

CMD ["./app"]
