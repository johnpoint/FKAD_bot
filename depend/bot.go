package depend

import (
	"FkAdBot/app/infra"
	"FkAdBot/app/logic"
	"FkAdBot/config"
	"context"

	"github.com/johnpoint/go-bootstrap/core"
)

type Bot struct{}

var _ core.Component = (*Bot)(nil)

func (d *Bot) Init(ctx context.Context) error {
	err := infra.InitBotAPI(config.Config.TelegramBot)
	if err != nil {
		return err
	}

	go logic.Bot.Run(infra.Bot.StartWebhook())
	return nil
}
