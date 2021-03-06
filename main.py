import re
import telebot
from apscheduler.schedulers.background import BackgroundScheduler
import config
from hashlib import sha1
import time

bot = telebot.TeleBot(config.TOKEN)
userList = {}
BOTPATH = config.BOTPATH
LOGCHATID = config.LOGCHATID


def update_ban_username():
    f = open(BOTPATH + '/black.list')
    black_list = f.readlines()
    f.close()
    for i in range(len(black_list)):
        black_list[i] = black_list[i].rstrip('\n')
    global black_lists
    black_lists = black_list
    return black_list


def add_black_username(name):
    black_list = update_ban_username()
    black_list.append(name)
    global black_lists
    black_lists = black_list
    f = open(BOTPATH + '/black.list', 'w')
    for i in black_list:
        f.write(i)
        f.write('\n')
    f.close()


def ver_black(name):
    pre = re.compile(u'[\u4e00-\u9fa5]')
    res = re.findall(pre, name)
    name_clean = ''.join(res)
    for i in range(len(black_lists)):
        if black_lists[i] in name_clean:
            return True
    return False


# 新群员加入进行人机验证
@bot.message_handler(content_types=['new_chat_members'])
def welcome_new(message):
    global userList
    if str(message.chat.id) not in userList:
        userList[str(message.chat.id)] = {}
    NewMemberID = message.new_chat_members[0].id
    try:
        bot.restrict_chat_member(message.chat.id, NewMemberID, until_date=None, can_send_messages=False,
                                 can_add_web_page_previews=False, can_send_media_messages=False,
                                 can_send_other_messages=False)
    except BaseException:
        pass
    update_ban_username()
    try:
        if ver_black(message.new_chat_members[0].first_name):
            bot.kick_chat_member(message.chat.id, NewMemberID, until_date=None)
            bot.send_message(LOGCHATID,
                             "== BAN AD ==\nID: [" + str(NewMemberID) +
                             "](tg://user?id=" + str(NewMemberID) + ")",
                             parse_mode="Markdown")
            bot.delete_message(message.chat.id, message.message_id)
            return
        if message.new_chat_members[0].last_name != None and ver_black(message.new_chat_members[0].last_name):
            bot.kick_chat_member(message.chat.id, NewMemberID, until_date=None)
            bot.send_message(LOGCHATID,
                             "== BAN AD ==\nID: [" + str(NewMemberID) +
                             "](tg://user?id=" + str(NewMemberID) + ")",
                             parse_mode="Markdown")
            bot.delete_message(message.chat.id, message.message_id)
            return
    except BaseException:
        pass
    try:
        bot.restrict_chat_member(message.chat.id, NewMemberID, until_date=None, can_send_messages=True,
                                 can_add_web_page_previews=False, can_send_media_messages=False,
                                 can_send_other_messages=False)
    except BaseException:
        pass
    if str(NewMemberID) in userList:
        userList.pop(str(NewMemberID))
    msg1 = bot.send_message(message.chat.id,
                            "请 [" + message.new_chat_members[0].first_name + "](tg://user?id=" + str(
                                NewMemberID) + ") 点击 [链接](" + config.VERURL + "#" + getUrl(
                                NewMemberID, message.chat.id) + ") 后在本群回复验证码进行人机检验,回复其他内容将会被立即踢出！",
                            parse_mode="Markdown").message_id
    bot.delete_message(message.chat.id, message.message_id)
    userList[str(message.chat.id)][str(NewMemberID)].append(NewMemberID)
    userList[str(message.chat.id)][str(NewMemberID)].append(msg1)
    userList[str(message.chat.id)][str(NewMemberID)].append(int(time.time()))


def getUrl(userID, chatID):
    global userList
    if str(chatID) not in userList:
        userList[str(chatID)] = {}
    url = sha1((str(userID) + config.SALT + str(time.time()) + str(chatID)).encode("utf-8")).hexdigest()
    print(url)
    userList[str(chatID)][str(userID)] = [sha1(
        ("#" + url).encode("utf-8")).hexdigest().upper()]
    return url


@bot.message_handler(regexp='.+')
def scan_message(message):
    global userList
    if str(message.chat.id) not in userList:
        userList[str(message.chat.id)] = {}
    if str(message.from_user.id) in userList[str(message.chat.id)]:
        try:
            bot.delete_message(message.chat.id, userList[str(message.chat.id)][str(message.from_user.id)][2])
        except:
            pass
        if len(message.text) < 6 or message.text not in userList[str(message.chat.id)][str(message.from_user.id)][0]:
            try:
                bot.kick_chat_member(
                    message.chat.id, message.from_user.id, until_date=None)
            except:
                pass
        else:
            bot.restrict_chat_member(message.chat.id, message.from_user.id, until_date=None,
                                     can_send_other_messages=True, can_send_media_messages=True,
                                     can_add_web_page_previews=True, can_send_messages=True)
            userList[str(message.chat.id)].pop(str(message.from_user.id))
            try:
                msg = bot.reply_to(message, "验证成功~").message_id
            except:
                pass
            time.sleep(10)
            bot.delete_message(message.chat.id, msg)
        bot.delete_message(message.chat.id, message.message_id)


def clean_list():
    print("Check user list")
    global userList  # url mid msg time
    uL = userList.copy()
    for i in uL.keys():  # chat id
        for j in uL[i].keys():
            if uL[i][j][3] + 15 <= int(time.time()):
                try:
                    bot.delete_message(int(i), uL[i][j][2])
                except:
                    pass
            if uL[i][j][3] + 60 <= int(time.time()):
                bot.kick_chat_member(int(i), int(j), until_date=None)
                bot.restrict_chat_member(
                    int(i), int(j), until_date=None, can_send_messages=True)
                bot.restrict_chat_member(
                    int(i), int(j), until_date=None, can_send_messages=False)
                userList[i].pop(j)
        if len(userList[i]) == 0:
            userList.pop(i)


if __name__ == '__main__':
    scheduler = BackgroundScheduler()
    scheduler.add_job(clean_list, 'interval', seconds=15)
    scheduler.start()
    try:
        bot.polling()
    except BaseException:
        pass
