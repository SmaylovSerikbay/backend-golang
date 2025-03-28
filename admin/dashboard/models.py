from django.db import models
from django.utils import timezone


class APIModel(models.Model):
    class Meta:
        abstract = True
        managed = False

    def save(self, *args, **kwargs):
        pass

    def delete(self, *args, **kwargs):
        pass


class User(APIModel):
    id = models.IntegerField(primary_key=True)
    email = models.EmailField()
    first_name = models.CharField(max_length=255)
    last_name = models.CharField(max_length=255)
    phone = models.CharField(max_length=20)
    photo_url = models.URLField(blank=True, null=True)
    role = models.CharField(max_length=20)
    fcm_token = models.TextField(blank=True, null=True)
    created_at = models.DateTimeField()
    updated_at = models.DateTimeField()

    class Meta(APIModel.Meta):
        verbose_name = 'Пользователь'
        verbose_name_plural = 'Пользователи'
        db_table = 'users'

    def __str__(self):
        return f"{self.first_name} {self.last_name} ({self.email})"


class DriverDocument(APIModel):
    STATUS_CHOICES = [
        ('pending', 'На модерации'),
        ('approved', 'Принят'),
        ('rejected', 'Отказ'),
        ('revision', 'Доработка'),
    ]

    id = models.IntegerField(primary_key=True)
    user = models.ForeignKey(User, on_delete=models.DO_NOTHING)
    car_brand = models.CharField(max_length=100)
    car_model = models.CharField(max_length=100)
    car_year = models.CharField(max_length=4)
    car_color = models.CharField(max_length=50)
    car_number = models.CharField(max_length=20)
    driver_license_front = models.URLField()
    driver_license_back = models.URLField()
    car_registration_front = models.URLField()
    car_registration_back = models.URLField()
    car_photo_front = models.URLField(blank=True, null=True)
    car_photo_side = models.URLField(blank=True, null=True)
    car_photo_interior = models.URLField(blank=True, null=True)
    status = models.CharField(max_length=20, choices=STATUS_CHOICES, default='pending')
    rejection_reason = models.TextField(blank=True, null=True)
    created_at = models.DateTimeField()
    updated_at = models.DateTimeField()

    class Meta(APIModel.Meta):
        verbose_name = 'Документ водителя'
        verbose_name_plural = 'Документы водителей'
        db_table = 'driver_documents'

    def __str__(self):
        return f"Документы водителя {self.user.first_name} {self.user.last_name}"


class Ride(APIModel):
    STATUS_CHOICES = [
        ('active', 'Активная'),
        ('started', 'Начата'),
        ('completed', 'Завершена'),
        ('cancelled', 'Отменена'),
    ]

    id = models.IntegerField(primary_key=True)
    driver = models.ForeignKey(User, on_delete=models.DO_NOTHING, related_name='driver_rides')
    passenger = models.ForeignKey(User, on_delete=models.DO_NOTHING, related_name='passenger_rides', null=True)
    from_address = models.CharField(max_length=255)
    to_address = models.CharField(max_length=255)
    from_location = models.CharField(max_length=100)
    to_location = models.CharField(max_length=100)
    status = models.CharField(max_length=20, choices=STATUS_CHOICES)
    price = models.DecimalField(max_digits=10, decimal_places=2)
    seats_count = models.IntegerField()
    booked_seats = models.IntegerField(default=0)
    departure_date = models.DateTimeField()
    comment = models.TextField(blank=True)
    front_seat_price = models.DecimalField(max_digits=10, decimal_places=2, null=True, blank=True)
    back_seat_price = models.DecimalField(max_digits=10, decimal_places=2, null=True, blank=True)
    cancellation_reason = models.TextField(blank=True)
    created_at = models.DateTimeField()
    updated_at = models.DateTimeField()

    class Meta(APIModel.Meta):
        verbose_name = 'Поездка'
        verbose_name_plural = 'Поездки'
        db_table = 'rides'

    def __str__(self):
        return f"Поездка {self.from_address} -> {self.to_address}"


class Booking(APIModel):
    STATUS_CHOICES = [
        ('pending', 'Ожидает подтверждения'),
        ('approved', 'Подтверждено'),
        ('started', 'Начато'),
        ('rejected', 'Отклонено'),
        ('cancelled', 'Отменено'),
        ('completed', 'Завершено'),
    ]

    BOOKING_TYPE_CHOICES = [
        ('regular', 'Обычное'),
        ('front_seat', 'Переднее сиденье'),
        ('back_seat', 'Заднее сиденье'),
    ]

    id = models.IntegerField(primary_key=True)
    ride = models.ForeignKey(Ride, on_delete=models.DO_NOTHING)
    passenger = models.ForeignKey(User, on_delete=models.DO_NOTHING)
    pickup_address = models.CharField(max_length=255)
    dropoff_address = models.CharField(max_length=255)
    pickup_location = models.CharField(max_length=100)
    dropoff_location = models.CharField(max_length=100)
    seats_count = models.IntegerField()
    status = models.CharField(max_length=20, choices=STATUS_CHOICES)
    booking_type = models.CharField(max_length=20, choices=BOOKING_TYPE_CHOICES)
    price = models.DecimalField(max_digits=10, decimal_places=2)
    comment = models.TextField(blank=True)
    reject_reason = models.TextField(blank=True)
    created_at = models.DateTimeField()
    updated_at = models.DateTimeField()

    class Meta(APIModel.Meta):
        verbose_name = 'Бронирование'
        verbose_name_plural = 'Бронирования'
        db_table = 'bookings'

    def __str__(self):
        return f"Бронирование {self.passenger.first_name} {self.passenger.last_name}" 