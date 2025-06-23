# Настройка интеграции с AmoCRM

## Предварительные требования

1. Аккаунт AmoCRM с правами администратора
2. Доступ к разделу "Интеграции" в AmoCRM

## Шаг 1: Создание интеграции в AmoCRM

1. Войдите в ваш аккаунт AmoCRM
2. Перейдите в **Настройки** → **Интеграции**
3. Нажмите **+ Добавить интеграцию**
4. Выберите **Создать интеграцию**

## Шаг 2: Настройка приватной интеграции

1. Заполните основную информацию:
    - **Название**: CRM-Dialer Integration
    - **Описание**: Интеграция с системой автообзвона
    - **Ссылка на сайт**: https://your-domain.com

2. В разделе **Права доступа** выберите:
    - ✅ Сделки (чтение, запись, удаление)
    - ✅ Контакты и компании (чтение, запись, удаление)
    - ✅ Списки (чтение)
    - ✅ Воронки и этапы (чтение)
    - ✅ Пользователи и роли (чтение)
    - ✅ Вебхуки (управление)
    - ✅ Входящие звонки (добавление)

3. В разделе **Redirect URI** добавьте:
   ```
   http://localhost:8080/api/v1/amocrm/auth/callback
   ```

   Для production используйте:
   ```
   https://your-domain.com/api/v1/amocrm/auth/callback
   ```

4. Сохраните интеграцию

## Шаг 3: Получение учетных данных

После создания интеграции вы получите:

- **ID интеграции** (Client ID)
- **Секретный ключ** (Client Secret)
- **Код авторизации** (временный, действует 20 минут)

## Шаг 4: Настройка окружения

1. Скопируйте файл `.env.example` в `.env`:
   ```bash
   cp .env.example .env
   ```

2. Заполните параметры AmoCRM:
   ```env
   AMOCRM_DOMAIN=your-subdomain.amocrm.ru
   AMOCRM_CLIENT_ID=ваш_client_id
   AMOCRM_CLIENT_SECRET=ваш_client_secret
   AMOCRM_REDIRECT_URI=http://localhost:8080/api/v1/amocrm/auth/callback
   ```

## Шаг 5: Первичная авторизация

### Вариант 1: Использование кода авторизации (рекомендуется для первого запуска)

1. Получите код авторизации в настройках интеграции AmoCRM
2. Добавьте его в `.env`:
   ```env
   AMOCRM_AUTH_CODE=def50200...ваш_код_авторизации
   ```
3. Запустите сервисы:
   ```bash
   make up
   ```
4. CRM сервис автоматически обменяет код на токены

### Вариант 2: OAuth авторизация через браузер

1. Запустите сервисы без кода авторизации:
   ```bash
   make up
   ```

2. Получите ссылку для авторизации:
   ```bash
   curl http://localhost:8080/api/v1/amocrm/auth
   ```

3. Откройте полученную ссылку в браузере и авторизуйтесь

4. После успешной авторизации вы будете перенаправлены на callback URL

## Шаг 6: Проверка подключения

1. Проверьте статус подключения:
   ```bash
   curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
        http://localhost:8080/api/v1/amocrm/status
   ```

2. Синхронизируйте поля:
   ```bash
   curl -X POST -H "Authorization: Bearer YOUR_JWT_TOKEN" \
        http://localhost:8080/api/v1/amocrm/fields/sync
   ```

3. Получите список сделок:
   ```bash
   curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
        http://localhost:8080/api/v1/amocrm/leads
   ```

## Шаг 7: Настройка вебхуков

1. В интерфейсе системы перейдите в **Настройки** → **Вебхуки**

2. Система автоматически подпишется на следующие события:
    - Создание сделки
    - Обновление сделки
    - Смена статуса сделки
    - Смена ответственного

3. URL для вебхуков:
   ```
   https://your-domain.com/api/v1/webhooks/amocrm/{event_type}
   ```

## Устранение проблем

### Ошибка "Invalid authorization code"

Код авторизации действителен только 20 минут. Получите новый код в настройках интеграции.

### Ошибка "Invalid client credentials"

Проверьте правильность `AMOCRM_CLIENT_ID` и `AMOCRM_CLIENT_SECRET`.

