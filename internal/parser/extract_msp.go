package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reContractsCount = regexp.MustCompile(`\((\d+)\s+договор`)
	reContractsSum   = regexp.MustCompile(`(\d[\d\s]*,\d{2})\s*руб`)
)

func extractMSPInfo(inputs []Input) *model.MSPInfo {
	found := false
	for _, in := range inputs {
		for _, p := range in.Paragraphs {
			if strings.Contains(p, "Декларация о соответствии") && strings.Contains(p, "МСП") {
				found = true
			}
			if strings.Contains(p, "субъектам малого и среднего предпринимательства") {
				found = true
			}
		}
	}
	if !found {
		return nil
	}

	out := &model.MSPInfo{}
	t := true
	out.IsSMP = &t

	for _, in := range inputs {
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				if len(row.Cells) < 3 {
					continue
				}
				// key — Cells[1], value — Cells[2].
				key := strings.TrimSpace(row.Cells[1])
				val := strings.TrimSpace(row.Cells[2])

				switch {
				case strings.Contains(key, "Среднесписочная численность") && out.AverageHeadcount == nil:
					// Значение в Cells[4] (последняя).
					last := strings.TrimSpace(row.Cells[len(row.Cells)-1])
					if n, err := strconv.ParseInt(onlyDigits(last), 10, 64); err == nil && n > 0 {
						out.AverageHeadcount = &n
					}
				case strings.Contains(key, "Доход за предшествующий") && out.RevenuePreviousYear == nil:
					last := strings.TrimSpace(row.Cells[len(row.Cells)-1])
					if m := regexp.MustCompile(`([\d\s]+,\d+)`).FindStringSubmatch(last); m != nil {
						raw := strings.ReplaceAll(strings.ReplaceAll(m[1], " ", ""), ",", ".")
						if f, err := strconv.ParseFloat(raw, 64); err == nil {
							v := int64(f * 1_000_000)
							out.RevenuePreviousYear = &v
						}
					}
				case strings.Contains(key, "Сведения о наличии у юридического лица") && out.Contracts44223Count == nil:
					if m := reContractsCount.FindStringSubmatch(val); m != nil {
						if n, err := strconv.ParseInt(m[1], 10, 64); err == nil {
							out.Contracts44223Count = &n
						}
					}
					if out.Contracts44223Sum == nil {
						valSum := strings.ReplaceAll(val, "\u00a0", " ")
						if m := reContractsSum.FindStringSubmatch(valSum); m != nil {
							if v := normalizeDecimal(m[1]); v != nil {
								out.Contracts44223Sum = v
							}
						}
					}
				case strings.Contains(key, "в реестрах недобросовестных поставщиков") && out.InRNP == nil:
					b := !strings.Contains(strings.ToLower(val), "нет")
					out.InRNP = &b
				}
			}
		}
	}
	out.SMPCategory = calcSMPCategory(out.AverageHeadcount, out.RevenuePreviousYear)
	return out
}

func calcSMPCategory(headcount *int64, revenue *int64) *string {
	if headcount == nil || revenue == nil {
		return nil
	}
	var cat string
	switch {
	case *headcount <= 15 && *revenue <= 120_000_000:
		cat = "micro"
	case *headcount <= 100 && *revenue <= 800_000_000:
		cat = "small"
	default:
		cat = "medium"
	}
	return &cat
}
