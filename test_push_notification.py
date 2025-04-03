#!/usr/bin/env python3
import requests
import json
import os
import sys
from datetime import datetime

# Конфигурация
API_URL = "http://192.168.0.169:80/api"
ADMIN_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjowLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NzM2NTM1MzEsImlhdCI6MTc0MjExNzUzMX0.9BOK0_cPzIRjJq6tdv8gdy6nyETHbCpxBeaj92IUqio"
FCM_SERVER_KEY = "BPB4eaquHe47fdthXUU3ogNCuR0aRhr9SQjoi2_rOd67KQe8O4l8AKPlkrYnYbyNMwo9TJEAGL2i03sODRrKlgs"
FCM_URL = "https://fcm.googleapis.com/v1/projects/astrataxiapp/messages:send"

def get_fcm_token_from_db():
    """Получает FCM токен напрямую из логов приложения."""
    print("Извлечение FCM токена из логов приложения...")
    # Это значение FCM токена из логов Flutter-приложения
    return "cyY-VNbIQkKAP_0D5qR4db:APA91bFYHSga9nSMSVJ2X2-Tpcob5yybwGWEdh6zSXurERa-nlHRhZa03NsYDnde2kaOTY6reVTEsc0nzCyLooL3xPsZj0GCHKjRAjhlSRolLx8xW7yGWIY"

def get_user_by_id(user_id):
    """Получает информацию о пользователе по ID."""
    headers = {
        "Authorization": f"Bearer {ADMIN_TOKEN}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.get(f"{API_URL}/user/{user_id}", headers=headers)
        if response.status_code == 200:
            user_data = response.json()
            print(f"Получены данные пользователя с ID {user_id}: {user_data}")
            return user_data
        else:
            print(f"Ошибка при получении пользователя с ID {user_id}: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Исключение при получении пользователя с ID {user_id}: {e}")
        return None

def get_user_fcm_token(user_id):
    """Получает FCM токен пользователя из API."""
    # Сначала пытаемся получить пользователя по ID
    user_data = get_user_by_id(user_id)
    
    if user_data and "fcmToken" in user_data and user_data["fcmToken"]:
        fcm_token = user_data["fcmToken"]
        print(f"FCM токен пользователя из API: {fcm_token}")
        return fcm_token
    
    # Если не удалось получить из API, используем токен из логов
    print("Не удалось получить FCM токен из API, используем токен из логов")
    return get_fcm_token_from_db()

def send_test_notification(fcm_token, title="Тестовое уведомление", body="Это тестовое push-уведомление"):
    """Отправляет тестовое уведомление на указанный FCM токен."""
    headers = {
        "Authorization": f"key={FCM_SERVER_KEY}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "to": fcm_token,
        "notification": {
            "title": title,
            "body": body,
        },
        "data": {
            "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "type": "TEST_NOTIFICATION",
            "click_action": "FLUTTER_NOTIFICATION_CLICK"
        }
    }
    
    try:
        print(f"Отправка уведомления: {json.dumps(payload, ensure_ascii=False)}")
        response = requests.post(FCM_URL, headers=headers, data=json.dumps(payload))
        print(f"Статус ответа: {response.status_code}")
        print(f"Ответ: {response.text}")
        return response.status_code == 200
    except Exception as e:
        print(f"Ошибка при отправке уведомления: {e}")
        return False

def get_user_id_from_args():
    """Получает ID пользователя из аргументов командной строки."""
    if len(sys.argv) > 1:
        try:
            return int(sys.argv[1])
        except ValueError:
            print("Ошибка: ID пользователя должен быть числом.")
    return 1  # По умолчанию пользователь с ID 1

def main():
    user_id = get_user_id_from_args()
    print(f"Получение FCM токена для пользователя с ID: {user_id}")
    
    # Получаем FCM токен
    fcm_token = get_user_fcm_token(user_id)
    
    if not fcm_token:
        print("FCM токен не найден. Невозможно отправить уведомление.")
        return
    
    # Отправляем тестовое уведомление
    title = "Астра Такси"
    body = f"Тестовое уведомление для пользователя {user_id}. Время: {datetime.now().strftime('%H:%M:%S')}"
    
    success = send_test_notification(fcm_token, title, body)
    
    if success:
        print("Уведомление успешно отправлено!")
    else:
        print("Не удалось отправить уведомление.")

if __name__ == "__main__":
    main() 