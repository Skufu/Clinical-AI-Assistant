# Multi-stage build for lightweight runtime
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY index\ \(3\).html /app/index\ \(3\).html
EXPOSE 8080
CMD ["/app/server"]

