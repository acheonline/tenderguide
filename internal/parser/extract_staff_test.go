package parser

import "testing"

func TestExtractStaffSummary(t *testing.T) {
	in := []Input{
		{
			Tables: [][]Row{
				{
					Row{Cells: []string{"Руководящее звено (руководитель и его заместители), чел. 2 1"}},
					Row{Cells: []string{"Специалисты (в том числе инженеры), чел. 6 4"}},
					Row{Cells: []string{"Рабочие и вспомогательный персонал 0 0"}},
					Row{Cells: []string{"ИТОГО: 8 5"}},
				},
			},
		},
	}
	got := extractStaffSummary(in)
	if got == nil {
		t.Fatal("summary = nil")
	}
	if got.ManagersCount == nil || *got.ManagersCount != 2 {
		t.Fatalf("managers = %v", got.ManagersCount)
	}
	if got.SpecialistsCount == nil || *got.SpecialistsCount != 6 {
		t.Fatalf("specialists = %v", got.SpecialistsCount)
	}
	if got.WorkersCount == nil || *got.WorkersCount != 0 {
		t.Fatalf("workers = %v", got.WorkersCount)
	}
	if got.TotalStaff == nil || *got.TotalStaff != 8 {
		t.Fatalf("total = %v", got.TotalStaff)
	}
}

func TestExtractStaffMembers(t *testing.T) {
	in := []Input{
		{
			Tables: [][]Row{
				{
					Row{Cells: []string{"Руководящее звено"}},
					Row{Cells: []string{"1. Дубинович И.В., заместитель генерального директора ФГБОУ ВО «УрГПУ», 2001, «Физика» 3"}},
					Row{Cells: []string{"Специалисты (в том числе исполнители)"}},
					Row{Cells: []string{"1. Кудрявцев Д.А., руководитель направления по информационной безопасности ФГБОУ ВО «ИТМО» 6"}},
				},
			},
		},
	}
	got := extractStaffMembers(in)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[0].Category == nil || *got[0].Category != "managers" {
		t.Fatalf("cat0 = %v", got[0].Category)
	}
	if got[1].Category == nil || *got[1].Category != "specialists" {
		t.Fatalf("cat1 = %v", got[1].Category)
	}
	if got[0].FullName == nil || *got[0].FullName != "Дубинович И.В." {
		t.Fatalf("fn0 = %v", got[0].FullName)
	}
}
