package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/acheonline/tenderguide/internal/model"
)

func getCustomer(ctx context.Context, db *sql.DB, id int64) (*model.Customer, error) {
	var (
		inn, name sql.NullString
	)
	err := db.QueryRowContext(ctx, `SELECT inn, name FROM customers WHERE id = ?`, id).Scan(&inn, &name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get customer: %w", err)
	}
	return &model.Customer{ID: id, INN: ns(inn), Name: ns(name)}, nil
}

func getParticipant(ctx context.Context, db *sql.DB, id int64) (*model.Participant, error) {
	var (
		inn, kpp, ogrn, okpo, oktmo, okopf, okopfName sql.NullString
		fullName, shortName, legalAddr, postalAddr    sql.NullString
		primaryOKVED, primaryOKVEDName                sql.NullString
		egrulDate, egrulCert, egrulAuth               sql.NullString
		taxDate, taxAuth                              sql.NullString
		phone, fax, email, website                    sql.NullString
		directorName, directorPos, chiefAcc           sql.NullString
		respName, respPhone                           sql.NullString
		hasBranches                                   sql.NullInt64
	)
	err := db.QueryRowContext(ctx, `
		SELECT inn, kpp, ogrn, okpo, oktmo, okopf, okopf_name,
		       full_name, short_name, legal_address, postal_address,
		       primary_okved, primary_okved_name,
		       egrul_registration_date, egrul_certificate, egrul_registration_authority,
		       tax_registration_date_current, tax_authority_current,
		       phone, fax, email, website,
		       director_name, director_position, chief_accountant_name,
		       responsible_person_name, responsible_person_phone, has_branches
		FROM participants WHERE id = ?`, id,
	).Scan(
		&inn, &kpp, &ogrn, &okpo, &oktmo, &okopf, &okopfName,
		&fullName, &shortName, &legalAddr, &postalAddr,
		&primaryOKVED, &primaryOKVEDName,
		&egrulDate, &egrulCert, &egrulAuth,
		&taxDate, &taxAuth,
		&phone, &fax, &email, &website,
		&directorName, &directorPos, &chiefAcc,
		&respName, &respPhone, &hasBranches,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get participant: %w", err)
	}
	return &model.Participant{
		ID:                         id,
		INN:                        ns(inn),
		KPP:                        ns(kpp),
		OGRN:                       ns(ogrn),
		OKPO:                       ns(okpo),
		OKTMO:                      ns(oktmo),
		OKOPF:                      ns(okopf),
		OKOPFName:                  ns(okopfName),
		FullName:                   ns(fullName),
		ShortName:                  ns(shortName),
		LegalAddress:               ns(legalAddr),
		PostalAddress:              ns(postalAddr),
		PrimaryOKVED:               ns(primaryOKVED),
		PrimaryOKVEDName:           ns(primaryOKVEDName),
		EGRULRegistrationDate:      ns(egrulDate),
		EGRULCertificate:           ns(egrulCert),
		EGRULRegistrationAuthority: ns(egrulAuth),
		TaxRegistrationDateCurrent: ns(taxDate),
		TaxAuthorityCurrent:        ns(taxAuth),
		Phone:                      ns(phone),
		Fax:                        ns(fax),
		Email:                      ns(email),
		Website:                    ns(website),
		DirectorName:               ns(directorName),
		DirectorPosition:           ns(directorPos),
		ChiefAccountantName:        ns(chiefAcc),
		ResponsiblePersonName:      ns(respName),
		ResponsiblePersonPhone:     ns(respPhone),
		HasBranches:                nb(hasBranches),
	}, nil
}

