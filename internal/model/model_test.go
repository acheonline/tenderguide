package model

import "testing"

func TestNewAll(t *testing.T) {
	a := NewAll("req-001")
	if a == nil {
		t.Fatal("NewAll returned nil")
	}
	if a.Request == nil {
		t.Fatal("Request is nil")
	}
	if a.Request.RequestID != "req-001" {
		t.Errorf("RequestID = %q, want %q", a.Request.RequestID, "req-001")
	}
	if a.Customer != nil {
		t.Errorf("Customer = %v, want nil", a.Customer)
	}
	if a.Participant != nil {
		t.Errorf("Participant = %v, want nil", a.Participant)
	}
	if a.BankAccounts != nil {
		t.Errorf("BankAccounts = %v, want nil", a.BankAccounts)
	}
	if a.MSPInfo != nil {
		t.Errorf("MSPInfo = %v, want nil", a.MSPInfo)
	}
	if a.StaffSummary != nil {
		t.Errorf("StaffSummary = %v, want nil", a.StaffSummary)
	}
	if a.StaffMembers != nil {
		t.Errorf("StaffMembers = %v, want nil", a.StaffMembers)
	}
	if a.Beneficiaries != nil {
		t.Errorf("Beneficiaries = %v, want nil", a.Beneficiaries)
	}
	if !a.IsEmpty() {
		t.Error("IsEmpty() = false, want true")
	}
}

func TestAllIsEmpty(t *testing.T) {
	inn := "7700000000"
	name := "ООО Ромашка"

	tests := []struct {
		name   string
		mutate func(a *All)
		want   bool
	}{
		{
			name:   "empty",
			mutate: func(a *All) {},
			want:   true,
		},
		{
			name: "participant INN set",
			mutate: func(a *All) {
				a.Participant = &Participant{INN: &inn}
			},
			want: false,
		},
		{
			name: "customer name set",
			mutate: func(a *All) {
				a.Customer = &Customer{Name: &name}
			},
			want: false,
		},
		{
			name: "request procedure name set",
			mutate: func(a *All) {
				a.Request.ProcedureName = &name
			},
			want: false,
		},
		{
			name: "empty participant struct",
			mutate: func(a *All) {
				a.Participant = &Participant{}
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAll("req-001")
			tt.mutate(a)
			if got := a.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
