# Этап 1: сборка
FROM golang:1.25.2-alpine AS builder

WORKDIR /app

# Git нужен для загрузки некоторых Go-модулей
RUN apk add --no-cache git

# Копируем корневой go.mod и модули конкретного сервиса
COPY books-service/go.mod books-service/go.sum ./books-service/

# Копируем исходный код общих пакетов
COPY pkg/ ./pkg/

# Создаем Go Workspace и подключаем оба модуля (сервис и общие пакеты)
RUN go work init && go work use ./pkg ./books-service
RUN cd books-service && go mod download

# Копируем исходный код конкретного сервиса
COPY books-service/ ./books-service/

# Статический бинарник для Linux (точка входа — cmd/server, как в этапах проекта)
RUN CGO_ENABLED=0 GOOS=linux go build -o /books-service ./books-service/cmd/server

# Этап 2: рантайм
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /books-service .

EXPOSE 8082

CMD ["./books-service"]
