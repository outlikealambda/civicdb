package data

import (
	"fmt"
	"github.com/jmcvetta/neoism"
	"log"
)

type Neo4jConnection struct {
	db *neoism.Database
}

func ConnectGraphDb() *Neo4jConnection {
	db, err := neoism.Connect("http://localhost:7474/db/data")
	if err != nil {
		log.Fatal(err)
	}
	return &Neo4jConnection{db: db}
}

func (graph *Neo4jConnection) Clean() {
	qs := []*neoism.CypherQuery{
		&neoism.CypherQuery{
			Statement: `
				MATCH (n)
				OPTIONAL MATCH (n)-[r]-()
				DELETE n,r
			`,
			Parameters: neoism.Props{},
		},
		/*&neoism.CypherQuery{
			Statement: `
				MATCH (n)
				DELETE n
			`,
			Parameters: neoism.Props{},
		},*/
	}
	tx, _ := graph.db.Begin(qs)
	tx.Commit()
}

func (graph *Neo4jConnection) Init() {
	graph.db.CreateIndex("Committee", "regNo")
	graph.db.CreateIndex("Person", "personId")
	graph.db.CreateIndex("Office", "officeId")
}

func (graph *Neo4jConnection) AddNonCandidateCommittee(committee *NonCandidateCommittee) {

	chairpersonNode := graph.findOrCreatePerson(committee.Chairperson)
	treasurerNode := graph.findOrCreatePerson(committee.Treasurer)

	committeeNode := graph.createCommittee(&committee.Committee)
	committeeNode.SetProperty("nctype", committee.NCType)
	committeeNode.SetProperty("area", committee.Area)
	committeeNode.SetProperty("issue", committee.Issue)

	chairpersonNode.Relate("chairperson for", committeeNode.Id(), neoism.Props{})
	treasurerNode.Relate("treasurer for", committeeNode.Id(), neoism.Props{})

	if committee.SingleCandidate != nil {
		candidateNode := graph.findOrCreatePerson(committee.SingleCandidate)
		committeeNode.Relate("dedicated to", candidateNode.Id(), neoism.Props{})
	}
}

func (graph *Neo4jConnection) AddCandidateCommittee(committee *CandidateCommittee) {

	// findPerson
	candidateNode := graph.findOrCreatePerson(committee.Candidate)
	chairpersonNode := graph.findOrCreatePerson(committee.Chairperson)
	treasurerNode := graph.findOrCreatePerson(committee.Treasurer)

	/*chairpersonNode := graph.findPerson(committee.Chairperson)
	if chairpersonNode == nil {
		chairpersonNode = graph.createPerson(committee.Chairperson)
	}

	treasurerNode := graph.findPerson(committee.Treasurer)
	if treasurerNode == nil {
		treasurerNode = graph.createPerson(committee.Treasurer)
	}*/

	// create candidate committee
	committeeNode := graph.createCommittee(&committee.Committee)
	committeeNode.SetProperty("inOffice", fmt.Sprintf("%v", committee.InOffice))

	// N.B. Empty Props{} is okay
	candidateNode.Relate("candidate for", committeeNode.Id(), neoism.Props{})
	chairpersonNode.Relate("chairperson for", committeeNode.Id(), neoism.Props{})
	treasurerNode.Relate("treasurer for", committeeNode.Id(), neoism.Props{})

	/*
		if office != nil {
			officeNode := graph.findOffice(office)
			if officeNode == nil {
				officeNode = graph.createOffice(office)
			}

			committeeNode.Relate("ran for", officeNode.Id(), neoism.Props{})
		}
	*/
}

func (graph *Neo4jConnection) AddCandidacy(committee *CandidateCommittee, office *Office, term string) {
	if office != nil {
		officeNode := graph.findOffice(office)
		if officeNode == nil {
			officeNode = graph.createOffice(office)
		}

		committeeNode := graph.findCommittee(&committee.Committee)
		if committeeNode != nil {
			committeeNode.Relate("ran for", officeNode.Id(), neoism.Props{"in": term})
		}
	}
}

func (graph *Neo4jConnection) findOffice(office *Office) *neoism.Node {
	result := []struct {
		N neoism.Node
	}{}
	var query *neoism.CypherQuery
	/*if office.Id > 0 {
		query = &neoism.CypherQuery{
			// Use backticks for long statements - Cypher is whitespace indifferent
			Statement: `
				MATCH (n:Office)
				WHERE WHERE n.Id = {id}
				RETURN n
			`,
			Parameters: neoism.Props{"id": office.Id},
			Result:     &result,
		}
	} else {*/
	query = &neoism.CypherQuery{
		// Use backticks for long statements - Cypher is whitespace indifferent
		Statement: `
				MATCH (n:Office)
				WHERE n.officeId = {officeId}
				RETURN n
			`,
		Parameters: neoism.Props{"officeId": office.Id},
		Result:     &result,
	}
	//}
	graph.db.Cypher(query)
	var officeNode *neoism.Node
	if len(result) > 0 {
		officeNode = &result[0].N
		officeNode.Db = graph.db
	}
	return officeNode
}

