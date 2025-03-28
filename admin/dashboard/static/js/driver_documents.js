function getCookie(name) {
    let cookieValue = null;
    if (document.cookie && document.cookie !== '') {
        const cookies = document.cookie.split(';');
        for (let i = 0; i < cookies.length; i++) {
            const cookie = cookies[i].trim();
            if (cookie.substring(0, name.length + 1) === (name + '=')) {
                cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
                break;
            }
        }
    }
    return cookieValue;
}

function approveDocument(documentId) {
    if (!confirm('Вы уверены, что хотите одобрить эти документы?')) {
        return;
    }

    console.log('Отправка запроса на одобрение документа ID:', documentId);
    
    const csrftoken = getCookie('csrftoken');
    
    // Создаем форму для отправки
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = window.location.pathname;
    form.style.display = 'none';

    // Добавляем CSRF токен
    const csrfInput = document.createElement('input');
    csrfInput.type = 'hidden';
    csrfInput.name = 'csrfmiddlewaretoken';
    csrfInput.value = csrftoken;
    form.appendChild(csrfInput);

    // Добавляем ID документа
    const idInput = document.createElement('input');
    idInput.type = 'hidden';
    idInput.name = 'document_id';
    idInput.value = documentId;
    form.appendChild(idInput);

    // Добавляем статус
    const statusInput = document.createElement('input');
    statusInput.type = 'hidden';
    statusInput.name = 'status';
    statusInput.value = 'approved';
    form.appendChild(statusInput);

    // Добавляем действие
    const actionInput = document.createElement('input');
    actionInput.type = 'hidden';
    actionInput.name = 'action';
    actionInput.value = 'update_status';
    form.appendChild(actionInput);

    document.body.appendChild(form);
    console.log('Форма создана и отправляется:', form);
    form.submit();
}

function rejectDocument(documentId) {
    const reason = prompt('Пожалуйста, укажите причину отклонения:');
    if (reason === null) {
        return;
    }

    console.log('Отправка запроса на отклонение документа ID:', documentId, 'Причина:', reason);
    
    const csrftoken = getCookie('csrftoken');
    
    // Создаем форму для отправки
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = window.location.pathname;
    form.style.display = 'none';

    // Добавляем CSRF токен
    const csrfInput = document.createElement('input');
    csrfInput.type = 'hidden';
    csrfInput.name = 'csrfmiddlewaretoken';
    csrfInput.value = csrftoken;
    form.appendChild(csrfInput);

    // Добавляем ID документа
    const idInput = document.createElement('input');
    idInput.type = 'hidden';
    idInput.name = 'document_id';
    idInput.value = documentId;
    form.appendChild(idInput);

    // Добавляем статус
    const statusInput = document.createElement('input');
    statusInput.type = 'hidden';
    statusInput.name = 'status';
    statusInput.value = 'rejected';
    form.appendChild(statusInput);

    // Добавляем причину отклонения
    const reasonInput = document.createElement('input');
    reasonInput.type = 'hidden';
    reasonInput.name = 'rejection_reason';
    reasonInput.value = reason;
    form.appendChild(reasonInput);

    // Добавляем действие
    const actionInput = document.createElement('input');
    actionInput.type = 'hidden';
    actionInput.name = 'action';
    actionInput.value = 'update_status';
    form.appendChild(actionInput);

    document.body.appendChild(form);
    console.log('Форма создана и отправляется:', form);
    form.submit();
} 