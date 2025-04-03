#!/bin/bash

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/.."

echo "Обновление зависимостей Go..."

# Добавляем новые зависимости
go get -u github.com/gin-contrib/cors
go get -u github.com/gin-gonic/gin
go get -u github.com/go-redis/redis/v8
go get -u github.com/prometheus/client_golang/prometheus
go get -u github.com/prometheus/client_golang/prometheus/promauto
go get -u github.com/prometheus/client_golang/prometheus/promhttp

# Выполняем go mod tidy для обновления go.mod и go.sum
go mod tidy

echo "Зависимости успешно обновлены!" 