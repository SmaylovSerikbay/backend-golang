import requests
import json
from typing import Optional, List, Dict
from datetime import datetime
from dataclasses import dataclass
from enum import Enum

# Константы
API_URL = "http://localhost:8080/api"
ADMIN_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjowLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NzM2NTM1MzEsImlhdCI6MTc0MjExNzUzMX0.9BOK0_cPzIRjJq6tdv8gdy6nyETHbCpxBeaj92IUqio"

class DocumentStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    REVISION = "revision"

@dataclass
class User:
    id: int
    email: str
    first_name: str
    last_name: str
    phone: str
    photo_url: str
    role: str

@dataclass
class DriverDocument:
    id: int
    user_id: int
    user: Optional[User]
    car_brand: str
    car_model: str
    car_year: str
    car_color: str
    car_number: str
    driver_license_front: str
    driver_license_back: str
    car_registration_front: str
    car_registration_back: str
    car_photo_front: Optional[str]
    car_photo_side: Optional[str]
    car_photo_interior: Optional[str]
    status: DocumentStatus
    rejection_reason: Optional[str]
    created_at: datetime
    updated_at: datetime

    @classmethod
    def from_dict(cls, data: Dict) -> 'DriverDocument':
        user = None
        if 'user' in data and data['user']:
            user = User(
                id=data['user']['id'],
                email=data['user']['email'],
                first_name=data['user']['firstName'],
                last_name=data['user']['lastName'],
                phone=data['user']['phone'],
                photo_url=data['user']['photoUrl'],
                role=data['user']['role']
            )
        
        return cls(
            id=data['id'],
            user_id=data['user_id'],
            user=user,
            car_brand=data['car_brand'],
            car_model=data['car_model'],
            car_year=data['car_year'],
            car_color=data['car_color'],
            car_number=data['car_number'],
            driver_license_front=data['driver_license_front'],
            driver_license_back=data['driver_license_back'],
            car_registration_front=data['car_registration_front'],
            car_registration_back=data['car_registration_back'],
            car_photo_front=data.get('car_photo_front'),
            car_photo_side=data.get('car_photo_side'),
            car_photo_interior=data.get('car_photo_interior'),
            status=DocumentStatus(data['status']),
            rejection_reason=data.get('rejection_reason'),
            created_at=datetime.fromisoformat(data['created_at'].replace('Z', '+00:00')),
            updated_at=datetime.fromisoformat(data['updated_at'].replace('Z', '+00:00'))
        )

class DriverDocumentsAPI:
    def __init__(self, base_url: str, admin_token: str):
        self.base_url = base_url
        self.admin_token = admin_token
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {admin_token}',
            'Content-Type': 'application/json'
        })

    def get_documents(self) -> List[DriverDocument]:
        """Получить список всех документов водителей"""
        response = self.session.get(f"{self.base_url}/driver/documents")
        response.raise_for_status()
        return [DriverDocument.from_dict(doc) for doc in response.json()]

    def update_document_status(self, doc_id: int, status: DocumentStatus, rejection_reason: Optional[str] = None) -> DriverDocument:
        """Обновить статус документа"""
        data = {'status': status}
        if rejection_reason:
            data['rejection_reason'] = rejection_reason

        response = self.session.put(
            f"{self.base_url}/driver/documents/{doc_id}/status",
            json=data
        )
        response.raise_for_status()
        return DriverDocument.from_dict(response.json())

def main():
    api = DriverDocumentsAPI(API_URL, ADMIN_TOKEN)

    # Получаем список всех документов
    print("Получение списка документов...")
    try:
        documents = api.get_documents()
        print(f"Найдено документов: {len(documents)}")
        for doc in documents:
            print(f"\nДокумент ID: {doc.id}")
            print(f"Пользователь: {doc.user.first_name} {doc.user.last_name} ({doc.user.email})")
            print(f"Автомобиль: {doc.car_brand} {doc.car_model} {doc.car_year}, цвет: {doc.car_color}")
            print(f"Гос. номер: {doc.car_number}")
            print(f"Статус: {doc.status}")
            if doc.rejection_reason:
                print(f"Причина отказа: {doc.rejection_reason}")
            print(f"Создан: {doc.created_at}")
            print(f"Обновлен: {doc.updated_at}")

        # Для первого документа со статусом "pending" пробуем обновить статус
        pending_docs = [doc for doc in documents if doc.status == DocumentStatus.PENDING]
        if pending_docs:
            doc = pending_docs[0]
            print(f"\nОбновление статуса для документа {doc.id}...")
            
            # Пробуем одобрить документ
            updated_doc = api.update_document_status(doc.id, DocumentStatus.APPROVED)
            print(f"Статус обновлен на: {updated_doc.status}")

            # Пробуем отклонить документ
            updated_doc = api.update_document_status(
                doc.id,
                DocumentStatus.REJECTED,
                rejection_reason="Тестовое отклонение"
            )
            print(f"Статус обновлен на: {updated_doc.status}")
            print(f"Причина отказа: {updated_doc.rejection_reason}")

    except requests.exceptions.RequestException as e:
        print(f"Ошибка при работе с API: {e}")
        if hasattr(e.response, 'text'):
            print(f"Ответ сервера: {e.response.text}")

if __name__ == "__main__":
    main() 