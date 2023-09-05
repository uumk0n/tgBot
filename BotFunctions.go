package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type BotConfig struct {
	APIKey      string `json:"apiKey"`
	ChannelID   string `json:"channelID"`
	BotToken    string `json:"botToken"`
	TgLink      string `json:"tgLink"`
	YouTubeLink string `json:"youTubeLink"`
}

func getBot(botConfig *BotConfig) (*tgbotapi.BotAPI, error) {

	botToken := botConfig.BotToken

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		writeToClickHouse(fmt.Sprintf("error creating bot API: %v", err))
		return nil, err
	}

	bot.Debug = true

	writeToClickHouse(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	return bot, nil
}

func readBotConfig(filePath string) (*BotConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		writeToClickHouse(fmt.Sprintf("error opening bot config file: %v", err))
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		writeToClickHouse(fmt.Sprintf("error reading bot config data: %v", err))
		return nil, err
	}

	var config BotConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		writeToClickHouse(fmt.Sprintf("error decoding bot config JSON: %v", err))
		return nil, err
	}

	return &config, nil
}

func runBot(bot *tgbotapi.BotAPI, config *BotConfig) error {
	defer func() {
		if r := recover(); r != nil {
			err := writeToClickHouse("Bot is stopping...")

			if err != nil {
				log.Printf("Error writing log to ClickHouse: %v", err)
			}
			os.Exit(1)
		}
	}()

	handleUpdates(bot, config)

	return nil
}

func restartBotAfterError(err error) {
	logMsg := fmt.Sprintf("Error: %v. Restarting in 3 seconds...", err)
	if writeErr := writeToClickHouse(logMsg); writeErr != nil {
		log.Printf("Error writing log to ClickHouse: %v", writeErr)
	}
	time.Sleep(3 * time.Second)
}

func searchVideos(query string, config *BotConfig) *youtube.SearchListResponse {
	client := &http.Client{
		Transport: &transport.APIKey{Key: config.APIKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		writeToClickHouse(fmt.Sprintf("error creating YouTube service: %v", err))
		return nil
	}

	call := service.Search.List([]string{"id", "snippet"}).
		Q(query).
		ChannelId(config.ChannelID).
		Type("video").
		MaxResults(10)

	response, err := call.Do()
	if err != nil {
		writeToClickHouse(fmt.Sprintf("error searching videos: %v", err))
		return nil
	}

	return response
}
