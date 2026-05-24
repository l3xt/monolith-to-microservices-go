# Этап 1: сборка
FROM golang:1.25.2-alpine AS builder

WORKDIR /app

# Git нужен для загрузки некоторых Go-модулей
RUN apk add --no-cache git

# Копируем корневой go.mod и модули конкретного сервиса
COPY users-service/go.mod users-service/go.sum ./users-service/

# Копируем исходный код общих пакетов
COPY pkg/ ./pkg/

# Создаем Go Workspace и подключаем оба модуля (сервис и общие пакеты)
RUN go work init && go work use ./pkg ./users-service
RUN cd users-service && go mod download

# Копируем исходный код конкретного сервиса
COPY users-service/ ./users-service/

# Статический бинарник для Linux (точка входа — cmd/server, как в этапах проекта)
RUN CGO_ENABLED=0 GOOS=linux go build -o /users-service ./users-service/cmd/server

# Этап 2: рантайм
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /users-service .

EXPOSE 8081

CMD ["./users-service"]
