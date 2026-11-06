package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reContractsCount = regexp.MustCompile(`да\s*\((\d+)\s+договор`)
	reContractsSum   = regexp.MustCompile(`(\d[\d\s]*,\d{2})\s*руб`)
	reMillionValue   = regexp.MustCompile(`([\d\s]+,\d+)`)
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
		for _, p := range in.Paragraphs {
			p = stripGarbage(p)
			if out.Contracts44223Count == nil {
				if m := reContractsCount.FindStringSubmatch(p); m != nil {
					if n, err := strconv.ParseInt(m[1], 10, 64); err == nil {
						out.Contracts44223Count = &n
					}
				}
			}
			if out.Contracts44223Sum == nil {
				if m := reContractsSum.FindStringSubmatch(p); m != nil {
					if v := normalizeDecimal(m[1]); v != nil {
						out.Contracts44223Sum = v
					}
				}
			}
			if out.InRNP == nil && strings.Contains(p, "Реестр недобросовестных") {
				b := !strings.Contains(strings.ToLower(p), "нет")
				out.InRNP = &b
			}
		}
		for _, tbl := range in.Tables {
			for _, row := range tbl {
				if len(row.Cells) == 0 {
					continue
				}
				joined := strings.Join(row.Cells, " ")
				if out.InRNP == nil && strings.Contains(joined, "Реестр недобросовестных") {
					b := !strings.Contains(strings.ToLower(joined), "нет")
					out.InRNP = &b
				}
				if len(row.Cells) < 2 {
					continue
				}
				last := strings.TrimSpace(row.Cells[len(row.Cells)-1])
				if strings.Contains(joined, "Среднесписочная численность") && out.AverageHeadcount == nil {
					if n, err := strconv.ParseInt(onlyDigits(last), 10, 64); err == nil && n > 0 {
						out.AverageHeadcount = &n
					}
				}
				if strings.Contains(joined, "Доход за предшествующий") && out.RevenuePreviousYear == nil {
					if m := reMillionValue.FindStringSubmatch(last); m != nil {
						raw := strings.ReplaceAll(strings.ReplaceAll(m[1], " ", ""), ",", ".")
						if f, err := strconv.ParseFloat(raw, 64); err == nil {
							v := int64(f * 1_000_000)
							out.RevenuePreviousYear = &v
						}
					}
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
