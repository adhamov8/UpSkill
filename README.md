# UpSkill Backend (Go + Kafka + PostgreSQL)

## Описание
**UpSkill** — это платформа для персонализированного обучения. Предполагается, что пользователь проходит тесты, получает **рекомендации** (персональный план), отслеживает прогресс, а также взаимодействует (в будущем) с менторами. Проект разбит на **четыре микросервиса**:
1. **AuthService** — регистрация, авторизация, JWT.
2. **UserService** — CRUD по пользователям.
3. **BadgeService** — управление наградами.
4. **RecommendationService** — логика для генерации персонального плана (упрощённо).

 Всё поднимается одной командой через общий `main.go`.

## Структура проекта
```
UpSkill/
├─ cmd/
│   └─ gateway/
│       └─ main.go               # Точка входа, запускающая все микросервисы
├─ internal/
│   ├─ config/                   # Конфигурация (env)
│   ├─ db/                       # Подключение к PostgreSQL и создание таблиц
│   └─ events/                   # Модуль для работы с Kafka
├─ service/
│   ├─ auth/                     # AuthService (регистрация, логин, refresh)
│   │   ├─ service.go
│   │   └─ validation.go         # Валидация email, password, и тд.
│   ├─ user/                     # UserService (CRUD пользователей)
│   ├─ badge/                    # BadgeService (добавление, получение наград)
│   └─ recommend/                # RecommendationService (заглушка рекомендаций)
├─ go.mod
├─ go.sum
└─ README.md
```

### AuthService (порт :8081)
- **POST /auth/register** — регистрация.
- **POST /auth/login** — логин (выдача `access_token` + `refresh_token`).
- **POST /auth/refresh** — обновление `access_token` по `refresh_token`.

### UserService (порт :8082)
- **GET /users** — получить список.
- **PUT /users/update?id=...** — обновить имя/фамилию.
- **DELETE /users/delete?id=...** — удалить.

### BadgeService (порт :8083)
- **GET /badges** — получить список наград.
- **POST /badges/create** — создать новую награду.

### RecommendationService (порт :8084)
- **POST /recommendations** — создать план для конкретного пользователя.
- **GET /recommendations** — получить все рекомендации.

Все они публикуют события в **Kafka**.

---

## Что сделано
1. **Микросервисная** архитектура (Auth, User, Badge, Recommendation).
2. **JWT** (регистрация, логин, refresh-токены).
3. **Валидация** (имя, email, пароль) в `service/auth/validation.go`.
4. **База**: **PostgreSQL** + `database/sql` для хранения пользователей, наград, рекомендаций.
5. **Kafka** (через [github.com/segmentio/kafka-go](https://github.com/segmentio/kafka-go)) для публикации событий.

---

## Что осталось сделать
1. **Чат с менторами**.
2. **AI/ML**.
3. **Admin-панель**.
4. **Восстановление пароля.
5. **Миграции**.
6. **Тесты**:.

---

## Запуск

### 1. Склонировать репозиторий
```bash
git clone https://github.com/adhamov8/UpSkill.git
cd UpSkill
```

### 2. Поднять PostgreSQL + Kafka
- **Docker Compose** 

### 3. Запустить

#### Вариант A: Локально
```bash
cd cmd/gateway
go mod tidy
go run .
```

#### Вариант B: Docker
```bash
docker-compose up --build
```

Запустятся сервисы:
- AuthService :8081
- UserService :8082
- BadgeService :8083
- RecommendationService :8084
- Gateway :8080 (возвращает простую страницу "UpSkill Gateway...")

---

## Примеры запросов

### Регистрация
```bash
curl -X POST http://localhost:8081/auth/register \
     -H "Content-Type: application/json" \
     -d '{
         "first_name":"Ali",
         "last_name":"Adhamov",
         "email":"ali@example.com",
         "password":"Qwerty123"
     }'
```

### Логин
```bash
curl -X POST http://localhost:8081/auth/login \
     -H "Content-Type: application/json" \
     -d '{
         "email":"ali@example.com",
         "password":"Qwerty123"
     }'
```
Ответ:
```json
{
  "access_token": "jwt...",
  "refresh_token": "rf_..."
}
```

### Получение пользователей
```bash
curl -X GET http://localhost:8082/users \
     -H "Authorization: Bearer <jwt_here>"
```

### Создать награду
```bash
curl -X POST http://localhost:8083/badges/create \
     -H "Content-Type: application/json" \
     -d '{"name":"Student","desc":"Complete advanced tasks"}'
```

### Добавить рекомендацию
```bash
curl -X POST http://localhost:8084/recommendations \
     -H "Content-Type: application/json" \
     -d '{"user_id":1}'
```

