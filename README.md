# Bank Ledger Core

Микросервис для банковских транзакций на Go с использованием Gin, GORM и PostgreSQL.

## Структура проекта

```
bank-ledger-core/
├── config/
│   └── database.go          # Конфигурация базы данных
├── handlers/
│   ├── account_handler.go   # Обработчики для счетов
│   └── transfer_handler.go  # Обработчики для переводов
├── models/
│   ├── account.go           # Модель счета
│   └── transfer.go          # Модель перевода
├── routes/
│   └── routes.go            # Маршруты API
├── services/
│   ├── decimal.go           # Вспомогательные функции для decimal
│   └── transfer_service.go  # Сервис переводов
├── docker-compose.yml       # Docker Compose конфигурация
├── Dockerfile              # Docker конфигурация
├── go.mod                  # Go модуль
├── go.sum                  # Зависимости
├── main.go                 # Точка входа
└── README.md               # Документация
```

## API Эндпоинты

### Счета
- `POST /api/v1/accounts` - Создать новый счет
- `GET /api/v1/accounts` - Получить все счета
- `GET /api/v1/accounts/:id` - Получить счет по ID

### Переводы
- `POST /api/v1/transfers/money` - Выполнить перевод средств

### Health Check
- `GET /health` - Проверка состояния сервиса

## Примеры запросов

### Создание счета
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "currency": "USD",
    "balance": "1000.00"
  }'
```

### Перевод средств
```bash
curl -X POST http://localhost:8080/api/v1/transfers/money \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": "100.50"
  }'
```

## Запуск проекта

### С Docker Compose
```bash
docker-compose up --build
```

### Локально
1. Установите PostgreSQL
2. Настройте переменные окружения:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=password
   export DB_NAME=bank_ledger
   export DB_SSLMODE=disable
   ```
3. Запустите приложение:
   ```bash
   go run main.go
   ```

## Особенности реализации

- **Транзакции**: Метод `TransferMoney` использует `db.Transaction` из GORM для обеспечения атомарности операций
- **Валидация**: Проверка наличия средств, совпадения валют и существования счетов
- **Откат**: При любой ошибке все операции автоматически откатываются
- **Гибкость**: Легкое переключение между PostgreSQL и SQLite через переменную окружения `DB_DRIVER`
- **Десятичные числа**: Использование `big.Float` для точных финансовых расчетов

## Переменные окружения

- `DB_DRIVER` - драйвер базы данных (postgres/sqlite, по умолчанию: postgres)
- `DB_HOST` - хост базы данных (по умолчанию: localhost)
- `DB_PORT` - порт базы данных (по умолчанию: 5432)
- `DB_USER` - пользователь базы данных (по умолчанию: postgres)
- `DB_PASSWORD` - пароль базы данных (по умолчанию: password)
- `DB_NAME` - имя базы данных (по умолчанию: bank_ledger)
- `DB_SSLMODE` - режим SSL (по умолчанию: disable)
- `PORT` - порт приложения (по умолчанию: 8080)
