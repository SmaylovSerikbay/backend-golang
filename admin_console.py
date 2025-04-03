import requests
import json
import webbrowser
import os
from typing import Optional, List, Dict
from datetime import datetime
from dataclasses import dataclass
from enum import Enum
from PIL import Image
import io
import tempfile
import subprocess
from tabulate import tabulate
from colorama import init, Fore, Style

# Инициализация colorama для Windows
init()

# Константы
API_URL = "http://localhost/api"
NGINX_URL = "http://localhost"  # Базовый URL для nginx
ADMIN_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjowLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NzM2NTM1MzEsImlhdCI6MTc0MjExNzUzMX0.9BOK0_cPzIRjJq6tdv8gdy6nyETHbCpxBeaj92IUqio"

class DocumentStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    REVISION = "revision"

@dataclass
class User:
    id: int
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
        data = response.json()
        if data is None:
            return []
        return [DriverDocument.from_dict(doc) for doc in data]

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

    def get_image(self, url: str) -> bytes:
        """Получить изображение по URL"""
        response = self.session.get(url)
        response.raise_for_status()
        return response.content

def get_status_color(status: DocumentStatus) -> str:
    """Получить цвет для статуса"""
    colors = {
        DocumentStatus.PENDING: Fore.YELLOW,
        DocumentStatus.APPROVED: Fore.GREEN,
        DocumentStatus.REJECTED: Fore.RED,
        DocumentStatus.REVISION: Fore.BLUE
    }
    return colors.get(status, Fore.WHITE)

def create_html_gallery(images: List[Dict[str, str]], title: str) -> str:
    """Создать HTML страницу с галереей изображений"""
    html = f"""
    <!DOCTYPE html>
    <html>
    <head>
        <title>{title}</title>
        <style>
            body {{ font-family: Arial, sans-serif; margin: 20px; background: #f0f0f0; }}
            .gallery {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }}
            .image-container {{ background: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }}
            .image-container img {{ width: 100%; height: auto; border-radius: 4px; }}
            .image-title {{ margin: 10px 0; font-size: 14px; color: #333; }}
            h1 {{ color: #333; text-align: center; }}
        </style>
    </head>
    <body>
        <h1>{title}</h1>
        <div class="gallery">
    """
    
    for img in images:
        html += f"""
            <div class="image-container">
                <img src="{img['url']}" alt="{img['title']}">
                <div class="image-title">{img['title']}</div>
            </div>
        """
    
    html += """
        </div>
    </body>
    </html>
    """
    return html

def view_all_images(api: DriverDocumentsAPI, doc: DriverDocument):
    """Открыть все изображения документа на одной странице"""
    images = []
    
    # Собираем все изображения документа
    image_map = {
        'Водительское удостоверение (перед)': doc.driver_license_front,
        'Водительское удостоверение (зад)': doc.driver_license_back,
        'Свидетельство о регистрации (перед)': doc.car_registration_front,
        'Свидетельство о регистрации (зад)': doc.car_registration_back,
        'Фото автомобиля (перед)': doc.car_photo_front,
        'Фото автомобиля (сбоку)': doc.car_photo_side,
        'Фото салона': doc.car_photo_interior
    }
    
    for title, url in image_map.items():
        if url:
            if not url.startswith('http'):
                url = f"{NGINX_URL}/{url.lstrip('/')}"
            images.append({'title': title, 'url': url})
    
    if not images:
        print(Fore.RED + "\nНет доступных изображений" + Style.RESET_ALL)
        return
    
    # Создаем временный HTML файл
    with tempfile.NamedTemporaryFile('w', delete=False, suffix='.html', encoding='utf-8') as f:
        html_content = create_html_gallery(
            images,
            f"Документы водителя {doc.user.first_name} {doc.user.last_name}"
        )
        f.write(html_content)
        temp_path = f.name
    
    # Открываем файл в браузере
    webbrowser.open('file://' + os.path.abspath(temp_path))

