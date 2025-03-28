import os
import requests
import logging
import re
from datetime import datetime
from django.conf import settings
from django.utils import timezone
from .models import User, DriverDocument, Ride, Booking

logger = logging.getLogger(__name__)

def parse_datetime(date_str):
    """
    Парсит строку даты в формате ISO 8601, включая варианты с 'Z' в конце
    и с микросекундами различной точности.
    """
    if not date_str:
        return timezone.now()
        
    try:
        # Если дата заканчивается на 'Z', заменяем на +00:00 (UTC)
        if date_str.endswith('Z'):
            date_str = date_str[:-1] + '+00:00'
            
        # Ограничиваем микросекунды до 6 цифр (максимум для Python)
        if '.' in date_str:
            parts = date_str.split('.')
            before_ms = parts[0]
            after_ms = parts[1]
            
            # Находим часть с микросекундами (до первой буквы или символа '+'/'-')
            ms_match = re.search(r'^(\d+)([^0-9].*)?$', after_ms)
            if ms_match:
                ms = ms_match.group(1)[:6]  # Берем только первые 6 цифр
                rest = ms_match.group(2) or ''
                date_str = f"{before_ms}.{ms}{rest}"
        
        # Пробуем разные форматы парсинга
        try:
            return datetime.fromisoformat(date_str)
        except ValueError:
            # Если не удалось, пробуем через strptime
            try:
                return datetime.strptime(date_str, "%Y-%m-%dT%H:%M:%S.%f%z")
            except ValueError:
                try:
                    return datetime.strptime(date_str, "%Y-%m-%dT%H:%M:%S%z")
                except ValueError:
                    return timezone.now()
    except Exception as e:
        logger.error(f"Ошибка при парсинге даты '{date_str}': {str(e)}")
        return timezone.now()

