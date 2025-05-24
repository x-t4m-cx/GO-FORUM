document.addEventListener('DOMContentLoaded', function() {
    // Элементы DOM
    const topicsContainer = document.getElementById('topicsContainer');
    const topicDetails = document.getElementById('topicDetails');
    const topicsList = document.getElementById('topicsList');
    const backToTopics = document.getElementById('backToTopics');
    const createTopicForm = document.getElementById('createTopicForm');
    const topicForm = document.getElementById('topicForm');
    const createCommentForm = document.getElementById('createCommentForm');
    const commentForm = document.getElementById('commentForm');
    const loginBtn = document.getElementById('loginBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    const usernameDisplay = document.getElementById('usernameDisplay');
    const authModal = new bootstrap.Modal(document.getElementById('authModal'));
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const updateTopicBtn = document.getElementById("updateTopicBtn");
    const deleteTopicBtn = document.getElementById('deleteTopicBtn');
    const authTabs = document.getElementById('authTabs');
    const newTopicBtn = document.getElementById('newTopicBtn');
    const sendMsgBtn = document.getElementById("send-btn");
    const gopher = document.getElementById("number1");

    // Текущий пользователь и токен
    let currentUser = null;
    let authToken = null;
    let chatSocket = null;
    // Инициализация
    checkAuth();
    loadTopics()
    updateTopicBtn.addEventListener('click',() =>{
        const topicId = document.getElementById("commentTopicId").value;
        showUpdateTopicForm(topicId)
    })
    function showUpdateTopicForm(topicId) {
        // Создаем модальное окно для редактирования
        const modalHtml = `
    <div class="modal fade" id="updateTopicModal" tabindex="-1" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Редактирование темы</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <form id="updateTopicForm">
                        <div class="mb-3">
                            <label for="updateTopicTitle" class="form-label">Заголовок</label>
                            <input type="text" class="form-control" id="updateTopicTitle" required>
                        </div>
                        <div class="mb-3">
                            <label for="updateTopicContent" class="form-label">Содержание</label>
                            <textarea class="form-control" id="updateTopicContent" rows="5" required></textarea>
                        </div>
                        <input type="hidden" id="updateTopicId" value="${topicId}">
                        <button type="submit" class="btn btn-primary">Сохранить</button>
                    </form>
                </div>
            </div>
        </div>
    </div>
    `;

        // Добавляем модальное окно в DOM
        document.body.insertAdjacentHTML('beforeend', modalHtml);

        // Получаем элемент модального окна
        const modalElement = document.getElementById('updateTopicModal');

        // Инициализируем модальное окно
        const updateModal = new bootstrap.Modal(modalElement);

        // Показываем модальное окно
        updateModal.show();

        // Загружаем текущие данные темы
        fetch(`/topics/${topicId}`)
            .then(response => response.json())
            .then(data => {
                if (data.data) {
                    document.getElementById('updateTopicTitle').value = data.data.title;
                    document.getElementById('updateTopicContent').value = data.data.content;
                }
            })
            .catch(error => {
                console.error('Ошибка при загрузке темы:', error);
                alert('Не удалось загрузить данные темы');
            });

        // Обработчик отправки формы
        document.getElementById('updateTopicForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            await updateTopic();
            updateModal.hide();
        });

        // Удаляем модальное окно при закрытии
        modalElement.addEventListener('hidden.bs.modal', () => {
            modalElement.remove();
        });
    }
    async function updateTopic() {
        const topicId = document.getElementById('updateTopicId').value;
        const title = document.getElementById('updateTopicTitle').value;
        const content = document.getElementById('updateTopicContent').value;

        try {
            const response = await makeRequest(`/topics/${topicId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + authToken
                },
                body: JSON.stringify({
                    title: title,
                    content: content
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка при обновлении темы');
            }

            // Закрываем модальное окно
            bootstrap.Modal.getInstance(document.getElementById('updateTopicModal')).hide();

            // Обновляем отображаемую тему
            showTopicDetails(topicId);
            loadTopics();
        } catch (error) {
            console.error('Ошибка:', error);
            alert('Не удалось обновить тему: ' + error.message);
        }
    }
    let flag = true
    newTopicBtn.addEventListener('click', (e) => {
        (flag) ?  createTopicForm.classList.remove("hidden") :  createTopicForm.classList.add("hidden")
        flag = !flag
        e.preventDefault();
    });

    backToTopics.addEventListener('click', () => {
        topicDetails.classList.add('hidden');
        topicsList.classList.remove('hidden');
        gopher.classList.remove("hidden");
    });

    loginBtn.addEventListener('click', (e) => {
        e.preventDefault();
        authModal.show();
    });

    logoutBtn.addEventListener('click', (e) => {
        e.preventDefault();
        logout();
    });

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;
        await login(username, password);
    });

    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('registerUsername').value;
        const password = document.getElementById('registerPassword').value;
        await register(username, password);
    });

    topicForm.addEventListener('submit', (e) => {
        e.preventDefault();
        createTopic();
        createTopicForm.classList.add("hidden")
    });

    commentForm.addEventListener('submit', (e) => {
        e.preventDefault();
        createComment();
    });

    deleteTopicBtn.addEventListener('click', () => {
        const topicId = document.getElementById('commentTopicId').value;
        deleteTopic(topicId);
    });

    // Переключение между вкладками входа и регистрации
    authTabs.addEventListener('click', (e) => {
        if (e.target.classList.contains('nav-link')) {
            const tabLinks = authTabs.querySelectorAll('.nav-link');
            tabLinks.forEach(link => link.classList.remove('active'));
            e.target.classList.add('active');
        }
    });

    // Функции
    function checkAuth() {
        const token = localStorage.getItem('forumToken');
        const username = localStorage.getItem('forumUsername');

        if (token && username) {
            currentUser = username;
            authToken = token;
            updateAuthUI(true);
            sendMsgBtn.classList.remove
            connectChatWebSocket(); // Подключаем чат при авторизации
        } else {
            updateAuthUI(false);
        }
    }


    function updateAuthUI(isAuthenticated) {
        if (isAuthenticated) {
            loginBtn.classList.add('hidden');
            logoutBtn.classList.remove('hidden');
            usernameDisplay.classList.remove('hidden');
            newTopicBtn.classList.remove('hidden'); // Показываем кнопку новой темы
            usernameDisplay.textContent = currentUser;
            createCommentForm.classList.remove('hidden');
            document.getElementById('message-input').disabled = false;
            document.getElementById("message-input").placeholder = "Введите сообщение..."
            document.getElementById('send-btn').disabled = false;
        } else {
            loginBtn.classList.remove('hidden');
            logoutBtn.classList.add('hidden');
            usernameDisplay.classList.add('hidden');
            newTopicBtn.classList.add('hidden');
            createTopicForm.classList.add('hidden');
            createCommentForm.classList.add('hidden');
            document.getElementById('message-input').disabled = true;
            document.getElementById("message-input").placeholder = "Необходимо войти"
            document.getElementById('send-btn').disabled = true;
        }
    }

    async function login(username, password) {
        try {
            const response = await fetch('/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка входа');
            }

            // Получаем токен из заголовка Authorization
            const authHeader = response.headers.get('Authorization');
            if (authHeader && authHeader.startsWith('Bearer ')) {
                authToken = authHeader.substring(7);
            }

            // Сохраняем данные аутентификации
            currentUser = username;
            localStorage.setItem('forumToken', authToken);
            localStorage.setItem('forumUsername', username);

            updateAuthUI(true);
            authModal.hide();
            loginForm.reset();
            loadTopics();
        } catch (error) {
            console.error('Ошибка входа:', error);
            alert('Ошибка входа: ' + error.message);
        }
    }

    async function register(username, password) {
        try {
            const response = await makeRequest('/auth/register',{
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка регистрации');
            }

            // После успешной регистрации автоматически входим
            await login(username, password);
        } catch (error) {
            console.error('Ошибка регистрации:', error);
            alert('Ошибка регистрации: ' + error.message);
        }
    }

    async function logout() {
        try {
            const response = await makeRequest('/auth/logout', {
                method: 'POST',
                headers: {
                    'Authorization': 'Bearer ' + authToken
                }
            });

            if (!response.ok) {
                throw new Error('Ошибка выхода');
            }
            if (chatSocket) {
                chatSocket.close();
            }

            // Очищаем данные аутентификации
            currentUser = null;
            authToken = null;
            localStorage.removeItem('forumToken');
            localStorage.removeItem('forumUsername');

            updateAuthUI(false);
            loadTopics();
        } catch (error) {
            console.error('Ошибка выхода:', error);
            alert('Ошибка выхода: ' + error.message);
        }
    }

    async function loadTopics() {
        try {

            const response = await makeRequest('/topics/');
            const data = await response.json();

            topicsContainer.innerHTML = '';

            if (data.data && data.data.length > 0) {
                data.data.forEach(topic => {
                    let updated = (checkUpdated(new Date(topic.updated_at))) ?
                        "" :
                        "Изменено: " + new Date(topic.updated_at).toLocaleString()
                    const topicElement = document.createElement('div');
                    topicElement.className = 'card topic-card';
                    topicElement.innerHTML = `
                        <div class="card-body">
                            <h5 class="card-title">${topic.title}</h5>
                            <p class="card-text">${topic.content.substring(0, 100)}${topic.content.length > 100 ? '...' : ''}</p>
                            <div class="d-flex justify-content-between align-items-center">
                                <small class="text-muted">Автор: ${topic.username}</small>
                                <small class="text-muted">
                                      Создано: ${new Date(topic.created_at).toLocaleString()}<br>
                                    ${updated}
                                </small>
                             
                                
                            </div>
                        </div>
                    `;

                    topicElement.addEventListener('click', () => showTopicDetails(topic.id));
                    topicsContainer.appendChild(topicElement);
                });
            } else {
                topicsContainer.innerHTML = '<p>Пока нет ни одной темы. Будьте первым!</p>';
            }
        } catch (error) {
            console.error('Ошибка при загрузке тем:', error);
            topicsContainer.innerHTML = '<p class="text-danger">Ошибка при загрузке тем</p>';
        }
    }


    async function showTopicDetails(topicId) {
        try {
            // Загрузка темы
            const topicResponse = await makeRequest(`/topics/${topicId}`);
            const topicData = await topicResponse.json();
            console.log(topicResponse)
            if (!topicData) {
                throw new Error('Тема не найдена');
            }

            const topic = topicData;

            // Заполнение данных темы
            document.getElementById('topicTitleDetail').textContent = topic.title;
            document.getElementById('topicContentDetail').textContent = topic.content;
            document.getElementById('topicAuthor').textContent = `Автор: ${topic.username}`;
            document.getElementById('topicCrDate').textContent = new Date(topic.created_at).toLocaleString();

            // Проверка updated_at на "01.01.1, 03:04:36"
            const updatedAt = new Date(topic.updated_at);


            if (!checkUpdated(updatedAt)) {
                document.getElementById('topicUpDate').textContent = "Изменено: " + updatedAt.toLocaleString();
            } else {
                document.getElementById('topicUpDate').textContent = ''; // или можно скрыть весь элемент
            }

            document.getElementById('commentTopicId').value = topic.id;

            // Показываем кнопку удаления только если пользователь - автор темы
            if (currentUser === topic.username ||
                currentUser === "admin") {
                deleteTopicBtn.classList.remove('hidden');
                updateTopicBtn.classList.remove("hidden");
            } else {
                deleteTopicBtn.classList.add('hidden');
                updateTopicBtn.classList.add("hidden");
            }

            // Загрузка комментариев
            const commentsResponse = await makeRequest(`/topics/comments/${topicId}`);
            const commentsData = await commentsResponse.json();

            const commentsContainer = document.getElementById('commentsContainer');
            commentsContainer.innerHTML = '';

            if (commentsData.data && commentsData.data.length > 0) {
                commentsData.data.forEach(comment => {
                    const commentElement = document.createElement('div');
                    commentElement.className = 'comment';
                    // В функции showTopicDetails, где создаются комментарии, изменим HTML:
                    commentElement.innerHTML = `
                        <div class="d-flex justify-content-between">
                            <strong>${comment.username}</strong>
                            <small class="text-muted">${new Date(comment.created_at).toLocaleString()}</small>
                        </div>
                        <p id="comment-content-${comment.id}">${comment.content}</p>
                        ${currentUser === comment.username || currentUser === "admin" ?
                                            `<div class="comment-actions">
                                <button class="btn btn-sm btn-primary edit-comment" data-id="${comment.id}">Редактировать</button>
                                <button class="btn btn-sm btn-danger delete-comment" data-id="${comment.id}">Удалить</button>
                            </div>` : ''}
                    `;
                    commentsContainer.appendChild(commentElement);
                });

                // Добавляем обработчики для кнопок удаления комментариев
                document.querySelectorAll('.delete-comment').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        e.stopPropagation();
                        deleteComment(btn.dataset.id);
                    });
                });
                document.querySelectorAll('.edit-comment').forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        e.stopPropagation();
                        editComment(btn.dataset.id);
                    });
                });
            } else {
                commentsContainer.innerHTML = '<p>Пока нет комментариев. Будьте первым!</p>';
            }

            // Переключаем видимость
            topicsList.classList.add('hidden');
            gopher.classList.add("hidden");
            topicDetails.classList.remove('hidden');
        } catch (error) {
            console.error('Ошибка при загрузке темы:', error);
            alert('Не удалось загрузить тему');
        }
    }
    function checkUpdated(updatedAt){
        return updatedAt.getFullYear() === 1 &&
            updatedAt.getMonth() === 0 &&
            updatedAt.getDate() === 1 &&
            updatedAt.getHours() === 3 &&
            updatedAt.getMinutes() === 4 &&
            updatedAt.getSeconds() === 36;
    }
    async function editComment(commentId) {
        try {
            // Получаем текущий комментарий
            const response = await makeRequest(`/comments/${commentId}`);
            const commentData = await response.json();

            if (!commentData) {
                throw new Error('Комментарий не найден');
            }

            const comment = commentData;

            // Создаем модальное окно для редактирования
            const modalHtml = `
            <div class="modal fade" id="editCommentModal" tabindex="-1" aria-hidden="true">
                <div class="modal-dialog">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">Редактирование комментария</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                        </div>
                        <div class="modal-body">
                            <form id="editCommentForm">
                                <div class="mb-3">
                                    <label for="editCommentContent" class="form-label">Комментарий</label>
                                    <textarea class="form-control" id="editCommentContent" rows="3" required>${comment.content}</textarea>
                                </div>
                                <input type="hidden" id="editCommentId" value="${commentId}">
                                <button type="submit" class="btn btn-primary">Сохранить</button>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        `;

            // Добавляем модальное окно в DOM
            document.body.insertAdjacentHTML('beforeend', modalHtml);

            // Инициализируем модальное окно
            const editModal = new bootstrap.Modal(document.getElementById('editCommentModal'));
            editModal.show();

            // Обработчик отправки формы
            document.getElementById('editCommentForm').addEventListener('submit', async (e) => {
                e.preventDefault();
                await updateComment(commentId);
                editModal.hide();
            });

            // Удаляем модальное окно при закрытии
            document.getElementById('editCommentModal').addEventListener('hidden.bs.modal', () => {
                document.getElementById('editCommentModal').remove();
            });

        } catch (error) {
            console.error('Ошибка при редактировании комментария:', error);
            alert('Не удалось загрузить комментарий для редактирования');
        }
    }
    async function updateComment(commentId) {
        try {
            const content = document.getElementById('editCommentContent').value;

            if (!content) {
                throw new Error('Содержание комментария не может быть пустым');
            }

            console.log('Updating comment:', {
                commentId: commentId,
                content: content
            });

            const response = await makeRequest(`/comments/${commentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + authToken
                },
                body: JSON.stringify({
                    content: content
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Ошибка при обновлении комментария');
            }

            // Обновляем комментарий на странице без перезагрузки
            document.getElementById(`comment-content-${commentId}`).textContent = content;

        } catch (error) {
            console.error('Ошибка при обновлении комментария:', error);
            alert('Не удалось обновить комментарий: ' + error.message);
        }
    }
    async function createTopic() {
        const title = document.getElementById('topicTitle').value;
        const content = document.getElementById('topicContent').value;

        try {
            const response = await makeRequest('/topics/', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + authToken
                },
                body: JSON.stringify({
                    title: title,
                    content: content
                })
            });

            const data = await response.json();

            if (response.ok) {
                // Очищаем форму
                topicForm.reset();
                // Перезагружаем список тем
                loadTopics();
            } else {
                throw new Error(data.error || 'Ошибка при создании темы');
            }
        } catch (error) {
            console.error('Ошибка:', error);
            alert('Не удалось создать тему: ' + error.message);
        }
    }

    async function createComment() {
        const topicId = document.getElementById('commentTopicId').value;
        const content = document.getElementById('commentContent').value;

        try {
            const response = await makeRequest('/comments/', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer ' + authToken
                },
                body: JSON.stringify({
                    topic_id: topicId,
                    content: content
                })
            });

            const data = await response.json();

            if (response.ok) {
                // Очищаем форму
                commentForm.reset();
                // Перезагружаем комментарии
                showTopicDetails(topicId);
            } else {
                throw new Error(data.error || 'Ошибка при создании комментария');
            }
        } catch (error) {
            console.error('Ошибка:', error);
            alert('Не удалось создать комментарий: ' + error.message);
        }
    }

    async function deleteTopic(topicId) {
        if (!confirm('Вы уверены, что хотите удалить эту тему?')) return;

        try {
            const response = await makeRequest(`/topics/${topicId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': 'Bearer ' + authToken
                }
            });

            if (response.ok) {
                // Возвращаемся к списку тем
                topicDetails.classList.add('hidden');
                topicsList.classList.remove('hidden');
                // Обновляем список тем
                loadTopics();
            } else {
                const data = await response.json();
                throw new Error(data.error || 'Ошибка при удалении темы');
            }
        } catch (error) {
            console.error('Ошибка:', error);
            alert('Не удалось удалить тему: ' + error.message);
        }
    }

    async function deleteComment(commentId) {
        if (!confirm('Вы уверены, что хотите удалить этот комментарий?')) return;

        try {
            const response = await makeRequest(`/comments/${commentId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': 'Bearer ' + authToken
                }
            });

            if (response.ok) {
                // Перезагружаем комментарии
                const topicId = document.getElementById('commentTopicId').value;
                showTopicDetails(topicId);
            } else {
                const data = await response.json();
                throw new Error(data.error || 'Ошибка при удалении комментария');
            }
        } catch (error) {
            console.error('Ошибка:', error);
            alert('Не удалось удалить комментарий: ' + error.message);
        }
    }
    async function makeRequest(url, options) {
        try {
            const response = await fetch("http://localhost:8080"+url, options);

            // Проверяем, есть ли новый токен в заголовках
            const newToken = response.headers.get('Authorization');
            if (newToken && newToken.startsWith('Bearer ')) {
                const token = newToken.substring(7);
                authToken = token;
                localStorage.setItem('forumToken', token);
            }

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Request failed');
            }

            return response;
        } catch (error) {
            console.error('Request error:', error);
            throw error;
        }
    }
    // Чат
    function connectChatWebSocket() {
        if (!currentUser) return;

        chatSocket = new WebSocket(`ws://${window.location.hostname}:8083/api/ws?username=${currentUser}`);

        chatSocket.onopen = () => {
            console.log('Chat WebSocket connected');
            fetchRecentChatMessages();
            setInterval(fetchRecentChatMessages, 60000); // Обновлять сообщения каждую минуту
        };

        chatSocket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            const messageDate = new Date(message.CreatedAt);
            const now = new Date();

            if (now - messageDate <= 60000) {
                addMessageToChat(message);
            }
        };

        chatSocket.onclose = () => {
            console.log('Chat WebSocket disconnected');
            setTimeout(connectChatWebSocket, 1000);
        };

        chatSocket.onerror = (error) => {
            console.error('Chat WebSocket error:', error);
        };
    }

    async function fetchRecentChatMessages() {
        try {
            const response = await fetch(`http://${window.location.hostname}:8083/api/messages`);
            const messages = await response.json();
            const messagesContainer = document.getElementById('messages');
            messagesContainer.innerHTML = '';

            const now = new Date();
            if(!Array.isArray(messages)){
                return
            }

            messages.forEach(message => {
                const messageDate = new Date(message.CreatedAt);
                if (now - messageDate <= 60000) {
                    addMessageToChat(message);
                }
            });
        } catch (error) {
            console.error('Error fetching chat messages:', error);
        }
    }

    function addMessageToChat(message) {
        const messagesContainer = document.getElementById('messages');
        const messageElement = document.createElement('div');
        messageElement.className = 'message';
        const messageDate = new Date(message.CreatedAt);
        messageElement.setAttribute('data-created-at', messageDate.getTime());

        messageElement.innerHTML = `
        <div class="username">${message.Username}</div>
        <div class="text">${message.Message}</div>
        <div class="time">${messageDate.toLocaleTimeString()}</div>
    `;
        messagesContainer.appendChild(messageElement);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    function removeExpiredChatMessages() {
        const messages = document.querySelectorAll('#messages .message');
        const now = new Date();

        messages.forEach(message => {
            const createdAt = parseInt(message.getAttribute('data-created-at'));
            if (now - new Date(createdAt) > 60000) {
                message.remove();
            }
        });
    }

    function sendChatMessage() {
        const messageInput = document.getElementById('message-input');
        const messageText = messageInput.value.trim();

        if (messageText && chatSocket && chatSocket.readyState === WebSocket.OPEN) {
            chatSocket.send(messageText);
            messageInput.value = '';
        }
    }

    sendMsgBtn.addEventListener('click', sendChatMessage);
    document.getElementById('message-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendChatMessage();
        }
    });
    document.getElementById('openChatBtn').addEventListener('click', () => {
        const chatContainer = document.getElementById('chatContainer');
        chatContainer.classList.toggle('hidden');

        if (!chatContainer.classList.contains('hidden') && currentUser) {
            fetchRecentChatMessages();
        }
    });

    document.getElementById('closeChatBtn').addEventListener('click', () => {
        document.getElementById('chatContainer').classList.add('hidden');
    });
    setInterval(removeExpiredChatMessages, 30000);
});