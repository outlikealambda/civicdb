package data

import "time"

type Address struct {
	streetAddress1 string
	streetAddress2 string
	city           string
	state          string
	zip            int // or string to make it generic for all postal codes?

	lat float64
	lon float64
}

// this covers org reports and contributions in and out of candidate and non-candidate committees

type Organization struct {
	name    string
	address *Address
}

func NewOrganization(name string) *Organization {
	return &Organization{name: name}
}

func (c *Organization) Name() string {
	return c.name
}

type Person struct {
	FirstName     string
	LastName      string
	address       *Address
	businessPhone int // requires normalization

	// only PAC have this data
	occupation      string
	placeOfBusiness string // often an address, but sometimes a bizname

	// only on the contribution
	employer string
}

func NewPerson(firstName string, lastName string) *Person {
	return &Person{FirstName: firstName, LastName: lastName}
}

func (p *Person) Name() string {
	return p.LastName + ", " + p.FirstName
}

type Committee struct {
	RegNo       string
	name        string
	address     *Address
	phone       int     // normalize and make an int?
	Candidate   *Person // or should this be a link to a person?
	Chairperson *Person
	Treasurer   *Person
	party       string // should pull from a master list
	terminated  bool
}

func (c *Committee) Name() string {
	return c.name
}

// per election period
type Candidacy struct {
	Candidate *Person // or should this be a link to a person?
	office    string  // should have a master list
	district  string  // should pull from a master enum list
	county    string  // should pull from a master list
	inOffice  bool

	// name (or individual) and office overlap with profiles
	electionPeriodStart time.Time
	electionPeriodEnd   time.Time
}

func NewCandidacy(candidate *Person, office string) *Candidacy {
	return &Candidacy{Candidate: candidate, office: office}
}

type NonCandidateCommittee struct {
	Committee
	ncType                 string     // what is this? should probably have an enum list
	area                   string     //area, scope, jurisdiction
	ballotIssueDescription string     // what is this?
	singleCandidate        *Candidacy // or should this link to the Person, since the candidacy is linked to an election period
}

// per election period implied by candidacy
type CandidateCommittee struct {
	Candidate *Person
	Committee
}

func NewCandidateCommittee(regNo string, name string, candidate *Person, chairperson *Person, treasurer *Person) *CandidateCommittee {
	return &CandidateCommittee{
		candidate,
		Committee{
			RegNo:       regNo,
			name:        name,
			Chairperson: chairperson,
			Treasurer:   treasurer,
		},
	}
}

type Contributor interface {
	Name() string
}

// election period implied by committee contributed to? (candidate committees are per election period)
type Contribution struct {
	recipient       *Committee
	contributor     Contributor // TODO: this could be a person, committee or organization
	ContributorType string
	date            time.Time
	amount          int
	aggregate       int // what is this?

	mappingLocation string // what is this? street address of the commitee
	outOfState      bool

	//range should be implied by amount?
}

func NewContribution() *Contribution {
	return &Contribution{}
}

func (c *Contribution) SetContributor(contributor Contributor, contributorType string) {
	c.contributor = contributor
	c.ContributorType = contributorType
}

type NonMonetaryContribution struct {
	Contribution
	category    string
	description string
}
