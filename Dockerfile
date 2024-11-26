FROM golang:1.22.2-alpine3.19 AS builder

WORKDIR /app

RUN apk update && apk add --no-cache git
RUN apk add --no-cache gcc musl-dev curl

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate /usr/local/bin/migrate

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN cp .env.example .env
RUN go test -v ./...
RUN CGO_ENABLED=1 go build -o /app/main ./cmd/main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache curl
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/internal/infra/db/schema/migrations ./internal/infra/db/schema/migrations

EXPOSE 8080

CMD ["./main"]