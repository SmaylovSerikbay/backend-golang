from django.contrib import admin
from django.utils.html import format_html
from django.urls import reverse
from django.utils.safestring import mark_safe
from django.core.paginator import Paginator
from django.db.models import Q
from .models import User, DriverDocument, Ride, Booking
from .services import api_service
import logging


class APIModelAdmin(admin.ModelAdmin):
    def get_queryset(self, request):
        # Возвращаем MockQuerySet вместо пустого списка
        class MockQuery:
            def __init__(self):
                self.order_by = []
                self.select_related = []
                self.distinct_fields = []
                self.where = None
                self.low_mark = 0
                self.high_mark = None
                self.used_aliases = set()
                self.has_filters = False
                self.is_sliced = False
        
        class MockQuerySet(list):
            def __init__(self, *args, **kwargs):
                super().__init__(*args, **kwargs)
                self.query = MockQuery()
                self.model = None
                self._prefetch_related_lookups = []
                self._result_cache = None
                self._db = None
                self._hints = {}
                self._known_related_objects = {}
                self._iterable_class = None
                self._fields = None
            
            def filter(self, *args, **kwargs):
                return self
            
            def order_by(self, *args, **kwargs):
                return self
            
            def distinct(self, *args, **kwargs):
                return self
            
            def select_related(self, *args, **kwargs):
                return self
            
            def prefetch_related(self, *args, **kwargs):
                return self
            
            def all(self):
                return self
            
            def none(self):
                return MockQuerySet([])
            
            def count(self):
                return len(self)
            
            def exclude(self, *args, **kwargs):
                return self
            
            def values(self, *args, **kwargs):
                return self
            
            def values_list(self, *args, **kwargs):
                return self
            
            def annotate(self, *args, **kwargs):
                return self
            
            def aggregate(self, *args, **kwargs):
                return {}
            
            def exists(self):
                return len(self) > 0
            
            def get(self, *args, **kwargs):
                if len(self) > 0:
                    return self[0]
                return None
            
            def first(self):
                if len(self) > 0:
                    return self[0]
                return None
            
            def last(self):
                if len(self) > 0:
                    return self[-1]
                return None
            
            def __getitem__(self, key):
                if isinstance(key, slice):
                    return MockQuerySet(super().__getitem__(key))
                return super().__getitem__(key)
            
            def _clone(self):
                return MockQuerySet(self)
            
            def using(self, alias):
                return self
            
            def complex_filter(self, filter_obj):
                return self
            
            def defer(self, *fields):
                return self
            
            def only(self, *fields):
                return self
            
            def ordered(self):
                return True
            
            def iterator(self):
                return iter(self)
            
            def _fetch_all(self):
                pass
            
            def explain(self, **options):
                return "Mock QuerySet"
        
        # Создаем пустой MockQuerySet
        mock_queryset = MockQuerySet([])
        mock_queryset.model = self.model
        return mock_queryset

    def create_mock_queryset(self, objects, model=None):
        """
        Создает MockQuerySet из списка объектов.
        
        Args:
            objects: Список объектов для включения в QuerySet
            model: Модель для QuerySet (по умолчанию self.model)
            
        Returns:
            MockQuerySet с объектами и настроенной моделью
        """
        if model is None:
            model = self.model
            
        class MockQuery:
            def __init__(self):
                self.order_by = []
                self.select_related = []
                self.distinct_fields = []
                self.where = None
                self.low_mark = 0
                self.high_mark = None
                self.used_aliases = set()
                self.has_filters = False
                self.is_sliced = False
        
        class MockQuerySet(list):
            def __init__(self, *args, **kwargs):
                super().__init__(*args, **kwargs)
                self.query = MockQuery()
                self.model = model
                self._prefetch_related_lookups = []
                self._result_cache = None
                self._db = None
                self._hints = {}
                self._known_related_objects = {}
                self._iterable_class = None
                self._fields = None
            
            def filter(self, *args, **kwargs):
                return self
            
            def order_by(self, *args, **kwargs):
                return self
            
            def distinct(self, *args, **kwargs):
                return self
            
            def select_related(self, *args, **kwargs):
                return self
            
            def prefetch_related(self, *args, **kwargs):
                return self
            
            def all(self):
                return self
            
            def none(self):
                return MockQuerySet([])
            
            def count(self):
                return len(self)
            
            def exclude(self, *args, **kwargs):
                return self
            
            def values(self, *args, **kwargs):
                return self
            
            def values_list(self, *args, **kwargs):
                return self
            
            def annotate(self, *args, **kwargs):
                return self
            
            def aggregate(self, *args, **kwargs):
                return {}
            
            def exists(self):
                return len(self) > 0
            
            def get(self, *args, **kwargs):
                if len(self) > 0:
                    return self[0]
                return None
            
            def first(self):
                if len(self) > 0:
                    return self[0]
                return None
            
            def last(self):
                if len(self) > 0:
                    return self[-1]
                return None
            
            def __getitem__(self, key):
                if isinstance(key, slice):
                    return MockQuerySet(super().__getitem__(key))
                return super().__getitem__(key)
                
            def _clone(self):
                return MockQuerySet(self)
                
            def using(self, alias):
                return self
                
            def complex_filter(self, filter_obj):
                return self
                
            def defer(self, *fields):
                return self
                
            def only(self, *fields):
                return self
                
            def ordered(self):
                return True
                
            def iterator(self):
                return iter(self)
                
            def _fetch_all(self):
                pass
                
            def explain(self, **options):
                return "Mock QuerySet"
        
        # Возвращаем MockQuerySet с объектами
        return MockQuerySet(objects)

    def get_search_results(self, request, queryset, search_term):
        objects = self.get_api_objects(request)
        if not search_term:
            return objects, False

        filtered_objects = []
        search_fields = self.get_search_fields(request)
        
        for obj in objects:
            for field in search_fields:
                value = self._get_field_value(obj, field)
                if value and search_term.lower() in str(value).lower():
                    filtered_objects.append(obj)
                    break
        
        # Возвращаем MockQuerySet вместо обычного списка
        return self.create_mock_queryset(filtered_objects), False

    def _get_field_value(self, obj, field):
        parts = field.split('__')
        value = obj
        for part in parts:
            if hasattr(value, part):
                value = getattr(value, part)
            else:
                return None
        return value

    def get_list_filter(self, request):
        return []

    def changelist_view(self, request, extra_context=None):
        from django.contrib.admin.views.main import ChangeList
        from django.template.response import TemplateResponse
        from django.contrib import messages
        
        logger = logging.getLogger(__name__)
        
        try:
            # Получаем объекты
            objects = self.get_api_objects(request) or []
            logger.debug(f"Получено {len(objects)} объектов из API")
            
            # Применяем поиск
            if 'q' in request.GET and request.GET['q']:
                objects, _ = self.get_search_results(request, None, request.GET['q'])
                logger.debug(f"После поиска осталось {len(objects)} объектов")
            
            # Создаем пагинатор
            paginator = Paginator(objects, self.list_per_page)
            page = request.GET.get('p', 1)
            try:
                page_obj = paginator.page(page)
            except:
                page_obj = paginator.page(1)
            
            # Создаем объект ChangeList без использования filter
            list_display = self.get_list_display(request)
            list_display_links = self.get_list_display_links(request, list_display)
            
            # Оборачиваем наш список в MockQuerySet
            mock_queryset = self.create_mock_queryset(objects)
            logger.debug(f"Создан MockQuerySet с {len(mock_queryset)} объектами")
            
            try:
                # Создаем ChangeList с нашим mock_queryset вместо None
                cl = ChangeList(
                    request,
                    self.model,
                    list_display,
                    list_display_links,
                    self.list_filter,
                    self.date_hierarchy,
                    self.search_fields,
                    self.list_select_related,
                    self.list_per_page,
                    self.list_max_show_all,
                    self.list_editable,
                    self,
                    self.sortable_by,
                    self.search_help_text,
                )
                
                # Важно: устанавливаем queryset до того, как будут использованы другие атрибуты
                cl.queryset = mock_queryset
                
                # Затем устанавливаем остальные атрибуты
                cl.result_list = page_obj.object_list
                cl.result_count = len(objects)
                cl.page_obj = page_obj
                cl.paginator = paginator
                cl.show_all = False
                cl.can_show_all = False
                cl.multi_page = paginator.num_pages > 1
                cl.show_full_result_count = True
                cl.formset = None
                cl.opts = self.model._meta
                cl.has_add_permission = self.has_add_permission(request)
                cl.has_change_permission = self.has_change_permission(request)
                cl.has_delete_permission = self.has_delete_permission(request)
                cl.has_view_permission = self.has_view_permission(request)
                cl.actions = []
                cl.actions_on_top = self.actions_on_top
                cl.actions_on_bottom = self.actions_on_bottom
                cl.actions_selection_counter = self.actions_selection_counter
                cl.model_admin = self
                cl.preserved_filters = self.get_preserved_filters(request)
                cl.sortable_by = self.sortable_by
                cl.search_help_text = self.search_help_text
                cl.date_hierarchy = self.date_hierarchy
                cl.list_filter = self.list_filter
                cl.list_display = list_display
                cl.list_display_links = list_display_links
                cl.list_select_related = self.list_select_related
                cl.list_per_page = self.list_per_page
                cl.list_max_show_all = self.list_max_show_all
                cl.list_editable = self.list_editable
                cl.search_fields = self.search_fields
                
                # Добавляем метод get_filters_params
                def get_filters_params(request=None):
                    return {}
                cl.get_filters_params = get_filters_params
                
                # Добавляем метод get_results
                def get_results(request):
                    return cl.result_list
                cl.get_results = get_results
                
                # Добавляем метод url_for_result
                def url_for_result(result):
                    pk_name = cl.opts.pk.attname
                    return reverse('admin:%s_%s_change' % (cl.opts.app_label, cl.opts.model_name),
                                  args=(getattr(result, pk_name),),
                                  current_app=self.admin_site.name)
                cl.url_for_result = url_for_result
                
                # Добавляем метод get_ordering_field
                def get_ordering_field(field_name):
                    return None
                cl.get_ordering_field = get_ordering_field
                
                # Добавляем метод get_ordering_field_columns
                def get_ordering_field_columns():
                    return {}
                cl.get_ordering_field_columns = get_ordering_field_columns
                
                # Добавляем метод get_query_string
                def get_query_string(new_params=None, remove=None):
                    if new_params is None:
                        new_params = {}
                    if remove is None:
                        remove = []
                    p = dict(request.GET.items())
                    for r in remove:
                        for k in list(p):
                            if k == r or k.startswith(r + '_'):
                                del p[k]
                    for k, v in new_params.items():
                        if v is None:
                            if k in p:
                                del p[k]
                        else:
                            p[k] = v
                    return '?' + '&'.join([f'{k}={v}' for k, v in p.items()])
                cl.get_query_string = get_query_string
                
                context = {
                    **self.admin_site.each_context(request),
                    'module_name': self.model._meta.model_name,
                    'title': self.model._meta.verbose_name_plural,
                    'cl': cl,
                    'media': self.media,
                    'has_add_permission': self.has_add_permission(request),
                    'has_change_permission': self.has_change_permission(request),
                    'has_delete_permission': self.has_delete_permission(request),
                    'has_view_permission': self.has_view_permission(request),
                    'has_editable_inline_admin_formsets': False,
                    'opts': self.model._meta,
                    'app_label': self.model._meta.app_label,
                    'action_form': None,
                    'actions_on_top': self.actions_on_top,
                    'actions_on_bottom': self.actions_on_bottom,
                    'actions_selection_counter': self.actions_selection_counter,
                    'preserved_filters': self.get_preserved_filters(request),
                    **(extra_context or {}),
                }
                
                request.current_app = self.admin_site.name
                
                return TemplateResponse(request, self.change_list_template or [
                    f'admin/{self.model._meta.app_label}/{self.model._meta.model_name}/change_list.html',
                    f'admin/{self.model._meta.app_label}/change_list.html',
                    'admin/change_list.html'
                ], context)
            except Exception as e:
                logger.error(f"Ошибка при создании ChangeList: {str(e)}")
                messages.error(request, f'Ошибка при создании списка: {str(e)}')
                return super().changelist_view(request, extra_context)
            
        except Exception as e:
            logger.error(f"Ошибка при получении данных: {str(e)}")
            messages.error(request, f'Ошибка при получении данных: {str(e)}')
            return super().changelist_view(request, extra_context)


