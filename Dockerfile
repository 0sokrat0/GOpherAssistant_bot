# Используем официальный образ Go
FROM golang:1.23.4-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем модульные файлы
COPY app/go.mod app/go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем весь код проекта
COPY app ./

# Копируем файл конфигурации из абсолютного пути
COPY config/config.yaml ./config/config.yaml

# Собираем приложение
RUN go build -o gopher-assistant-bot ./cmd/main.go

# Указываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./gopher-assistant-bot"]