func getBankAccounts(ctx context.Context, db *sql.DB, pid int64) ([]*model.BankAccount, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, bank_name, account, correspondent_account, bik
		FROM bank_accounts WHERE participant_id = ?`, pid)
	if err != nil {
		return nil, fmt.Errorf("storage: get bank_accounts: %w", err)
	}
	defer rows.Close()
	var out []*model.BankAccount
	for rows.Next() {
		var (
			id                       int64
			bankName, acc, corr, bik sql.NullString
		)
		if err := rows.Scan(&id, &bankName, &acc, &corr, &bik); err != nil {
			return nil, err
		}
		out = append(out, &model.BankAccount{
			ID:                   id,
			ParticipantID:        pid,
			BankName:             ns(bankName),
			Account:              ns(acc),
			CorrespondentAccount: ns(corr),
			BIK:                  ns(bik),
		})
	}
	return out, rows.Err()
}

func getMSPInfo(ctx context.Context, db *sql.DB, pid int64) (*model.MSPInfo, error) {
	var (
		id                         int64
		isSMP, revenue, cnt, inRNP sql.NullInt64
		category, sum              sql.NullString
		avgHeadcount               sql.NullInt64
	)
	err := db.QueryRowContext(ctx, `
		SELECT id, is_smp, smp_category, average_headcount, revenue_previous_year,
		       contracts_44_223_count, contracts_44_223_sum, in_rnp
		FROM msp_info WHERE participant_id = ?`, pid,
	).Scan(&id, &isSMP, &category, &avgHeadcount, &revenue, &cnt, &sum, &inRNP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get msp_info: %w", err)
	}
	return &model.MSPInfo{
		ID:                  id,
		ParticipantID:       pid,
		IsSMP:               nb(isSMP),
		SMPCategory:         ns(category),
		AverageHeadcount:    ni(avgHeadcount),
		RevenuePreviousYear: ni(revenue),
		Contracts44223Count: ni(cnt),
		Contracts44223Sum:   ns(sum),
		InRNP:               nb(inRNP),
	}, nil
}

func getStaffSummary(ctx context.Context, db *sql.DB, reqID int64) (*model.StaffSummary, error) {
	var (
		id                                int64
		total, managers, specialists, wrk sql.NullInt64
	)
	err := db.QueryRowContext(ctx, `
		SELECT id, total_staff, managers_count, specialists_count, workers_count
		FROM staff_summary WHERE request_id = ?`, reqID,
	).Scan(&id, &total, &managers, &specialists, &wrk)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: get staff_summary: %w", err)
	}
	return &model.StaffSummary{
		ID:               id,
		RequestID:        reqID,
		TotalStaff:       ni(total),
		ManagersCount:    ni(managers),
		SpecialistsCount: ni(specialists),
		WorkersCount:     ni(wrk),
	}, nil
}

func getStaffMembers(ctx context.Context, db *sql.DB, reqID int64) ([]*model.StaffMember, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, category, full_name, position, education, experience_years
		FROM staff_members WHERE request_id = ?`, reqID)
	if err != nil {
		return nil, fmt.Errorf("storage: get staff_members: %w", err)
	}
	defer rows.Close()
	var out []*model.StaffMember
	for rows.Next() {
		var (
			id                     int64
			cat, fn, pos, edu, exp sql.NullString
		)
		if err := rows.Scan(&id, &cat, &fn, &pos, &edu, &exp); err != nil {
			return nil, err
		}
		out = append(out, &model.StaffMember{
			ID:              id,
			RequestID:       reqID,
			Category:        ns(cat),
			FullName:        ns(fn),
			Position:        ns(pos),
			Education:       ns(edu),
			ExperienceYears: ns(exp),
		})
	}
	return out, rows.Err()
}

func getBeneficiaries(ctx context.Context, db *sql.DB, pid int64) ([]*model.Beneficiary, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, full_name, inn, passport, registration_address, role, share_gt_10pct
		FROM beneficiaries WHERE participant_id = ?`, pid)
	if err != nil {
		return nil, fmt.Errorf("storage: get beneficiaries: %w", err)
	}
	defer rows.Close()
	var out []*model.Beneficiary
	for rows.Next() {
		var (
			id                            int64
			fn, inn, passport, addr, role sql.NullString
			share                         sql.NullInt64
		)
		if err := rows.Scan(&id, &fn, &inn, &passport, &addr, &role, &share); err != nil {
			return nil, err
		}
		out = append(out, &model.Beneficiary{
			ID:                  id,
			ParticipantID:       pid,
			FullName:            ns(fn),
			INN:                 ns(inn),
			Passport:            ns(passport),
			RegistrationAddress: ns(addr),
			Role:                ns(role),
			ShareGT10Pct:        nb(share),
		})
	}
	return out, rows.Err()
}
