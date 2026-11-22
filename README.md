# WhoDB

![WhoDB](./logo.png)

WhoDB — это инструмент для управления базами данных с веб-интерфейсом, GraphQL API, CLI и desktop-приложением. Проект помогает подключаться к разным источникам данных, просматривать структуру и содержимое, выполнять запросы и работать со схемой в одном репозитории.

## ✨ Возможности

- Веб-приложение для работы с базами данных на Go и React
- GraphQL API и HTTP-эндпоинты для работы с данными
- Подключение к PostgreSQL, MySQL, MariaDB, SQLite3, MongoDB, Redis, ClickHouse, ElasticSearch и Memcached
- Просмотр таблиц и других объектов источника данных через разделы `storage-unit`
- Визуализация схемы и связей в разделе `graph`
- Выполнение запросов в редакторе `scratchpad`
- CLI-интерфейс для терминала
- Desktop-приложение на Wails
- Экспорт данных в CSV, Excel и NDJSON
- Поддержка AI-функций для чата и работы с запросами, если провайдер настроен

## 🚀 Быстрый старт

### Запуск через Docker

```bash
docker run -it -p 8080:8080 clidey/whodb
```

После запуска откройте `http://localhost:8080`.

### Локальный запуск из исходников

Подтверждено по файлам проекта:

- Go `1.26.1`
- `pnpm` `10.29.3`
- Версия Node.js в проекте отдельно не указана

1. Клонируйте репозиторий:

```bash
git clone https://github.com/QuestGuideRF/whodb.git
cd whodb
```

2. Установите зависимости frontend:

```bash
cd frontend
pnpm install
```

3. В отдельном терминале запустите backend:

```bash
cd core
go run ./cmd/whodb
```

4. В отдельном терминале запустите frontend:

```bash
cd frontend
pnpm start
```

Во время локальной разработки:

- backend по умолчанию работает на `http://localhost:8080`
- frontend запускается на `http://localhost:3000`
- запросы `frontend` к `/api` проксируются на backend

### Запуск CLI

```bash
cd cli
go run .
```

### Запуск desktop-приложения

```bash
cd desktop-ce
make install-wails
make dev
```

## 📁 Структура проекта

```text
core/            Backend на Go, GraphQL-схема, роутер и плагины источников данных
frontend/        Web-интерфейс на React + TypeScript
cli/             CLI и TUI-интерфейс
desktop-ce/      Desktop-приложение Community Edition на Wails
desktop-common/  Общие desktop-скрипты и Makefile-логика
dev/             Вспомогательные dev-скрипты
flatpak/         Файлы упаковки для Flatpak
go.work          Go workspace для модулей core и cli
README.md        Описание проекта
```

## 🛠 Использование

После запуска веб-версии откройте `http://localhost:3000` при разработке или `http://localhost:8080` при запуске собранного backend/контейнера.

Типовой сценарий работы:

1. Откройте страницу входа и укажите параметры подключения к источнику данных.
2. Перейдите в `Storage Unit`, чтобы просматривать таблицы, коллекции и строки.
3. Используйте `Graph`, чтобы посмотреть схему и связи между объектами.
4. Используйте `Scratchpad`, чтобы выполнять запросы вручную.
5. Используйте `Chat`, если в окружении настроен AI-провайдер.
6. При необходимости экспортируйте данные через доступные функции backend.

Полезные адреса, подтвержденные кодом:

```bash
curl http://localhost:8080/health
```

- `GET /health` — проверка доступности сервера
- `POST /api/query` — GraphQL API
- `/api/*` — дополнительные HTTP-эндпоинты проекта

Для frontend также присутствует генерация GraphQL-клиента:

```bash
cd frontend
pnpm run generate
```

## 🧪 Технологии

- Go
- Chi Router
- gqlgen / GraphQL
- GORM
- React
- TypeScript
- Vite
- Apollo Client
- Redux Toolkit
- Tailwind CSS
- Wails
- Bubble Tea
- Docker

________________________________________

## Настольное приложение

Общий код рабочего стола и сборочные утилиты Wails находятся в `desktop-common/`. Перед сборкой проверьте среду через `wails doctor`. В Windows требуется WebView2, в macOS — инструменты командной строки Xcode, в Linux — компилятор C, GTK 3 и WebKitGTK. Для выбранной поставки запускайте `make dev` или `make build` из каталога с её Makefile. Список целей выводит `make help`. Меню, диалоги файлов и события окна обслуживаются общим слоем Go.

## Командная строка

CLI доступен через `npx @clidey/whodb-cli`. Без аргументов открывается терминальный интерфейс. Также доступны команды для запросов, просмотра схем, таблиц и столбцов, экспорта и запуска MCP-сервера. Конкретный синтаксис выводит `npx @clidey/whodb-cli --help`.

Подключения можно задать переменными `WHODB_<DBTYPE>` с JSON-массивом или `WHODB_<DBTYPE>_N` с отдельным JSON-профилем. Сохраняйте доступы в конфигурации окружения. В режиме MCP предусмотрены отдельные операции для подключений, схем и запросов. Шаблоны для установки навыков MCP находятся в `cli/skills/`.

Подпись относится к текущей версии проекта.

---

Автор: legenda  
Telegram: [@LEGENDA_SD](https://t.me/LEGENDA_SD)
