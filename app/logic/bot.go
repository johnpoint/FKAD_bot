package logic

import (
	"FkAdBot/app/infra"
	"FkAdBot/config"
	"FkAdBot/pkg/log"
	"FkAdBot/pkg/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/guonaihong/gout"
	"strings"
	"sync"
	"time"
)

var BotProcessor *Bot

type Bot struct{}

type matchProcessor struct {
	MatchFunc func(tgbotapi.Update) bool
	Processor func(update tgbotapi.Update) error
}

var waitVerifyUserMap sync.Map

var matchProcessorSlice = make([]*matchProcessor, 0)

func init() {
	NewProcessor(func(update tgbotapi.Update) bool {
		return update.Message != nil && !update.Message.IsCommand()
	}, MessageProcessor) // 消息处理器
	NewProcessor(func(update tgbotapi.Update) bool {
		return update.Message != nil && len(update.Message.NewChatMembers) != 0
	}, NewMemberProcessor) // 新用户加入处理器

	go checkWaitMapLoop()
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
			log.Info("Bot.Run", log.Any("update", msg))
			for i := range matchProcessorSlice {
				if matchProcessorSlice[i].MatchFunc(msg) {
					err := matchProcessorSlice[i].Processor(msg)
					if err != nil {
						log.Error("Bot.Run", log.Any("update", msg), log.Any("error", err))
					}
					break
				}
			}
		}
	}
}

func MessageProcessor(update tgbotapi.Update) error {
	if update.Message.Chat.Type == "private" || update.Message.Chat.Type == "channel" {
		return nil
	}
	v, ok := waitVerifyUserMap.LoadAndDelete(update.Message.From.ID)
	if ok {
		joinData := v.(*JoinGroupVerifyData)
		ver, ok := verifyCodeMap.LoadAndDelete(update.Message.From.ID)
		if ok && !strings.Contains(ver.(string), update.Message.Text) {
			kickJoinGroupVerifyData(joinData)
		} else {
			passJoinGroupVerifyData(joinData)
		}
	}
	return nil
}

type JoinGroupVerifyData struct {
	ChatID    int64
	ExpireAt  int64
	MessageID int
	UserID    int64
	Url       string
}

func NewMemberProcessor(update tgbotapi.Update) error {
	for _, member := range update.Message.NewChatMembers {
		// 删除入群消息
		delMsg := tgbotapi.DeleteMessageConfig{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.MessageID,
		}
		send, err := infra.Bot.GetBot().Send(delMsg)
		if err != nil {
			log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
			return err
		}

		// 新用户进群，禁言
		req := tgbotapi.RestrictChatMemberConfig{
			UntilDate:   0,
			Permissions: &tgbotapi.ChatPermissions{},
		}
		req.ChatMemberConfig = tgbotapi.ChatMemberConfig{
			ChatID: update.Message.Chat.ID,
			UserID: member.ID,
		}
		send, err = infra.Bot.GetBot().Send(req)
		if err != nil {
			log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
			return err
		}

		hashUrl := getUrlHash(member.ID)

		// 发送欢迎消息
		helloMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf(
				"请 [新群友](tg://user?id=%d) 点击 [链接](%s/%s) 后在本群回复验证码进行人机检验，回复其他内容将会被立即踢出！",
				member.ID,
				config.Config.TelegramBot.VerifyPageUrl,
				hashUrl,
			),
		)
		helloMsg.ParseMode = "Markdown"

		send, err = infra.Bot.GetBot().Send(helloMsg)
		if err != nil {
			log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
			return err
		}

		waitVerifyUserMap.Store(member.ID, &JoinGroupVerifyData{
			ExpireAt: time.Now().Add(60 * time.Second).Unix(),
			ChatID:   update.Message.Chat.ID,
			UserID:   member.ID,
			Url:      hashUrl,
		})
	}
	return nil
}

var verifyCodeMap sync.Map

func getUrlHash(userID int64) string {
	hashUrl := utils.RandomString()
	passCode := utils.RandomString()
	verifyCodeMap.Store(userID, passCode)
	err := gout.GET(config.Config.TelegramBot.VerifyPageUrl + "/" + config.Config.TelegramBot.VerifyPageSecret + "/" + hashUrl + "/" + passCode).Do()
	if err != nil {
		log.Error("getUrlHash", log.Any("error", err))
	}
	return hashUrl
}

// 清理过期的验证请求
func checkWaitMapLoop() {
	for {
		var needDel []*JoinGroupVerifyData
		waitVerifyUserMap.Range(func(key, value interface{}) bool {
			if value.(*JoinGroupVerifyData).ExpireAt <= time.Now().Unix() {
				needDel = append(needDel, value.(*JoinGroupVerifyData))
			}
			return true
		})
		for i := range needDel {
			kickJoinGroupVerifyData(needDel[i])
		}
		time.Sleep(15 * time.Second)
	}
}

func kickJoinGroupVerifyData(data *JoinGroupVerifyData) {
	waitVerifyUserMap.Delete(data.UserID)
	// 删除入群欢迎消息
	delMsg := tgbotapi.DeleteMessageConfig{
		MessageID: data.MessageID,
		ChatID:    data.ChatID,
	}
	send, err := infra.Bot.GetBot().Send(delMsg)
	if err != nil {
		log.Error("kickJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		return
	}

	// 踢用户
	kickUser := tgbotapi.KickChatMemberConfig{
		UntilDate:      time.Now().Add(60 * time.Second).Unix(),
		RevokeMessages: true,
	}
	kickUser.ChatMemberConfig = tgbotapi.ChatMemberConfig{
		ChatID: data.ChatID,
		UserID: data.UserID,
	}
	send, err = infra.Bot.GetBot().Send(kickUser)
	if err != nil {
		log.Error("kickJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		return
	}
}

func passJoinGroupVerifyData(data *JoinGroupVerifyData) {
	waitVerifyUserMap.Delete(data.UserID)
	// 开启权限
	req := tgbotapi.RestrictChatMemberConfig{
		UntilDate: 0,
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages:       true,
			CanSendMediaMessages:  true,
			CanSendPolls:          true,
			CanSendOtherMessages:  true,
			CanAddWebPagePreviews: true,
		},
	}
	req.ChatMemberConfig = tgbotapi.ChatMemberConfig{
		ChatID: data.ChatID,
		UserID: data.UserID,
	}
	send, err := infra.Bot.GetBot().Send(req)
	if err != nil {
		log.Error("passJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		return
	}

	// 发送欢迎消息
	helloMsg := tgbotapi.NewMessage(data.ChatID, "验证通过，欢迎新群友!")
	helloMsg.ParseMode = "Markdown"

	send, err = infra.Bot.GetBot().Send(helloMsg)
	if err != nil {
		log.Error("passJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		return
	}
}
