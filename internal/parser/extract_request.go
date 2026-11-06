package parser

import (
	"regexp"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reProcedureNumber = regexp.MustCompile(`за\s*№\s*(\d{6,})`)
	reLetterNumber    = regexp.MustCompile(`№\s*(\d+)\s+От\s+«`)
	reLetterDate      = regexp.MustCompile(`От\s+«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)
	reValidUntil      = regexp.MustCompile(`действует до\s+«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)
	reProcedureName   = regexp.MustCompile(`предлагает заключить договор на\s+(?:на\s+)?(.+?)(?:\s*,\s*\(|\.\s*$)`)
	rePublishedAt     = regexp.MustCompile(`от\s*«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)
)

func extractRequest(inputs []Input) *model.Request {
	out := &model.Request{}
	for _, in := range inputs {
		for _, p := range in.Paragraphs {
			p = stripGarbage(p)
			if out.ProcedureNumber == nil {
				if m := reProcedureNumber.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.ProcedureNumber = &v
				}
			}
			if out.ProcedureName == nil {
				if m := reProcedureName.FindStringSubmatch(p); m != nil {
					name := strings.TrimSpace(m[1])
					if name != "" {
						out.ProcedureName = &name
					}
				}
			}
			if out.LetterNumber == nil {
				if m := reLetterNumber.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.LetterNumber = &v
				}
			}
			if out.LetterDate == nil {
				if m := reLetterDate.FindStringSubmatch(p); m != nil {
					out.LetterDate = dateFromParts(m[1], m[2], m[3])
				}
			}
			if out.ValidUntil == nil {
				if m := reValidUntil.FindStringSubmatch(p); m != nil {
					out.ValidUntil = dateFromParts(m[1], m[2], m[3])
				}
			}
			if out.ProcedurePublishedAt == nil && out.ProcedureNumber != nil {
				if m := rePublishedAt.FindStringSubmatch(p); m != nil {
					out.ProcedurePublishedAt = dateFromParts(m[1], m[2], m[3])
				}
			}
			if strings.Contains(p, "с учетом НДС") || strings.Contains(p, "с учётом НДС") {
				if out.PriceWithVAT == nil {
					if v := normalizeDecimal(p); v != nil {
						out.PriceWithVAT = v
						cur := "RUB"
						out.PriceCurrency = &cur
					}
				}
			}
			if strings.Contains(p, "сумма НДС") {
				if out.VAT == nil {
					if v := normalizeDecimal(p); v != nil {
						out.VAT = v
					}
				}
			}
			if strings.Contains(p, "без учета НДС") || strings.Contains(p, "без учёта НДС") {
				if out.PriceWithoutVAT == nil {
					if v := normalizeDecimal(p); v != nil {
						out.PriceWithoutVAT = v
					}
				}
			}
		}
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				for _, cell := range row.Cells {
					if strings.Contains(cell, "Опыт участника") && out.AwardExperienceAmount == nil {
						// Берём последнюю ячейку той же строки.
						if len(row.Cells) > 0 {
							last := strings.TrimSpace(row.Cells[len(row.Cells)-1])
							if last != "" {
								out.AwardExperienceAmount = &last
							}
						}
					}
				}
			}
		}
	}
	return out
}
