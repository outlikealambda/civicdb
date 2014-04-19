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

	distanceToFixedReference float64 // no need to persist
	bearingToFixedReference  float64
}

func NewAddress(lat, lon, distance, bearing float64) *Address {
	return &Address{lat: lat, lon: lon, distanceToFixedReference: distance, bearingToFixedReference: bearing}
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
	Id            int
	FirstName     string
	LastName      string
	addresses     []*Address
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
	Id          int
	RegNo       string
	name        string
	address     *Address
	phone       int     // normalize and make an int?
	Candidate   *Person // or should this be a link to a person?
	Chairperson *Person
	Treasurer   *Person
	Party       string // should pull from a master list
	Terminated  bool
}

func (c *Committee) Name() string {
	return c.name
}

// per election period
type Candidacy struct {
	Candidate *Person // or should this be a link to a person?
	Office    *Office // should have a master list
	InOffice  bool

	// name (or individual) and office overlap with profiles
	electionPeriodStart time.Time
	electionPeriodEnd   time.Time
}

func NewCandidacy(candidate *Person, office *Office) *Candidacy {
	return &Candidacy{Candidate: candidate, Office: office}
}

type NonCandidateCommittee struct {
	NCType          string // TODO: enum list of valid types
	Area            string
	Issue           string
	SingleCandidate *Person
	Committee
}

func NewNonCandidateCommittee(regNo string, name string, chairperson *Person, treasurer *Person, nctype string, terminated bool) *NonCandidateCommittee {
	return &NonCandidateCommittee{
		NCType: nctype,
		Committee: Committee{
			RegNo:       regNo,
			name:        name,
			Chairperson: chairperson,
			Treasurer:   treasurer,
			Terminated:  terminated,
		},
	}
}

func (c *NonCandidateCommittee) SetSingleCandidate(person *Person) {
	c.SingleCandidate = person
}

func (c *NonCandidateCommittee) SetFocus(area string, party string, issue string) {
	c.Area = area
	c.Committee.Party = party
	c.Issue = issue
}

// per election period implied by candidacy
type CandidateCommittee struct {
	Candidate *Person
	Race      *Office
	InOffice  bool
	//otherOffices map[string]string
	Committee
}

func NewCandidateCommittee(regNo string, name string, candidate *Person, chairperson *Person, treasurer *Person, office *Office, party string, terminated bool, inOffice bool) *CandidateCommittee {
	return &CandidateCommittee{
		candidate,
		office,
		inOffice,
		//make(map[string]string),
		Committee{
			RegNo:       regNo,
			name:        name,
			Chairperson: chairperson,
			Treasurer:   treasurer,
			Party:       party,
			Terminated:  terminated,
		},
	}
}

//func (c *CandidateCommittee) AddOtherOffice(term string, otherOffice string) {
//	c.otherOffices[term] = otherOffice
//}

//func (c *CandidateCommittee) GetOtherOfficeForTerm(term string) string {
//	return c.otherOffices[term]
//}

type Contributor interface {
	Name() string
}

// election period implied by committee contributed to? (candidate committees are per election period)
type Contribution struct {
	Id              int
	Recipient       *Committee
	Contributor     Contributor // TODO: this could be a person, committee or organization
	ContributorType string
	Amount          int
	Aggregate       int
	Period          string

	// TODO
	//date            time.Time
	//mappingLocation string // what is this? street address of the commitee
	//outOfState      bool
	//range should be implied by amount?
}

func NewContribution(id int, recipient *Committee, amount int, aggregate int, period string) *Contribution {
	return &Contribution{Recipient: recipient, Amount: amount, Period: period}
}

func (c *Contribution) SetContributor(contributor Contributor, contributorType string) {
	c.Contributor = contributor
	c.ContributorType = contributorType
}

type NonMonetaryContribution struct {
	Contribution
	category    string
	description string
}

type Office struct {
	Id       int
	Title    string
	Region   string // HI or US
	District string
	County   string
}

func NewOffice(id int, title string, region string, district string, county string) *Office {
	return &Office{id, title, region, district, county}
}
