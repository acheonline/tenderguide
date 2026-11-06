package parser

import (
	"regexp"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reProcedureNumber  = regexp.MustCompile(`за\s*№\s*(\d{6,})`)
	reProcedureFromURL = regexp.MustCompile(`regNumber=(\d{6,})`)
	reLetterNumberLine = regexp.MustCompile(`^№\s*(\d+)\s*$`)
	reLetterNumberFull = regexp.MustCompile(`№\s*(\d+)\s+От\s+«`)
	reLetterDate       = regexp.MustCompile(`От\s+«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)
	reValidUntil       = regexp.MustCompile(`действует до\s*«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)
	rePublishedAt      = regexp.MustCompile(`от\s*«(\d{1,2})»\s+([а-я]+)\s+(\d{4})`)
	reProcNameStart    = regexp.MustCompile(`предлагает заключить договор на\s*$`)
	reProcNameNext     = regexp.MustCompile(`^(?:на\s+)?(.+?)\s*,\s*$`)
	reNumberLine       = regexp.MustCompile(`^\d[\d\s]*,\d{2}`)
)

func extractRequest(inputs []Input) *model.Request {
	out := &model.Request{}
	for _, in := range inputs {
		paras := make([]string, len(in.Paragraphs))
		for i, raw := range in.Paragraphs {
			paras[i] = stripGarbage(raw)
		}

		var letterNumberCandidate *string

		for i, p := range paras {
			if out.ProcedureNumber == nil {
				if m := reProcedureFromURL.FindStringSubmatch(in.Paragraphs[i]); m != nil {
					v := m[1]
					out.ProcedureNumber = &v
				} else if m := reProcedureNumber.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.ProcedureNumber = &v
				}
			}

			// letter_number: "№197" отдельной строкой, "От «01» ..." — следующей.
			if out.LetterNumber == nil {
				if m := reLetterNumberFull.FindStringSubmatch(p); m != nil {
					v := m[1]
					out.LetterNumber = &v
				} else if m := reLetterNumberLine.FindStringSubmatch(p); m != nil {
					v := m[1]
					letterNumberCandidate = &v
				} else if letterNumberCandidate != nil && reLetterDate.MatchString(p) {
					out.LetterNumber = letterNumberCandidate
					letterNumberCandidate = nil
				}
			}

			// procedure_name: "предлагает заключить договор на" + следующий абзац.
			if out.ProcedureName == nil && reProcNameStart.MatchString(p) && i+1 < len(paras) {
				next := paras[i+1]
				if m := reProcNameNext.FindStringSubmatch(next); m != nil {
					name := strings.TrimSpace(m[1])
					if name != "" {
						out.ProcedureName = &name
					}
				} else {
					name := strings.TrimSpace(strings.TrimRight(next, ",."))
					if name != "" {
						out.ProcedureName = &name
					}
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
		}

		// Цены: число и маркер могут быть в соседних абзацах.
		for i, p := range paras {
			next := ""
			if i+1 < len(paras) {
				next = paras[i+1]
			}
			window := p + " " + next

			if out.PriceWithVAT == nil &&
				(strings.Contains(window, "с учетом НДС") || strings.Contains(window, "с учётом НДС")) &&
				reNumberLine.MatchString(p) {
				if v := normalizeDecimal(p); v != nil {
					out.PriceWithVAT = v
					cur := "RUB"
					out.PriceCurrency = &cur
				}
			}
			if out.VAT == nil && strings.Contains(window, "сумма НДС") && reNumberLine.MatchString(p) {
				if v := normalizeDecimal(p); v != nil {
					out.VAT = v
				}
			}
			if out.PriceWithoutVAT == nil &&
				(strings.Contains(window, "без учета НДС") || strings.Contains(window, "без учёта НДС")) &&
				reNumberLine.MatchString(p) {
				if v := normalizeDecimal(p); v != nil {
					out.PriceWithoutVAT = v
				}
			}
		}

		for _, tbl := range in.Tables {
			for _, row := range tbl {
				for _, cell := range row.Cells {
					if strings.Contains(cell, "Опыт участника") && out.AwardExperienceAmount == nil {
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
