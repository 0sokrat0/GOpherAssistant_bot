# Используем официальный образ Go
FROM golang:1.23.4-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем `go.mod` и `go.sum` для загрузки зависимостей
COPY app/go.mod app/go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем весь код проекта
COPY app ./

# Собираем приложение
RUN go build -o gopher-assistant-bot ./cmd/main.go

# Указываем порт, который будет слушать приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./gopher-assistant-bot"]
