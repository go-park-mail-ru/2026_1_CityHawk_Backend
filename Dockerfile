FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cityhawk-backend ./cmd

FROM alpine:3.20

WORKDIR /app
RUN adduser -D -g '' appuser

COPY --from=builder /app/cityhawk-backend ./cityhawk-backend

ENV PORT=8080
EXPOSE 8080

USER appuser
CMD ["./cityhawk-backend"]
