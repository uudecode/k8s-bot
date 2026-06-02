package monitor

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *Service) RunTelegramCommands(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := s.Bot.GetUpdatesChan(updateConfig)
	defer s.Bot.StopReceivingUpdates()

	s.Logger.Info().Msg("telegram command listener started")

	for {
		select {
		case <-ctx.Done():
			s.Logger.Info().Msg("telegram command listener stopped")
			return nil

		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.Chat == nil {
				continue
			}

			if update.Message.Chat.ID != s.ChatID {
				s.Logger.Warn().
					Int64("chat_id", update.Message.Chat.ID).
					Str("username", update.Message.From.UserName).
					Msg("ignored telegram command from unauthorized chat")
				continue
			}

			if !update.Message.IsCommand() {
				continue
			}

			s.handleTelegramCommand(ctx, update.Message)
		}
	}
}

func (s *Service) handleTelegramCommand(ctx context.Context, message *tgbotapi.Message) {
	switch message.Command() {
	case "start", "status":
		s.sendStatusCommandResponse(ctx)

	case "help":
		help := "Available commands:\n\n" +
			"/status - show current cluster status\n" +
			"/start - show current cluster status\n" +
			"/help - show this help"

		if err := s.Notifier.Send(help); err != nil {
			s.Logger.Error().Err(err).Msg("failed to send help response")
		}

	default:
		if err := s.Notifier.Send("Unknown command. Use /help"); err != nil {
			s.Logger.Error().Err(err).Msg("failed to send unknown command response")
		}
	}
}

func (s *Service) sendStatusCommandResponse(ctx context.Context) {
	commandCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	report, err := s.BuildStatusReport(commandCtx)
	if err != nil {
		s.Logger.Error().Err(err).Msg("failed to build status report")

		if notifyErr := s.Notifier.Send("Failed to build status report: `" + err.Error() + "`"); notifyErr != nil {
			s.Logger.Error().Err(notifyErr).Msg("failed to send status error response")
		}

		return
	}

	if err := s.Notifier.Send(report); err != nil {
		s.Logger.Error().Err(err).Msg("failed to send status response")
	}
}

func (s *Service) registerBotCommands() {
	commands := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "status", Description: "Show current cluster status"},
		tgbotapi.BotCommand{Command: "start", Description: "Show current cluster status"},
		tgbotapi.BotCommand{Command: "help", Description: "Show available commands"},
	)

	if _, err := s.Bot.Request(commands); err != nil {
		s.Logger.Warn().Err(err).Msg("Failed to register telegram bot commands menu")
	} else {
		s.Logger.Info().Msg("Telegram bot commands menu registered successfully")
	}
}
