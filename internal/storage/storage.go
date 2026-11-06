package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/acheonline/tenderguide/internal/model"
)

// ErrDuplicateRequest возвращается, когда requests.request_id уже есть в БД.
var ErrDuplicateRequest = errors.New("storage: duplicate request_id")

// Store — обёртка над *sql.DB с методами сохранения и чтения.
type Store struct {
	db *sql.DB
}

// Open открывает/создаёт БД и инициализирует схему.
func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	if _, err := db.ExecContext(ctx, schemaDDL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: init schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close закрывает БД.
func (s *Store) Close() error { return s.db.Close() }

// Save пишет model.All в БД в одной транзакции.
func (s *Store) Save(ctx context.Context, a *model.All) (int64, error) {
	if a == nil || a.Request == nil {
		return 0, errors.New("storage: nil request")
	}
	if a.Request.RequestID == "" {
		return 0, errors.New("storage: empty request_id")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("storage: begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var customerID sql.NullInt64
	if a.Customer != nil {
		id, err := upsertCustomer(ctx, tx, a.Customer)
		if err != nil {
			return 0, err
		}
		customerID = sql.NullInt64{Int64: id, Valid: true}
	}

	var participantID sql.NullInt64
	if a.Participant != nil {
		id, err := upsertParticipant(ctx, tx, a.Participant)
		if err != nil {
			return 0, err
		}
		participantID = sql.NullInt64{Int64: id, Valid: true}

		if err := insertBankAccounts(ctx, tx, id, a.BankAccounts); err != nil {
			return 0, err
		}
		if err := insertMSPInfo(ctx, tx, id, a.MSPInfo); err != nil {
			return 0, err
		}
		if err := insertBeneficiaries(ctx, tx, id, a.Beneficiaries); err != nil {
			return 0, err
		}
	}

	createdAt := a.Request.CreatedAt
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO requests (
			request_id, procedure_number, procedure_name, procedure_published_at,
			letter_number, letter_date, valid_until,
			price_with_vat, vat, price_without_vat, price_currency,
			award_experience_amount, customer_id, participant_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Request.RequestID,
		a.Request.ProcedureNumber,
		a.Request.ProcedureName,
		a.Request.ProcedurePublishedAt,
		a.Request.LetterNumber,
		a.Request.LetterDate,
		a.Request.ValidUntil,
		a.Request.PriceWithVAT,
		a.Request.VAT,
		a.Request.PriceWithoutVAT,
		a.Request.PriceCurrency,
		a.Request.AwardExperienceAmount,
		customerID,
		participantID,
		createdAt,
	)
	if err != nil {
		if isUniqueViolation(err, "requests.request_id") {
			return 0, ErrDuplicateRequest
		}
		return 0, fmt.Errorf("storage: insert request: %w", err)
	}

	requestPK, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: last insert id: %w", err)
	}

	if err := insertStaffSummary(ctx, tx, requestPK, a.StaffSummary); err != nil {
		return 0, err
	}
	if err := insertStaffMembers(ctx, tx, requestPK, a.StaffMembers); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("storage: commit: %w", err)
	}
	committed = true
	return requestPK, nil
}

// GetByRequestID возвращает model.All по строковому request_id.
func (s *Store) GetByRequestID(ctx context.Context, requestID string) (*model.All, error) {
	var (
		reqPK                                int64
		procedureNumber, procedureName       sql.NullString
		procedurePublishedAt                 sql.NullString
		letterNumber, letterDate, validUntil sql.NullString
		priceWithVAT, vat, priceWithoutVAT   sql.NullString
		priceCurrency, awardExperienceAmount sql.NullString
		customerID, participantID            sql.NullInt64
		createdAt                            string
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT id, procedure_number, procedure_name, procedure_published_at,
		       letter_number, letter_date, valid_until,
		       price_with_vat, vat, price_without_vat, price_currency,
		       award_experience_amount, customer_id, participant_id, created_at
		FROM requests WHERE request_id = ?`, requestID,
	).Scan(
		&reqPK,
		&procedureNumber, &procedureName, &procedurePublishedAt,
		&letterNumber, &letterDate, &validUntil,
		&priceWithVAT, &vat, &priceWithoutVAT, &priceCurrency,
		&awardExperienceAmount, &customerID, &participantID, &createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("storage: get request: %w", sql.ErrNoRows)
		}
		return nil, fmt.Errorf("storage: get request: %w", err)
	}

	out := &model.All{
		Request: &model.Request{
			ID:                    reqPK,
			RequestID:             requestID,
			ProcedureNumber:       ns(procedureNumber),
			ProcedureName:         ns(procedureName),
			ProcedurePublishedAt:  ns(procedurePublishedAt),
			LetterNumber:          ns(letterNumber),
			LetterDate:            ns(letterDate),
			ValidUntil:            ns(validUntil),
			PriceWithVAT:          ns(priceWithVAT),
			VAT:                   ns(vat),
			PriceWithoutVAT:       ns(priceWithoutVAT),
			PriceCurrency:         ns(priceCurrency),
			AwardExperienceAmount: ns(awardExperienceAmount),
			CustomerID:            ni(customerID),
			ParticipantID:         ni(participantID),
			CreatedAt:             createdAt,
		},
	}

	if customerID.Valid {
		c, err := getCustomer(ctx, s.db, customerID.Int64)
		if err != nil {
			return nil, err
		}
		out.Customer = c
	}
	if participantID.Valid {
		p, err := getParticipant(ctx, s.db, participantID.Int64)
		if err != nil {
			return nil, err
		}
		out.Participant = p
		if out.BankAccounts, err = getBankAccounts(ctx, s.db, participantID.Int64); err != nil {
			return nil, err
		}
		if out.MSPInfo, err = getMSPInfo(ctx, s.db, participantID.Int64); err != nil {
			return nil, err
		}
		if out.Beneficiaries, err = getBeneficiaries(ctx, s.db, participantID.Int64); err != nil {
			return nil, err
		}
	}
	if out.StaffSummary, err = getStaffSummary(ctx, s.db, reqPK); err != nil {
		return nil, err
	}
	if out.StaffMembers, err = getStaffMembers(ctx, s.db, reqPK); err != nil {
		return nil, err
	}
	return out, nil
}

// --- helpers ---

func ns(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func ni(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	i := v.Int64
	return &i
}

func nb(v sql.NullInt64) *bool {
	if !v.Valid {
		return nil
	}
	b := v.Int64 != 0
	return &b
}

func boolToInt(b *bool) any {
	if b == nil {
		return nil
	}
	if *b {
		return 1
	}
	return 0
}

func isUniqueViolation(err error, needle string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") && strings.Contains(msg, needle)
}
