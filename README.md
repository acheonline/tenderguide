# tenderguide

CLI для автоматического разбора заявок на конкурс из DOCX в SQLite.

CLI-сервис для загрузки и разбора пакета заявки на конкурс. Читает директорию с DOCX-файлами комплекта (формы 1–13), извлекает атрибуты заявки, участника, заказчика, банковских реквизитов, МСП-статуса, штата и бенефициаров, сохраняет в SQLite.

## Getting Started

Краткое введение: требования, установка зависимостей и команды запуска описаны в разделах ниже — [Requirements](#requirements), [How to run](#how-to-run), [How to test](#how-to-test).

## Requirements

- Go 1.22+ (см. `go.mod`)
- SQLite-драйвер `modernc.org/sqlite` — pure Go, без CGO, отдельная установка не требуется
- `make` не используется — в репозитории нет `Makefile`

## How to run

### Установка зависимостей

```bash
git clone https://github.com/acheonline/tenderguide.git
cd tenderguide
go mod download
```

### Запуск

Без сборки:

```bash
go run ./cmd/tenderguide load --dir ./testdata --db /tmp/tenderguide.db
go run ./cmd/tenderguide show --request testdata --db /tmp/tenderguide.db
```

С предварительной сборкой бинарника:

```bash
go build -o tenderguide ./cmd/tenderguide
./tenderguide load --dir ./testdata --db /tmp/tenderguide.db
./tenderguide show --request testdata --db /tmp/tenderguide.db
```

После `load` в stdout выводится `OK: request <id> saved (id=<n>)`, данные сохраняются в SQLite-файл по пути `--db` (по умолчанию `./tenderguide.db`). После `show` в stdout печатается карточка заявки в формате `text` или `json`. Отдельный конфигурационный файл и переменные окружения не требуются — всё задаётся флагами CLI.

## How to test

Все тесты запускаются стандартной командой Go:

```bash
go test ./...
```

Дополнительно можно прогнать статический анализ:

```bash
go vet ./...
```

Тестовые данные для интеграционных сценариев (`load`/`show`) уже лежат в репозитории в каталоге `testdata/` — дополнительная подготовка не требуется. Отдельного разделения на юнит/интеграционные/e2e-категории и отдельных целей `make test` нет.

## Использование

### `load` — загрузить директорию с заявкой

```
tenderguide load --dir <path> [--db <path>]
```

| Флаг | Обязательный | По умолчанию | Описание |
|---|---|---|---|
| `--dir` | да | — | директория с DOCX-файлами одной заявки |
| `--db` | нет | `./tenderguide.db` | путь к файлу SQLite |

**Одна директория = одна заявка.** `request_id` = имя директории. Внутри — любое число `*.docx` (формы 1–13, не все обязательно).

Пример:

```bash
go run ./cmd/tenderguide load --dir ./testdata --db /tmp/tenderguide.db
# OK: request testdata saved (id=1)
```

Повторный запуск с тем же `request_id` → exit `1`, `duplicate request_id`, БД не меняется.

### `show` — вывести заявку из SQLite

```
tenderguide show --request <id> [--db <path>] [--format text|json]
```

| Флаг | Обязательный | По умолчанию | Описание |
|---|---|---|---|
| `--request` | да | — | `request_id` заявки (имя директории при `load`) |
| `--db` | нет | `./tenderguide.db` | путь к файлу SQLite |
| `--format` | нет | `text` | `text` или `json` |

Пример:

```bash
go run ./cmd/tenderguide show --request testdata --db /tmp/tenderguide.db
# === Заявка ===
# ID:                 testdata
# ...

go run ./cmd/tenderguide show --request testdata --db /tmp/tenderguide.db --format json
# {"request":{"request_id":"testdata",...}}
```

В формате `text` `NULL`-поля выводятся как `—`, пустые секции — с пометкой `(нет)`.

### `--version` — вывести версию

```bash
go run ./cmd/tenderguide --version
# tenderguide dev
```

Версию можно переопределить при сборке: `go build -ldflags "-X main.version=1.2.3" -o tenderguide ./cmd/tenderguide`.

## Что извлекается

Поля группируются по таблицам.

### `requests`

`request_id`, `procedure_number`, `procedure_name`, `procedure_published_at`, `letter_number`, `letter_date`, `valid_until`, `price_with_vat`, `vat`, `price_without_vat`, `price_currency`, `award_experience_amount`, `created_at`.

### `customers`

`inn` (в документах отсутствует — всегда `NULL`), `name`.

### `participants`

`inn`, `kpp`, `ogrn`, `okpo`, `oktmo`, `okopf`, `okopf_name`, `full_name`, `short_name`, `legal_address`, `postal_address`, `primary_okved`, `primary_okved_name`, `egrul_registration_date`, `egrul_certificate`, `egrul_registration_authority`, `tax_registration_date_current`, `tax_authority_current`, `phone`, `fax`, `email`, `website`, `director_name`, `director_position`, `chief_accountant_name`, `responsible_person_name`, `responsible_person_phone`, `has_branches`.

### `bank_accounts`

`bank_name`, `account`, `correspondent_account`, `bik`.

### `msp_info`

`is_smp`, `smp_category`, `average_headcount`, `revenue_previous_year`, `contracts_44_223_count`, `contracts_44_223_sum`, `in_rnp`.

### `staff_summary`

`total_staff`, `managers_count`, `specialists_count`, `workers_count`.

### `staff_members`

`category`, `full_name`, `position`, `education`, `experience_years`.

### `beneficiaries`

`full_name`, `inn`, `passport`, `registration_address`, `role`, `share_gt_10pct`.

## Принцип «сохраняем что смогли»

- В схеме БД нет `NOT NULL` на бизнес-полях — только технический `id`.
- Парсер не падает на отсутствии поля: не нашли — `NULL`.
- Нет формы — нет строк в соответствующей таблице.
- Заявка, в которой не распознано ни одного поля, всё равно сохраняется (кроме `id` и `request_id`).
- Единственная защита от повторной загрузки — `UNIQUE` на `requests.request_id`.

## Схема БД

Восемь таблиц, создаются через `CREATE TABLE IF NOT EXISTS` при `storage.Open`:

- `requests` — заявка (1:1 с директорией).
- `customers` — заказчик (0..1 на заявку).
- `participants` — участник (0..1 на заявку, дедуп по `inn`).
- `bank_accounts` — банковские реквизиты (0..N на участника).
- `msp_info` — МСП-статус (0..1 на участника).
- `staff_summary` — сводка по штату (0..1 на заявку).
- `staff_members` — сотрудники (0..N на заявку).
- `beneficiaries` — учредители/бенефициары (0..N на участника).

FK: `requests.participant_id → participants.id`, `requests.customer_id → customers.id`, `bank_accounts.participant_id → participants.id`, `msp_info.participant_id → participants.id`, `staff_summary.request_id → requests.id`, `staff_members.request_id → requests.id`, `beneficiaries.participant_id → participants.id`.

## Структура проекта

```
cmd/tenderguide/        точка входа CLI
internal/model/         Go-структуры (nullable через указатели)
internal/loader/        чтение DOCX → []parser.Input
internal/parser/        извлечение полей из текста (regex + нормализаторы)
internal/storage/       SQLite: схема, Save, GetByRequestID
testdata/               пример комплекта DOCX
```

## Разработка

### Проверка результата в БД

```bash
sqlite3 /tmp/tenderguide.db ".tables"
sqlite3 /tmp/tenderguide.db "SELECT request_id, procedure_number, price_with_vat FROM requests;"
sqlite3 /tmp/tenderguide.db "SELECT inn, kpp, ogrn, full_name FROM participants;"
sqlite3 /tmp/tenderguide.db "SELECT * FROM bank_accounts;"
```

### Проверка схемы

```bash
sqlite3 /tmp/tenderguide.db "PRAGMA table_info(participants);"
```

Все бизнес-столбцы должны иметь `notnull = 0`, кроме `id` (технический первичный ключ).

## Зависимости

- `github.com/fumiama/go-docx` — чтение DOCX (AGPL-3.0, согласовано мейнтейнером).
- `modernc.org/sqlite` — SQLite-драйвер (pure Go, без CGO).

Все зависимости ставятся через `go mod download`.

## Automation

### Overview

- issue + метка `agent-todo` → `RepoKeeper Agent` (Pi backend, до 5 попыток исправления) создаёт PR в ветке `repokeeper/issue-<N>-...`
- `PR Verifier (Auto-Regression)` запускается по `workflow_run` после успешного завершения `RepoKeeper Agent`
- Проверяет: сборку, `load` на `testdata/`, содержимое БД, `show --format json`, `go test ./...`
- При успехе: `gh pr merge --squash --delete-branch` + `gh issue close --reason completed`
- При сбое `RepoKeeper Agent`: `Repair Agent` анализирует лог, переписывает issue через LLM, закрывает старый (not planned) и создаёт новый с `agent-todo` (до 3 попыток)

### Labels

- `agent-todo` (на issue) — запускает `RepoKeeper Agent` для создания PR
- `agent-fix` (на PR) — запускает `RepoKeeper Agent` в режиме исправления PR

### Workflows

- `.github/workflows/repokeeper.yml` — `RepoKeeper Agent`: реализует issue или исправляет PR. Бэкенд `pi`, до 5 попыток исправления тестов. Конфиг — `repokeeper.yml` в корне.
- `.github/workflows/pr-verifier.yml` — `PR Verifier (Auto-Regression)`: запускается по `workflow_run` после `RepoKeeper Agent`. Проверяет сборку, `load`, БД, `show --format json`, `go test ./...`. При успехе мерджит PR и закрывает issue.
- `.github/workflows/repair-agent.yml` — `Repair Agent (RepoKeeper Failure Handler)`: запускается по `workflow_run` при сбое `RepoKeeper Agent`. Анализирует лог, переписывает issue через LLM, закрывает старый (not planned) и создаёт новый с `agent-todo`.

### Where to look on failure

- **RepoKeeper упал** — вкладка Actions → `RepoKeeper Agent` → шаг `Run RepoKeeper Agent` → искать `Verification failed` в логе.
- **Repair Agent не сработал** — проверить, что `conclusion == failure` (а не `success` с отказом от реализации).
- **PR Verifier упал** — Actions → `PR Verifier (Auto-Regression)` → красный шаг указывает, что именно не прошло.
- **PR не найден** — проверить префикс ветки в `find_pr` (ожидается `repokeeper/issue-<N>-`), увеличить таймаут (`seq 1 30`).
- **Merge не прошёл** — проверить `permissions: contents: write` в `pr-verifier.yml` и отсутствие branch protection с required approve.

### Self-hosted runner

Workflow запускаются на self-hosted runner с метками `[self-hosted, tenderguide]`. Требования к runner: Go 1.22+, `sqlite3`, `python3`, `gh` CLI, Node.js 22 (для Pi-бэкенда RepoKeeper).

## Вне скоупа текущей версии

- PDF, XLSX, сканы, OCR.
- Проверки 44-ФЗ / 223-ФЗ, свежесть справок, сверка ТЗ и спецификации.
- Несколько заявок за один запуск.
- Подкоманды `list`, `export`.
- Миграции (схема создаётся через `CREATE TABLE IF NOT EXISTS`).
- REST/gRPC API, UI.