#!/bin/bash

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/.."

if ! command -v hey &> /dev/null; then
    echo "Инструмент Hey не найден. Пробуем установить..."
    if command -v go &> /dev/null; then
        go install github.com/rakyll/hey@latest
        if [ $? -ne 0 ]; then
            echo "Не удалось установить Hey через Go. Пожалуйста, установите его вручную."
            exit 1
        fi
    else
        echo "Go не найден. Пожалуйста, установите Hey вручную."
        exit 1
    fi
fi

# База URL для тестирования
BASE_URL=${1:-"http://localhost:8080"}

# Создаем временный файл для хранения токена
TOKEN_FILE=$(mktemp)

# Функция для выполнения аутентификации и получения токена
get_auth_token() {
    echo "Получение токена авторизации..."
    curl -s -X POST -H "Content-Type: application/json" \
        -d '{"phone":"+77001234567","password":"password123"}' \
        "$BASE_URL/api/login" | grep -o '"token":"[^"]*"' | cut -d'"' -f4 > "$TOKEN_FILE"
    
    if [ ! -s "$TOKEN_FILE" ]; then
        echo "Не удалось получить токен авторизации. Проверьте, запущен ли сервер."
        exit 1
    fi
    
    TOKEN=$(cat "$TOKEN_FILE")
    echo "Токен получен: ${TOKEN:0:20}..."
}

# Тестирование открытых эндпоинтов
test_public_endpoints() {
    echo "Тестирование открытых эндпоинтов..."
    
    echo "1. Проверка работоспособности системы..."
    hey -n 1000 -c 100 "$BASE_URL/health"
    
    echo "2. Тестирование входа..."
    hey -n 200 -c 20 -m POST -H "Content-Type: application/json" \
        -d '{"phone":"+77001234567","password":"password123"}' \
        "$BASE_URL/api/login"
}

# Тестирование защищенных эндпоинтов
test_protected_endpoints() {
    TOKEN=$(cat "$TOKEN_FILE")
    echo "Тестирование защищенных эндпоинтов..."
    
    echo "1. Получение профиля пользователя..."
    hey -n 500 -c 50 -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/user/profile"
    
    echo "2. Поиск адреса..."
    hey -n 200 -c 20 -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL/api/address/search?query=Абая"
    
    echo "3. Создание бронирования (создает нагрузку на базу данных)..."
    hey -n 100 -c 10 -m POST -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{"pickup_address":"Абая 44","pickup_lat":51.1,"pickup_lng":71.4,"destination_address":"Кабанбай батыра 53","destination_lat":51.12,"destination_lng":71.42,"car_class":"economy","payment_method":"cash","scheduled_time":""}' \
        "$BASE_URL/api/bookings"
}

# Комплексный смешанный тест, имитирующий реальную нагрузку
complex_test() {
    TOKEN=$(cat "$TOKEN_FILE")
    echo "Запуск комплексного тестирования..."
    
    # Запускаем несколько параллельных тестов с разной интенсивностью
    hey -n 2000 -c 200 "$BASE_URL/health" &
    hey -n 1000 -c 100 -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/user/profile" &
    hey -n 500 -c 50 -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/address/search?query=Абая" &
    hey -n 200 -c 20 -m POST -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{"pickup_address":"Абая 44","pickup_lat":51.1,"pickup_lng":71.4,"destination_address":"Кабанбай батыра 53","destination_lat":51.12,"destination_lng":71.42,"car_class":"economy","payment_method":"cash","scheduled_time":""}' \
        "$BASE_URL/api/bookings" &
    
    wait
    echo "Комплексное тестирование завершено."
}

echo "=========================================================="
echo "  Начало нагрузочного тестирования для $BASE_URL"
echo "=========================================================="

# Получаем токен авторизации для тестирования
get_auth_token

# Тестируем публичные эндпоинты
test_public_endpoints

# Тестируем защищенные эндпоинты
test_protected_endpoints

# Запускаем комплексный тест
complex_test

# Удаляем временный файл
rm -f "$TOKEN_FILE"

echo "=========================================================="
echo "  Нагрузочное тестирование завершено"
echo "==========================================================" 