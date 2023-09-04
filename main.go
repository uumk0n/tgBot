package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	bot, config := getBot()
	if bot == nil {
		writeToClickHouse("Ошибка создания бота")
		return
	} else if config == nil {
		writeToClickHouse("Ошибка считывания конфиг файла")
		return
	} else {
		for attempts := 0; attempts < 10; attempts++ { // Ограничиваем количество попыток
			err := runBot(bot, config)
			if err != nil {
				logMsg := fmt.Sprintf("Error: %v. Restarting in 3 seconds...", err)
				err := writeToClickHouse(logMsg)
				if err != nil {
					log.Printf("Error writing log to ClickHouse: %v", err)
				}
				time.Sleep(3 * time.Second)
				continue
			}
			break // Выходим из цикла, если успешно запустили бота
		}
	}
}
