#!/bin/bash

go build -ldflags "-w -s" -o tgBot -buildvcs=false

if [ $? -eq 0 ]; then
    echo "Сборка завершена успешно."
    ./tgBot
else
    echo "Ошибка при сборке приложения."
fi