@admin.register(User)
class UserAdmin(APIModelAdmin):
    list_display = ('id', 'full_name', 'email', 'phone', 'role', 'created_at')
    search_fields = ('first_name', 'last_name', 'email', 'phone')
    readonly_fields = ('id', 'created_at', 'updated_at', 'user_photo')
    ordering = ('-created_at',)

    def get_api_objects(self, request):
        objects = api_service.get_users()
        if 'role' in request.GET:
            filtered_objects = [obj for obj in objects if obj.role == request.GET['role']]
        else:
            filtered_objects = objects
            
        # Используем метод create_mock_queryset вместо создания собственного MockQuerySet
        return self.create_mock_queryset(filtered_objects)

    def full_name(self, obj):
        return f"{obj.first_name} {obj.last_name}"
    full_name.short_description = 'Полное имя'

    def user_photo(self, obj):
        if obj.photo_url:
            return mark_safe(f'<img src="{obj.photo_url}" style="max-width: 300px;">')
        return "Нет фото"
    user_photo.short_description = 'Фото пользователя'

    def has_add_permission(self, request):
        return False

    def has_delete_permission(self, request, obj=None):
        return False


@admin.register(DriverDocument)
class DriverDocumentAdmin(APIModelAdmin):
    list_display = ('id', 'driver_name', 'car_info', 'status', 'created_at', 'verification_actions')
    search_fields = ('user__first_name', 'user__last_name', 'car_brand', 'car_model', 'car_number')
    readonly_fields = ('id', 'created_at', 'updated_at', 'document_images', 'car_info', 'driver_info')
    fields = (
        'driver_info', 'car_info', 'document_images',
        ('status', 'rejection_reason')
    )
    ordering = ('-created_at',)

    def get_api_objects(self, request):
        objects = api_service.get_driver_documents()
        if 'status' in request.GET:
            filtered_objects = [obj for obj in objects if obj.status == request.GET['status']]
        else:
            filtered_objects = objects
            
        # Используем метод create_mock_queryset вместо создания собственного MockQuerySet
        return self.create_mock_queryset(filtered_objects)

    def post(self, request, object_id=None):
        """
        Обрабатывает POST запросы для обновления статуса документов
        """
        import logging
        logger = logging.getLogger(__name__)
        
        logger.debug(f"Получен POST запрос: {request.POST}")
        
        if 'action' in request.POST and request.POST['action'] == 'update_status':
            try:
                document_id = request.POST.get('document_id')
                status = request.POST.get('status')
                rejection_reason = request.POST.get('rejection_reason', '')
                
                logger.debug(f"Обработка запроса на обновление статуса документа ID {document_id}: статус={status}, причина={rejection_reason}")
                
                if not document_id or not status:
                    error_msg = "Не указан ID документа или статус"
                    logger.error(error_msg)
                    self.message_user(request, error_msg, level='error')
                    return None
                
                # Отправляем запрос к API для обновления статуса
                data = {
                    'status': status,
                    'rejectionReason': rejection_reason
                }
                logger.debug(f"Отправка запроса к API: {data}")
                
                try:
                    response = api_service.update_driver_document(document_id, data)
                    logger.debug(f"Получен ответ от API: {response}")
                    
                    if response:
                        success_msg = f"Статус документа успешно обновлен на '{status}'"
                        logger.info(success_msg)
                        self.message_user(request, success_msg)
                    else:
                        error_msg = "Ошибка при обновлении статуса документа: пустой ответ от API"
                        logger.error(error_msg)
                        self.message_user(request, error_msg, level='error')
                except Exception as api_error:
                    error_msg = f"Ошибка при обращении к API: {str(api_error)}"
                    logger.exception(error_msg)
                    self.message_user(request, error_msg, level='error')
            except Exception as e:
                error_msg = f"Произошла ошибка: {str(e)}"
                logger.exception(f"Исключение при обновлении статуса документа: {str(e)}")
                self.message_user(request, error_msg, level='error')
        else:
            logger.debug(f"Получен POST запрос без действия 'update_status': {request.POST}")
        
        # Перенаправляем на страницу списка
        from django.http import HttpResponseRedirect
        return HttpResponseRedirect(request.path)

    def driver_name(self, obj):
        return f"{obj.user.first_name} {obj.user.last_name}"
    driver_name.short_description = 'Водитель'

    def car_info(self, obj):
        return format_html(
            '<div style="margin-bottom: 10px;">'
            '<strong>Марка:</strong> {}<br>'
            '<strong>Модель:</strong> {}<br>'
            '<strong>Год:</strong> {}<br>'
            '<strong>Цвет:</strong> {}<br>'
            '<strong>Номер:</strong> {}'
            '</div>',
            obj.car_brand, obj.car_model, obj.car_year,
            obj.car_color, obj.car_number
        )
    car_info.short_description = 'Информация об автомобиле'

    def driver_info(self, obj):
        return format_html(
            '<div style="margin-bottom: 10px;">'
            '<strong>Имя:</strong> {} {}<br>'
            '<strong>Email:</strong> {}<br>'
            '<strong>Телефон:</strong> {}'
            '</div>',
            obj.user.first_name, obj.user.last_name,
            obj.user.email, obj.user.phone
        )
    driver_info.short_description = 'Информация о водителе'

    def document_images(self, obj):
        html = '<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin-bottom: 20px;">'
        
        # Водительское удостоверение
        html += '<div class="document-section">'
        html += '<h3>Водительское удостоверение</h3>'
        html += '<div style="display: flex; gap: 10px;">'
        if obj.driver_license_front:
            html += f'<div><p>Лицевая сторона</p><img src="{obj.driver_license_front}" style="max-width: 300px;"></div>'
        if obj.driver_license_back:
            html += f'<div><p>Обратная сторона</p><img src="{obj.driver_license_back}" style="max-width: 300px;"></div>'
        html += '</div></div>'
        
        # Регистрация автомобиля
        html += '<div class="document-section">'
        html += '<h3>Регистрация автомобиля</h3>'
        html += '<div style="display: flex; gap: 10px;">'
        if obj.car_registration_front:
            html += f'<div><p>Лицевая сторона</p><img src="{obj.car_registration_front}" style="max-width: 300px;"></div>'
        if obj.car_registration_back:
            html += f'<div><p>Обратная сторона</p><img src="{obj.car_registration_back}" style="max-width: 300px;"></div>'
        html += '</div></div>'
        
        # Фотографии автомобиля
        html += '<div class="document-section" style="grid-column: span 2;">'
        html += '<h3>Фотографии автомобиля</h3>'
        html += '<div style="display: flex; gap: 10px;">'
        if obj.car_photo_front:
            html += f'<div><p>Спереди</p><img src="{obj.car_photo_front}" style="max-width: 300px;"></div>'
        if obj.car_photo_side:
            html += f'<div><p>Сбоку</p><img src="{obj.car_photo_side}" style="max-width: 300px;"></div>'
        if obj.car_photo_interior:
            html += f'<div><p>Салон</p><img src="{obj.car_photo_interior}" style="max-width: 300px;"></div>'
        html += '</div></div>'
        
        html += '</div>'
        return format_html(html)
    document_images.short_description = 'Документы и фотографии'

    def verification_actions(self, obj):
        if obj.status == 'pending':
            return format_html(
                '<a class="button" href="#" onclick="approveDocument({})">Одобрить</a> '
                '<a class="button" href="#" onclick="rejectDocument({})">Отклонить</a>',
                obj.id, obj.id
            )
        return obj.get_status_display()
    verification_actions.short_description = 'Действия'

    def save_model(self, request, obj, form, change):
        if change:
            api_service.update_driver_document(obj.id, {
                'status': obj.status,
                'rejection_reason': obj.rejection_reason or ''
            })

    def has_add_permission(self, request):
        return False

    def has_delete_permission(self, request, obj=None):
        return False

    class Media:
        js = ('js/driver_documents.js',)


