# 🚀 Go Language Showcase

Комплексный проект, демонстрирующий все основные возможности языка программирования Go.

## 📋 Содержание

Этот проект включает примеры и демонстрации следующих возможностей Go:

### 1. **Типы данных** (`types/`)
- Базовые типы (int, float, string, bool, complex)
- Структуры (structs)
- Массивы и срезы (arrays & slices)
- Ассоциативные массивы (maps)
- Константы и переменные

### 2. **Интерфейсы** (`interfaces/`)
- Определение и реализация интерфейсов
- Полиморфизм
- Type assertions и type switches
- Обработка ошибок (error interface)
- Кастомные ошибки

### 3. **Конкурентность** (`concurrency/`)
- Горутины (goroutines)
- Каналы (channels)
- Буферизованные и небуферизованные каналы
- Select statement
- Worker Pool паттерн
- Мьютексы и синхронизация (sync.Mutex, sync.WaitGroup)

### 4. **Дженерики** (`generics/`)
- Дженерик функции
- Дженерик структуры (Stack, Map)
- Ограничения типов (type constraints)
- Параметризованные типы

### 5. **Рефлексия** (`reflection/`)
- Получение типов и значений
- Инспекция структур
- Работа с тегами (struct tags)
- Изменение значений через рефлексию
- Вызов методов динамически

### 6. **База данных** (`database/`)
- Работа с SQLite
- CRUD операции
- Транзакции
- Prepared statements
- Экспорт данных в JSON

### 7. **HTTP Сервер** (`server/`)
- REST API с Gorilla Mux
- WebSocket для real-time коммуникации
- Rate Limiting (10 req/s)
- CORS middleware
- Security Headers
- Graceful Shutdown
- Structured Logging
- JSON encoding/decoding
- Красивый монохромный веб-интерфейс

### 8. **Продвинутые паттерны конкурентности** (`advanced/`)
- Pipeline Pattern
- Fan-Out/Fan-In Pattern
- Circuit Breaker Pattern
- Semaphore Pattern
- In-Memory Cache с TTL

### 8. **Прочие возможности** (`main.go`)
- Defer, panic, recover
- Работа с файлами (чтение, запись)
- Context (с таймаутом и отменой)
- Управление памятью

## 🔧 Установка

### Шаг 1: Установка Go

#### Windows
1. Установщик уже скачан в файл `go_installer.msi`
2. Запустите установщик (двойной клик)
3. Следуйте инструкциям мастера установки
4. После установки откройте новое окно терминала

Или скачайте вручную с официального сайта: https://go.dev/dl/

#### Linux
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### macOS
```bash
brew install go
```

### Шаг 2: Проверка установки
```bash
go version
```

Вы должны увидеть что-то вроде: `go version go1.21.5 windows/amd64`

### Шаг 3: Установка зависимостей проекта
```bash
go mod download
```

## 🚀 Запуск

### Запуск всего проекта
```bash
go run main.go
```

Это запустит все демонстрации последовательно и в конце запустит HTTP сервер.

### Запуск только HTTP сервера
После запуска `go run main.go`, HTTP сервер будет доступен по адресу:
```
http://localhost:8080
```

Откройте браузер и перейдите по этому адресу, чтобы увидеть красивый веб-интерфейс с документацией API.

### API Endpoints

- `GET /api/users` - Получить список всех пользователей
- `GET /api/users/{id}` - Получить пользователя по ID
- `POST /api/users` - Создать нового пользователя
- `PUT /api/users/{id}` - Обновить пользователя
- `DELETE /api/users/{id}` - Удалить пользователя
- `GET /api/stats` - Получить статистику сервера
- `WS /ws` - WebSocket подключение для real-time коммуникации

### Примеры использования API

#### Получить всех пользователей
```bash
curl http://localhost:8080/api/users
```

#### Создать пользователя
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Анна Иванова","email":"anna@example.com"}'
```

#### Получить пользователя по ID
```bash
curl http://localhost:8080/api/users/1
```

#### Обновить пользователя
```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Новое имя","email":"new@example.com"}'
```

#### Удалить пользователя
```bash
curl -X DELETE http://localhost:8080/api/users/1
```

#### Получить статистику
```bash
curl http://localhost:8080/api/stats
```

## 📦 Структура проекта

```
go-showcase/
├── main.go                 # Главный файл приложения
├── go.mod                  # Файл зависимостей
├── go.sum                  # Хэши зависимостей
├── README.md               # Этот файл
├── .gitignore              # Git игнорируемые файлы
├── types/                  # Демонстрация типов данных
│   └── basic_types.go
├── interfaces/             # Интерфейсы и обработка ошибок
│   └── interfaces.go
├── concurrency/            # Горутины и каналы
│   └── concurrency.go
├── generics/               # Дженерики (Go 1.18+)
│   └── generics.go
├── reflection/             # Рефлексия
│   └── reflection.go
├── database/               # Работа с БД
│   └── database.go
├── server/                 # HTTP сервер
│   └── server.go
├── advanced/               # Продвинутые паттерны
│   └── patterns.go
├── middleware/             # HTTP middleware
│   └── middleware.go
└── websocket/              # WebSocket hub
    └── hub.go
```

## 🎯 Что демонстрирует проект

### Базовые концепции
- ✅ Типы данных и переменные
- ✅ Структуры и методы
- ✅ Интерфейсы и полиморфизм
- ✅ Массивы, срезы и карты
- ✅ Указатели

### Продвинутые возможности
- ✅ Горутины и каналы
- ✅ Select statement
- ✅ Мьютексы и синхронизация
- ✅ Дженерики (Go 1.18+)
- ✅ Рефлексия
- ✅ Context

### Практическое применение
- ✅ HTTP сервер с REST API
- ✅ Работа с базой данных (SQLite)
- ✅ JSON encoding/decoding
- ✅ Файловые операции
- ✅ Middleware
- ✅ Error handling

### Паттерны и лучшие практики
- ✅ Worker Pool
- ✅ Dependency Injection
- ✅ Error wrapping
- ✅ Defer, panic, recover
- ✅ Package organization

## 📚 Дополнительные ресурсы

- [Официальная документация Go](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Tour](https://go.dev/tour/)

## 🤝 Вклад

Этот проект создан в образовательных целях для демонстрации возможностей Go.

## 📝 Лицензия

MIT License - используйте код свободно для обучения и практики!

## ⚡ Быстрый старт

1. Установите Go (используйте `go_installer.msi` или скачайте с go.dev)
2. Клонируйте или скачайте этот проект
3. Откройте терминал в папке проекта
4. Выполните: `go mod download`
5. Запустите: `go run main.go`
6. Откройте браузер: `http://localhost:8080`

Наслаждайтесь изучением Go! 🎉

