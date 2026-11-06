package parser

import (
	"regexp"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reINN12    = regexp.MustCompile(`\b\d{12}\b`)
	rePassport = regexp.MustCompile(`\d{4}\s*\d{6}`)
	reFIO3     = regexp.MustCompile(`[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+`)
	reRegAddr  = regexp.MustCompile(`\d{6},\s*г\.\s*[А-ЯЁ][а-яё-]+[^,]*,\s*[^,]+,\s*д\.\s*\d+[^,]*`)
)

func extractBeneficiaries(inputs []Input) []*model.Beneficiary {
	var out []*model.Beneficiary
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			// Таблица формы 8: в одной из ячеек есть "Руководитель/Участник" и ИНН12.
			for _, row := range tbl {
				if len(row.Cells) == 0 {
					continue
				}
				joined := strings.Join(row.Cells, " ")
				if !strings.Contains(joined, "Руководитель") && !strings.Contains(joined, "Участник") {
					continue
				}
				if !reINN12.MatchString(joined) {
					continue
				}
				b := &model.Beneficiary{}
				if m := reINN12.FindString(joined); m != "" {
					b.INN = &m
				}
				// ФИО бенефициара — последнее ФИО в строке (руководитель идёт раньше).
				fios := reFIO3.FindAllString(joined, -1)
				if len(fios) > 0 {
					fn := fios[len(fios)-1]
					b.FullName = &fn
				}
				if ms := rePassport.FindAllString(joined, -1); len(ms) > 0 {
					m := ms[len(ms)-1]
					b.Passport = &m
				}
				if m := reRegAddr.FindString(joined); m != "" {
					b.RegistrationAddress = &m
				}
				if strings.Contains(joined, "Руководитель/Участник") {
					r := "Руководитель/Участник"
					b.Role = &r
				}
				share := true
				b.ShareGT10Pct = &share
				out = append(out, b)
			}
		}
	}
	return out
}
