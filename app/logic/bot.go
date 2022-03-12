package logic

import (
	"FkAdBot/app/infra"
	"FkAdBot/pkg/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var BotProcessor *Bot

type Bot struct{}

type matchProcessor struct {
	MatchFunc func(tgbotapi.Update) bool
	Processor func(update tgbotapi.Update) error
}

var matchProcessorSlice = make([]*matchProcessor, 0)

func init() {
	NewProcessor(func(update tgbotapi.Update) bool {
		return update.Message != nil && !update.Message.IsCommand()
	}, MessageProcessor) // 消息处理器
}

func NewProcessor(match func(tgbotapi.Update) bool, processor func(update tgbotapi.Update) error) {
	matchProcessorSlice = append(matchProcessorSlice,
		&matchProcessor{
			MatchFunc: match,
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
			for i := range matchProcessorSlice {
				if matchProcessorSlice[i].MatchFunc(msg) {
					err := matchProcessorSlice[i].Processor(msg)
					if err != nil {
						log.Error("Bot.Run", log.Any("update", msg))
					}
					break
				}
			}
		}
	}
}

func MessageProcessor(update tgbotapi.Update) error {
	replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	send, err := infra.Bot.GetBot().Send(replyMsg)
	if err != nil {
		log.Error("MessageProcessor", log.Any("info", send))
		return err
	}
	return nil
}