def display_documents_table(documents: List[DriverDocument]):
    """Отображение списка документов в виде таблицы со всеми данными"""
    if not documents:
        print(Fore.YELLOW + "\nДокументы не найдены" + Style.RESET_ALL)
        return

    # Подготовка данных для таблицы
    table_data = []
    for doc in documents:
        status_colored = f"{get_status_color(doc.status)}{doc.status}{Style.RESET_ALL}"
        
        # Формируем строку с количеством загруженных фото
        total_photos = sum(1 for photo in [
            doc.driver_license_front,
            doc.driver_license_back,
            doc.car_registration_front,
            doc.car_registration_back,
            doc.car_photo_front,
            doc.car_photo_side,
            doc.car_photo_interior
        ] if photo)
        
        row = [
            doc.id,
            f"{doc.user.first_name} {doc.user.last_name}",
            doc.user.phone,
            f"{doc.car_brand} {doc.car_model} ({doc.car_year})",
            doc.car_color,
            doc.car_number,
            status_colored,
            doc.rejection_reason or "-",
            f"{total_photos}/7",
            doc.created_at.strftime("%Y-%m-%d %H:%M")
        ]
        table_data.append(row)

    # Заголовки таблицы
    headers = [
        "ID",
        "Водитель",
        "Телефон",
        "Автомобиль",
        "Цвет",
        "Номер",
        "Статус",
        "Причина отказа",
        "Фото",
        "Создан"
    ]

    # Вывод таблицы
    print("\n" + tabulate(
        table_data,
        headers=headers,
        tablefmt="grid",
        numalign="center",
        stralign="left"
    ))

def main():
    api = DriverDocumentsAPI(API_URL, ADMIN_TOKEN)

    while True:
        print("\n" + Fore.CYAN + "Меню:" + Style.RESET_ALL)
        print("1. Показать список документов")
        print("2. Просмотреть изображения документа")
        print("3. Изменить статус документа")
        print("4. Выход")

        choice = input("\nВыберите действие (1-4): ")

        if choice == '1':
            try:
                documents = api.get_documents()
                display_documents_table(documents)
            except Exception as e:
                print(Fore.RED + f"\nОшибка при получении документов: {e}" + Style.RESET_ALL)

        elif choice == '2':
            try:
                doc_id = int(input("Введите ID документа: "))
                documents = api.get_documents()
                doc = next((d for d in documents if d.id == doc_id), None)
                
                if not doc:
                    print(Fore.RED + "\nДокумент не найден" + Style.RESET_ALL)
                    continue

                view_all_images(api, doc)

            except ValueError:
                print(Fore.RED + "\nНекорректный ID документа" + Style.RESET_ALL)
            except Exception as e:
                print(Fore.RED + f"\nОшибка: {e}" + Style.RESET_ALL)

        elif choice == '3':
            try:
                doc_id = int(input("Введите ID документа: "))
                documents = api.get_documents()
                doc = next((d for d in documents if d.id == doc_id), None)
                
                if not doc:
                    print(Fore.RED + "\nДокумент не найден" + Style.RESET_ALL)
                    continue

                print("\nТекущий статус:", get_status_color(doc.status) + str(doc.status) + Style.RESET_ALL)
                print("\nДоступные статусы:")
                print("1. Одобрить")
                print("2. Отклонить")
                print("3. Отправить на доработку")
                print("4. Отмена")

                status_choice = input("\nВыберите действие (1-4): ")

                if status_choice == '1':
                    doc = api.update_document_status(doc.id, DocumentStatus.APPROVED)
                    print(Fore.GREEN + "\nДокументы успешно одобрены" + Style.RESET_ALL)
                elif status_choice == '2':
                    reason = input("Введите причину отказа: ")
                    doc = api.update_document_status(doc.id, DocumentStatus.REJECTED, reason)
                    print(Fore.YELLOW + "\nДокументы отклонены" + Style.RESET_ALL)
                elif status_choice == '3':
                    reason = input("Введите причину доработки: ")
                    doc = api.update_document_status(doc.id, DocumentStatus.REVISION, reason)
                    print(Fore.BLUE + "\nДокументы отправлены на доработку" + Style.RESET_ALL)

            except ValueError:
                print(Fore.RED + "\nНекорректный ID документа" + Style.RESET_ALL)
            except Exception as e:
                print(Fore.RED + f"\nОшибка: {e}" + Style.RESET_ALL)

        elif choice == '4':
            break

if __name__ == "__main__":
    main() 