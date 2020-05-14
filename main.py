import telebot
import config
from hashlib import md5

bot = telebot.TeleBot(config.TOKEN)
userList = []


# 新群员加入进行人机验证
@bot.message_handler(content_types=['new_chat_members'])
def welcome_new(message):
    global userList
    NewMemberID = message.new_chat_members[0].id
    secret = getSecret(NewMemberID)
    userList.append({NewMemberID, secret})
    try:
        bot.restrict_chat_member(message.chat.id, NewMemberID, until_date=None, can_send_messages=True,
                                 can_add_web_page_previews=False, can_send_media_messages=False,
                                 can_send_other_messages=False)
    except BaseException:
        pass
    bot.send_message(message.chat.id, "")


def getUrl(userID, secret):
    return ""


def getSecret(userID):
    str = str(userID) + config.SALT
    return md5(md5(str.encode('utf-8')).hexdigest())


if __name__ == '__main__':
    try:
        bot.polling()
    except BaseException:
        pass
