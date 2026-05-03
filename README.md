# GoBank

Микросервисный банковский сервис на Go. Позволяет регистрироваться, создавать кошельки и переводить деньги между пользователями.

## Архитектура

```
┌─────────────┐     HTTP      ┌─────────────────┐     gRPC      ┌─────────────────┐
│   Client    │ ───────────►  │  auth-service   │ ◄──────────── │ wallet-service  │
└─────────────┘               │   :8080         │               │   :8081         │
                              └─────────────────┘               └────────┬────────┘
                                       │                                   │
                                  PostgreSQL                          PostgreSQL
                                  (auth DB)                          (wallet DB)
                                                                           │
                                                                         Kafka
                                                                           │
                                                                 ┌─────────▼────────┐
                                                                 │notification-svc  │
                                                                 └──────────────────┘
```
## Стек технологий
 
| Слой | Технология |
|------|-----------|
| Язык | Go 1.22 |
| HTTP | Chi |
| БД | PostgreSQL 16 + pgx |
| Миграции | golang-migrate |
| Кэш / Rate limiting | Redis 7 |
| Очередь | Kafka |
| Межсервисная коммуникация | gRPC + Protobuf |
| Авторизация | JWT |
| Документация | Swagger |
| Тесты | testify + mockery |
| CI | GitHub Actions |
| Инфраструктура | Docker Compose |


## Основные фичи

- **Регистрация и авторизация** — JWT access токены
- **Кошельки** — создание, пополнение, просмотр баланса
- **Переводы** — атомарные транзакции через PostgreSQL, деньги не теряются
- **История транзакций** — все операции по кошельку
- **Уведомления** — событие о переводе публикуется в Kafka, notification-сервис его обрабатывает
- **Rate limiting** — защита от брутфорса, 100 запросов в минуту на IP через Redis
- **gRPC авторизация** — wallet-сервис валидирует токены через auth по gRPC

## Быстрый старт

### 1. Клонируй репозиторий

```bash
git clone https://github.com/AriartyyyA/gobank
cd gobank
```

### 2. Создай .env файл

```bash
cp .env.example .env
```

### 3. Запусти инфраструктуру

```bash
docker compose up -d
```

### 4. Примени миграции

```bash
migrate -path migrations/auth \
  -database "postgres://gobank:secret@localhost:5432/gobank_auth?sslmode=disable" up

migrate -path migrations/wallet \
  -database "postgres://gobank:secret@localhost:5433/gobank_wallet?sslmode=disable" up
```

```

## API

### Auth Service (localhost:8080)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/register` | Регистрация |
| POST | `/auth/login` | Вход, возвращает JWT токен |
| GET | `/users/me` | Данные текущего пользователя |

Swagger UI: `http://localhost:8080/swagger/index.html`

### Wallet Service (localhost:8081)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/wallets` | Создать кошелёк |
| GET | `/wallets/me` | Баланс кошелька |
| GET | `/wallet` | Данные кошелька |
| POST | `/wallets/deposit` | Пополнить баланс |
| POST | `/wallets/transfer` | Перевод между кошельками |
| GET | `/wallet/history` | История транзакций |

> Все запросы к wallet-сервису требуют заголовок `Authorization: Bearer <token>`

## Структура проекта

```
gobank/
├── cmd/
│   ├── auth/           # Точка входа auth-сервиса
│   ├── wallet/         # Точка входа wallet-сервиса
│   └── notification/   # Точка входа notification-сервиса
├── internal/
│   ├── auth/           # Auth микросервис
│   │   ├── domain/     # Сущности, интерфейсы, ошибки
│   │   ├── usecase/    # Бизнес-логика
│   │   ├── delivery/   # HTTP и gRPC хэндлеры
│   │   └── repository/ # PostgreSQL репозиторий
│   ├── wallet/         # Wallet микросервис (аналогично)
│   └── notification/   # Notification микросервис
├── pkg/                # Общий код
│   ├── middleware/     # JWT middleware
│   ├── ratelimit/      # Rate limiting через Redis
│   └── kafka/          # Producer и Consumer
├── proto/              # Protobuf определения
├── migrations/         # SQL миграции
└── deploy/             # Конфиги Prometheus и Grafana
```

## Архитектурные решения

**Clean Architecture** — зависимости идут только внутрь: `delivery → usecase → domain`. Слой domain ничего не знает о БД или HTTP.

**Транзакции через контекст** — паттерн `WithTx` позволяет usecase запускать атомарные операции не зная про pgx напрямую.

**Деньги в копейках** — баланс хранится как `int64` (копейки), без проблем с точностью float.

**gRPC для межсервисного общения** — wallet не дублирует JWT логику, а спрашивает у auth.