package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Замените на ваш токен бота и ID администратора
const (
	BotToken = "8130389933:AAF6UbNo6KYCF3mQiK-1T2dsfIiMdlvUaTI" // Замените на ваш токен бота
	AdminID  = 575225733                                        // Замените на ваш Telegram user ID
)

// Структуры для базы данных
type Answer struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Caption string `json:"caption"`
}

type DataItem struct {
	Text    string   `json:"text"`
	Answers []Answer `json:"answers"`
}

type Database struct {
	Step string     `json:"step"`
	Data []DataItem `json:"data"`
}

// Структуры для бизнес-сообщений
type BusinessMessage struct {
	MessageID            int                       `json:"message_id"`
	From                 *tgbotapi.User            `json:"from"`
	BusinessConnectionID int64                     `json:"business_connection_id"`
	Chat                 *tgbotapi.Chat            `json:"chat"`
	Date                 int                       `json:"date"`
	Text                 string                    `json:"text"`
	Entities             *[]tgbotapi.MessageEntity `json:"entities,omitempty"`
	Caption              string                    `json:"caption,omitempty"`
	// Добавьте другие необходимые поля, если нужно
}

// Кастомная структура обновления
type CustomUpdate struct {
	UpdateID        int               `json:"update_id"`
	Message         *tgbotapi.Message `json:"message,omitempty"`
	BusinessMessage *BusinessMessage  `json:"business_message,omitempty"`
}

// Функции для загрузки и сохранения базы данных
func loadDatabase() (*Database, error) {
	db := &Database{}
	file, err := ioutil.ReadFile("db.json")
	if err != nil {
		if os.IsNotExist(err) {
			db.Step = ""
			db.Data = []DataItem{}
			return db, nil
		}
		return nil, err
	}
	err = json.Unmarshal(file, db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func saveDatabase(db *Database) error {
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("db.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Установка вебхука
	// Замените на ваш актуальный ngrok URL
	webhookURL := "https://a13a-91-207-171-60.ngrok-free.app/" + bot.Token

	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	// Запускаем сервер без TLS
	go func() {
		err = http.ListenAndServe(":8443", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Start listening for updates...")

	for update := range updates {
		// Преобразуем обновление в JSON и обратно, чтобы захватить BusinessMessage
		rawData, err := json.Marshal(update)
		if err != nil {
			log.Println("Ошибка сериализации обновления:", err)
			continue
		}

		var customUpdate CustomUpdate
		err = json.Unmarshal(rawData, &customUpdate)
		if err != nil {
			log.Println("Ошибка десериализации обновления:", err)
			continue
		}

		handleUpdate(bot, &customUpdate)
	}
}

// Обработка обновлений
func handleUpdate(bot *tgbotapi.BotAPI, update *CustomUpdate) {
	db, err := loadDatabase()
	if err != nil {
		log.Println("Ошибка загрузки базы данных:", err)
		return
	}

	// Обработка сообщений от администратора
	if update.Message != nil && update.Message.From.ID == AdminID {
		handleAdminMessage(bot, update.Message, db)
	}

	// Обработка бизнес-сообщений
	if update.BusinessMessage != nil && update.BusinessMessage.BusinessConnectionID != 0 {
		handleBusinessMessage(bot, update.BusinessMessage, db)
	}
}

// Функция обработки сообщений от администратора
func handleAdminMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *Database) {
	chatID := message.Chat.ID
	text := message.Text

	// Определение клавиатур
	homeKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Add auto reply ✉️"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("remove auto reply 🚫"),
		),
	)
	homeKeyboard.ResizeKeyboard = true

	backKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Back 🔙"),
		),
	)
	backKeyboard.ResizeKeyboard = true

	doneKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Done!"),
		),
	)
	doneKeyboard.ResizeKeyboard = true

	// Обработка команд и шагов
	switch text {
	case "/start":
		msg := tgbotapi.NewMessage(chatID, "Привет! Добро пожаловать в бот-менеджер бизнес-аккаунта! 🤖\n\nЧтобы использовать бота, перейдите в раздел бизнес-аккаунта в вашем профиле Telegram, выберите раздел чат-бота и введите имя пользователя бота. 💼\n\nПримечание: Только премиум пользователи могут использовать эту опцию. ℹ️")
		msg.ReplyMarkup = homeKeyboard
		bot.Send(msg)
		db.Step = ""
		saveDatabase(db)
	case "Back 🔙", "Done!":
		msg := tgbotapi.NewMessage(chatID, "Вы вернулись в главное меню.")
		msg.ReplyMarkup = homeKeyboard
		bot.Send(msg)
		db.Step = ""
		saveDatabase(db)
	case "Add auto reply ✉️":
		msg := tgbotapi.NewMessage(chatID, "Чтобы установить автоответ, введите сообщение, на которое бот должен отвечать (на следующем шаге вы отправите ответ на этот текст)")
		msg.ReplyMarkup = backKeyboard
		bot.Send(msg)
		db.Step = "add-1"
		saveDatabase(db)
	case "remove auto reply 🚫":
		if len(db.Data) > 0 {
			var list strings.Builder
			for _, item := range db.Data {
				list.WriteString(fmt.Sprintf("<code>%s</code>\n---\n", item.Text))
			}
			msg1 := tgbotapi.NewMessage(chatID, list.String())
			msg1.ParseMode = "HTML"
			bot.Send(msg1)

			msg2 := tgbotapi.NewMessage(chatID, "Чтобы удалить элемент из автоответов, скопируйте и вставьте один из вышеперечисленных")
			msg2.ReplyMarkup = backKeyboard
			bot.Send(msg2)

			db.Step = "remove"
			saveDatabase(db)
		} else {
			msg := tgbotapi.NewMessage(chatID, "Список автоответов пуст!")
			msg.ReplyMarkup = homeKeyboard
			bot.Send(msg)
		}
	default:
		// Обработка шагов
		switch db.Step {
		case "add-1":
			// Сохранение триггерного текста
			newItem := DataItem{
				Text:    text,
				Answers: []Answer{},
			}
			db.Data = append(db.Data, newItem)
			db.Step = "add-2"
			saveDatabase(db)

			msg := tgbotapi.NewMessage(chatID, "✅ Успешно создано.\n\nОтправьте контент для ответа на этот текст (может включать любой тип контента: текст, фото, видео, GIF, стикер, голосовое сообщение и т.д.)")
			msg.ReplyMarkup = backKeyboard
			bot.Send(msg)
		case "add-2":
			// Получение последнего элемента
			if len(db.Data) == 0 {
				db.Step = ""
				saveDatabase(db)
				return
			}
			lastIndex := len(db.Data) - 1
			lastItem := &db.Data[lastIndex]

			// Проверка типа сообщения и сохранение ответа
			var answer Answer
			if message.Text != "" {
				answer.Type = "text"
				answer.Content = message.Text
			} else if message.Sticker != nil {
				answer.Type = "sticker"
				answer.Content = message.Sticker.FileID
			} else if message.Photo != nil {
				photo := message.Photo[len(message.Photo)-1]
				answer.Type = "photo"
				answer.Content = photo.FileID
			} else if message.Video != nil {
				answer.Type = "video"
				answer.Content = message.Video.FileID
			} else if message.Voice != nil {
				answer.Type = "voice"
				answer.Content = message.Voice.FileID
			} else if message.Document != nil {
				answer.Type = "file"
				answer.Content = message.Document.FileID
			} else if message.Audio != nil {
				answer.Type = "music"
				answer.Content = message.Audio.FileID
			} else if message.Animation != nil {
				answer.Type = "animation"
				answer.Content = message.Animation.FileID
			} else if message.VideoNote != nil {
				answer.Type = "video_note"
				answer.Content = message.VideoNote.FileID
			}

			if message.Caption != "" {
				answer.Caption = message.Caption
			}

			if answer.Type != "" {
				lastItem.Answers = append(lastItem.Answers, answer)
				saveDatabase(db)

				msg := tgbotapi.NewMessage(chatID, "✅ Ответ был добавлен к вашему тексту\n\nВы можете отправить больше контента или нажать 'Done!'")
				msg.ReplyMarkup = doneKeyboard
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Возникла проблема с отправленным вами контентом, пожалуйста, отправьте другой контент")
				msg.ReplyMarkup = doneKeyboard
				bot.Send(msg)
			}
		case "remove":
			// Удаление указанного элемента
			removed := false
			for i, item := range db.Data {
				if item.Text == text {
					db.Data = append(db.Data[:i], db.Data[i+1:]...)
					removed = true
					break
				}
			}
			if removed {
				saveDatabase(db)
				msg := tgbotapi.NewMessage(chatID, "✅ Успешно удалено")
				msg.ReplyMarkup = homeKeyboard
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Указанный элемент не найден")
				msg.ReplyMarkup = backKeyboard
				bot.Send(msg)
			}
			db.Step = ""
			saveDatabase(db)
		default:
			// Ничего не делать
		}
	}
}

