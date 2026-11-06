package model

// Request represents a procurement request parsed from a DOCX file.
// All nullable fields use pointers; absence of data is nil, not zero value.
type Request struct {
	ID                    int64
	RequestID             string  // derived from directory name, always present
	ProcedureNumber       *string
	ProcedureName         *string
	ProcedurePublishedAt  *string // ISO-8601
	LetterNumber          *string
	LetterDate            *string // ISO-8601
	ValidUntil            *string // ISO-8601
	PriceWithVAT          *string // decimal string
	VAT                   *string // decimal string
	PriceWithoutVAT       *string // decimal string
	PriceCurrency         *string
	AwardExperienceAmount *string
	CustomerID            *int64
	ParticipantID         *int64
	CreatedAt             string // ISO-8601
}

// Customer represents the customer (заказчик) of a request.
type Customer struct {
	ID   int64
	INN  *string
	Name *string
}

// Participant represents the participant (участник) of a request.
type Participant struct {
	ID                         int64
	INN                        *string
	KPP                        *string
	OGRN                       *string
	OKPO                       *string
	OKTMO                      *string
	OKOPF                      *string
	OKOPFName                  *string
	FullName                   *string
	ShortName                  *string
	LegalAddress               *string
	PostalAddress              *string
	PrimaryOKVED               *string
	PrimaryOKVEDName           *string
	EGRULRegistrationDate      *string // ISO-8601
	EGRULCertificate           *string
	EGRULRegistrationAuthority *string
	TaxRegistrationDateCurrent *string // ISO-8601
	TaxAuthorityCurrent        *string
	Phone                      *string
	Fax                        *string
	Email                      *string
	Website                    *string
	DirectorName               *string
	DirectorPosition           *string
	ChiefAccountantName        *string
	ResponsiblePersonName      *string
	ResponsiblePersonPhone     *string
	HasBranches                *bool
}

// BankAccount represents a bank account of a participant.
type BankAccount struct {
	ID                   int64
	ParticipantID        int64
	BankName             *string
	Account              *string
	CorrespondentAccount *string
	BIK                  *string
}

// MSPInfo represents small/medium business information of a participant.
type MSPInfo struct {
	ID                  int64
	ParticipantID       int64
	IsSMP               *bool
	SMPCategory         *string
	AverageHeadcount    *int64
	RevenuePreviousYear *int64
	Contracts44223Count *int64
	Contracts44223Sum   *string // decimal string
	InRNP               *bool
}

// StaffSummary represents aggregated staff counts for a request.
type StaffSummary struct {
	ID               int64
	RequestID        int64
	TotalStaff       *int64
	ManagersCount    *int64
	SpecialistsCount *int64
	WorkersCount     *int64
}

// StaffMember represents a single staff member listed in a request.
type StaffMember struct {
	ID              int64
	RequestID       int64
	Category        *string
	FullName        *string
	Position        *string
	Education       *string
	ExperienceYears *string
}

// Beneficiary represents a beneficiary of a participant.
type Beneficiary struct {
	ID                  int64
	ParticipantID       int64
	FullName            *string
	INN                 *string
	Passport            *string
	RegistrationAddress *string
	Role                *string
	ShareGT10Pct        *bool
}

// All aggregates all entities of a single request for passing between layers.
type All struct {
	Request       *Request
	Customer      *Customer
	Participant   *Participant
	BankAccounts  []*BankAccount
	MSPInfo       *MSPInfo
	StaffSummary  *StaffSummary
	StaffMembers  []*StaffMember
	Beneficiaries []*Beneficiary
}

// NewAll creates an All with a Request initialized with the given requestID.
func NewAll(requestID string) *All {
	return &All{
		Request: &Request{RequestID: requestID},
	}
}

// IsEmpty reports whether the request, participant and customer have no non-empty fields.
func (a *All) IsEmpty() bool {
	if a == nil {
		return true
	}
	return requestIsEmpty(a.Request) && participantIsEmpty(a.Participant) && customerIsEmpty(a.Customer)
}

func requestIsEmpty(r *Request) bool {
	if r == nil {
		return true
	}
	return r.ProcedureNumber == nil &&
		r.ProcedureName == nil &&
		r.ProcedurePublishedAt == nil &&
		r.LetterNumber == nil &&
		r.LetterDate == nil &&
		r.ValidUntil == nil &&
		r.PriceWithVAT == nil &&
		r.VAT == nil &&
		r.PriceWithoutVAT == nil &&
		r.PriceCurrency == nil &&
		r.AwardExperienceAmount == nil &&
		r.CustomerID == nil &&
		r.ParticipantID == nil
}

func customerIsEmpty(c *Customer) bool {
	if c == nil {
		return true
	}
	return c.INN == nil && c.Name == nil
}

func participantIsEmpty(p *Participant) bool {
	if p == nil {
		return true
	}
	return p.INN == nil &&
		p.KPP == nil &&
		p.OGRN == nil &&
		p.OKPO == nil &&
		p.OKTMO == nil &&
		p.OKOPF == nil &&
		p.OKOPFName == nil &&
		p.FullName == nil &&
		p.ShortName == nil &&
		p.LegalAddress == nil &&
		p.PostalAddress == nil &&
		p.PrimaryOKVED == nil &&
		p.PrimaryOKVEDName == nil &&
		p.EGRULRegistrationDate == nil &&
		p.EGRULCertificate == nil &&
		p.EGRULRegistrationAuthority == nil &&
		p.TaxRegistrationDateCurrent == nil &&
		p.TaxAuthorityCurrent == nil &&
		p.Phone == nil &&
		p.Fax == nil &&
		p.Email == nil &&
		p.Website == nil &&
		p.DirectorName == nil &&
		p.DirectorPosition == nil &&
		p.ChiefAccountantName == nil &&
		p.ResponsiblePersonName == nil &&
		p.ResponsiblePersonPhone == nil &&
		p.HasBranches == nil
}
