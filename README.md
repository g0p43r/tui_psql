# tui_psql

TUI-клиент для PostgreSQL на Go.

Сейчас проект находится на стадии MVP: можно подключиться к базе, увидеть список таблиц и preview данных, а также открыть выбранную запись в отдельном окне.

## Возможности

- форма подключения к PostgreSQL
- асинхронная проверка подключения через `pgx`
- список таблиц и views после успешного connect
- preview строк выбранной таблицы
- развёрнутый просмотр выбранной записи
- форматирование PostgreSQL-типов для TUI:
  - `uuid`
  - `json/jsonb`
  - `bytea`
  - `date`
  - `timestamp`
  - `timestamptz`
  - `numeric`

## Стек

- Go
- `bubbletea`
- `bubbles`
- `lipgloss`
- `pgx/v5`

## Запуск

Требования:

- Go `1.26.2`
- локально доступная PostgreSQL база

Запуск через `Taskfile`:

```bash
task start
```

Проверка сборки:

```bash
task build
```

Прямой запуск без `task`:

```bash
env GOCACHE=/tmp/tui_psql-gocache go run ./cmd/tui_psql
```

## Текущий UX

### Экран подключения

- `Tab` / `Shift+Tab` переключают поля
- `Enter` на поле `Password` запускает подключение
- `Ctrl+C` завершает приложение

Поля по умолчанию:

- `Host`: `localhost`
- `Port`: `5432`
- `Database`: `postgres`
- `User`: `postgres`
- `SSLMode`: сейчас жёстко `disable`

### Экран browser

- слева: список таблиц
- справа: preview выбранной таблицы

Управление:

- `Tab` переключает фокус между списком таблиц и preview
- `Up/Down` или `j/k` двигают выделение
- `Enter` на правой панели открывает выбранную запись
- `Esc` или `Enter` закрывают окно записи
- `Ctrl+C` завершает приложение

## Структура проекта

```text
cmd/tui_psql/                # entrypoint
internal/app/                # root Bubble Tea model и orchestration
internal/domain/             # доменные типы
internal/pg/                 # подключение, introspection, preview queries
internal/pg/formatter/       # приведение PostgreSQL значений к строкам для UI
internal/ui/screens/connection/
internal/ui/screens/browser/
internal/ui/styles/          # общие lipgloss стили
```

## Что уже реализовано архитектурно

- UI не ходит в PostgreSQL напрямую
- подключение и запросы запускаются через `tea.Cmd`
- форматирование значений вынесено в отдельный helper package
- экраны разделены на `model/update/view`

## Ограничения текущего MVP

- нет SQL editor
- нет history запросов
- нет скролла по большим result sets
- preview использует `SELECT * ... LIMIT 50`
- `sslmode` пока не настраивается из UI
- нет сохранения connection profiles

## Следующие шаги

- SQL editor
- полноценный result viewer
- scroll по строкам и колонкам
- disconnect / reconnect
- сохранение профилей подключений
- query history