class APIService:
    def __init__(self):
        self.base_url = settings.GO_API_URL.rstrip('/')
        self.token = settings.ADMIN_API_TOKEN
        if not self.token:
            logger.error("ADMIN_API_TOKEN не установлен в настройках")
            raise ValueError("API token не установлен в настройках")

    def _get_headers(self):
        if not self.token:
            raise ValueError("API token не установлен в настройках")
        return {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {self.token.strip()}'
        }

    def _make_request(self, method, endpoint, data=None, params=None):
        url = f"{self.base_url}{endpoint}"
        try:
            logger.debug(f"Отправка {method} запроса к {url}")
            logger.debug(f"Заголовки запроса: {self._get_headers()}")
            if data:
                logger.debug(f"Данные запроса: {data}")
            if params:
                logger.debug(f"Параметры запроса: {params}")
            
            response = requests.request(
                method,
                url,
                headers=self._get_headers(),
                json=data,
                params=params,
                timeout=10,
                verify=True
            )
            
            logger.debug(f"Получен ответ от {url}: статус={response.status_code}")
            
            if response.status_code == 401:
                error_msg = f"Ошибка авторизации при запросе к {url}: {response.text}"
                logger.error(error_msg)
                return []
            
            if response.status_code == 403:
                error_msg = f"Ошибка доступа при запросе к {url}: {response.text}"
                logger.error(error_msg)
                return []
                
            if response.status_code == 500:
                error_msg = f"Внутренняя ошибка сервера при запросе к {url}: {response.text}"
                logger.error(error_msg)
                return []
            
            if response.status_code == 404:
                error_msg = f"Ресурс не найден при запросе к {url}: {response.text}"
                logger.error(error_msg)
                return []
            
            try:
                response.raise_for_status()
                
                # Логируем ответ API
                response_data = response.json()
                logger.debug(f"Ответ API ({url}): {response_data}")
                
                return response_data
            except ValueError:
                # Если ответ не содержит JSON
                logger.warning(f"Ответ не содержит JSON: {response.text}")
                return {"message": response.text}
            
        except requests.exceptions.RequestException as e:
            error_msg = f"Ошибка API при запросе к {url}: {str(e)}"
            logger.error(error_msg)
            return []
        except Exception as e:
            error_msg = f"Неожиданная ошибка при запросе к {url}: {str(e)}"
            logger.error(error_msg)
            return []

    def _create_user_from_data(self, data):
        if not data:
            logger.warning("Попытка создать пользователя из пустых данных")
            return None
        try:
            # Логируем входные данные
            logger.debug(f"Создание пользователя из данных: {data}")
            
            # Проверяем наличие обязательных полей
            if 'id' not in data:
                logger.warning("В данных пользователя отсутствует поле 'id'")
                # Для админского пользователя устанавливаем id=0
                if data.get('role') == 'admin':
                    user_id = 0
                else:
                    return None
            else:
                user_id = data.get('id')
            
            # Создаем пользователя с обработкой возможных отсутствующих полей
            user = User(
                id=user_id,
                email=data.get('email', ''),
                first_name=data.get('firstName', ''),
                last_name=data.get('lastName', ''),
                phone=data.get('phone', ''),
                photo_url=data.get('photoUrl', ''),
                role=data.get('role', ''),
                fcm_token=data.get('fcmToken', ''),
                created_at=parse_datetime(data.get('created_at')),
                updated_at=parse_datetime(data.get('updated_at'))
            )
            
            # Логируем созданный объект
            logger.debug(f"Создан пользователь: id={user.id}, email={user.email}, role={user.role}")
            
            return user
        except Exception as e:
            logger.error(f"Ошибка при создании пользователя из данных: {e}")
            logger.error(f"Данные пользователя: {data}")
            return None

    def _create_driver_document_from_data(self, data):
        if not data:
            logger.warning("Попытка создать документ водителя из пустых данных")
            return None
        try:
            # Создаем пользователя из данных, если они есть
            user = None
            if data.get('user'):
                user = self._create_user_from_data(data.get('user'))
            elif data.get('user_id'):
                # Если есть только ID пользователя, создаем минимальный объект
                user = User(
                    id=data.get('user_id'),
                    first_name="Пользователь",
                    last_name=f"ID: {data.get('user_id')}",
                    email="",
                    phone="",
                    role="driver"
                )
            else:
                logger.warning("Попытка создать пользователя из пустых данных")
                # Создаем заглушку пользователя
                user = User(
                    id=0,
                    first_name="Неизвестный",
                    last_name="пользователь",
                    email="",
                    phone="",
                    role="driver"
                )
                
            # Обрабатываем как camelCase, так и snake_case ключи
            return DriverDocument(
                id=data.get('id'),
                user=user,
                car_brand=data.get('car_brand', data.get('carBrand', '')),
                car_model=data.get('car_model', data.get('carModel', '')),
                car_year=data.get('car_year', data.get('carYear', '')),
                car_color=data.get('car_color', data.get('carColor', '')),
                car_number=data.get('car_number', data.get('carNumber', '')),
                driver_license_front=data.get('driver_license_front', data.get('driverLicenseFront', '')),
                driver_license_back=data.get('driver_license_back', data.get('driverLicenseBack', '')),
                car_registration_front=data.get('car_registration_front', data.get('carRegistrationFront', '')),
                car_registration_back=data.get('car_registration_back', data.get('carRegistrationBack', '')),
                car_photo_front=data.get('car_photo_front', data.get('carPhotoFront', '')),
                car_photo_side=data.get('car_photo_side', data.get('carPhotoSide', '')),
                car_photo_interior=data.get('car_photo_interior', data.get('carPhotoInterior', '')),
                status=data.get('status', 'pending'),
                rejection_reason=data.get('rejection_reason', data.get('rejectionReason', '')),
                created_at=parse_datetime(data.get('created_at')),
                updated_at=parse_datetime(data.get('updated_at'))
            )
        except Exception as e:
            logger.error(f"Ошибка при создании документа водителя из данных: {e}")
            logger.error(f"Данные документа: {data}")
            return None

    def _create_ride_from_data(self, data):
        if not data:
            return None
        try:
            driver = self._create_user_from_data(data.get('driver'))
            passenger = self._create_user_from_data(data.get('passenger'))
            return Ride(
                id=data.get('id'),
                driver=driver,
                passenger=passenger,
                from_address=data.get('fromAddress', ''),
                to_address=data.get('toAddress', ''),
                from_location=data.get('fromLocation', ''),
                to_location=data.get('toLocation', ''),
                status=data.get('status', 'pending'),
                price=data.get('price', 0),
                seats_count=data.get('seatsCount', 0),
                booked_seats=data.get('bookedSeats', 0),
                departure_date=parse_datetime(data.get('departureDate')),
                comment=data.get('comment', ''),
                front_seat_price=data.get('frontSeatPrice', 0),
                back_seat_price=data.get('backSeatPrice', 0),
                cancellation_reason=data.get('cancellationReason', ''),
                created_at=parse_datetime(data.get('created_at')),
                updated_at=parse_datetime(data.get('updated_at'))
            )
        except Exception as e:
            logger.error(f"Error creating ride from data: {e}")
            return None

    def _create_booking_from_data(self, data):
        if not data:
            return None
        try:
            passenger = self._create_user_from_data(data.get('passenger'))
            ride = self._create_ride_from_data(data.get('ride'))
            return Booking(
                id=data.get('id'),
                ride=ride,
                passenger=passenger,
                pickup_address=data.get('pickupAddress', ''),
                dropoff_address=data.get('dropoffAddress', ''),
                pickup_location=data.get('pickupLocation', ''),
                dropoff_location=data.get('dropoffLocation', ''),
                seats_count=data.get('seatsCount', 0),
                status=data.get('status', 'pending'),
                booking_type=data.get('bookingType', 'standard'),
                price=data.get('price', 0),
                comment=data.get('comment', ''),
                reject_reason=data.get('rejectReason', ''),
                created_at=parse_datetime(data.get('created_at')),
                updated_at=parse_datetime(data.get('updated_at'))
            )
        except Exception as e:
            logger.error(f"Error creating booking from data: {e}")
            return None

    def get_users(self):
        try:
            data = self._make_request('GET', '/profile')
            logger.debug(f"Получены данные пользователей: {data}")
            
            if not data:
                logger.warning("API вернул пустые данные для пользователей")
                return []
                
            # Проверяем, является ли data словарем (одиночный объект)
            if isinstance(data, dict):
                logger.debug(f"API вернул одиночный объект пользователя: {data}")
                user = self._create_user_from_data(data)
                if user:
                    logger.debug(f"Создан объект пользователя: {user.id}, {user.email}, {user.first_name}")
                    return [user]
                logger.warning("Не удалось создать объект пользователя из данных")
                return []
            
            # Если data - список, обрабатываем каждый элемент
            if isinstance(data, list):
                logger.debug(f"API вернул список пользователей длиной {len(data)}")
                users = [self._create_user_from_data(user_data) for user_data in data if user_data]
                logger.debug(f"Создано {len(users)} объектов пользователей")
                return users
                
            # Если data не словарь и не список, возвращаем пустой список
            logger.error(f"Неожиданный формат данных от API: {type(data)}")
            return []
        except Exception as e:
            logger.error(f"Ошибка при получении пользователей: {str(e)}")
            return []

    def get_user(self, user_id):
        data = self._make_request('GET', f'/profile/{user_id}')
        return self._create_user_from_data(data)

    def get_driver_documents(self, params=None):
        try:
            data = self._make_request('GET', '/driver/documents', params=params)
            if not data:
                return []
                
            # Проверяем, является ли data словарем (одиночный объект)
            if isinstance(data, dict):
                doc = self._create_driver_document_from_data(data)
                if doc:
                    return [doc]
                return []
            
            # Если data - список, обрабатываем каждый элемент
            if isinstance(data, list):
                return [self._create_driver_document_from_data(doc_data) for doc_data in data if doc_data]
                
            # Если data не словарь и не список, возвращаем пустой список
            logger.error(f"Неожиданный формат данных от API: {type(data)}")
            return []
        except Exception as e:
            logger.error(f"Ошибка при получении документов водителя: {str(e)}")
            return []

    def get_driver_document(self, doc_id):
        data = self._make_request('GET', f'/driver/documents/{doc_id}')
        return self._create_driver_document_from_data(data)

    def update_driver_document(self, doc_id, data):
        logger.debug(f"Отправка запроса на обновление статуса документа ID {doc_id}: {data}")
        return self._make_request('PUT', f'/driver/documents/{doc_id}/status', data=data)

    def get_rides(self, params=None):
        try:
            data = self._make_request('GET', '/rides', params=params)
            if not data:
                return []
                
            # Проверяем, является ли data словарем (одиночный объект)
            if isinstance(data, dict):
                ride = self._create_ride_from_data(data)
                if ride:
                    return [ride]
                return []
            
            # Если data - список, обрабатываем каждый элемент
            if isinstance(data, list):
                return [self._create_ride_from_data(ride_data) for ride_data in data if ride_data]
                
            # Если data не словарь и не список, возвращаем пустой список
            logger.error(f"Неожиданный формат данных от API: {type(data)}")
            return []
        except Exception as e:
            logger.error(f"Ошибка при получении поездок: {str(e)}")
            return []

    def get_ride(self, ride_id):
        data = self._make_request('GET', f'/rides/{ride_id}')
        return self._create_ride_from_data(data)

    def update_ride(self, ride_id, data):
        return self._make_request('PUT', f'/rides/{ride_id}', data=data)

    def get_bookings(self, params=None):
        try:
            data = self._make_request('GET', '/bookings', params=params)
            if not data:
                return []
                
            # Проверяем, является ли data словарем (одиночный объект)
            if isinstance(data, dict):
                booking = self._create_booking_from_data(data)
                if booking:
                    return [booking]
                return []
            
            # Если data - список, обрабатываем каждый элемент
            if isinstance(data, list):
                return [self._create_booking_from_data(booking_data) for booking_data in data if booking_data]
                
            # Если data не словарь и не список, возвращаем пустой список
            logger.error(f"Неожиданный формат данных от API: {type(data)}")
            return []
        except Exception as e:
            logger.error(f"Ошибка при получении бронирований: {str(e)}")
            return []

    def get_booking(self, booking_id):
        data = self._make_request('GET', f'/bookings/{booking_id}')
        return self._create_booking_from_data(data)

    def update_booking(self, booking_id, data):
        return self._make_request('PUT', f'/bookings/{booking_id}', data=data)


api_service = APIService() 