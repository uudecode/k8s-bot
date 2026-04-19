package notifier

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	Bot    *tgbotapi.BotAPI
	ChatID int64
}

func (t *TelegramNotifier) Send(message string) error {
	msg := tgbotapi.NewMessage(t.ChatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := t.Bot.Send(msg)
	return err
}
