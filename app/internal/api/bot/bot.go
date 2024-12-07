package bot

import (
	"context"
	"fmt"
	"sync"

	tb "gopkg.in/telebot.v4"
)

// Wrapper представляет обертку над ботом.
type Wrapper struct {
	bot       *tb.Bot
	config    *Config
	aiService Service
	states    map[int64]string // Хранилище состояний пользователей
	mu        sync.Mutex       // Защита для работы с состояниями
}

// NewWrapper создаёт нового бота с обработчиками.
func NewWrapper(config *Config, aiService Service) (*Wrapper, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	settings := tb.Settings{
		Token:  config.Token,
		Poller: &tb.LongPoller{Timeout: config.Timeout},
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		return nil, err
	}

	w := &Wrapper{
		bot:       bot,
		config:    config,
		aiService: aiService,
		states:    make(map[int64]string), // Инициализация состояний
	}

	w.setupHandlers()
	return w, nil
}

// Start запускает бота.
func (w *Wrapper) Start(_ context.Context) error {
	w.bot.Start()
	return nil
}

// setupHandlers настраивает обработчики событий.
func (w *Wrapper) setupHandlers() {
	// Меню с кнопками
	menu := &tb.ReplyMarkup{}
	btnGPT := menu.Data("💬 ChatGPT", "chat_gpt")
	btnBack := menu.Data("⬅️ Назад", "back")

	menu.Inline(
		menu.Row(btnGPT),
	)

	// Обработчик команды /start
	w.bot.Handle("/start", func(c tb.Context) error {
		message := "Привет! Меня зовут Гоша, ваш личный ассистент. Давайте начнём!"
		stickerID := "CAACAgIAAxkBAAENSmpnVLOt0C0CvGTQByda2SQiIJK4-gACqRcAAtoIAUn-P0sCoVKCnzYE" // Пример FileID

		// Отправляем стикер
		sticker := &tb.Sticker{File: tb.File{FileID: stickerID}}
		if err := c.Send(sticker); err != nil {
			return fmt.Errorf("Ошибка отправки стикера: %w", err)
		}

		// Отправляем сообщение с кнопками
		return c.Send(message, &tb.SendOptions{
			ReplyMarkup: menu,
			ParseMode:   tb.ModeMarkdown,
		})
	})

	// Обработчик кнопки ChatGPT
	w.bot.Handle(&btnGPT, func(c tb.Context) error {
		userID := c.Sender().ID

		// Устанавливаем состояние ожидания текста
		w.setState(userID, "awaiting_text")

		// Кнопка "Назад"
		backMenu := &tb.ReplyMarkup{}
		backMenu.Inline(backMenu.Row(btnBack))

		// Подтверждаем нажатие кнопки
		if err := c.Respond(&tb.CallbackResponse{
			Text: "Введите текст для ChatGPT.",
		}); err != nil {
			return err
		}

		// Сообщаем пользователю, что ожидается ввод текста
		return c.Send("Теперь введите текст, который вы хотите отправить ChatGPT.", &tb.SendOptions{
			ReplyMarkup: backMenu,
		})
	})

	// Обработчик кнопки "Назад"
	w.bot.Handle(&btnBack, func(c tb.Context) error {
		userID := c.Sender().ID

		// Сбрасываем состояние
		w.setState(userID, "")

		// Возвращаем главное меню
		return c.Send("Вы вернулись в главное меню. Выберите действие:", &tb.SendOptions{
			ReplyMarkup: menu,
		})
	})

	// Обработчик текстовых сообщений
	w.bot.Handle(tb.OnText, func(c tb.Context) error {
		userID := c.Sender().ID

		// Проверяем состояние пользователя
		if w.getState(userID) == "awaiting_text" {
			// Обрабатываем текст через ChatGPT
			return w.handleText(c)
		}

		// Если пользователь не в состоянии ожидания текста
		return c.Send("Я вас не понял. Используйте /start, чтобы начать.")
	})
}

// handleText обрабатывает текстовые запросы для ChatGPT.
func (w *Wrapper) handleText(c tb.Context) error {
	txt := c.Text()

	// Запрос к GPT
	ctx := context.TODO()
	response, err := w.aiService.ChatCompletion(ctx, txt)
	if err != nil {
		return c.Send("Ошибка при обработке запроса.")
	}

	// Форматируем ответ
	formattedResponse := formatResponse(response)

	// Отправляем сообщение
	return sendLongMessage(c, formattedResponse)
}

// Управление состояниями пользователей
func (w *Wrapper) setState(userID int64, state string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.states[userID] = state
}

func (w *Wrapper) getState(userID int64) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.states[userID]
}