@admin.register(Ride)
class RideAdmin(APIModelAdmin):
    list_display = ('id', 'driver_name', 'route', 'departure_date', 'status', 'price', 'seats_info')
    search_fields = ('driver__first_name', 'driver__last_name', 'from_address', 'to_address')
    readonly_fields = ('id', 'created_at', 'updated_at', 'driver_info', 'ride_details')
    ordering = ('-created_at',)

    def get_api_objects(self, request):
        objects = api_service.get_rides()
        if 'status' in request.GET:
            filtered_objects = [obj for obj in objects if obj.status == request.GET['status']]
        else:
            filtered_objects = objects
            
        # Используем метод create_mock_queryset вместо создания собственного MockQuerySet
        return self.create_mock_queryset(filtered_objects)

    def driver_name(self, obj):
        return f"{obj.driver.first_name} {obj.driver.last_name}"
    driver_name.short_description = 'Водитель'

    def route(self, obj):
        return f"{obj.from_address} → {obj.to_address}"
    route.short_description = 'Маршрут'

    def seats_info(self, obj):
        return f"{obj.booked_seats}/{obj.seats_count}"
    seats_info.short_description = 'Места (занято/всего)'

    def driver_info(self, obj):
        return format_html(
            '<div style="margin-bottom: 10px;">'
            '<strong>Водитель:</strong> {} {}<br>'
            '<strong>Email:</strong> {}<br>'
            '<strong>Телефон:</strong> {}'
            '</div>',
            obj.driver.first_name, obj.driver.last_name,
            obj.driver.email, obj.driver.phone
        )
    driver_info.short_description = 'Информация о водителе'

    def ride_details(self, obj):
        return format_html(
            '<div style="margin-bottom: 10px;">'
            '<strong>Откуда:</strong> {}<br>'
            '<strong>Куда:</strong> {}<br>'
            '<strong>Дата отправления:</strong> {}<br>'
            '<strong>Цена:</strong> {} тг<br>'
            '<strong>Места:</strong> {} из {}<br>'
            '<strong>Комментарий:</strong> {}'
            '</div>',
            obj.from_address, obj.to_address,
            obj.departure_date.strftime('%d.%m.%Y %H:%M'),
            obj.price, obj.booked_seats, obj.seats_count,
            obj.comment or '-'
        )
    ride_details.short_description = 'Детали поездки'

    def has_add_permission(self, request):
        return False

    def has_delete_permission(self, request, obj=None):
        return False


