package data

import (
	"github.com/megesdal/melodispurences/bed"
	"strings"
	"time"
)

type PersonIndex struct {
	lastId            int
	firstNameTree     *bed.BPlusTree
	lastNameTree      *bed.BPlusTree
	queryDurationSum  time.Duration
	insertDurationSum time.Duration
	count             int
	people            map[int]*Person
}

func NewPersonIndex() *PersonIndex {
	branchFactor := 32
	lastNameTree := bed.New(branchFactor, bed.CompareDictionaryOrder)
	firstNameTree := bed.New(branchFactor, bed.CompareDictionaryOrder)
	lastId := 0
	people := make(map[int]*Person)
	return &PersonIndex{
		lastId:        lastId,
		firstNameTree: firstNameTree,
		lastNameTree:  lastNameTree,
		people:        people,
	}
}

func (index *PersonIndex) QueryTimeSpent() time.Duration {
	return index.queryDurationSum
}

func (index *PersonIndex) InsertTimeSpent() time.Duration {
	return index.insertDurationSum
}

func (index *PersonIndex) Size() int {
	return index.count
}

func (index *PersonIndex) GetOrCreatePerson(name string) (*Person, bool) {

	foundId := -1
	exactMatchLN := false
	exactMatchFN := false

	firstLast := strings.SplitN(name, ",", 2)

	lastName := strings.Trim(firstLast[0], "\r\n\t ")
	var firstName string
	if len(firstLast) == 2 {
		firstName = strings.Trim(firstLast[1], "\r\n\t ")
	}

	beforeLastQueryTS := time.Now()
	lastNameResults := index.lastNameTree.RangeQuery(lastName, 0.2)
	index.queryDurationSum += time.Now().Sub(beforeLastQueryTS)

	if len(lastNameResults) > 0 {

		var possibleIds map[int]float64
		possibleLastNameResultIds := make(map[int]float64)
		for _, lastNameResult := range lastNameResults {
			for _, lastNameResultId := range lastNameResult.Values {
				possibleLastNameResultIds[lastNameResultId.(int)] = lastNameResult.Distance
			}
		}

		if firstName != "" {

			beforeFirstQueryTS := time.Now()
			firstNameResults := index.firstNameTree.RangeQuery(firstName, 0.2)
			index.queryDurationSum += time.Now().Sub(beforeFirstQueryTS)

			if len(firstNameResults) > 0 {
				possibleIds = make(map[int]float64)
				for _, firstNameResult := range firstNameResults {
					for _, firstNameResultId := range firstNameResult.Values {
						// find intersection...
						_, existsInLastName := possibleLastNameResultIds[firstNameResultId.(int)]
						if existsInLastName {
							possibleIds[firstNameResultId.(int)] = firstNameResult.Distance
						}
					}
				}
			}
		} else {
			possibleIds = possibleLastNameResultIds
		}

		if len(possibleIds) > 0 {
			minDistanceSum := float64(-1)
			for possibleId, distance := range possibleIds {

				distanceSum := distance
				if firstName != "" {
					distanceSum += possibleLastNameResultIds[possibleId]
				}

				if minDistanceSum < 0 || distanceSum < minDistanceSum {
					minDistanceSum = distanceSum
					foundId = possibleId

					if firstName != "" {
						exactMatchFN = distance == 0
						exactMatchLN = possibleLastNameResultIds[possibleId] == 0
					} else {
						exactMatchFN = false
						exactMatchLN = distance == 0
					}
				}
			}
		}
	}

	isNew := false
	var person *Person
	if foundId < 0 {
		foundId = index.lastId
		person = NewPerson(firstName, lastName)
		person.Id = foundId
		index.people[foundId] = person
		index.lastId++
		index.count++
		isNew = true
	} else {
		person = index.people[foundId]
	}

	if firstName != "" && !exactMatchFN {
		beforeInsertTS := time.Now()
		index.firstNameTree.Put(firstName, foundId)
		index.insertDurationSum += time.Now().Sub(beforeInsertTS)
	}

	if !exactMatchLN {
		beforeInsertTS := time.Now()
		index.lastNameTree.Put(lastName, foundId)
		index.insertDurationSum += time.Now().Sub(beforeInsertTS)
	}

	return person, isNew
}
