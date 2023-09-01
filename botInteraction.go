package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleUpdates(bot *tgbotapi.BotAPI, config *BotConfig) {
	var updateMutex sync.Mutex

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Error getting updates channel: %v", err)
		return
	}

	for update := range updates {
		updateMutex.Lock()

		if update.Message == nil && update.CallbackQuery == nil {
			updateMutex.Unlock()
			continue // Продолжаем получать обновления
		}

		if update.Message != nil && update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				handleStart(update, bot, true)
			}
		} else if update.Message != nil && update.Message.Text == "Список соцсетей" {
			handleSocialMedia(update, bot, config)
		} else if update.Message != nil && update.Message.Text == "Поиск видео" {
			findVideos(update, bot, config, updates)
		}

		updateMutex.Unlock()
	}

	logmsg := "Bot instance terminated, restarting..."

	writeToClickHouse(logmsg)
}

func handleStart(update tgbotapi.Update, bot *tgbotapi.BotAPI, sendGreeting bool) {
	replyMarkup := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Список соцсетей"),
			tgbotapi.NewKeyboardButton("Поиск видео"),
		),
	)

	if sendGreeting {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я PruKon_Bot. Сейчас вы находитесь в меню")
		msg.ReplyMarkup = replyMarkup

		bot.Send(msg)
	}
}

func handleSocialMedia(update tgbotapi.Update, bot *tgbotapi.BotAPI, config *BotConfig) {
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Телеграмм канал", config.TgLink),
			tgbotapi.NewInlineKeyboardButtonURL("YouTube", config.YouTubeLink),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список соцсетей:")
	msg.ReplyMarkup = replyMarkup

	bot.Send(msg)
}

func findVideos(update tgbotapi.Update, bot *tgbotapi.BotAPI, config *BotConfig, updatesChan tgbotapi.UpdatesChannel) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите запрос:")
	bot.Send(msg)

	answerCh := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case update := <-updatesChan: // Используйте канал updates напрямую
				if update.Message != nil && update.Message.Text != "" {
					answerCh <- update.Message.Text
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	userInput := <-answerCh

	MSG := tgbotapi.NewMessage(update.Message.Chat.ID, "Ищу видео относительно вашего запроса!")
	bot.Send(MSG)

	response := searchVideos(userInput, config)
	if response == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка при выполнении поиска.")
		bot.Send(msg)
		return
	}

	if len(response.Items) == 0 {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось найти видео по вашему запросу")
		bot.Send(msg)
		return
	}

	msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Результаты поиска:")
	for _, item := range response.Items {
		msg.Text += fmt.Sprintf("\nTitle: %s\n", item.Snippet.Title)
		msg.Text += fmt.Sprintf("URL: https://www.youtube.com/watch?v=%s\n", item.Id.VideoId)
		if len(item.Snippet.Description) != 0 {
			msg.Text += fmt.Sprintf("Description: %s\n", item.Snippet.Description)
		}
		msg.Text += "-------------------------------"
	}

	bot.Send(msg)
}
