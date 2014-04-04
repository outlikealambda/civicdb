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