// Функция обработки бизнес-сообщений
func handleBusinessMessage(bot *tgbotapi.BotAPI, bMessage *BusinessMessage, db *Database) {
	bText := bMessage.Text
	bChatID := bMessage.Chat.ID
	bID := strconv.FormatInt(bMessage.BusinessConnectionID, 10)
	bMessageID := strconv.Itoa(bMessage.MessageID)

	for _, item := range db.Data {
		if item.Text == bText {
			for index, answer := range item.Answers {
				params := url.Values{}
				params.Add("chat_id", strconv.FormatInt(bChatID, 10))
				params.Add("business_connection_id", bID)
				params.Add("parse_mode", "HTML")
				params.Add("disable_web_page_preview", "true")
				if index == 0 {
					params.Add("reply_to_message_id", bMessageID)
				}
				if answer.Caption != "" {
					params.Add("caption", answer.Caption)
				}
				switch answer.Type {
				case "text":
					params.Add("text", answer.Content)
					sendBotAPIRequest(BotToken, "sendMessage", params)
				case "sticker":
					params.Add("sticker", answer.Content)
					sendBotAPIRequest(BotToken, "sendSticker", params)
				case "photo":
					params.Add("photo", answer.Content)
					sendBotAPIRequest(BotToken, "sendPhoto", params)
				case "video":
					params.Add("video", answer.Content)
					sendBotAPIRequest(BotToken, "sendVideo", params)
				case "voice":
					params.Add("voice", answer.Content)
					sendBotAPIRequest(BotToken, "sendVoice", params)
				case "file":
					params.Add("document", answer.Content)
					sendBotAPIRequest(BotToken, "sendDocument", params)
				case "music":
					params.Add("audio", answer.Content)
					sendBotAPIRequest(BotToken, "sendAudio", params)
				case "animation":
					params.Add("animation", answer.Content)
					sendBotAPIRequest(BotToken, "sendAnimation", params)
				case "video_note":
					params.Add("video_note", answer.Content)
					sendBotAPIRequest(BotToken, "sendVideoNote", params)
				}
			}
		}
	}
}

// Функция отправки запросов к API Telegram
func sendBotAPIRequest(token, method string, params url.Values) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		log.Println("Ошибка отправки запроса:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// Здесь можно добавить обработку ошибок, проверив ответ
	log.Println("Ответ:", string(body))
}
