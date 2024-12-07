# Используем официальный образ Go
FROM golang:1.23.4-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта
COPY . .

# Загружаем зависимости
RUN go mod tidy

# Собираем бинарный файл
RUN go build -o bot .

# Указываем переменные среды
ENV BOT_TOKEN=8130389933:AAGdMGjRpoLoVjhy_i2WLwPJ7tr3F-_kxYk

# Запускаем приложение
CMD ["./bot"]
