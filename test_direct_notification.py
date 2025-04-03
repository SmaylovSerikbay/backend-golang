#!/usr/bin/env python3
import requests
import json
import os
import sys
from datetime import datetime

# Конфигурация
API_URL = "http://192.168.0.169:80/api"  # Используем Nginx
ADMIN_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjowLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NzM2NTM1MzEsImlhdCI6MTc0MjExNzUzMX0.9BOK0_cPzIRjJq6tdv8gdy6nyETHbCpxBeaj92IUqio"

def get_user_profile():
    """Получает профиль текущего пользователя."""
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
            print(f"Ошибка при получении профиля: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Исключение при получении профиля: {e}")
        return None

def get_users_list():
    """Получает список всех пользователей."""
    headers = {
        "Authorization": f"Bearer {ADMIN_TOKEN}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.get(f"{API_URL}/admin/users", headers=headers)
        if response.status_code == 200:
            users_data = response.json()
            print(f"Получен список пользователей, всего {len(users_data)} пользователей")
            return users_data
        else:
            print(f"Ошибка при получении списка пользователей: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Исключение при получении списка пользователей: {e}")
        return None

def update_document_status(doc_id, status, reason=None):
    """Обновляет статус документа водителя."""
    headers = {
        "Authorization": f"Bearer {ADMIN_TOKEN}",
        "Content-Type": "application/json"
    }
    
    data = {
        "status": status
    }
    
    if reason:
        data["rejection_reason"] = reason
    
    try:
        url = f"{API_URL}/driver/documents/{doc_id}/status"
        print(f"Отправка запроса на {url} с данными {data}")
        
        response = requests.put(url, headers=headers, json=data)
        if response.status_code == 200:
            result = response.json()
            print(f"Статус документа успешно обновлен: {result}")
            return True
        else:
            print(f"Ошибка при обновлении статуса документа: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"Исключение при обновлении статуса документа: {e}")
        return False

def get_document_info():
    """Получает информацию о документах водителя."""
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

def get_document_id_from_args():
    """Получает ID документа из аргументов командной строки."""
    if len(sys.argv) > 1:
        try:
            return int(sys.argv[1])
        except ValueError:
            print("Ошибка: ID документа должен быть числом.")
    return 2  # По умолчанию документ с ID 2 (из логов)

def main():
    if len(sys.argv) < 3:
        print("Использование: python test_direct_notification.py <doc_id> <статус> [причина]")
        print("Возможные статусы: approved, rejected, pending")
        return
    
    try:
        doc_id = int(sys.argv[1])
    except ValueError:
        print("Ошибка: ID документа должен быть числом.")
        return
    
    status = sys.argv[2]
    if status not in ["approved", "rejected", "pending"]:
        print(f"Предупреждение: Неизвестный статус '{status}'. Продолжаем...")
    
    reason = sys.argv[3] if len(sys.argv) > 3 else None
    
    # Обновляем статус документа
    print(f"Обновление статуса документа (ID: {doc_id}) на '{status}'")
    success = update_document_status(doc_id, status, reason)
    
    if success:
        print("Статус документа успешно обновлен!")
    else:
        print("Не удалось обновить статус документа.")

if __name__ == "__main__":
    main() 