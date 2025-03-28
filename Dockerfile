FROM golang:1.21.13-alpine

# Устанавливаем необходимые зависимости
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Создаем директорию uploads с правильными правами
RUN mkdir -p /app/uploads && chmod 777 /app/uploads

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN go build -o main .

EXPOSE 8080

# Добавляем скрипт для ожидания готовности БД и Redis
COPY wait-for.sh /wait-for.sh
RUN chmod +x /wait-for.sh

CMD ["./main"] 