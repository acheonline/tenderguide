# tenderguide

CLI-сервис для загрузки и разбора пакета заявки на конкурс. Читает директорию с DOCX-файлами комплекта (формы 1–13), извлекает атрибуты заявки, участника, заказчика, банковских реквизитов, МСП-статуса, штата и бенефициаров, сохраняет в SQLite.

## Getting Started

### Требования

- Go 1.22+
- `make` (опционально, для удобных команд)
- SQLite (драйвер `modernc.org/sqlite` — pure Go, без CGO, отдельная установка не требуется)

### Клонирование и установка зависимостей

```bash
git clone https://github.com/acheonline/tenderguide.git
cd tenderguide
go mod download
```

### Локальный запуск

```bash
go run ./cmd/tenderguide load --dir ./testdata --db /tmp/tenderguide.db
go run ./cmd/tenderguide show --request testdata --db /tmp/tenderguide.db
```

## Требования

- Go 1.22+
- SQLite (драйвер `modernc.org/sqlite` — pure Go, без CGO)

## Сборка

```bash
go build -o tenderguide ./cmd/tenderguide
```

Или запуск без сборки:

```bash
go run ./cmd/tenderguide <command> [flags]
```

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
```

В формате `text` `NULL`-поля выводятся как `—`, пустые секции — с пометкой `(нет)`.

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

### Тесты

```bash
go test ./...
```

### Линт

```bash
go vet ./...
```

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

## Вне скоупа текущей версии

- PDF, XLSX, сканы, OCR.
- Проверки 44-ФЗ / 223-ФЗ, свежесть справок, сверка ТЗ и спецификации.
- Несколько заявок за один запуск.
- Подкоманды `list`, `export`.
- Миграции (схема создаётся через `CREATE TABLE IF NOT EXISTS`).
- REST/gRPC API, UI.