FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/main ./...

FROM alpine:latest

WORKDIR /app/

COPY --from=builder /app/bin/main .

COPY --from=builder /app/templates ./templates

RUN chmod +x /app/main

EXPOSE 8080

CMD ["/app/main"]
