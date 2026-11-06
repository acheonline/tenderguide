package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/acheonline/tenderguide/internal/model"
)

// ExtractDoc обрабатывает все Input (весь комплект DOCX документации)
// и возвращает собранную модель. Никогда не возвращает nil.
func ExtractDoc(docID string, inputs []Input) *model.DocAll {
	out := &model.DocAll{
		Info: &model.DocInfo{
			DocID:     docID,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	reRegNumber := regexp.MustCompile(`regNumber=(\d{6,})`)

	for _, in := range inputs {
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				extractDocInfoRow(out.Info, row)
				joined := strings.Join(row.Cells, " ")
				extractDocRequirementRow(out, stripGarbage(joined))
			}
		}
		for _, raw := range in.Paragraphs {
			if out.Info.ProcedureNumber == nil {
				if m := reDocProcedureNumber.FindStringSubmatch(raw); m != nil {
					v := m[1]
					out.Info.ProcedureNumber = &v
				} else if m := reRegNumber.FindStringSubmatch(raw); m != nil {
					v := m[1]
					out.Info.ProcedureNumber = &v
				}
			}
			p := stripGarbage(raw)
			if out.Info.CustomerINN == nil {
				if m := reDocINN.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.Info.CustomerINN = &v
				}
			}
			extractDocRequirementRow(out, p)
		}
	}

	return out
}

var (
	reDocCustomerName    = regexp.MustCompile(`(?:АО|ПАО|ООО|ЗАО)\s*«[^»]+»`)
	reDocOrganizerName   = regexp.MustCompile(`Организатор:\s*([^.]+?)(?:\.|$)`)
	reDocProcedureName   = regexp.MustCompile(`Предмет Конкурса:\s*(.+?)(?:\n|$)`)
	reDocNMCK            = regexp.MustCompile(`(\d[\d\s]*,\d{2})\s*руб`)
	reDocVAT             = regexp.MustCompile(`НДС[^:]*:\s*(\d[\d\s]*,\d{2})\s*руб`)
	reDocSubmissionDL    = regexp.MustCompile(`до\s*(\d{2}\.\d{2}\.\d{4})[^\d]*(\d{2}:\d{2})`)
	reDocReviewDL        = regexp.MustCompile(`(\d{2}\.\d{2}\.\d{4})[^\d]*(\d{2}:\d{2})`)
	reDocBidSecurityAmt  = regexp.MustCompile(`составляет:\s*(\d[\d\s]*,\d{2})\s*руб`)
	reDocBidSecurityPct  = regexp.MustCompile(`(\d+)\s*%`)
	reDocLawBasis        = regexp.MustCompile(`(44-ФЗ|223-ФЗ)`)
	reDocCriterionWeight = regexp.MustCompile(`(\d+)%`)
	reDocProcedureNumber = regexp.MustCompile(`№\s*(\d{6,})`)
	reDocINN             = regexp.MustCompile(`ИНН\s*(\d{10})`)
)

