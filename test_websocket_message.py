#!/usr/bin/env python3
import requests
import json
import os
import sys
from datetime import datetime

# Конфигурация
FCM_SERVER_KEY = "BPB4eaquHe47fdthXUU3ogNCuR0aRhr9SQjoi2_rOd67KQe8O4l8AKPlkrYnYbyNMwo9TJEAGL2i03sODRrKlgs"
FCM_URL = "https://fcm.googleapis.com/fcm/send"

def get_fcm_token_from_logs():
    """Получаем FCM токен из логов приложения."""
    return "cyY-VNbIQkKAP_0D5qR4db:APA91bFYHSga9nSMSVJ2X2-Tpcob5yybwGWEdh6zSXurERa-nlHRhZa03NsYDnde2kaOTY6reVTEsc0nzCyLooL3xPsZj0GCHKjRAjhlSRolLx8xW7yGWIY"

def simulate_websocket_message(fcm_token, doc_id, user_id, status, reason=None):
    """Отправляет имитацию WebSocket сообщения через FCM."""
    headers = {
        "Authorization": f"key={FCM_SERVER_KEY}",
        "Content-Type": "application/json"
    }
    
    # Создаем сообщение в формате, который ожидает приложение
    payload = {
        "to": fcm_token,
        "data": {
            "type": "DOCUMENT_STATUS_UPDATE",
            "payload": {
                "document_id": doc_id,
                "status": status,
                "rejection_reason": reason,
                "user_id": user_id
            }
        }
    }
    
    try:
        print(f"Отправка имитации WebSocket сообщения через FCM:")
        print(json.dumps(payload, indent=2, ensure_ascii=False))
        
        response = requests.post(FCM_URL, headers=headers, json=payload)
        print(f"Статус ответа: {response.status_code}")
        print(f"Ответ: {response.text}")
        
        return response.status_code == 200
    except Exception as e:
        print(f"Ошибка при отправке имитации WebSocket сообщения: {e}")
        return False

def main():
    if len(sys.argv) < 4:
        print("Использование: python test_websocket_message.py <документ_id> <пользователь_id> <статус> [причина]")
        print("Возможные статусы: approved, rejected, pending")
        return
    
    try:
        doc_id = int(sys.argv[1])
    except ValueError:
        print("Ошибка: ID документа должен быть числом.")
        return
    
    try:
        user_id = int(sys.argv[2])
    except ValueError:
        print("Ошибка: ID пользователя должен быть числом.")
        return
    
    status = sys.argv[3]
    if status not in ["approved", "rejected", "pending"]:
        print(f"Предупреждение: Неизвестный статус '{status}'. Продолжаем...")
    
    reason = sys.argv[4] if len(sys.argv) > 4 else None
    
    # Получаем FCM токен
    fcm_token = get_fcm_token_from_logs()
    
    # Отправляем имитацию WebSocket сообщения
    success = simulate_websocket_message(fcm_token, doc_id, user_id, status, reason)
    
    if success:
        print("Имитация WebSocket сообщения успешно отправлена!")
    else:
        print("Не удалось отправить имитацию WebSocket сообщения.")

if __name__ == "__main__":
    main() 