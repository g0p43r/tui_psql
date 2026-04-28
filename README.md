# tui_psql

TUI-клиент для PostgreSQL на Go.

Сейчас это рабочий MVP: можно подключиться к базе, сохранить профиль подключения, переключаться между профилями, просматривать таблицы, смотреть результат в viewer и открывать SQL-шаблоны для `INSERT`, `UPDATE`, `DELETE`.

## Возможности

- форма подключения к PostgreSQL
- асинхронное подключение через `pgx`
- сохранение connection profiles в локальный `profiles.json`
- пароль в профили не сохраняется
- список таблиц и views после успешного подключения
- result viewer:
  - вертикальная навигация по строкам
  - горизонтальная навигация по колонкам
  - sticky first column
  - row/column viewport status
- развёрнутый просмотр выбранной записи
- SQL editor overlay с шаблонами:
  - `INSERT`
  - `UPDATE`
  - `DELETE`
- lifecycle подключения:
  - disconnect
  - reconnect
- форматирование PostgreSQL-типов для UI:
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

Запуск:

```bash
task start
```

Сборка:

```bash
task build
```

Прямой запуск:

```bash
env GOCACHE=/tmp/tui_psql-gocache go run ./cmd/tui_psql
```

## Профили

Профили сохраняются через `os.UserConfigDir()`.

На macOS путь обычно такой:

```text
~/Library/Application Support/tui_psql/profiles.json
```

Важно:

- пароль не сохраняется
- активная сессия не сохраняется
- текущее подключение живёт только в памяти процесса

## Текущий UX

### Экран подключения

Поля по умолчанию:

- `Host`: `localhost`
- `Port`: `5432`
- `Database`: `postgres`
- `User`: `postgres`
- `SSLMode`: сейчас жёстко `disable`

Управление:

- `Tab` / `Shift+Tab` — навигация по форме
- `Enter` на поле `Password` — подключиться
- `Ctrl+P` — фокус на список профилей
- `Ctrl+F` — вернуть фокус на форму
- `Up/Down` или `j/k` — навигация по профилям
- `Enter` на списке профилей — применить профиль
- `Ctrl+D` — удалить выбранный профиль
- `Ctrl+C` — выход

### Экран browser

Layout:

- слева — список таблиц
- справа — result viewer выбранной таблицы

Управление:

- `Tab` — переключение фокуса между таблицами и viewer
- `Up/Down` или `j/k` — навигация по строкам / таблицам
- `Left/Right` или `h/l` — горизонтальная навигация по колонкам
- `PgUp/PgDn` — page navigation по строкам
- `Home/End` — перейти в начало/конец результата
- `Enter` — открыть выбранную запись
- `Esc` или `Enter` в record view — закрыть record view
- `F2` — открыть SQL template для `INSERT`
- `F3` — открыть SQL template для `UPDATE`
- `F4` — открыть SQL template для `DELETE`
- `Ctrl+P` — вернуться к выбору профилей
- `Ctrl+X` — disconnect
- `Ctrl+R` — reconnect
- `Ctrl+C` — выход

### SQL Editor

Пока это только editor с шаблонами, без выполнения SQL.

Сейчас:

- открывается как modal overlay
- заполняется шаблоном по выбранной таблице
- для `UPDATE` / `DELETE` использует текущую выбранную строку как подсказку в комментариях
- `Esc` закрывает editor

## Структура проекта

```text
cmd/tui_psql/                  # entrypoint
internal/app/                  # root Bubble Tea model и orchestration
internal/config/               # profiles storage
internal/domain/               # доменные типы
internal/pg/                   # connect, preview, introspection
internal/pg/formatter/         # приведение PostgreSQL значений к строкам для UI
internal/ui/screens/connection/
internal/ui/screens/browser/
internal/ui/styles/            # общие lipgloss стили
```

## Что уже реализовано архитектурно

- UI не ходит в PostgreSQL напрямую
- подключение и запросы запускаются через `tea.Cmd`
- formatter PostgreSQL-типов вынесен отдельно
- экраны разделены на `model/update/view`
- browser разбит на отдельные render-файлы

## Ограничения текущего MVP

- SQL editor пока не выполняет запросы
- viewer сейчас работает на preview `SELECT * ... LIMIT 50`
- нет history запросов
- нет сохранения активной session state
- `sslmode` пока не настраивается из UI

## Следующие шаги

- выполнение SQL из editor
- отдельный result mode для `INSERT/UPDATE/DELETE`
- `rows affected` и SQL errors в editor flow
- query history