func extractDocInfoRow(info *model.DocInfo, row Row) {
	if len(row.Cells) == 0 {
		return
	}
	joined := strings.Join(row.Cells, " ")
	joinedLower := strings.ToLower(joined)

	if info.PriceCriterionWeight == nil && strings.Contains(joinedLower, "цена договора") {
		if m := reDocCriterionWeight.FindStringSubmatch(joined); m != nil {
			v := m[1]
			info.PriceCriterionWeight = &v
		}
	}
	if info.QualificationCriterionWeight == nil && strings.Contains(joinedLower, "квалификация участников") {
		if m := reDocCriterionWeight.FindStringSubmatch(joined); m != nil {
			v := m[1]
			info.QualificationCriterionWeight = &v
		}
	}

	var text string
	if len(row.Cells) >= 2 {
		text = strings.TrimSpace(row.Cells[len(row.Cells)-1])
	} else {
		text = strings.TrimSpace(row.Cells[0])
	}
	text = stripGarbage(text)
	if text == "" {
		return
	}
	lower := strings.ToLower(text)

	if info.CustomerName == nil && strings.Contains(lower, "заказчик") {
		if m := reDocCustomerName.FindString(text); m != "" {
			info.CustomerName = &m
		}
	}
	if info.OrganizerName == nil && strings.Contains(lower, "организатор") {
		if m := reDocOrganizerName.FindStringSubmatch(text); m != nil {
			v := strings.TrimSpace(m[1])
			info.OrganizerName = &v
		}
	}
	if info.ProcedureName == nil && strings.Contains(lower, "предмет конкурса") {
		if m := reDocProcedureName.FindStringSubmatch(text); m != nil {
			v := strings.TrimSpace(m[1])
			info.ProcedureName = &v
		}
	}
	if info.NMCK == nil && strings.Contains(lower, "начальная (максимальная) цена") {
		if m := reDocNMCK.FindStringSubmatch(text); m != nil {
			if v := normalizeDecimal(m[1]); v != nil {
				info.NMCK = v
			}
		}
	}
	if info.VAT == nil && strings.Contains(lower, "ндс") {
		if m := reDocVAT.FindStringSubmatch(text); m != nil {
			if v := normalizeDecimal(m[1]); v != nil {
				info.VAT = v
			}
		}
	}
	if info.SubmissionDeadline == nil && strings.Contains(lower, "дата окончания срока подачи") {
		if m := reDocSubmissionDL.FindStringSubmatch(text); m != nil {
			if v := dateTimeToUTC(m[1], m[2]); v != nil {
				info.SubmissionDeadline = v
			}
		}
	}
	if info.ReviewDeadline == nil && strings.Contains(lower, "дата рассмотрения") {
		if m := reDocReviewDL.FindStringSubmatch(text); m != nil {
			if v := dateTimeToUTC(m[1], m[2]); v != nil {
				info.ReviewDeadline = v
			}
		}
	}
	if info.BidSecurityAmount == nil && strings.Contains(lower, "размер обеспечения заявки") {
		if m := reDocBidSecurityAmt.FindStringSubmatch(text); m != nil {
			if v := normalizeDecimal(m[1]); v != nil {
				info.BidSecurityAmount = v
			}
		}
		if m := reDocBidSecurityPct.FindStringSubmatch(text); m != nil {
			v := m[1]
			info.BidSecurityPercent = &v
		}
		if strings.Contains(lower, "денежных средств") && strings.Contains(lower, "банковской гарантии") {
			v := "cash_or_bank_guarantee"
			info.BidSecurityType = &v
		} else if strings.Contains(lower, "денежных средств") {
			v := "cash"
			info.BidSecurityType = &v
		} else if strings.Contains(lower, "банковской гарантии") {
			v := "bank_guarantee"
			info.BidSecurityType = &v
		}
	}
	if info.ContractSecurityRequired == nil && strings.Contains(lower, "требование к обеспечению договора") {
		b := !strings.Contains(lower, "не установлено")
		info.ContractSecurityRequired = &b
	}
	if info.LawBasis == nil {
		if m := reDocLawBasis.FindStringSubmatch(text); m != nil {
			v := m[1]
			info.LawBasis = &v
		}
	}
}

var (
	reReqSection6 = regexp.MustCompile(`(6\.\d+(?:\.\d+)?)\.\s+(.+)`)
	reReqSection7 = regexp.MustCompile(`(7\.1\.\d+)\.\s+(.+)`)
	reDocForm     = regexp.MustCompile(`•\s*форма\s*(\d+)\s*[–-]\s*«([^»]+)»(\s*\(([^)]+)\))?;?`)
)

func extractDocRequirementRow(out *model.DocAll, p string) {
	if m := reReqSection6.FindStringSubmatch(p); m != nil {
		code := m[1]
		name := strings.TrimSpace(m[2])
		req := true
		sec := "6.1"
		out.Requirements = append(out.Requirements, &model.DocRequirement{
			DocID: out.Info.DocID, Kind: "requirement",
			Code: &code, Name: &name, Required: &req, SourceSection: &sec,
		})
		return
	}
	if m := reReqSection7.FindStringSubmatch(p); m != nil {
		code := m[1]
		name := strings.TrimSpace(m[2])
		req := true
		sec := "7.1"
		out.Requirements = append(out.Requirements, &model.DocRequirement{
			DocID: out.Info.DocID, Kind: "document",
			Code: &code, Name: &name, Required: &req, SourceSection: &sec,
		})
		return
	}
	if m := reDocForm.FindStringSubmatch(p); m != nil {
		code := "форма " + m[1]
		name := strings.TrimSpace(m[2])
		req := true
		if m[4] != "" && strings.Contains(strings.ToLower(m[4]), "не предоставление не является основанием") {
			req = false
		}
		sec := "8.1.1"
		out.Requirements = append(out.Requirements, &model.DocRequirement{
			DocID: out.Info.DocID, Kind: "document",
			Code: &code, Name: &name, Required: &req, SourceSection: &sec,
		})
	}
}

func dateTimeToUTC(date, timeStr string) *string {
	d := normalizeDotDate(date)
	if d == nil {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04", *d+" "+timeStr)
	if err != nil {
		return nil
	}
	utc := t.Add(-3 * time.Hour).UTC().Format(time.RFC3339)
	return &utc
}
