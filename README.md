# Counter Test Task

Микросервис на Go для учёта кликов по баннерам с поминутной статистикой.  
Сервис хранит данные в PostgreSQL и предоставляет REST API для регистрации кликов и получения статистики.

## Стек технологий
- **Go** — основной язык разработки
- **PostgreSQL** — хранение данных
- **Docker / Docker Compose** — контейнеризация
- **[goose](https://github.com/pressly/goose)** — управление миграциями базы данных

## 2. Запуск миграций базы данных

В проекте используется утилита **[goose](https://github.com/pressly/goose)** для управления миграциями PostgreSQL.

### 2.1. Установка goose
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -version
```
### 2.2. Подготовка подключения
Проверяем, что контейнер с БД запущен:
```bash
docker-compose up -d postgres
```
Экспортируем переменные окружения (порты и креды соответствуют docker-compose.yml):
```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=localhost port=55432 user=postgres password=postgres dbname=postgres sslmode=disable"
```

### 2.3 Запуск нагрузки
Примеры вызовов:
```bash
go run ./cmd/loader -mode=single -banner=1 -rps=500 -duration=60s
go run ./cmd/loader -mode=range -min=1 -max=100 -rps=300
```