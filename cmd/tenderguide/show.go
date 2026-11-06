package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/acheonline/tenderguide/internal/model"
	"github.com/acheonline/tenderguide/internal/storage"
)

const nullPlaceholder = "—"

func runShow(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	request := fs.String("request", "", "request_id to show")
	db := fs.String("db", defaultDBPath, "path to SQLite database")
	format := fs.String("format", "text", "output format: text or json")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *request == "" {
		fmt.Fprintln(os.Stderr, "--request is required")
		fmt.Fprintln(os.Stderr, "usage: tenderguide show --request <id> [--db <path>] [--format text|json]")
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(os.Stderr, "invalid format: %s\n", *format)
		return 2
	}

	return show(ctx, *request, *db, *format, os.Stdout, os.Stderr)
}

func show(ctx context.Context, requestID, dbPath, format string, stdout, stderr io.Writer) int {
	dsn := "file:" + dbPath + "?_pragma=foreign_keys(1)"
	store, err := storage.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintf(stderr, "show: %v\n", err)
		return 1
	}
	defer store.Close()

	all, err := store.GetByRequestID(ctx, requestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Fprintf(stderr, "request not found: %s\n", requestID)
			return 1
		}
		fmt.Fprintf(stderr, "show: %v\n", err)
		return 1
	}

	if format == "json" {
		data, err := json.MarshalIndent(all, "", "  ")
		if err != nil {
			fmt.Fprintf(stderr, "show: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(data))
		return 0
	}

	writeText(stdout, all)
	return 0
}

func writeText(w io.Writer, a *model.All) {
	writeRequest(w, a.Request)
	writeCustomer(w, a.Customer)
	writeParticipant(w, a.Participant)
	writeBankAccounts(w, a.BankAccounts)
	writeMSPInfo(w, a.MSPInfo)
	writeStaffSummary(w, a.StaffSummary)
	writeStaffMembers(w, a.StaffMembers)
	writeBeneficiaries(w, a.Beneficiaries)
}

func writeRequest(w io.Writer, r *model.Request) {
	fmt.Fprintln(w, "=== Заявка ===")
	if r == nil {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	field(w, "ID", r.RequestID)
	field(w, "Номер процедуры", str(r.ProcedureNumber))
	field(w, "Предмет", str(r.ProcedureName))
	field(w, "Опубликовано", str(r.ProcedurePublishedAt))
	field(w, "Письмо №", str(r.LetterNumber))
	field(w, "Дата письма", str(r.LetterDate))
	field(w, "Действует до", str(r.ValidUntil))
	field(w, "Цена с НДС", price(r.PriceWithVAT, r.PriceCurrency))
	field(w, "Цена без НДС", price(r.PriceWithoutVAT, r.PriceCurrency))
	field(w, "НДС", str(r.VAT))
	field(w, "Опыт участника", str(r.AwardExperienceAmount))
	field(w, "Создано", r.CreatedAt)
	fmt.Fprintln(w)
}

func writeCustomer(w io.Writer, c *model.Customer) {
	fmt.Fprintln(w, "=== Заказчик ===")
	if c == nil {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	field(w, "ИНН", str(c.INN))
	field(w, "Наименование", str(c.Name))
	fmt.Fprintln(w)
}

func writeParticipant(w io.Writer, p *model.Participant) {
	fmt.Fprintln(w, "=== Участник ===")
	if p == nil {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	field(w, "ИНН", str(p.INN))
	field(w, "КПП", str(p.KPP))
	field(w, "ОГРН", str(p.OGRN))
	field(w, "ОКПО", str(p.OKPO))
	field(w, "ОКТМО", str(p.OKTMO))
	field(w, "ОКОПФ", str(p.OKOPF))
	field(w, "Полное название", str(p.FullName))
	field(w, "Краткое название", str(p.ShortName))
	field(w, "Юр. адрес", str(p.LegalAddress))
	field(w, "Почтовый адрес", str(p.PostalAddress))
	field(w, "Телефон", str(p.Phone))
	field(w, "Email", str(p.Email))
	field(w, "Сайт", str(p.Website))
	field(w, "Руководитель", join(str(p.DirectorName), str(p.DirectorPosition)))
	field(w, "Гл. бухгалтер", str(p.ChiefAccountantName))
	field(w, "Ответственное лицо", join(str(p.ResponsiblePersonName), str(p.ResponsiblePersonPhone)))
	field(w, "Филиалы", boolean(p.HasBranches))
	field(w, "ОГРН дата", str(p.EGRULRegistrationDate))
	field(w, "ОГРН свидетельство", str(p.EGRULCertificate))
	field(w, "ОГРН орган", str(p.EGRULRegistrationAuthority))
	field(w, "Налог. дата", str(p.TaxRegistrationDateCurrent))
	field(w, "ОКВЭД", join(str(p.PrimaryOKVED), str(p.PrimaryOKVEDName)))
	field(w, "ЕГРЮЛ", str(p.TaxAuthorityCurrent))
	fmt.Fprintln(w)
}

func writeBankAccounts(w io.Writer, accounts []*model.BankAccount) {
	fmt.Fprintln(w, "=== Банковские счета ===")
	if len(accounts) == 0 {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	for i, a := range accounts {
		fmt.Fprintf(w, "  #%d\n", i+1)
		indentedField(w, "Банк", str(a.BankName))
		indentedField(w, "Расчётный счёт", str(a.Account))
		indentedField(w, "Корр. счёт", str(a.CorrespondentAccount))
		indentedField(w, "БИК", str(a.BIK))
	}
	fmt.Fprintln(w)
}

func writeMSPInfo(w io.Writer, m *model.MSPInfo) {
	fmt.Fprintln(w, "=== МСП ===")
	if m == nil {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	field(w, "МСП", boolean(m.IsSMP))
	field(w, "Категория", str(m.SMPCategory))
	field(w, "Среднесписочная", intVal(m.AverageHeadcount))
	field(w, "Доход (млн)", intVal(m.RevenuePreviousYear))
	field(w, "Контракты 44/223", intVal(m.Contracts44223Count))
	field(w, "Сумма контрактов", str(m.Contracts44223Sum))
	field(w, "В РНП", boolean(m.InRNP))
	fmt.Fprintln(w)
}

func writeStaffSummary(w io.Writer, s *model.StaffSummary) {
	fmt.Fprintln(w, "=== Штат ===")
	if s == nil {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	field(w, "Всего", intVal(s.TotalStaff))
	field(w, "Руководящее звено", intVal(s.ManagersCount))
	field(w, "Специалисты", intVal(s.SpecialistsCount))
	field(w, "Рабочие", intVal(s.WorkersCount))
	fmt.Fprintln(w)
}

func writeStaffMembers(w io.Writer, members []*model.StaffMember) {
	fmt.Fprintln(w, "=== Сотрудники ===")
	if len(members) == 0 {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	for i, m := range members {
		fmt.Fprintf(w, "  #%d [%s]\n", i+1, str(m.Category))
		indentedField(w, "ФИО", str(m.FullName))
		indentedField(w, "Должность", str(m.Position))
		indentedField(w, "Образование", str(m.Education))
		indentedField(w, "Стаж", str(m.ExperienceYears))
	}
	fmt.Fprintln(w)
}

func writeBeneficiaries(w io.Writer, bs []*model.Beneficiary) {
	fmt.Fprintln(w, "=== Бенефициары ===")
	if len(bs) == 0 {
		fmt.Fprintln(w, "(нет)")
		fmt.Fprintln(w)
		return
	}
	for i, b := range bs {
		fmt.Fprintf(w, "  #%d\n", i+1)
		indentedField(w, "ФИО", str(b.FullName))
		indentedField(w, "ИНН", str(b.INN))
		indentedField(w, "Паспорт", str(b.Passport))
		indentedField(w, "Адрес", str(b.RegistrationAddress))
		indentedField(w, "Роль", str(b.Role))
		indentedField(w, "Доля > 10%", boolean(b.ShareGT10Pct))
	}
	fmt.Fprintln(w)
}

func field(w io.Writer, label, value string) {
	fmt.Fprintf(w, "%-20s%s\n", label+":", value)
}

func indentedField(w io.Writer, label, value string) {
	fmt.Fprintf(w, "    %-16s%s\n", label+":", value)
}

func str(p *string) string {
	if p == nil {
		return nullPlaceholder
	}
	return *p
}

func intVal(p *int64) string {
	if p == nil {
		return nullPlaceholder
	}
	return fmt.Sprintf("%d", *p)
}

func boolean(p *bool) string {
	if p == nil {
		return nullPlaceholder
	}
	if *p {
		return "да"
	}
	return "нет"
}

func price(amount, currency *string) string {
	if amount == nil {
		return nullPlaceholder
	}
	if currency == nil {
		return *amount
	}
	return *amount + " " + *currency
}

func join(a, b string) string {
	if a == nullPlaceholder && b == nullPlaceholder {
		return nullPlaceholder
	}
	if a == nullPlaceholder {
		return b
	}
	if b == nullPlaceholder {
		return a
	}
	return a + " (" + b + ")"
}
