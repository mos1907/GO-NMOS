FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-nmos-api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN adduser -D appuser
COPY --from=builder /go-nmos-api /app/go-nmos-api
COPY --from=builder /app/migrations /app/migrations
USER appuser
EXPOSE 8080
CMD ["/app/go-nmos-api"]
