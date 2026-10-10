FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o cas-server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cas-server .
EXPOSE 8082
ENTRYPOINT ["./cas-server"]
