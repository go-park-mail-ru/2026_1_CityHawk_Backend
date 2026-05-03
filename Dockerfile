FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG BUILD_TARGET=./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -o cityhawk-service ${BUILD_TARGET}

FROM alpine:3.20

WORKDIR /app
RUN adduser -D -g '' appuser

COPY --from=builder /app/cityhawk-service ./cityhawk-service

ENV PORT=8080
EXPOSE 8080 50051 50052 50053 50054 50055 9101 9102 9103 9104 9105

USER appuser
CMD ["./cityhawk-service"]
