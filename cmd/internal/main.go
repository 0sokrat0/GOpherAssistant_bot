package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	BotToken = "8130389933:AAGdMGjRpoLoVjhy_i2WLwPJ7tr3F-_kxYk" // Замените на ваш токен
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

// Функции для загрузки и сохранения базы данных
func loadDatabase() (*Database, error) {
	db := &Database{}
	file, err := ioutil.ReadFile("db.json")
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, создаем пустую базу данных
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
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("db.json", data, 0644)
	return err
}

// Обработка сообщений от администратора
func handleAdminMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *Database) {
	chatID := message.Chat.ID
	text := message.Text

	// Клавиатуры
	homeKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Add auto reply ✉️")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Remove auto reply 🚫")),
	)
	homeKeyboard.ResizeKeyboard = true

	backKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Back 🔙")),
	)
	backKeyboard.ResizeKeyboard = true

	switch text {
	case "/start":
		msg := tgbotapi.NewMessage(chatID, "Welcome to Business Manager Bot! 🤖")
		msg.ReplyMarkup = homeKeyboard
		bot.Send(msg)
		db.Step = ""
		saveDatabase(db)

	case "Add auto reply ✉️":
		msg := tgbotapi.NewMessage(chatID, "Enter the trigger text for the auto-reply.")
		msg.ReplyMarkup = backKeyboard
		bot.Send(msg)
		db.Step = "add-1"
		saveDatabase(db)

	case "Remove auto reply 🚫":
		if len(db.Data) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Auto-reply list is empty!")
			msg.ReplyMarkup = homeKeyboard
			bot.Send(msg)
		} else {
			var list string
			for _, item := range db.Data {
				list += fmt.Sprintf("<code>%s</code>\n---\n", item.Text)
			}
			msg := tgbotapi.NewMessage(chatID, list)
			msg.ParseMode = "HTML"
			bot.Send(msg)

			msg2 := tgbotapi.NewMessage(chatID, "Copy the text you want to remove.")
			msg2.ReplyMarkup = backKeyboard
			bot.Send(msg2)
			db.Step = "remove"
			saveDatabase(db)
		}

	case "Back 🔙":
		msg := tgbotapi.NewMessage(chatID, "Main menu.")
		msg.ReplyMarkup = homeKeyboard
		bot.Send(msg)
		db.Step = ""
		saveDatabase(db)

	default:
		handleStep(bot, message, db)
	}
}

// Обработка шагов
func handleStep(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *Database) {
	chatID := message.Chat.ID
	text := message.Text

	switch db.Step {
	case "add-1":
		newItem := DataItem{
			Text:    text,
			Answers: []Answer{},
		}
		db.Data = append(db.Data, newItem)
		db.Step = "add-2"
		saveDatabase(db)

		msg := tgbotapi.NewMessage(chatID, "Trigger added! Send a reply (text, photo, etc.) for this trigger.")
		bot.Send(msg)

	case "add-2":
		if len(db.Data) == 0 {
			db.Step = ""
			saveDatabase(db)
			return
		}
		lastIndex := len(db.Data) - 1
		lastItem := &db.Data[lastIndex]

		answer := Answer{
			Type:    "text",
			Content: text,
		}
		lastItem.Answers = append(lastItem.Answers, answer)
		saveDatabase(db)

		msg := tgbotapi.NewMessage(chatID, "Reply added! Send more replies or type 'Back 🔙' to finish.")
		bot.Send(msg)

	case "remove":
		for i, item := range db.Data {
			if item.Text == text {
				db.Data = append(db.Data[:i], db.Data[i+1:]...)
				saveDatabase(db)

				msg := tgbotapi.NewMessage(chatID, "Trigger removed!")
				bot.Send(msg)
				break
			}
		}
		db.Step = ""
		saveDatabase(db)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Долгий опрос
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	db, _ := loadDatabase()

	for update := range updates {
		if update.Message != nil {
			if update.Message.From.ID == AdminID {
				handleAdminMessage(bot, update.Message, db)
			}
		}
	}
}
