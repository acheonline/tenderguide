package parser

import (
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

func extractBeneficiaries(inputs []Input) []*model.Beneficiary {
	var out []*model.Beneficiary
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				// Форма 8: 14 ячеек на строку бенефициара.
				if len(row.Cells) < 13 {
					continue
				}
				// Cells[7] — ИНН бенефициара (12 цифр).
				inn := strings.TrimSpace(row.Cells[7])
				if !isINN12(inn) {
					continue
				}
				b := &model.Beneficiary{}
				v := inn
				b.INN = &v

				// Cells[9] — ФИО бенефициара.
				if fio := strings.TrimSpace(row.Cells[9]); fio != "" {
					b.FullName = &fio
				}
				// Cells[10] — адрес регистрации.
				if addr := strings.TrimSpace(row.Cells[10]); addr != "" {
					b.RegistrationAddress = &addr
				}
				// Cells[11] — паспорт бенефициара.
				if passport := strings.TrimSpace(row.Cells[11]); passport != "" {
					b.Passport = &passport
				}
				// Cells[12] — роль.
				if role := strings.TrimSpace(row.Cells[12]); role != "" {
					b.Role = &role
				}
				// share_gt_10pct = true, если это учредитель (форма 2 п.2).
				share := true
				b.ShareGT10Pct = &share
				out = append(out, b)
			}
		}
	}
	return out
}

func isINN12(s string) bool {
	if len(s) != 12 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
