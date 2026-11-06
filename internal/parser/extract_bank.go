package parser

import (
	"regexp"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reAccount     = regexp.MustCompile(`расчетный счет\s*(\d{20})`)
	reCorrAccount = regexp.MustCompile(`корреспондентский счет\s*(\d{20})`)
	reBIK         = regexp.MustCompile(`БИК банка\s*(\d{9})`)
)

func extractBankAccounts(inputs []Input) []*model.BankAccount {
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				if len(row.Cells) < 2 {
					continue
				}
				found := false
				for _, c := range row.Cells {
					if strings.Contains(c, "Банковские реквизиты") {
						found = true
						break
					}
				}
				if !found {
					continue
				}
				var parts []string
				for _, c := range row.Cells {
					if strings.Contains(c, "Банковские реквизиты") {
						continue
					}
					parts = append(parts, c)
				}
				val := stripGarbage(strings.Join(parts, " "))
				if val == "" {
					continue
				}
				b := &model.BankAccount{}
				if i := strings.Index(val, ","); i > 0 {
					name := strings.TrimSpace(val[:i])
					b.BankName = &name
				}
				if m := reAccount.FindStringSubmatch(val); m != nil {
					v := m[1]
					b.Account = &v
				}
				if m := reCorrAccount.FindStringSubmatch(val); m != nil {
					v := m[1]
					b.CorrespondentAccount = &v
				}
				if m := reBIK.FindStringSubmatch(val); m != nil {
					v := m[1]
					b.BIK = &v
				}
				if b.BankName != nil || b.Account != nil || b.CorrespondentAccount != nil || b.BIK != nil {
					return []*model.BankAccount{b}
				}
			}
		}
	}
	return nil
}
