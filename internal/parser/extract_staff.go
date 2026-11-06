package parser

import (
	"strconv"
	"strings"

	"github.com/acheonline/tenderguide/internal/model"
)

func extractStaffSummary(inputs []Input) *model.StaffSummary {
	for _, in := range inputs {
		for _, tbl := range in.Tables {
			out := &model.StaffSummary{}
			hasAny := false
			for _, row := range tbl {
				if len(row.Cells) < 2 {
					continue
				}
				joined := strings.Join(row.Cells, " ")
				last := strings.TrimSpace(row.Cells[len(row.Cells)-1])
				n, err := strconv.ParseInt(onlyDigits(last), 10, 64)
				if err != nil {
					continue
				}
				switch {
				case strings.Contains(joined, "ИТОГО"):
					out.TotalStaff = &n
					hasAny = true
				case strings.Contains(joined, "Руководящее звено"):
					out.ManagersCount = &n
					hasAny = true
				case strings.Contains(joined, "Специалисты"):
					out.SpecialistsCount = &n
					hasAny = true
				case strings.Contains(joined, "Рабочие"):
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
				if len(row.Cells) < 1 {
					continue
				}
				if len(row.Cells) == 1 {
					joined := row.Cells[0]
					switch {
					case strings.Contains(joined, "Руководящее звено"):
						category = "managers"
					case strings.Contains(joined, "Специалисты"):
						category = "specialists"
					case strings.Contains(joined, "Рабочие"):
						category = "workers"
					}
					continue
				}
				if category == "" {
					continue
				}
				fullNamePos := row.Cells[1]
				if fullNamePos == "" || !strings.Contains(fullNamePos, ".") {
					continue
				}
				cat := category
				m := &model.StaffMember{Category: &cat}

				name := fullNamePos
				pos := ""
				if i := strings.Index(fullNamePos, ","); i > 0 {
					name = strings.TrimSpace(fullNamePos[:i])
					pos = strings.TrimSpace(fullNamePos[i+1:])
				}
				if name != "" {
					m.FullName = &name
				}
				if pos != "" {
					m.Position = &pos
				}
				if len(row.Cells) > 2 && row.Cells[2] != "" {
					edu := strings.TrimSpace(row.Cells[2])
					m.Education = &edu
				}
				if len(row.Cells) > 3 && row.Cells[3] != "" {
					exp := strings.ReplaceAll(strings.TrimSpace(row.Cells[3]), ",", ".")
					m.ExperienceYears = &exp
				}
				out = append(out, m)
			}
		}
	}
	return out
}
