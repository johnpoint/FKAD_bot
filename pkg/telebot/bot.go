package telebot

import (
	"github.com/johnpoint/go-bootstrap/log"
	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	matchProcessorSlice []*matchProcessor
}

type matchProcessor struct {
	MatchFunc func(tgbotapi.Update) bool
	Processor func(update tgbotapi.Update) (bool, error) // isBreak,error
}

func (b *Bot) NewProcessor(match func(tgbotapi.Update) bool, processor func(update tgbotapi.Update) (bool, error)) {
	b.matchProcessorSlice = append(b.matchProcessorSlice,
		&matchProcessor{
			MatchFunc: match,
			Processor: processor,
		},
	)
}

func (b *Bot) NewCommandProcessor(command string, processor func(update tgbotapi.Update) (bool, error)) {
	b.matchProcessorSlice = append(b.matchProcessorSlice,
		&matchProcessor{
			MatchFunc: func(update tgbotapi.Update) bool {
				return update.Message != nil && update.Message.IsCommand() && update.Message.Command() == command
			},
			Processor: processor,
		},
	)
}

func (b *Bot) Run(updates tgbotapi.UpdatesChannel) {
	if updates == nil {
		panic("updates is nil")
	}
	for {
		select {
		case msg := <-updates:
			log.Info("Bot.Run", zap.Any("update", msg))
			for i := range b.matchProcessorSlice {
				if b.matchProcessorSlice[i].MatchFunc(msg) {
					log.Info("Bot.Run", zap.Any("match!", i))
					isBreak, err := b.matchProcessorSlice[i].Processor(msg)
					if err != nil {
						log.Error("Bot.Run", zap.Any("update", msg), zap.Any("error", err))
					}
					if isBreak {
						break
					}
				}
			}
		}
	}
}
