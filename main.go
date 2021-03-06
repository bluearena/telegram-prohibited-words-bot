package main

import (
	"encoding/json"
	"github.com/gameraccoon/telegram-prohibited-words-bot/database"
	"github.com/gameraccoon/telegram-prohibited-words-bot/processing"
	"github.com/gameraccoon/telegram-prohibited-words-bot/telegramChat"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nicksnyder/go-i18n/i18n"
	"io/ioutil"
	"log"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	i18n.MustLoadTranslationFile("./data/strings/ru-ru.all.json")
}

func getFileStringContent(filePath string) (content string, err error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err == nil {
		content = strings.TrimSpace(string(fileContent))
	}
	return
}

func loadConfig(path string) (config processing.StaticConfiguration, err error) {
	jsonString, err := getFileStringContent(path)
	if err == nil {
		dec := json.NewDecoder(strings.NewReader(jsonString))
		err = dec.Decode(&config)
	}
	return
}

func getApiToken() (token string, err error) {
	return getFileStringContent("./telegramApiToken.txt")
}

func updateBot(bot *tgbotapi.BotAPI, staticData *processing.StaticProccessStructs) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Fatal(err.Error())
	}

	processors := Processors{
		Main: makeUserCommandProcessors(),
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		processUpdate(&update, staticData, &processors)
	}
}

func main() {
	apiToken, err := getApiToken()
	if err != nil {
		log.Fatal(err.Error())
	}

	config, err := loadConfig("./config.json")
	if err != nil {
		log.Fatal(err.Error())
	}

	trans, err := i18n.Tfunc(config.DefaultLanguage)
	if err != nil {
		log.Fatal(err.Error())
	}

	db := &database.Database{}
	err = db.Connect("./prohibited-words-data.db")
	defer db.Disconnect()

	if err != nil {
		log.Fatal("Can't connect database")
	}

	database.UpdateVersion(db)

	chat, err := telegramChat.MakeTelegramChat(apiToken)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Authorized on account %s", chat.GetBotUsername())

	chat.SetDebugModeEnabled(config.ExtendedLog)

	staticData := &processing.StaticProccessStructs{
		Config:      &config,
		Chat:        chat,
		Db:          db,
		Trans:       trans,
		CachedWords: map[int64][]string{},
	}

	updateBot(chat.GetBot(), staticData)
}
