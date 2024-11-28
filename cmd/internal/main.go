package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Chat struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type AutoReply struct {
	Trigger  string `json:"trigger"`
	Response string `json:"response"`
}

var (
	filename    = "db.json"
	adminID     int64 // ID администратора, задайте позже
	botToken    string
	autoReplies []AutoReply
)

func main() {
	// Загружаем переменные окружения
	err := godotenv.Load("bot.env")
	if err != nil {
		log.Fatal("Ошибка загрузки переменных окружения:", err)
	}

	botToken = os.Getenv("TOKEN")
	if botToken == "" {
		log.Fatal("Токен бота не установлен.")
	}

	adminID = parseAdminID()

	autoReplies, err = loadAutoReplies(filename)
	if err != nil {
		log.Fatalf("Ошибка загрузки автоответов: %v", err)
	}

	offset := 0
	for {
		updates, err := getUpdates(offset)
		if err != nil {
			log.Printf("Ошибка получения обновлений: %s", err)
			continue
		}

		for _, update := range updates {
			log.Printf("Обрабатываю сообщение ID: %d, Текст: %s", update.UpdateID, update.Message.Text)

			// Автоответ
			for _, reply := range autoReplies {
				if strings.EqualFold(update.Message.Text, reply.Trigger) {
					log.Printf("Триггер найден: %s. Отправляю ответ: %s", reply.Trigger, reply.Response)
					err := sendMessage(update.Message.Chat.ID, reply.Response)
					if err != nil {
						log.Printf("Ошибка отправки сообщения: %s", err)
					}
					break
				}
			}
		}

		if len(updates) > 0 {
			offset = updates[len(updates)-1].UpdateID + 1
		}
	}
}

func parseAdminID() int64 {
	admin := os.Getenv("ADMIN_ID")
	if admin == "" {
		log.Fatal("ADMIN_ID не установлен в .env файле.")
	}

	var id int64
	_, err := fmt.Sscanf(admin, "%d", &id)
	if err != nil {
		log.Fatal("ADMIN_ID должен быть числом.")
	}

	return id
}

func processUpdate(update Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	// Обработка команд администратора
	if chatID == adminID {
		handleAdminCommands(chatID, text)
		return
	}

	// Обработка автоответов для других пользователей
	if update.Message.Chat.Type == "private" {
		for _, reply := range autoReplies {
			if strings.EqualFold(reply.Trigger, text) {
				sendMessage(chatID, reply.Response)
				break
			}
		}
	}
}

func handleAdminCommands(chatID int64, text string) {
	switch text {
	case "/start":
		sendMessageWithButtons(chatID, "Добро пожаловать! Выберите действие:", [][]string{
			{"Add auto reply ✉️", "Remove auto reply 🚫"},
		})
	case "Add auto reply ✉️":
		sendMessage(chatID, "Введите триггер для автоответа:")
	case "Remove auto reply 🚫":
		listAutoReplies(chatID)
	default:
		if strings.HasPrefix(text, "REMOVE:") {
			trigger := strings.TrimPrefix(text, "REMOVE:")
			removeAutoReply(chatID, strings.TrimSpace(trigger))
		} else {
			addAutoReply(chatID, text)
		}
	}
}

func sendMessage(chatID int64, text string) error {
	params := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	_, err := callTelegramAPI("sendMessage", params)
	return err
}

func sendMessageWithButtons(chatID int64, text string, buttons [][]string) {
	keyboard := map[string]interface{}{
		"keyboard":          buttons,
		"resize_keyboard":   true,
		"one_time_keyboard": true,
	}

	params := map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": keyboard,
	}

	_, err := callTelegramAPI("sendMessage", params)
	if err != nil {
		log.Println("Ошибка отправки сообщения с кнопками:", err)
	}
}

func listAutoReplies(chatID int64) {
	if len(autoReplies) == 0 {
		sendMessage(chatID, "Список автоответов пуст.")
		return
	}

	var list string
	for _, reply := range autoReplies {
		list += fmt.Sprintf("Trigger: %s\nResponse: %s\n\n", reply.Trigger, reply.Response)
	}

	sendMessage(chatID, list+"\nЧтобы удалить, отправьте: REMOVE:<trigger>")
}

func addAutoReply(chatID int64, text string) {
	if len(autoReplies) > 0 {
		last := &autoReplies[len(autoReplies)-1]
		if last.Response == "" {
			last.Response = text
			saveAutoReplies()
			sendMessage(chatID, "Автоответ успешно добавлен!")
			return
		}
	}

	autoReplies = append(autoReplies, AutoReply{Trigger: text})
	sendMessage(chatID, "Введите ответ на этот триггер.")
}

func removeAutoReply(chatID int64, trigger string) {
	for i, reply := range autoReplies {
		if strings.EqualFold(reply.Trigger, trigger) {
			autoReplies = append(autoReplies[:i], autoReplies[i+1:]...)
			saveAutoReplies()
			sendMessage(chatID, "Автоответ успешно удалён!")
			return
		}
	}

	sendMessage(chatID, "Триггер не найден.")
}

func loadAutoReplies(filename string) ([]AutoReply, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var replies []AutoReply
	err = json.Unmarshal(data, &replies)
	return replies, err
}

func saveAutoReplies() {
	data, err := json.Marshal(autoReplies)
	if err != nil {
		log.Println("Ошибка сохранения автоответов:", err)
		return
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Println("Ошибка записи автоответов в файл:", err)
	}
}

func getUpdates(offset int) ([]Update, error) {
	params := map[string]string{
		"offset": fmt.Sprintf("%d", offset),
		"limit":  "5",
	}

	response, err := callTelegramAPI("getUpdates", params)
	if err != nil {
		return nil, err
	}

	var updatesResponse UpdatesResponse
	err = json.Unmarshal(response, &updatesResponse)
	if err != nil {
		return nil, err
	}

	return updatesResponse.Result, nil
}

func callTelegramAPI(method string, params interface{}) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)

	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации параметров: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API Telegram: %s", body)
	}

	return body, nil
}
