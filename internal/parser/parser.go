package parser

import (
	"log"
	"time"

	"github.com/acheonline/tenderguide/internal/model"
)

// Input — то, что отдаёт loader для одного DOCX-файла.
type Input struct {
	FileName   string
	Paragraphs []string
	Tables     [][]Row
}

// Row — одна строка таблицы.
type Row struct {
	Cells []string
}

// Extract обрабатывает все Input (весь комплект DOCX в директории)
// и возвращает собранную модель. Никогда не возвращает nil.
func Extract(requestID string, inputs []Input) *model.All {
	out := &model.All{
		Request: &model.Request{
			RequestID: requestID,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	req := extractRequest(inputs)
	cust := extractCustomer(req)
	part := extractParticipant(inputs)
	banks := extractBankAccounts(inputs)
	msp := extractMSPInfo(inputs)
	summary := extractStaffSummary(inputs)
	members := extractStaffMembers(inputs)
	bens := extractBeneficiaries(inputs)

	if req != nil {
		if req.ProcedureNumber != nil {
			out.Request.ProcedureNumber = req.ProcedureNumber
		}
		if req.ProcedureName != nil {
			out.Request.ProcedureName = req.ProcedureName
		}
		if req.ProcedurePublishedAt != nil {
			out.Request.ProcedurePublishedAt = req.ProcedurePublishedAt
		}
		if req.LetterNumber != nil {
			out.Request.LetterNumber = req.LetterNumber
		}
		if req.LetterDate != nil {
			out.Request.LetterDate = req.LetterDate
		}
		if req.ValidUntil != nil {
			out.Request.ValidUntil = req.ValidUntil
		}
		if req.PriceWithVAT != nil {
			out.Request.PriceWithVAT = req.PriceWithVAT
		}
		if req.VAT != nil {
			out.Request.VAT = req.VAT
		}
		if req.PriceWithoutVAT != nil {
			out.Request.PriceWithoutVAT = req.PriceWithoutVAT
		}
		if req.PriceCurrency != nil {
			out.Request.PriceCurrency = req.PriceCurrency
		}
		if req.AwardExperienceAmount != nil {
			out.Request.AwardExperienceAmount = req.AwardExperienceAmount
		}
	}
	if cust != nil {
		out.Customer = cust
	}
	if part != nil {
		out.Participant = part
	}
	out.BankAccounts = banks
	out.MSPInfo = msp
	out.StaffSummary = summary
	out.StaffMembers = members
	out.Beneficiaries = bens

	log.Printf("parser: extract done for %q: participant=%v, customer=%v, procedure_number=%v",
		requestID, part != nil, cust != nil, out.Request.ProcedureNumber)
	return out
}

// firstNonEmpty возвращает первый не-nil указатель.
func firstNonEmpty(vals ...*string) *string {
	for _, v := range vals {
		if v != nil && *v != "" {
			return v
		}
	}
	return nil
}
