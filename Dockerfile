FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o ec2diff .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ec2diff .
ENTRYPOINT ["./ec2diff"]
