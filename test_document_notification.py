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
FCM_URL = "https://fcm.googleapis.com/fcm/send"

def get_user_info(user_id):
    """Получает информацию о пользователе из API."""
    headers = {
        "Authorization": f"Bearer {ADMIN_TOKEN}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.get(f"{API_URL}/profile", headers=headers)
        if response.status_code == 200:
            user_data = response.json()
            print(f"Получены данные пользователя: {user_data}")
            return user_data
        else:
            print(f"Ошибка при получении профиля пользователя: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Исключение при получении информации о пользователе: {e}")
        return None

def get_document_info(user_id):
    """Получает информацию о документах пользователя из API."""
    headers = {
        "Authorization": f"Bearer {ADMIN_TOKEN}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.get(f"{API_URL}/driver/documents", headers=headers)
        if response.status_code == 200:
            doc_data = response.json()
            print(f"Получены данные о документах: {doc_data}")
            return doc_data
        else:
            print(f"Ошибка при получении документов: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Исключение при получении информации о документах: {e}")
        return None

def send_document_status_notification(fcm_token, doc_id, status, reason=None):
    """Отправляет уведомление об изменении статуса документа."""
    headers = {
        "Authorization": f"key={FCM_SERVER_KEY}",
        "Content-Type": "application/json"
    }
    
    # Формируем заголовок и текст уведомления в зависимости от статуса
    if status == "approved":
        title = "Документы одобрены"
        body = "Ваши документы проверены и одобрены. Теперь вы можете начать работу водителем."
    elif status == "rejected":
        title = "Документы отклонены"
        body = f"Ваши документы были отклонены. Причина: {reason or 'не указана'}"
    elif status == "pending":
        title = "Документы на проверке"
        body = "Ваши документы отправлены на проверку. Мы сообщим о результате."
    else:
        title = "Статус документов изменен"
        body = f"Статус ваших документов изменен на '{status}'."
    
    payload = {
        "to": fcm_token,
        "notification": {
            "title": title,
            "body": body,
        },
        "data": {
            "type": "DOCUMENT_STATUS_UPDATE",
            "time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "document_id": doc_id,
            "status": status,
            "rejection_reason": reason,
            "click_action": "FLUTTER_NOTIFICATION_CLICK"
        }
    }
    
    try:
        print(f"Отправка уведомления об изменении статуса документа: {json.dumps(payload, ensure_ascii=False)}")
        response = requests.post(FCM_URL, headers=headers, data=json.dumps(payload))
        print(f"Статус ответа: {response.status_code}")
        print(f"Ответ: {response.text}")
        return response.status_code == 200
    except Exception as e:
        print(f"Ошибка при отправке уведомления: {e}")
        return False

def simulate_websocket_message(fcm_token, doc_id, status, reason=None):
    """Отправляет уведомление, имитирующее WebSocket сообщение."""
    headers = {
        "Authorization": f"key={FCM_SERVER_KEY}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "to": fcm_token,
        "data": {
            "type": "DOCUMENT_STATUS_UPDATE",
            "payload": {
                "document_id": doc_id,
                "status": status,
                "rejection_reason": reason,
                "user_id": "1"  # ID пользователя
            }
        }
    }
    
    try:
        print(f"Отправка имитации WebSocket сообщения: {json.dumps(payload, ensure_ascii=False)}")
        response = requests.post(FCM_URL, headers=headers, data=json.dumps(payload))
        print(f"Статус ответа: {response.status_code}")
        print(f"Ответ: {response.text}")
        return response.status_code == 200
    except Exception as e:
        print(f"Ошибка при отправке имитации WebSocket сообщения: {e}")
        return False

def main():
    if len(sys.argv) < 3:
        print("Использование: python test_document_notification.py <пользователь_id> <статус> [причина]")
        print("Возможные статусы: approved, rejected, pending")
        return
    
    try:
        user_id = int(sys.argv[1])
    except ValueError:
        print("Ошибка: ID пользователя должен быть числом.")
        return
    
    status = sys.argv[2]
    if status not in ["approved", "rejected", "pending"]:
        print(f"Предупреждение: Неизвестный статус '{status}'. Продолжаем...")
    
    reason = sys.argv[3] if len(sys.argv) > 3 else None
    
    print(f"Получение информации о пользователе с ID: {user_id}")
    user_info = get_user_info(user_id)
    
    if not user_info:
        print("Информация о пользователе не найдена. Невозможно отправить уведомление.")
        return
    
    fcm_token = user_info.get("fcmToken")
    if not fcm_token:
        print("FCM токен не найден. Невозможно отправить уведомление.")
        return
    
    print(f"Получение информации о документах пользователя")
    doc_info = get_document_info(user_id)
    
    if not doc_info or not isinstance(doc_info, dict):
        print("Информация о документах не найдена. Создаем тестовое уведомление.")
        doc_id = 999  # Фиктивный ID
    else:
        doc_id = doc_info.get("id", 999)
    
    # Отправляем обычное push-уведомление
    print(f"Отправка уведомления об изменении статуса документа (ID: {doc_id}) на '{status}'")
    success1 = send_document_status_notification(fcm_token, doc_id, status, reason)
    
    # Имитируем WebSocket сообщение
    print(f"Отправка имитации WebSocket сообщения")
    success2 = simulate_websocket_message(fcm_token, doc_id, status, reason)
    
    if success1 and success2:
        print("Все уведомления успешно отправлены!")
    elif success1:
        print("Push-уведомление отправлено, но имитация WebSocket не удалась.")
    elif success2:
        print("Имитация WebSocket отправлена, но push-уведомление не удалось.")
    else:
        print("Не удалось отправить уведомления.")

if __name__ == "__main__":
    main() 