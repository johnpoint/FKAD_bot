import telebot
import config
from hashlib import md5
from hashlib import sha1

bot = telebot.TeleBot(config.TOKEN)
userList = {}


# 新群员加入进行人机验证
@bot.message_handler(content_types=['new_chat_members'])
def welcome_new(message):
    global userList
    NewMemberID = message.new_chat_members[0].id
    secret = getSecret(str(NewMemberID))
    userList.append({str(NewMemberID): secret})
    try:
        bot.restrict_chat_member(message.chat.id, NewMemberID, until_date=None, can_send_messages=True,
                                 can_add_web_page_previews=False, can_send_media_messages=False,
                                 can_send_other_messages=False)
    except BaseException:
        pass
    bot.send_message(message.chat.id,
                     "请点击 [链接](" + config.VERURL + "#" + getUrl(NewMemberID,
                                                                secret) + ") 后在本群回复验证码进行人机检验,回复其他内容将会被立即踢出！",
                     parse_mode="Markdown")


def getUrl(userID, secret):
    global userList
    userList[str(userID)] = sha1(sha1((str(userID) + secret)).hexdigest()).hexdigest()
    return sha1((str(userID) + secret).encode('utf-8')).hexdigest()


def getSecret(userID):
    str1 = str(userID) + config.SALT
    return md5(md5(str1).hexdigest())


@bot.message_handler(regexp=[''])
def scan_message(message):
    global userList
    if str(message.from_user.id) in userList:
        if message.message_text not in userList[str(message.from_user.id)]:
            try:
                bot.kick_chat_member(message.chat.id, message.from_user.id, until_date=None)
            except:
                pass


if __name__ == '__main__':
    try:
        bot.polling()
    except BaseException:
        pass
