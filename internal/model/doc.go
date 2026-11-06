// Package model contains domain types for tenderguide.
package model

// DocAll aggregates documentation entities for passing between layers.
type DocAll struct {
	Info         *DocInfo
	Requirements []*DocRequirement
}

// DocInfo represents the informational card and evaluation criteria of a documentation set.
type DocInfo struct {
	ID                           int64
	DocID                        string
	CustomerName                 *string
	CustomerINN                  *string
	OrganizerName                *string
	ProcedureNumber              *string
	ProcedureName                *string
	NMCK                         *string
	NMCKWithoutVAT               *string
	VAT                          *string
	SubmissionDeadline           *string
	ReviewDeadline               *string
	BidSecurityAmount            *string
	BidSecurityPercent           *string
	BidSecurityType              *string
	ContractSecurityAmount       *string
	ContractSecurityRequired     *bool
	PriceCriterionWeight         *string
	QualificationCriterionWeight *string
	LawBasis                     *string
	CreatedAt                    string
}

// DocRequirement represents a requirement or document listed in the documentation.
type DocRequirement struct {
	ID            int64
	DocID         string
	Kind          string // "requirement" | "document"
	Code          *string
	Name          *string
	Required      *bool
	SourceSection *string
}
