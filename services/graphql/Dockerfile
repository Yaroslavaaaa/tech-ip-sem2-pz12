FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY services/graphql/go.mod services/graphql/go.sum ./
RUN go mod download

# Копируем исходный код
COPY services/graphql/ .

# Копируем shared и pkg (если нужны)
COPY shared/ ../shared/
COPY pkg/ ../pkg/

# Сборка
WORKDIR /app/cmd/graphql
RUN go build -o graphql-server .

# Финальный образ
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/cmd/graphql/graphql-server .

EXPOSE 8090

CMD ["./graphql-server"]