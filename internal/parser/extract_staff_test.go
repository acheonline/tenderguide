package parser

import "testing"

func TestExtractStaffSummary(t *testing.T) {
	in := []Input{
		{
			Tables: [][]Row{
				{
					Row{Cells: []string{"Руководящее звено", "2", "1"}},
					Row{Cells: []string{"Специалисты", "6", "4"}},
					Row{Cells: []string{"Рабочие", "0", "0"}},
					Row{Cells: []string{"ИТОГО:", "8", "5"}},
				},
			},
		},
	}
	got := extractStaffSummary(in)
	if got == nil {
		t.Fatal("summary = nil")
	}
	if got.ManagersCount == nil || *got.ManagersCount != 1 {
		t.Fatalf("managers = %v", got.ManagersCount)
	}
	if got.SpecialistsCount == nil || *got.SpecialistsCount != 4 {
		t.Fatalf("specialists = %v", got.SpecialistsCount)
	}
	if got.TotalStaff == nil || *got.TotalStaff != 5 {
		t.Fatalf("total = %v", got.TotalStaff)
	}
}

func TestExtractStaffMembers(t *testing.T) {
	in := []Input{
		{
			Tables: [][]Row{
				{
					Row{Cells: []string{"Руководящее звено"}},
					Row{Cells: []string{"1.", "Дубинович И.В., заместитель генерального директора", "ФГБОУ ВО ...", "3"}},
					Row{Cells: []string{"Специалисты"}},
					Row{Cells: []string{"1.", "Кудрявцев Д.А., руководитель направления", "ФГБОУ ...", "6"}},
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
