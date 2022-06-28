package logic

import (
	"FkAdBot/app/infra"
	"FkAdBot/config"
	"FkAdBot/pkg/log"
	"FkAdBot/pkg/telebot"
	"FkAdBot/pkg/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/guonaihong/gout"
	"strings"
	"sync"
	"time"
)

var Bot = &telebot.Bot{}

var waitVerifyUserMap sync.Map

func init() {
	Bot.NewProcessor(func(update tgbotapi.Update) bool {
		return update.Message != nil && !update.Message.IsCommand() && len(update.Message.NewChatMembers) == 0
	}, MessageProcessor) // 消息处理器
	Bot.NewProcessor(func(update tgbotapi.Update) bool {
		return update.Message != nil && len(update.Message.NewChatMembers) != 0
	}, NewMemberProcessor) // 新用户加入处理器

	go checkWaitMapLoop()
}

func MessageProcessor(update tgbotapi.Update) (bool, error) {
	if update.Message.Chat.Type == "private" || update.Message.Chat.Type == "channel" {
		return true, nil
	}
	v, ok := waitVerifyUserMap.LoadAndDelete(update.Message.From.ID)
	if ok {
		joinData := v.(*JoinGroupVerifyData)
		ver, _ := verifyCodeMap.LoadAndDelete(update.Message.From.ID)
		if !strings.Contains(ver.(string), update.Message.Text) || len(update.Message.Text) < 5 {
			kickJoinGroupVerifyData(joinData)
		} else {
			passJoinGroupVerifyData(joinData)
		}
		// 删除入群欢迎消息
		delMsg := tgbotapi.DeleteMessageConfig{
			MessageID: update.Message.MessageID,
			ChatID:    update.Message.Chat.ID,
		}
		send, err := infra.Bot.GetBot().Request(delMsg)
		if err != nil {
			log.Error("MessageProcessor", log.Any("info", send), log.Any("error", err))
			//return
		}
	}
	return true, nil
}

type JoinGroupVerifyData struct {
	ChatID    int64
	ExpireAt  int64
	MessageID int
	UserID    int64
	Url       string
}

func NewMemberProcessor(update tgbotapi.Update) (bool, error) {
	for _, member := range update.Message.NewChatMembers {
		// 新用户进群，禁言
		req := tgbotapi.RestrictChatMemberConfig{
			Permissions: &tgbotapi.ChatPermissions{},
		}
		req.ChatMemberConfig = tgbotapi.ChatMemberConfig{
			ChatID: update.Message.Chat.ID,
			UserID: member.ID,
		}
		send, err := infra.Bot.GetBot().Request(req)
		if err != nil {
			log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
			return true, nil
		}

		// 删除入群消息
		delMsg := tgbotapi.DeleteMessageConfig{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.MessageID,
		}
		send, err = infra.Bot.GetBot().Request(delMsg)
		if err != nil {
			log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
			return true, nil
		}

		go func() {
			time.Sleep(10 * time.Second)
			req := tgbotapi.RestrictChatMemberConfig{
				Permissions: &tgbotapi.ChatPermissions{
					CanSendMessages: true,
				},
			}
			req.ChatMemberConfig = tgbotapi.ChatMemberConfig{
				ChatID: update.Message.Chat.ID,
				UserID: member.ID,
			}
			send, err = infra.Bot.GetBot().Request(req)
			if err != nil {
				log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
				//return err
			}
		}()

		hashUrl := getUrlHash(member.ID)
		expireAt := time.Now().Add(60 * time.Second)

		// 发送欢迎消息
		helloMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf(
				"请 [新群友](tg://user?id=%d) 点击 [链接](%s/%s) 后在本群回复验证码进行人机检验，回复其他内容将会被立即踢出！(60s)",
				member.ID,
				config.Config.TelegramBot.VerifyPageUrl,
				hashUrl,
			),
		)
		helloMsg.ParseMode = "Markdown"

		sendResp, err := infra.Bot.GetBot().Send(helloMsg)
		if err != nil {
			log.Error("NewMemberProcessor", log.Any("info", send), log.Any("error", err))
			//return err
		}

		waitVerifyUserMap.Store(member.ID, &JoinGroupVerifyData{
			ExpireAt:  expireAt.Unix(),
			ChatID:    update.Message.Chat.ID,
			UserID:    member.ID,
			Url:       hashUrl,
			MessageID: sendResp.MessageID,
		})
	}
	return true, nil
}

var verifyCodeMap sync.Map

func getUrlHash(userID int64) string {
	hashUrl := utils.RandomString()
	passCode := utils.RandomString()
	verifyCodeMap.Store(userID, passCode)
	var body string
	err := gout.GET(config.Config.TelegramBot.VerifyPageUrl + "/" + config.Config.TelegramBot.VerifyPageSecret + "/" + hashUrl + "/" + passCode[:7]).BindBody(&body).Do()
	if err != nil {
		log.Error("getUrlHash", log.Any("error", err))
	}
	log.Info("getUrlHash", log.String("body", body))
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
	send, err := infra.Bot.GetBot().Request(delMsg)
	if err != nil {
		log.Error("kickJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		//return
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
	send, err = infra.Bot.GetBot().Request(kickUser)
	if err != nil {
		log.Error("kickJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		//return
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
	send, err := infra.Bot.GetBot().Request(req)
	if err != nil {
		log.Error("passJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		//return
	}

	// 发送欢迎消息
	helloMsg := tgbotapi.NewMessage(data.ChatID, "验证通过，欢迎新群友!")
	helloMsg.ParseMode = "Markdown"

	sendResp, err := infra.Bot.GetBot().Send(helloMsg)
	if err != nil {
		log.Error("passJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		//return
	}

	go func() {
		time.Sleep(5 * time.Second)
		// 删除入群欢迎消息
		delMsg := tgbotapi.DeleteMessageConfig{
			MessageID: sendResp.MessageID,
			ChatID:    sendResp.Chat.ID,
		}
		send, err = infra.Bot.GetBot().Request(delMsg)
		if err != nil {
			log.Error("passJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
			//return
		}
	}()

	// 删除入群欢迎消息
	delMsg := tgbotapi.DeleteMessageConfig{
		MessageID: data.MessageID,
		ChatID:    data.ChatID,
	}
	send, err = infra.Bot.GetBot().Request(delMsg)
	if err != nil {
		log.Error("passJoinGroupVerifyData", log.Any("info", send), log.Any("error", err))
		//return
	}
}
