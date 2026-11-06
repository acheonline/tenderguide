package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/acheonline/tenderguide/internal/model"
)

func upsertCustomer(ctx context.Context, tx *sql.Tx, c *model.Customer) (int64, error) {
	if c.INN != nil && *c.INN != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM customers WHERE inn = ?`, *c.INN).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errorsIsNoRows(err) {
			return 0, fmt.Errorf("storage: select customer: %w", err)
		}
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO customers (inn, name) VALUES (?, ?)`, c.INN, c.Name)
	if err != nil {
		return 0, fmt.Errorf("storage: insert customer: %w", err)
	}
	return res.LastInsertId()
}

func upsertParticipant(ctx context.Context, tx *sql.Tx, p *model.Participant) (int64, error) {
	if p.INN != nil && *p.INN != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM participants WHERE inn = ?`, *p.INN).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errorsIsNoRows(err) {
			return 0, fmt.Errorf("storage: select participant: %w", err)
		}
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO participants (
			inn, kpp, ogrn, okpo, oktmo, okopf, okopf_name,
			full_name, short_name, legal_address, postal_address,
			primary_okved, primary_okved_name,
			egrul_registration_date, egrul_certificate, egrul_registration_authority,
			tax_registration_date_current, tax_authority_current,
			phone, fax, email, website,
			director_name, director_position, chief_accountant_name,
			responsible_person_name, responsible_person_phone, has_branches
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.INN, p.KPP, p.OGRN, p.OKPO, p.OKTMO, p.OKOPF, p.OKOPFName,
		p.FullName, p.ShortName, p.LegalAddress, p.PostalAddress,
		p.PrimaryOKVED, p.PrimaryOKVEDName,
		p.EGRULRegistrationDate, p.EGRULCertificate, p.EGRULRegistrationAuthority,
		p.TaxRegistrationDateCurrent, p.TaxAuthorityCurrent,
		p.Phone, p.Fax, p.Email, p.Website,
		p.DirectorName, p.DirectorPosition, p.ChiefAccountantName,
		p.ResponsiblePersonName, p.ResponsiblePersonPhone, boolToInt(p.HasBranches),
	)
	if err != nil {
		return 0, fmt.Errorf("storage: insert participant: %w", err)
	}
	return res.LastInsertId()
}

func insertBankAccounts(ctx context.Context, tx *sql.Tx, pid int64, items []*model.BankAccount) error {
	for _, b := range items {
		if b == nil {
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO bank_accounts (participant_id, bank_name, account, correspondent_account, bik)
			VALUES (?, ?, ?, ?, ?)`,
			pid, b.BankName, b.Account, b.CorrespondentAccount, b.BIK,
		)
		if err != nil {
			// Конфликт по UNIQUE(participant_id, account) — пропускаем.
			if isUniqueViolation(err, "bank_accounts") {
				continue
			}
			return fmt.Errorf("storage: insert bank_account: %w", err)
		}
	}
	return nil
}

func insertMSPInfo(ctx context.Context, tx *sql.Tx, pid int64, m *model.MSPInfo) error {
	if m == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO msp_info (
			participant_id, is_smp, smp_category, average_headcount,
			revenue_previous_year, contracts_44_223_count, contracts_44_223_sum, in_rnp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		pid, boolToInt(m.IsSMP), m.SMPCategory, m.AverageHeadcount,
		m.RevenuePreviousYear, m.Contracts44223Count, m.Contracts44223Sum, boolToInt(m.InRNP),
	)
	if err != nil {
		if isUniqueViolation(err, "msp_info") {
			return nil
		}
		return fmt.Errorf("storage: insert msp_info: %w", err)
	}
	return nil
}

func insertStaffSummary(ctx context.Context, tx *sql.Tx, reqID int64, s *model.StaffSummary) error {
	if s == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO staff_summary (request_id, total_staff, managers_count, specialists_count, workers_count)
		VALUES (?, ?, ?, ?, ?)`,
		reqID, s.TotalStaff, s.ManagersCount, s.SpecialistsCount, s.WorkersCount,
	)
	if err != nil {
		return fmt.Errorf("storage: insert staff_summary: %w", err)
	}
	return nil
}

func insertStaffMembers(ctx context.Context, tx *sql.Tx, reqID int64, items []*model.StaffMember) error {
	for _, m := range items {
		if m == nil {
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO staff_members (request_id, category, full_name, position, education, experience_years)
			VALUES (?, ?, ?, ?, ?, ?)`,
			reqID, m.Category, m.FullName, m.Position, m.Education, m.ExperienceYears,
		)
		if err != nil {
			if isUniqueViolation(err, "staff_members") {
				continue
			}
			return fmt.Errorf("storage: insert staff_member: %w", err)
		}
	}
	return nil
}

func insertBeneficiaries(ctx context.Context, tx *sql.Tx, pid int64, items []*model.Beneficiary) error {
	for _, b := range items {
		if b == nil {
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO beneficiaries (participant_id, full_name, inn, passport, registration_address, role, share_gt_10pct)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			pid, b.FullName, b.INN, b.Passport, b.RegistrationAddress, b.Role, boolToInt(b.ShareGT10Pct),
		)
		if err != nil {
			return fmt.Errorf("storage: insert beneficiary: %w", err)
		}
	}
	return nil
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows || (err != nil && err.Error() == "sql: no rows in result set")
}
