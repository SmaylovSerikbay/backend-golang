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

    const csrftoken = getCookie('csrftoken');
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = window.location.pathname;

    const csrfInput = document.createElement('input');
    csrfInput.type = 'hidden';
    csrfInput.name = 'csrfmiddlewaretoken';
    csrfInput.value = csrftoken;
    form.appendChild(csrfInput);

    const statusInput = document.createElement('input');
    statusInput.type = 'hidden';
    statusInput.name = 'status';
    statusInput.value = 'approved';
    form.appendChild(statusInput);

    document.body.appendChild(form);
    form.submit();
}

function rejectDocument(documentId) {
    const reason = prompt('Пожалуйста, укажите причину отклонения:');
    if (reason === null) {
        return;
    }

    const csrftoken = getCookie('csrftoken');
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = window.location.pathname;

    const csrfInput = document.createElement('input');
    csrfInput.type = 'hidden';
    csrfInput.name = 'csrfmiddlewaretoken';
    csrfInput.value = csrftoken;
    form.appendChild(csrfInput);

    const statusInput = document.createElement('input');
    statusInput.type = 'hidden';
    statusInput.name = 'status';
    statusInput.value = 'rejected';
    form.appendChild(statusInput);

    const reasonInput = document.createElement('input');
    reasonInput.type = 'hidden';
    reasonInput.name = 'rejection_reason';
    reasonInput.value = reason;
    form.appendChild(reasonInput);

    document.body.appendChild(form);
    form.submit();
} 