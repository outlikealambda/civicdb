package data

import "time"

type Address struct {
	streetAddress1 string
	streetAddress2 string
	city string
	state string
	zip int  // or string to make it generic for all postal codes?

	lat float64
	lon float64
}

// this covers org reports and contributions in and out of candidate and non-candidate committees

type Person struct {
	name string // TODO: split into first+ and last?
	address *Address
	businessPhone int // requires normalization

	// only PAC have this data
	occupation string
	placeOfBusiness string  // often an address, but sometimes a bizname

	// only on the contribution
	employer string
}

func (p *Person) Name() string {
	return p.name
}

type Committee struct {
	regNo string
	name string
	address *Address
	phone int  // normalize and make an int?
	chairperson *Person
	treasurer *Person
	party string // should pull from a master list
	terminated bool
}

func (c *Committee) Name() string {
	return c.name
}

// per election period
type Candidacy struct {
	candidate *Person  // or should this be a link to a person?
	office string  // should have a master list
	district string // should pull from a master enum list
	county string // should pull from a master list
	inOffice bool

	// name (or individual) and office overlap with profiles
	electionPeriodStart time.Time
	electionPeriodEnd time.Time
}

type NonCandidateCommittee struct {
	Committee
	ncType string // what is this? should probably have an enum list
	area string //area, scope, jurisdiction
	ballotIssueDescription string // what is this?
	singleCandidate *Candidacy // or should this link to the Person, since the candidacy is linked to an election period
}

// per election period implied by candidacy
type CandidateCommitee struct {
	Candidacy
	Committee
}

type Contributor interface {
	Name() string
}

// election period implied by committee contributed to? (candidate committees are per election period)
type Contribution struct {
	recipient *Committee
	contributor Contributor  // TODO: this could be a person, committee or organization
	contrubutorType string
	date time.Time
	amount int
	aggregate int // what is this?

	mappingLocation string // what is this? street address of the commitee
	outOfState bool

	//range should be implied by amount?
}

type NonMonetaryContribution struct {
	Contribution
	category string
	description string
}