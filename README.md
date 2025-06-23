# CRM-Dialer Integration System

Модульная система интеграции между CRM системами (AmoCRM) и системой автообзвона.

## Архитектура

Система построена на микросервисной архитектуре с использованием следующих технологий:

- **Frontend**: React + TypeScript + MUI + React Flow + MobX
- **Backend**: Go + gRPC + REST + NATS + GoFiber
- **База данных**: PostgreSQL + Redis
- **Контейнеризация**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **Мониторинг**: Grafana + Prometheus + Loki

## Компоненты системы

### Backend сервисы

1. **API Gateway** - единая точка входа, маршрутизация, аутентификация
2. **Webhook Service** - обработка входящих вебхуков от AmoCRM
3. **CRM Service** - интеграция с API AmoCRM
4. **Queue Service** - управление очередями с учетом rate limiting
5. **Flow Engine Service** - обработка бизнес-логики на основе настроек
6. **Dialer Service** - интеграция с системой автообзвона
7. **Config Service** - управление настройками интеграции

### Frontend

Веб-интерфейс для настройки интеграции с визуальным редактором бизнес-процессов на основе React Flow.

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+ (для локальной разработки)
- Node.js 18+ (для локальной разработки)

### Быстрый старт

1. Клонируйте репозиторий:
```bash
git clone https://github.com/yourcompany/crm-dialer-integration.git
cd crm-dialer-integration
```

2. Создайте файл `.env` на основе примера:
```bash
cp .env.example .env
```

3. Отредактируйте `.env` файл, добавив необходимые параметры:
```env
# AmoCRM
AMOCRM_DOMAIN=your-domain.amocrm.ru
AMOCRM_CLIENT_ID=your-client-id
AMOCRM_CLIENT_SECRET=your-client-secret
AMOCRM_REDIRECT_URI=http://localhost:8080/api/v1/amocrm/auth/callback

# Dialer System
DIALER_API_URL=https://your-dialer-api.com
DIALER_API_KEY=your-dialer-api-key

# Security
JWT_SECRET=your-secure-jwt-secret
POSTGRES_PASSWORD=your-secure-password
```

4. Запустите систему:
```bash
make up
```

5. Примените миграции базы данных:
```bash
make migrate
```

6. Откройте веб-интерфейс:
- Frontend: http://localhost:8080
- Grafana: http://localhost:3001 (admin/admin)
- Prometheus: http://localhost:9090

## Использование

### Настройка интеграции с AmoCRM

1. Перейдите в раздел "Настройки" в веб-интерфейсе
2. Нажмите "Подключить AmoCRM"
3. Следуйте инструкциям для авторизации
4. После успешной авторизации синхронизируйте поля и справочники

### Создание потока обработки

1. Перейдите в раздел "Потоки"
2. Нажмите "Создать новый поток"
3. Используйте визуальный редактор для построения логики:
    - Добавьте узел "Старт"
    - Добавьте условия для фильтрации сделок
    - Добавьте действия (отправка в автообзвон, обновление сделки и т.д.)
    - Соедините узлы стрелками
4. Сохраните и активируйте поток

### Настройка вебхуков

Система автоматически подписывается на следующие события AmoCRM:
- Создание сделки
- Обновление сделки
- Удаление сделки
- Изменение статуса сделки
- Изменение ответственного

URL для вебхуков: `https://your-domain.com/api/v1/webhooks/amocrm/{event_type}`

## API Документация

### Аутентификация

Все API запросы (кроме вебхуков) требуют JWT токен в заголовке:
```
Authorization: Bearer <token>
```

### Основные эндпоинты

#### AmoCRM
- `GET /api/v1/amocrm/fields` - получить список полей
- `POST /api/v1/amocrm/fields/sync` - синхронизировать поля
- `GET /api/v1/amocrm/leads` - получить список сделок
- `GET /api/v1/amocrm/leads/{id}` - получить сделку по ID

#### Flows
- `GET /api/v1/flows` - получить список потоков
- `GET /api/v1/flows/{id}` - получить поток по ID
- `POST /api/v1/flows` - создать новый поток
- `PUT /api/v1/flows/{id}` - обновить поток
- `DELETE /api/v1/flows/{id}` - удалить поток

#### Dialer
- `GET /api/v1/dialer/schedulers` - получить список расписаний
- `GET /api/v1/dialer/campaigns` - получить список кампаний
- `GET /api/v1/dialer/buckets` - получить список корзин

## Разработка

### Структура проекта

```
.
├── cmd/                    # Точки входа для сервисов
│   ├── api-gateway/
│   ├── webhook-service/
│   ├── crm-service/
│   └── ...
├── internal/              # Внутренний код приложения
│   ├── gateway/          # API Gateway handlers
│   ├── models/           # Модели данных
│   ├── repository/       # Слой работы с БД
│   └── services/         # Бизнес-логика
├── pkg/                  # Переиспользуемые пакеты
│   ├── config/
│   ├── logger/
│   └── nats/
├── web/                  # Frontend приложение
│   ├── src/
│   └── public/
├── migrations/           # SQL миграции
├── monitoring/           # Конфигурации мониторинга
└── docker-compose.yml    # Docker конфигурация
```

### Локальная разработка

#### Backend
```bash
# Установка зависимостей
go mod download

# Запуск сервиса
go run cmd/api-gateway/main.go

# Запуск тестов
go test ./...

# Линтинг
golangci-lint run
```

#### Frontend
```bash
cd web

# Установка зависимостей
npm install

# Запуск в режиме разработки
npm start

# Сборка
npm run build

# Тесты
npm test
```

## Мониторинг и логирование

### Метрики

Система экспортирует метрики в формате Prometheus:
- HTTP запросы (количество, длительность, статусы)
- Обработка вебхуков
- Очередь запросов к AmoCRM
- Отправка контактов в автообзвон

### Логи

Все сервисы пишут структурированные логи, которые собираются Loki и доступны в Grafana.

### Дашборды

В Grafana предустановлены дашборды для:
- Общего состояния системы
- Производительности API
- Статистики интеграций
- Ошибок и предупреждений

## Troubleshooting

### Проблемы с авторизацией AmoCRM

1. Проверьте правильность `AMOCRM_CLIENT_ID` и `AMOCRM_CLIENT_SECRET`
2. Убедитесь, что `AMOCRM_REDIRECT_URI` совпадает с настройками в AmoCRM
3. Проверьте логи CRM Service: `docker-compose logs crm-service`

### Rate limiting AmoCRM

Система автоматически управляет ограничениями API AmoCRM (7 запросов/сек).
Если возникают ошибки 429, проверьте Queue Service.

### Проблемы с NATS

1. Проверьте, что NATS запущен: `docker-compose ps nats`
2. Проверьте подключение: `docker-compose exec nats nats-cli`

## Contributing

1. Создайте feature branch: `git checkout -b feature/amazing-feature`
2. Commit изменения: `git commit -m 'Add amazing feature'`
3. Push в branch: `git push origin feature/amazing-feature`
4. Откройте Pull Request

## License

MIT License

Copyright (c) 2024 Your Company

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Поддержка

Для получения помощи:
- Создайте issue в репозитории
- Обратитесь в техподдержку: support@yourcompany.com
- Документация API: 