@admin.register(Booking)
class BookingAdmin(APIModelAdmin):
    list_display = ('id', 'passenger_name', 'ride_info', 'booking_type', 'status', 'price', 'created_at')
    search_fields = ('passenger__first_name', 'passenger__last_name', 'ride__from_address', 'ride__to_address')
    readonly_fields = ('id', 'created_at', 'updated_at', 'passenger_info', 'booking_details')
    ordering = ('-created_at',)

    def get_api_objects(self, request):
        objects = api_service.get_bookings()
        if 'status' in request.GET:
            filtered_objects = [obj for obj in objects if obj.status == request.GET['status']]
        elif 'booking_type' in request.GET:
            filtered_objects = [obj for obj in objects if obj.booking_type == request.GET['booking_type']]
        else:
            filtered_objects = objects
            
        # Используем метод create_mock_queryset вместо создания собственного MockQuerySet
        return self.create_mock_queryset(filtered_objects)

    def passenger_name(self, obj):
        return f"{obj.passenger.first_name} {obj.passenger.last_name}"
    passenger_name.short_description = 'Пассажир'

    def ride_info(self, obj):
        return f"{obj.ride.from_address} → {obj.ride.to_address} ({obj.ride.departure_date.strftime('%d.%m.%Y %H:%M')})"
    ride_info.short_description = 'Информация о поездке'

    def passenger_info(self, obj):
        return format_html(
            '<div style="margin-bottom: 10px;">'
            '<strong>Пассажир:</strong> {} {}<br>'
            '<strong>Email:</strong> {}<br>'
            '<strong>Телефон:</strong> {}'
            '</div>',
            obj.passenger.first_name, obj.passenger.last_name,
            obj.passenger.email, obj.passenger.phone
        )
    passenger_info.short_description = 'Информация о пассажире'

    def booking_details(self, obj):
        return format_html(
            '<div style="margin-bottom: 10px;">'
            '<strong>Тип бронирования:</strong> {}<br>'
            '<strong>Место посадки:</strong> {}<br>'
            '<strong>Место высадки:</strong> {}<br>'
            '<strong>Количество мест:</strong> {}<br>'
            '<strong>Цена:</strong> {} тг<br>'
            '<strong>Комментарий:</strong> {}'
            '</div>',
            obj.get_booking_type_display(),
            obj.pickup_address, obj.dropoff_address,
            obj.seats_count, obj.price,
            obj.comment or '-'
        )
    booking_details.short_description = 'Детали бронирования'

    def has_add_permission(self, request):
        return False

    def has_delete_permission(self, request, obj=None):
        return False 