### Ошибка "Invalid redirect URI"

Убедитесь, что `AMOCRM_REDIRECT_URI` в `.env` точно совпадает с URI в настройках интеграции.

### Токен истек

Система автоматически обновляет токены каждый час. Если автообновление не работает:

1. Проверьте логи CRM сервиса:
   ```bash
   docker-compose logs crm-service
   ```

2. Удалите файл с токенами и повторите авторизацию:
   ```bash
   docker-compose exec crm-service rm /tmp/amocrm_token.json
   ```

## Безопасность

1. **Никогда** не коммитьте файл `.env` в репозиторий
2. Используйте разные учетные данные для dev и production окружений
3. Регулярно ротируйте секретные ключи
4. Ограничьте доступ к callback URL только для AmoCRM IP-адресов в production

## Дополнительные ресурсы

- [Документация AmoCRM API](https://www.amocrm.ru/developers/content/crm_platform/api-reference)
- [Библиотека amocrm для Go](https://github.com/2010kira2010/amocrm)
- [Настройка вебхуков AmoCRM](https://www.amocrm.ru/developers/content/crm_platform/webhooks-api)

## Примеры использования API

### Получение JWT токена для тестирования

```bash
# Логин в систему
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "demo123"
  }'
```

### Работа с полями

```bash
# Получить список полей сделок
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/amocrm/fields?entity_type=leads

# Получить список полей контактов
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/amocrm/fields?entity_type=contacts
```

### Работа со сделками

```bash
# Получить список сделок с пагинацией
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/v1/amocrm/leads?page=1&limit=50"

# Получить сделку по ID
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/amocrm/leads/123456

# Обновить сделку
curl -X PUT -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/v1/amocrm/leads/123456 \
  -d '{
    "name": "Обновленная сделка",
    "price": 100000,
    "status_id": 142
  }'
```

### Работа с контактами

```bash
# Поиск контактов
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8080/api/v1/amocrm/contacts?query=Иван"

# Получить контакт по ID
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/amocrm/contacts/123456
```

## Структура токенов

Токены сохраняются в файле `/tmp/amocrm_token.json` внутри контейнера:

```json
{
  "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOi...",
  "refresh_token": "def50200abc...",
  "token_type": "Bearer",
  "expires_at": "2024-01-15T10:30:00Z"
}
```

## Мониторинг интеграции

1. **Проверка логов**:
   ```bash
   # Все логи CRM сервиса
   docker-compose logs -f crm-service
   
   # Последние 100 строк
   docker-compose logs --tail=100 crm-service
   ```

2. **Метрики в Grafana**:
    - Откройте http://localhost:3001
    - Используйте дашборд "AmoCRM Integration"
    - Отслеживайте:
        - Количество API запросов
        - Время ответа
        - Ошибки авторизации
        - Статус токенов

3. **Проверка состояния токенов**:
   ```bash
   # Войти в контейнер
   docker-compose exec crm-service sh
   
   # Проверить файл токенов
   cat /tmp/amocrm_token.json | jq .
   ```

## Расширенная настройка

### Настройка Rate Limiting

AmoCRM ограничивает количество запросов до 7 в секунду. Система автоматически управляет этим через Queue Service.

### Настройка батчинга

При массовых операциях система автоматически разбивает запросы на батчи по 200 сущностей.

### Настройка таймаутов

В файле `pkg/config/config.go` можно настроить:

```go
// Таймаут для HTTP запросов
HTTPTimeout: 30 * time.Second,

// Интервал обновления токенов
TokenRefreshInterval: 1 * time.Hour,
```

## Часто задаваемые вопросы

**Q: Как часто обновляются токены?**
A: Автоматически каждый час. Refresh token действителен 3 месяца.

**Q: Где хранятся токены в production?**
A: Рекомендуется использовать внешнее хранилище (Redis, PostgreSQL) вместо файловой системы.

**Q: Как добавить новые права доступа?**
A: Измените настройки интеграции в AmoCRM и повторите авторизацию.

**Q: Можно ли использовать несколько аккаунтов AmoCRM?**
A: Да, но потребуется доработка для multi-tenancy поддержки.