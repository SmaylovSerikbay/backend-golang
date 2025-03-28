from django.conf import settings
from django.contrib import messages
import logging

logger = logging.getLogger(__name__)

class APITokenMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response
        if not settings.ADMIN_API_TOKEN:
            logger.error("ADMIN_API_TOKEN не установлен в настройках")

    def __call__(self, request):
        if not settings.ADMIN_API_TOKEN:
            messages.error(request, 'API токен не настроен. Пожалуйста, проверьте настройки ADMIN_API_TOKEN.')
        else:
            request.api_token = settings.ADMIN_API_TOKEN.strip()
            logger.debug(f"API токен установлен: {request.api_token[:10]}...")
        
        response = self.get_response(request)
        return response 