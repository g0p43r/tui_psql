# tui_psql

A terminal UI PostgreSQL client written in Go.

The project is currently an MVP, but it already supports connection profiles, table browsing, result viewing, SQL templates, a standalone SQL workbench, query history, and basic table/row operations.

## Features

- PostgreSQL connection form with `sslmode` support.
- Local connection profiles stored as JSON.
- Passwords are not saved to profiles.
- Profile navigation and profile deletion.
- Disconnect and reconnect actions.
- Table and view browser after a successful connection.
- Table preview with:
  - vertical row navigation;
  - horizontal column navigation;
  - sticky first column;
  - row/column viewport status;
  - horizontal offset indicator.
- Expanded record view for the selected row.
- SQL templates for:
  - `INSERT`;
  - `UPDATE`;
  - `DELETE`;
  - `CREATE TABLE`;
  - `ALTER TABLE`;
  - `DROP TABLE`.
- SQL execution from editor flows.
- Standalone SQL Workbench opened with `Ctrl+E`.
- Query Workbench split layout:
  - top panel for SQL input;
  - bottom panel for query results.
- Query result navigation inside the workbench.
- Query history panel for reusing previous queries.
- Schema refresh after `CREATE TABLE`, `ALTER TABLE`, and `DROP TABLE`.
- Table preview reload after `INSERT`, `UPDATE`, and `DELETE`.
- Query result row limit with truncated result indication.
- Safer generated SQL templates with quoted identifiers.
- PostgreSQL value formatting for:
  - `uuid`;
  - `json` / `jsonb`;
  - `bytea`;
  - `date`;
  - `timestamp`;
  - `timestamptz`;
  - `numeric`.

## Stack

- Go
- Bubble Tea
- Bubbles
- Lip Gloss
- pgx/v5
- pgxpool

## Run

Requirements:

- Go `1.26.2`
- A reachable PostgreSQL database
- Task, if you want to use `Taskfile.yml`

Start the app:

```bash
task start
```

Build:

```bash
task build
```

Run directly:

```bash
env GOCACHE=/tmp/tui_psql-gocache go run ./cmd/tui_psql
```

Verify:

```bash
env GOCACHE=/tmp/tui_psql-gocache go build ./...
env GOCACHE=/tmp/tui_psql-gocache go test ./...
```

## Profiles

Profiles are stored through `os.UserConfigDir()`.

On macOS the file is usually located at:

```text
~/Library/Application Support/tui_psql/profiles.json
```

Important behavior:

- passwords are not saved;
- active sessions are not saved;
- the current database connection lives only in the running process;
- reconnect uses the current profile/form values.

## Connection Screen

Default fields:

- `Host`: `localhost`
- `Port`: `5432`
- `Database`: `postgres`
- `User`: `postgres`
- `SSLMode`: `disable`
- `Password`: empty

Supported `sslmode` values:

- `disable`
- `allow`
- `prefer`
- `require`
- `verify-ca`
- `verify-full`

Hotkeys:

- `Tab` / `Shift+Tab`: move through form fields
- `Enter` on the password field: connect
- `Ctrl+P`: focus saved profiles
- `Ctrl+F`: focus the connection form
- `Up` / `Down` or `k` / `j`: move through profiles or form fields
- `Enter` in profiles list: apply selected profile
- `Ctrl+D`: delete selected profile
- `Ctrl+Q`: quit

## Browser Screen

Layout:

- left panel: database tables and views;
- right panel: table preview/result viewer;
- bottom panel: hotkey help.

General navigation:

- `Tab`: switch focus between tables and result viewer
- `Up` / `Down` or `k` / `j`: move through tables or rows
- `Left` / `Right` or `h` / `l`: move across columns in the result viewer
- `PgUp` / `PgDown`: page through rows
- `Home` / `End`: jump to first/last row
- `Enter`: open expanded record view from the result viewer
- `Esc` or `Enter`: close expanded record view
- `Alt+Up` / `Alt+Down`: scroll the whole layout when needed
- `Ctrl+Q`: quit

