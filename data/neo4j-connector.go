package data

import (
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
				MATCH (n)-[r]->()
				DELETE r
			`,
			Parameters: neoism.Props{},
		},
		&neoism.CypherQuery{
			Statement: `
				MATCH n
				DELETE n
			`,
			Parameters: neoism.Props{},
		},
	}
	tx, _ := graph.db.Begin(qs)
	tx.Commit()
}

func (graph *Neo4jConnection) AddCandidateCommittee(committee *CandidateCommittee) {

	// findPerson
	candidateNode := graph.findPerson(committee.Candidate)
	if candidateNode == nil {
		candidateNode = graph.createPerson(committee.Candidate)
	}

	chairpersonNode := graph.findPerson(committee.Chairperson)
	if chairpersonNode == nil {
		chairpersonNode = graph.createPerson(committee.Chairperson)
	}

	treasurerNode := graph.findPerson(committee.Treasurer)
	if treasurerNode == nil {
		treasurerNode = graph.createPerson(committee.Treasurer)
	}

	// create candidate committee
	committeeNode := graph.createCommittee(&committee.Committee)

	// N.B. Empty Props{} is okay
	candidateNode.Relate("candidate for", committeeNode.Id(), neoism.Props{})
	chairpersonNode.Relate("chairperson for", committeeNode.Id(), neoism.Props{})
	treasurerNode.Relate("treasurer for", committeeNode.Id(), neoism.Props{})
}

func (graph *Neo4jConnection) createCommittee(committee *Committee) *neoism.Node {
	result := []struct {
		N neoism.Node // Column "n" gets automagically unmarshalled into field N
	}{}
	query := neoism.CypherQuery{
		Statement: "CREATE (n:Committee {regNo: {regNo}, name: {name}}) RETURN n",
		// Use parameters instead of constructing a query string
		Parameters: neoism.Props{"regNo": committee.RegNo, "name": committee.Name()},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	node := result[0].N // Only one row of data returned
	node.Db = graph.db  // Must manually set Db with objects returned from Cypher query
	return &node
}

func (graph *Neo4jConnection) createPerson(person *Person) *neoism.Node {
	result := []struct {
		N neoism.Node // Column "n" gets automagically unmarshalled into field N
	}{}
	query := neoism.CypherQuery{
		Statement: "CREATE (n:Person {firstName: {firstName}, lastName: {lastName}}) RETURN n",
		// Use parameters instead of constructing a query string
		Parameters: neoism.Props{"firstName": person.FirstName, "lastName": person.LastName},
		Result:     &result,
	}
	graph.db.Cypher(&query)
	node := result[0].N // Only one row of data returned
	node.Db = graph.db  // Must manually set Db with objects returned from Cypher query
	return &node
}

func (graph *Neo4jConnection) findPerson(person *Person) *neoism.Node {
	result := []struct {
		N neoism.Node
	}{}
	query := neoism.CypherQuery{
		// Use backticks for long statements - Cypher is whitespace indifferent
		Statement: `
			MATCH (n:Person)
			WHERE n.firstName = {firstName} AND n.lastName ={lastName}
			RETURN n
		`,
		Parameters: neoism.Props{"firstName": person.FirstName, "lastName": person.LastName},
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

func (graph *Neo4jConnection) PopulateGraphWithPersonContribution(person *Person, contribution *Contribution) {

	nodePerson, err := graph.db.CreateNode(neoism.Props{"lastName": person.LastName})
	if err != nil {
		log.Println(err)
	}
	nodeContribution, err2 := graph.db.CreateNode(neoism.Props{"contributorType": contribution.ContributorType})
	if err2 != nil {
		log.Println(err2)
	}

	// TODO: should the contribution be properties on the relation "contributed to"
	nodePerson.Relate("contributed as", nodeContribution.Id(), neoism.Props{}) // Empty Props{} is okay
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
