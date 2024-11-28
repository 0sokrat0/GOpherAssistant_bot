package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Ошибка поискапеременных окружения: ", err)
	}

	TOKEN := os.Getenv("TOKEN") // получаем токен бота

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", TOKEN)

	resp, err := http.Get(url) // получаем данные
	if err != nil {
		log.Fatal("Ошибка получения данных: ", err)
		defer resp.Body.Close() // закрываем соединениес сервером

		fmt.Println("Статус ответа:", resp.Status) // выводим статус ответа
	}
}
