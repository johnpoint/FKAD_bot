package infra

import (
	"FkAdBot/config"
	"FkAdBot/pkg/log"
	"FkAdBot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Bot *BotAPI

func InitBotAPI(tgConfig *config.TelegramBotConfig) error {
	bot, err := tgbotapi.NewBotAPI(tgConfig.Token)
	if err != nil {
		return err
	}
	Bot = &BotAPI{
		api:    bot,
		config: tgConfig,
	}
	return nil
}

type BotAPI struct {
	api    *tgbotapi.BotAPI
	config *config.TelegramBotConfig
}

func (b *BotAPI) StartWebhook() tgbotapi.UpdatesChannel {
	var webhookUrl = b.config.Url + utils.Md5(utils.RandomString())
	wh, _ := tgbotapi.NewWebhook(webhookUrl)

	_, err := b.api.Request(wh)
	if err != nil {
		panic(err)
	}

	info, err := b.api.GetWebhookInfo()
	if err != nil {
		panic(err)
	}

	if info.LastErrorDate != 0 {
		panic("Telegram callback failed: " + info.LastErrorMessage)
	}

	log.Info("BotAPI.StartWebhook", log.String("url", webhookUrl))
	return b.api.ListenForWebhook("/" + utils.Md5(utils.RandomString()))
}

func (b *BotAPI) GetBot() *tgbotapi.BotAPI {
	return b.api
}
