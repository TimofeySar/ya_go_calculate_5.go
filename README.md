# 📊 Распределённый вычислитель арифметических выражений

Проект `ya_go_calculate.go` реализует распределённую систему для вычисления арифметических выражений с использованием оркестратора и агентов. Пользователь отправляет выражение через REST API, система разбивает его на задачи, которые асинхронно выполняются агентами. Поддерживаются сложные выражения со скобками и авторизация пользователей. 🚀

---

## 📌 Возможности

- Поддержка операций: `+`, `-`, `*`, `/`, а также скобок.
- Асинхронное выполнение арифметических операций с настраиваемыми задержками.
- REST API для регистрации, авторизации, отправки выражений и получения результатов.
- Масштабируемость через настройку числа агентов (`COMPUTING_POWER`).
- Настраиваемое время выполнения операций через переменные окружения.
- Хранение данных в SQLite с поддержкой пользователей и выражений.
- Полное покрытие юнит-тестами для парсинга и обработки выражений.

---

## ⚙ Установка и запуск

### Требования

- **Go** (версия 1.20 или выше): [golang.org/dl/](https://golang.org/dl/)
- **protoc** (Protobuf компилятор): [github.com/protocolbuffers/protobuf/releases](https://github.com/protocolbuffers/protobuf/releases)
- **WinLibs** (для компиляции Protobuf на Windows): [winlibs.com](https://winlibs.com)
- **Git**: [git-scm.com/downloads](https://git-scm.com/downloads)
- **SQLite** (встроен в Go, установка не требуется)

### Установка `protoc`

1. Скачай `protoc` с [официальной страницы](https://github.com/protocolbuffers/protobuf/releases) (например, `protoc-27.3-win64.zip`)
2. Распакуй в `C:\protoc`
3. Добавь `C:\protoc\bin` в переменную окружения `PATH`:
   - Нажми `Win + R`, введи `sysdm.cpl`
   - Перейди во вкладку **Дополнительно** → **Переменные среды**
   - В разделе **Системные переменные** найди `Path`, нажми **Изменить**
   - Добавь: `C:\WinLibs\mingw64\bin`
4. Проверь:

```bash
protoc --version
```
Ожидаемый вывод: libprotoc 27.3

## 🧰 Установка WinLibs (для компиляции Protobuf)

1. Перейди на [https://winlibs.com](https://winlibs.com).
2. Скачай последнюю версию, например:  
   `winlibs-x86_64-posix-seh-gcc-13.2.0-llvm-17.0.6-mingw-w64ucrt-11.0.1-r4.zip`
3. Распакуй архив в папку, например:  
   `C:\WinLibs`
4. Добавь `C:\WinLibs\mingw64\bin` в переменную окружения `PATH`:
   - Нажми `Win + R`, введи `sysdm.cpl`
   - Перейди во вкладку **Дополнительно** → **Переменные среды**
   - В разделе **Системные переменные** найди `Path`, нажми **Изменить**
   - Добавь: `C:\WinLibs\mingw64\bin`
5. Проверь установку в PowerShell или CMD:

   ```bash
   gcc --version
   ```

   Ожидаемый вывод:

   ```
   gcc (x86_64-posix-seh-rev4, Built by MinGW-W64 project) 13.2.0
   ```

---

## 📦 Установка проекта

### 1. Клонирование репозитория

```bash
git clone https://github.com/TimofeySar/ya_go_calculate.go.git
cd ya_go_calculate.go
```

### 2. Установка зависимостей

```bash
go mod tidy
go get github.com/gorilla/mux
go get google.golang.org/protobuf
go get google.golang.org/grpc
```

### 3. Запуск оркестратора

```bash
go run ./cmd/orchestrator/main.go
```

Оркестратор запустится на:  
[http://localhost:8080](http://localhost:8080)

### 4. Запуск агента

#### Установка переменных окружения (пример для PowerShell):

```powershell
$env:TIME_ADDITION_MS=500
$env:TIME_SUBTRACTION_MS=500
$env:TIME_MULTIPLICATIONS_MS=1000
$env:TIME_DIVISIONS_MS=1000
$env:COMPUTING_POWER=2
```

#### Запуск агента:

```bash
go run ./cmd/agent/main.go
```

💡 По умолчанию:  
1000 мс для `+` и `-`, 2000 мс для `*` и `/`, 1 агент.

---

## 📚 Использование API

### Регистрация

- Метод: `POST`
- URL: `http://localhost:8080/api/v1/register`

#### Тело запроса:

```json
{
  "login": "user1",
  "password": "pass1"
}
```

#### Ответ:

```json
{
  "id": "1"
}
```

---

### Логин

- Метод: `POST`
- URL: `http://localhost:8080/api/v1/login`

#### Тело запроса:

```json
{
  "login": "user1",
  "password": "pass1"
}
```

#### Ответ:

```json
{
  "token": "<jwt-token>"
}
```

---

### Отправка выражения

- Метод: `POST`
- URL: `http://localhost:8080/api/v1/calculate`
- Заголовок: `Authorization: Bearer <jwt-token>`

#### Тело запроса:

```json
{
  "expression": "(2+2)*3"
}
```

#### Ответ:

```json
{
  "id": "expr-123456789"
}
```

---

### Получение всех выражений

- Метод: `GET`
- URL: `http://localhost:8080/api/v1/expressions`
- Заголовок: `Authorization: Bearer <jwt-token>`

#### Ответ:

```json
{
  "expressions": [
    {
      "id": "expr-123456789",
      "status": "completed",
      "result": 12
    }
  ]
}
```

---

### Получение результата по ID

- Метод: `GET`
- URL: `http://localhost:8080/api/v1/expressions/{id}`
- Заголовок: `Authorization: Bearer <jwt-token>`

#### Ответ:

```json
{
  "expression": {
    "id": "expr-123456789",
    "status": "completed",
    "result": 12
  }
}
```

---

## 🧪 Примеры запросов

### CMD (cURL)

```bash
curl -X POST http://localhost:8080/api/v1/register -H "Content-Type: application/json" -d "{\"login\":\"user1\",\"password\":\"pass1\"}"

curl -X POST http://localhost:8080/api/v1/login -H "Content-Type: application/json" -d "{\"login\":\"user1\",\"password\":\"pass1\"}"

curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -H "Authorization: Bearer <jwt-token>" -d "{\"expression\":\"(2+2)*3\"}"

curl -X GET http://localhost:8080/api/v1/expressions/expr-123456789 -H "Authorization: Bearer <jwt-token>"
```

### PowerShell

```powershell
# Регистрация нового пользователя
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/register" -Method POST -Headers @{ "Content-Type" = "application/json" } -Body '{"login":"12345","password":"12345"}'

# Логин и получение токена
$response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/login" -Method POST -Headers @{ "Content-Type" = "application/json" } -Body '{"login":"12345","password":"12345"}'
$user1Token = ($response.Content | ConvertFrom-Json).token

# Отправка выражения (например, (2+2)*5)
$calcResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/calculate" -Method POST -Headers @{ "Content-Type" = "application/json"; "Authorization" = "Bearer $user1Token" } -Body '{"expression":"(2+2)*5"}'
$exprId = ($calcResponse.Content | ConvertFrom-Json).id

# Ожидание
Start-Sleep -Seconds 4

# Получение результата
$response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/expressions/$exprId" -Method GET -Headers @{ "Authorization" = "Bearer $user1Token" }
Write-Output "Result for (2+2)*5: $($response.Content)"
```

#### Ожидаемый вывод:

```
Result for (2+2)*5: {"expression":{"ID":"expr-<id>","UserID":"1","Expr":"(2+2)*5","Status":"completed","Result":10}}
```

---

## 🗂️ Структура проекта

```
ya_go_calculate.go/
├── cmd/
│   ├── orchestrator/
│   │   └── main.go
│   └── agent/
│       └── main.go
├── internal/
│   ├── orchestrator/
│   │   ├── server.go
│   │   ├── expression.go
│   │   └── orchestrator_test.go
│   ├── agent/
│   │   └── worker.go
│   └── calculation/
│       ├── calc.go
│       └── calc_test.go
├── proto/
│   └── calculator.proto
├── go.mod
└── README.md
```

---

## 📈 Как это работает

### Оркестратор

- Принимает выражения через REST API.
- Парсит выражения в постфиксную форму (`InfixToPostfix`)
- Генерирует задачи (`GenerateTasks`)
- Сохраняет выражения в SQLite
- Отправляет задачи агентам через канал
- Собирает результаты и обновляет статус

### Агент

- Получает задачу от оркестратора
- Выполняет операцию с задержкой
- Возвращает результат

---

## 🔁 Схема взаимодействия

```plaintext
[Клиент] ---- POST /calculate ----> [Оркестратор]
    |                                   |
    |                                   | 1. Парсинг (InfixToPostfix)
    |                                   | 2. Разбиение (GenerateTasks)
    |                                   | 3. Сохранение в БД
    |                                   | 4. Отправка в канал задач
    |                                   v
    |                           [Канал задач]
    |                                   |
    |                            [Агенты] <--- GET задачи
    |                                   |
    |                            POST результат
    |                                   v
[Оркестратор] --- обновление статуса --> SQLite
```

## 🛠️ Тестирование

### Юнит-тесты

Проект покрыт тестами для:
- парсинга выражений,
- генерации задач,
- обработки выражений в `orchestrator`.

📌 **Важно:** тесты не работают при запущенном `orchestrator`, но **агент должен быть запущен**.

Запуск юнит-тестов:
```bash
go test ./internal/orchestrator -v
go test ./internal/integration
```