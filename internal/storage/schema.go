package storage

// schemaDDL содержит DDL всех таблиц. Выполняется при Open через CREATE TABLE IF NOT EXISTS.
// Все бизнес-поля nullable. Технический id — единственный неявный NOT NULL.
// NOT NULL стоит только там, где строка без значения бессмысленна:
//   - requests.request_id (из имени директории, есть всегда)
//   - requests.created_at (проставляется кодом)
//   - FK-поля в дочерних таблицах
const schemaDDL = `
CREATE TABLE IF NOT EXISTS customers (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    inn  TEXT UNIQUE,
    name TEXT
);

CREATE TABLE IF NOT EXISTS participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inn TEXT UNIQUE,
    kpp TEXT,
    ogrn TEXT,
    okpo TEXT,
    oktmo TEXT,
    okopf TEXT,
    okopf_name TEXT,
    full_name TEXT,
    short_name TEXT,
    legal_address TEXT,
    postal_address TEXT,
    primary_okved TEXT,
    primary_okved_name TEXT,
    egrul_registration_date TEXT,
    egrul_certificate TEXT,
    egrul_registration_authority TEXT,
    tax_registration_date_current TEXT,
    tax_authority_current TEXT,
    phone TEXT,
    fax TEXT,
    email TEXT,
    website TEXT,
    director_name TEXT,
    director_position TEXT,
    chief_accountant_name TEXT,
    responsible_person_name TEXT,
    responsible_person_phone TEXT,
    has_branches INTEGER
);

CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id TEXT UNIQUE NOT NULL,
    procedure_number TEXT,
    procedure_name TEXT,
    procedure_published_at TEXT,
    letter_number TEXT,
    letter_date TEXT,
    valid_until TEXT,
    price_with_vat TEXT,
    vat TEXT,
    price_without_vat TEXT,
    price_currency TEXT,
    award_experience_amount TEXT,
    customer_id INTEGER REFERENCES customers(id),
    participant_id INTEGER REFERENCES participants(id),
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS bank_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    participant_id INTEGER NOT NULL REFERENCES participants(id),
    bank_name TEXT,
    account TEXT,
    correspondent_account TEXT,
    bik TEXT,
    UNIQUE (participant_id, account)
);

CREATE TABLE IF NOT EXISTS msp_info (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    participant_id INTEGER NOT NULL UNIQUE REFERENCES participants(id),
    is_smp INTEGER,
    smp_category TEXT,
    average_headcount INTEGER,
    revenue_previous_year INTEGER,
    contracts_44_223_count INTEGER,
    contracts_44_223_sum TEXT,
    in_rnp INTEGER
);

CREATE TABLE IF NOT EXISTS staff_summary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER NOT NULL UNIQUE REFERENCES requests(id),
    total_staff INTEGER,
    managers_count INTEGER,
    specialists_count INTEGER,
    workers_count INTEGER
);

CREATE TABLE IF NOT EXISTS staff_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER NOT NULL REFERENCES requests(id),
    category TEXT,
    full_name TEXT,
    position TEXT,
    education TEXT,
    experience_years TEXT,
    UNIQUE (request_id, full_name)
);

CREATE TABLE IF NOT EXISTS beneficiaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    participant_id INTEGER NOT NULL REFERENCES participants(id),
    full_name TEXT,
    inn TEXT,
    passport TEXT,
    registration_address TEXT,
    role TEXT,
    share_gt_10pct INTEGER
);

CREATE TABLE IF NOT EXISTS doc_info (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_id TEXT UNIQUE NOT NULL,
    customer_name TEXT,
    customer_inn TEXT,
    organizer_name TEXT,
    procedure_number TEXT,
    procedure_name TEXT,
    nmck TEXT,
    nmck_without_vat TEXT,
    vat TEXT,
    submission_deadline TEXT,
    review_deadline TEXT,
    bid_security_amount TEXT,
    bid_security_percent TEXT,
    bid_security_type TEXT,
    contract_security_amount TEXT,
    contract_security_required INTEGER,
    price_criterion_weight TEXT,
    qualification_criterion_weight TEXT,
    law_basis TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS doc_requirements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_id TEXT NOT NULL REFERENCES doc_info(doc_id),
    kind TEXT NOT NULL,
    code TEXT,
    name TEXT,
    required INTEGER,
    source_section TEXT,
    UNIQUE (doc_id, kind, code)
);
`
