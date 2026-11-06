package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

var (
	reSummaryRow = regexp.MustCompile(`^(.+?)\s+(\d+)\s+(\d+)\s*$`)
	reMemberNum  = regexp.MustCompile(`^(\d+)\.\s*(.+)$`)
	reEduStart   = regexp.MustCompile(`(ФГБОУ|МГТУ|Учебный центр|Санкт-Петербургская|Профессиональный лицей|Автономная некоммерческая)`)
)

func extractStaffSummary(inputs []Input) *model.StaffSummary {
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			out := &model.StaffSummary{}
			hasAny := false
			for _, row := range tbl {
				if len(row.Cells) == 0 {
					continue
				}
				text := strings.Join(row.Cells, " ")
				text = strings.Join(strings.Fields(text), " ")
				m := reSummaryRow.FindStringSubmatch(text)
				if m == nil {
					continue
				}
				label := strings.ToLower(m[1])
				n, err := strconv.ParseInt(m[2], 10, 64)
				if err != nil {
					continue
				}
				switch {
				case strings.Contains(label, "итого"):
					out.TotalStaff = &n
					hasAny = true
				case strings.Contains(label, "руководящее звено"):
					out.ManagersCount = &n
					hasAny = true
				case strings.Contains(label, "специалисты"):
					out.SpecialistsCount = &n
					hasAny = true
				case strings.Contains(label, "рабочие"):
					out.WorkersCount = &n
					hasAny = true
				}
			}
			if hasAny {
				return out
			}
		}
	}
	return nil
}

func extractStaffMembers(inputs []Input) []*model.StaffMember {
	var out []*model.StaffMember
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			var category string
			for _, row := range tbl {
				if len(row.Cells) == 0 {
					continue
				}
				text := strings.Join(row.Cells, " ")
				text = strings.Join(strings.Fields(text), " ")
				if text == "" {
					continue
				}
				switch {
				case strings.Contains(text, "Руководящее звено"):
					category = "managers"
					continue
				case strings.Contains(text, "Специалисты"):
					category = "specialists"
					continue
				case strings.Contains(text, "Рабочие"):
					category = "workers"
					continue
				}
				if category == "" {
					continue
				}
				if !reMemberNum.MatchString(text) {
					continue
				}
				m := &model.StaffMember{}
				cat := category
				m.Category = &cat
				// Формат: "1. Дубинович И.В., заместитель … <образование> <стаж>"
				// Стаж — последнее число (или число с запятой) в тексте.
				expRe := regexp.MustCompile(`(\d+(?:[,.]\d+)?)\s*$`)
				if em := expRe.FindStringSubmatch(text); em != nil {
					exp := strings.ReplaceAll(em[1], ",", ".")
					m.ExperienceYears = &exp
					text = strings.TrimSpace(text[:len(text)-len(em[0])])
				}
				// ФИО и должность: до первой запятой после номера.
				body := reMemberNum.ReplaceAllString(text, "$2")
				if i := strings.Index(body, ","); i > 0 {
					fn := strings.TrimSpace(body[:i])
					pos := strings.TrimSpace(body[i+1:])
					if fn != "" {
						m.FullName = &fn
					}
					if loc := reEduStart.FindStringIndex(pos); loc != nil {
						edu := strings.TrimSpace(pos[loc[0]:])
						pos = strings.TrimSpace(pos[:loc[0]])
						if edu != "" {
							m.Education = &edu
						}
					}
					if pos != "" {
						m.Position = &pos
					}
				}
				if m.FullName != nil {
					out = append(out, m)
				}
			}
		}
	}
	return out
}
