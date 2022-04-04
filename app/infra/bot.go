package infra

import (
	"FkAdBot/config"
	"FkAdBot/pkg/log"
	"FkAdBot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"time"
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
	var err error
	randomPath := utils.Md5(utils.RandomString())
	var webhookUrl = b.config.Url + randomPath

	var wh tgbotapi.WebhookConfig
	if len(b.config.CustomTLSCert) != 0 {
		wh, _ = tgbotapi.NewWebhookWithCert(webhookUrl, tgbotapi.FilePath(b.config.CustomTLSCert))
	} else {
		wh, _ = tgbotapi.NewWebhook(webhookUrl)
	}

	log.Info("BotAPI.StartWebhook", log.String("url", webhookUrl))

	updateChan := b.api.ListenForWebhook("/" + randomPath)

	go func() {
		log.Info("BotAPI.StartWebhook", log.String("info", b.config.Listen))
		if len(b.config.CustomTLSCert) != 0 && len(b.config.CustomTLSKey) != 0 {
			err = http.ListenAndServeTLS(b.config.Listen, b.config.CustomTLSCert, b.config.CustomTLSKey, nil)
		} else {
			err = http.ListenAndServe(b.config.Listen, nil)
		}
		if err != nil {
			log.Error("http", log.Err(err))
			return
		}
	}()

	_, err = b.api.Request(wh)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	info, err := b.api.GetWebhookInfo()
	if err != nil {
		panic(err)
	}

	if info.LastErrorDate != 0 {
		log.Error("WebhookInfo", log.Any("i", info))
		panic("Telegram callback failed: " + info.LastErrorMessage)
	}

	return updateChan
}

func (b *BotAPI) GetBot() *tgbotapi.BotAPI {
	return b.api
}