Connection actions:

- `Ctrl+P`: go back to profile selection
- `Ctrl+X`: disconnect
- `Ctrl+R`: reconnect

Table actions, when the left table panel is focused:

- `Ctrl+C`: create table
- `Ctrl+U`: alter table
- `Ctrl+D`: drop table

Row actions, when the right result viewer is focused:

- `Ctrl+C`: create row
- `Ctrl+U`: update row
- `Ctrl+D`: delete row

SQL actions:

- `Ctrl+E`: open standalone SQL Workbench
- `Ctrl+H`: open SQL history
- `Alt+Enter`: execute SQL from an editor/workbench
- `Ctrl+T`: cycle query type in editor/workbench

## SQL Editor

The SQL editor is used for table and row workflows.

It can generate templates for:

- inserting a row into the selected table;
- updating the selected row;
- deleting the selected row;
- creating a table;
- altering a table;
- dropping a table.

Template behavior:

- identifiers are quoted in generated SQL;
- column types are included in comments;
- `UPDATE` and `DELETE` templates include current selected row values;
- schema-qualified objects are detected for refresh/reload flows.

Execution behavior:

- `Alt+Enter` executes the SQL;
- `Ctrl+T` cycles query type: `auto`, `select`, `insert`, `update`, `delete`, `exec`;
- successful DDL refreshes schema objects;
- successful row mutations reload the relevant table preview when possible;
- errors are shown in the editor.

## SQL Workbench

The SQL Workbench is a standalone full-screen query screen opened with `Ctrl+E`.

Layout:

- top 20%: SQL input;
- bottom 80%: query output.

Hotkeys:

- `Alt+Enter`: execute SQL
- `Ctrl+T`: cycle query type
- `Alt+Up` / `Alt+Down`: move through result rows
- `Alt+Left` / `Alt+Right`: move through result columns
- `PgUp` / `PgDown`: page through result rows
- `Esc`: close the workbench

Result behavior:

- `SELECT`, `WITH`, and `SHOW` run as read queries in auto mode;
- read query results are limited to 500 rows;
- truncated results are marked in the result viewer;
- workbench results stay in the workbench result panel.

## SQL History

The app stores recent SQL executions in memory for the current process.

History behavior:

- successful queries are stored;
- failed queries are stored with error status;
- history is limited to the latest 200 entries;
- history is not persisted across app restarts.

Hotkeys:

- `Ctrl+H`: open history from the browser screen
- `Up` / `Down` or `k` / `j`: move through history
- `Enter`: open selected query in the SQL editor
- `Esc`: close history

## Database Layer

The app uses `pgxpool.Pool` for database access.

Important details:

- database operations run through Bubble Tea commands;
- the UI does not call PostgreSQL directly;
- request IDs prevent stale table/preview responses from replacing newer UI state;
- session IDs prevent old SQL responses from affecting the UI after reconnect/disconnect;
- arbitrary read query output is capped for TUI safety.

## Project Structure

```text
cmd/tui_psql/                  # application entrypoint
internal/app/                  # root Bubble Tea model and orchestration
internal/config/               # profile storage
internal/domain/               # domain types
internal/errs/                 # unified application errors
internal/pg/                   # connect, execute, preview, introspection
internal/pg/formatter/         # PostgreSQL value formatting for UI
internal/ui/screens/connection/
internal/ui/screens/browser/
internal/ui/styles/            # shared Lip Gloss styles
```

## Current Limitations

- Query history is in-memory only.
- Table preview uses `SELECT * ... LIMIT 50`.
- Arbitrary read query results are capped at 500 rows.
- There is no persisted session state.
- There is no full SQL parser; statement detection is intentionally lightweight.
- Result editing is template-driven, not a spreadsheet-style inline editor.

## Useful Next Steps

- Persist SQL history.
- Add configurable preview/query limits.
- Add schema browser and object search.
- Add cancelable long-running queries.
- Add transaction-aware editor workflows.
- Split browser state into smaller nested models.