func (graph *Neo4jConnection) createOffice(office *Office) *neoism.Node {
	result := []struct {
		N neoism.Node // Column "n" gets automagically unmarshalled into field N
	}{}
	query := neoism.CypherQuery{
		Statement: "CREATE (n:Office {officeId: {officeId}, title: {title}, region: {region}, district: {district}, county: {county}}) RETURN n",
		// Use parameters instead of constructing a query string
		Parameters: neoism.Props{"officeId": office.Id, "title": office.Title, "region": office.Region, "district": office.District, "county": office.County},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	node := result[0].N // Only one row of data returned
	node.Db = graph.db  // Must manually set Db with objects returned from Cypher query
	//office.Id = node.Id()
	return &node
}

func (graph *Neo4jConnection) findCommittee(committee *Committee) *neoism.Node {
	result := []struct {
		N neoism.Node
	}{}
	query := neoism.CypherQuery{
		// Use backticks for long statements - Cypher is whitespace indifferent
		Statement: `
			MATCH (n:Committee)
			WHERE n.regNo = {regNo}
			RETURN n
		`,
		Parameters: neoism.Props{"regNo": committee.RegNo},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	var committeeNode *neoism.Node
	if len(result) > 0 {
		committeeNode = &result[0].N
		committeeNode.Db = graph.db
	}
	return committeeNode
}

func (graph *Neo4jConnection) createCommittee(committee *Committee) *neoism.Node {
	result := []struct {
		N neoism.Node // Column "n" gets automagically unmarshalled into field N
	}{}
	query := neoism.CypherQuery{
		Statement: "CREATE (n:Committee {regNo: {regNo}, name: {name}, party: {party}, terminated: {terminated}}) RETURN n",
		// Use parameters instead of constructing a query string
		Parameters: neoism.Props{"regNo": committee.RegNo, "name": committee.Name(), "party": committee.Party, "terminated": committee.Terminated},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	node := result[0].N // Only one row of data returned
	node.Db = graph.db  // Must manually set Db with objects returned from Cypher query
	committee.Id = node.Id()
	return &node
	/*node, _, err := graph.db.CreateNode(neoism.Props{"regNo": committee.RegNo, "name": committee.Name()})
	if err != nil {
		log.Println(err)
	}
	return node*/
}

func (graph *Neo4jConnection) createPerson(person *Person) *neoism.Node {
	result := []struct {
		N neoism.Node // Column "n" gets automagically unmarshalled into field N
	}{}
	var zipCode int
	if person.addresses != nil {
		zipCode = person.addresses[0].zip
	}
	query := neoism.CypherQuery{
		Statement: "CREATE (n:Person {personId: {personId}, firstName: {firstName}, lastName: {lastName}, zipCode: {zipCode}}) RETURN n",
		// Use parameters instead of constructing a query string
		Parameters: neoism.Props{"personId": person.Id, "firstName": person.FirstName, "lastName": person.LastName, "zipCode": zipCode},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	node := result[0].N // Only one row of data returned
	node.Db = graph.db  // Must manually set Db with objects returned from Cypher query
	return &node
}

func (graph *Neo4jConnection) findOrCreatePerson(person *Person) *neoism.Node {
	/*node, _, err := graph.db.GetOrCreateNode("Person", "personId", neoism.Props{"personId": person.Id, "firstName": person.FirstName, "lastName": person.LastName})
	if err != nil {
		log.Println(err)
	}
	return node*/
	personNode := graph.findPerson(person)
	if personNode == nil {
		personNode = graph.createPerson(person)
	}
	return personNode
}

func (graph *Neo4jConnection) findPerson(person *Person) *neoism.Node {
	result := []struct {
		N neoism.Node
	}{}
	query := neoism.CypherQuery{
		// Use backticks for long statements - Cypher is whitespace indifferent
		Statement: `
			MATCH (n:Person)
			WHERE n.personId = {personId}
			RETURN n
		`,
		Parameters: neoism.Props{"personId": person.Id, "lastName": person.LastName},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	var personNode *neoism.Node
	if len(result) > 0 {
		personNode = &result[0].N
		personNode.Db = graph.db
	}
	return personNode
}

func (graph *Neo4jConnection) AddContribution(contribution *Contribution) {

	if recipientNode := graph.findCommittee(contribution.Recipient); recipientNode != nil {
		var contributorNode *neoism.Node
		if person, ok := contribution.Contributor.(*Person); ok {
			contributorNode = graph.findPerson(person)
			if contributorNode == nil {
				contributorNode = graph.createPerson(person)
			}
		}

		if committee, ok := contribution.Contributor.(*NonCandidateCommittee); ok {
			contributorNode = graph.findCommittee(&committee.Committee)
		}

		if contributorNode != nil {
			contributorNode.Relate("contributed to", recipientNode.Id(), neoism.Props{"aggregate": contribution.Aggregate, "contributionId": contribution.Id, "amount": contribution.Amount, "in": contribution.Period, "type": contribution.ContributorType})
		} else {
			log.Println("Unable to find contributor:", contribution.Contributor.Name())
		}
	}

	// TODO: should the contribution be properties on the relation "contributed to"

	/*
		res1 := []struct {
			A   string `json:"a.lastName"` // `json` tag matches column name in query
			Rel string `json:"type(r)"`
			B   string `json:"b.contributorType"`
		}{}
		cq1 := neoism.CypherQuery{
			// Use backticks for long statements - Cypher is whitespace indifferent
			Statement: `
				MATCH (a)-[r]->(b)
				WHERE a.lastName = {name}
				RETURN a.lastName, type(r), b.contributorType
			`,
			Parameters: neoism.Props{"name": person.LastName},
			Result:     &res1,
		}
		graph.db.Cypher(&cq1)
		r := res1[0]
		fmt.Println(r.A, r.Rel, r.B)
	*/
}
