package parser

import (
	"regexp"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var rePassport = regexp.MustCompile(`\d{4}\s*\d{6}`)

func extractBeneficiaries(inputs []Input) []*model.Beneficiary {
	var out []*model.Beneficiary
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				if len(row.Cells) < 3 {
					continue
				}
				// Ищем строку бенефициара: в ячейках есть ФИО и ИНН 12 цифр.
				joined := strings.Join(row.Cells, " ")
				if !regexp.MustCompile(`[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+\s+[А-ЯЁ][а-яё]+`).MatchString(joined) {
					continue
				}
				if !reINN12.MatchString(joined) && !rePassport.MatchString(joined) {
					continue
				}
				b := &model.Beneficiary{}
				if m := reINN12.FindString(joined); m != "" {
					b.INN = &m
				}
				if m := rePassport.FindString(joined); m != "" {
					b.Passport = &m
				}
				// Роль.
				if strings.Contains(joined, "Руководитель") {
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
