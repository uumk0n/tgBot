package main

func main() {

	filePath := "config.json"

	botConfig, err := readBotConfig(filePath)
	if err != nil {
		writeToClickHouse("Ошибка считывания конфиг файла")
		return
	}

	telegramBot, err := getBot(botConfig)
	if err != nil {
		writeToClickHouse("Ошибка создания бота")
		return
	}

	for attempts := 0; attempts < 10; attempts++ {
		err := runBot(telegramBot, botConfig)
		if err != nil {
			restartBotAfterError(err)
			continue
		}
		break
	}
}
