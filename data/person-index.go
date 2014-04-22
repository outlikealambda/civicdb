package data

import (
	"fmt"
	"github.com/peg-one/civicdb/bed"
	"github.com/peg-one/civicdb/wmaddress"
	"math"
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

func scoreLastNameDistance(distance float64) float64 {
	// return 0.2 - distance
	return 5 * (0.2 - distance)
}

func scoreFirstNameDistance(distance float64) float64 {
	// return 0.2 - distance
	return 2.5 * (0.2 - distance)
}

func scoreGeographicDistance(distance float64) float64 {
	return 1
}

func (index *PersonIndex) findPossibleLastNameMatches(lastName string) (possibleIds map[int]float64) {
	lastNameResults := index.lastNameTree.RangeQuery(lastName, 0.2)
	possibleIds = make(map[int]float64)

	// fmt.Printf("found %v possible last name matches for %v\n", len(lastNameResults), lastName)
	for _, result := range lastNameResults {
		for _, personId := range result.Values {
			possibleIds[personId.(int)] = scoreLastNameDistance(result.Distance)
		}
	}

	return
}

func (index *PersonIndex) findPossibleFirstNameMatches(firstName string) (possibleIds map[int]float64) {
	firstNameResults := index.firstNameTree.RangeQuery(firstName, 0.2)
	possibleIds = make(map[int]float64)

	for _, result := range firstNameResults {
		for _, personId := range result.Values {
			possibleIds[personId.(int)] = scoreFirstNameDistance(result.Distance)
		}
	}

	return
}

func (index *PersonIndex) findPossibleAddressMatches(lat, lon, fixedLat, fixedLon float64) (possibleIds map[int]float64, distance, bearing float64) {
	distance, bearing = wmaddress.CalculateDistance(lat, lon, fixedLat, fixedLon)
	possibleIds = make(map[int]float64)

	debug := false
	// if lat-21.285 < 0.001 {
	// 	debug = false
	// 	fmt.Println("")
	// }

	for key, person := range index.people {
		for _, address := range person.addresses {
			if debug {
				fmt.Printf("distance: %6.5f, existing: %6.5f", distance, address.distanceToFixedReference)
			}
			if math.Abs(distance-address.distanceToFixedReference) < 50 {
				d := wmaddress.CalculateApproximateDistance(distance, bearing, address.distanceToFixedReference, address.bearingToFixedReference)

				if debug {
					fmt.Printf(" | calculated triangle distance: %.2f", d)
				}

				if prevVal, isSet := possibleIds[key]; d < 50 && (!isSet || scoreGeographicDistance(d) < prevVal) {
					if debug {
						fmt.Printf(" | setting distance to: %f", scoreGeographicDistance(d))
					}
					possibleIds[key] = scoreGeographicDistance(d)
				}
			}
			if debug {
				fmt.Println("")
			}
		}
	}

	return
}

func extractLastNameFirstName(lastCommaFirst string) (lastName, firstName string) {
	firstLast := strings.SplitN(lastCommaFirst, ",", 2)

	lastName = strings.Trim(firstLast[0], "\r\n\t ")
	if len(firstLast) == 2 {
		firstName = strings.Trim(firstLast[1], "\r\n\t ")
	}

	return
}

func normalizeNameString(name string) string {
	// remove whitespace
	name = strings.Replace(name, " ", "", -1)
	// remove full-stops
	name = strings.Replace(name, ".", "", -1)
	// lower case
	name = strings.ToLower(name)

	return name
}

type keyScorePair struct {
	key   int
	score float64
}

func sortMatchedPersons(matchedScores map[int]float64) (sortedMatches []*keyScorePair) {
	sortedMatches = make([]*keyScorePair, 0, len(matchedScores))

	for key, score := range matchedScores {
		keyScore := &keyScorePair{key, score}
		inserted := false

		for i, pair := range sortedMatches {
			if pair.score < score {
				//insert before pair
				sortedMatches = append(sortedMatches, nil)
				copy(sortedMatches[i+1:], sortedMatches[i:])
				sortedMatches[i] = keyScore
				inserted = true
				break
			}
		}

		if !inserted {
			// append to end
			sortedMatches = append(sortedMatches, keyScore)
		}
	}

	return
}

func (index *PersonIndex) ExtractAndGetOrCreatePerson(fullName, coordString string) (person *Person, isNew bool) {
	lastName, firstName := extractLastNameFirstName(fullName)
	lat, lon, err := wmaddress.ExtractCoordinates(coordString)

	if err != nil {
		// name only
		lat, lon = 0, 0
	}

	return index.GetOrCreatePerson(firstName, lastName, lat, lon)
}

func (index *PersonIndex) GetOrCreatePerson(firstName, lastName string, lat, lon float64) (person *Person, isNew bool) {
	possibleLastNameIds := index.findPossibleLastNameMatches(normalizeNameString(lastName))

	possibleFirstNameIds := index.findPossibleFirstNameMatches(normalizeNameString(firstName))

	bjLat, bjLon := 21.296834, -157.85665
	possibleAddressIds, distance, bearing := index.findPossibleAddressMatches(lat, lon, bjLat, bjLon)

	possibleMatches := make(map[int]float64)

	for key, score := range possibleLastNameIds {
		// fmt.Printf("found %v last names, ", len(possibleLastNameIds))
		// i believe the map returns zero if un-initialized/used key is entered
		possibleMatches[key] = possibleMatches[key] + score
	}
	for key, score := range possibleFirstNameIds {
		// fmt.Printf("found %v first names, ", len(possibleFirstNameIds))
		// i believe the map returns zero if un-initialized/used key is entered
		possibleMatches[key] = possibleMatches[key] + score
	}
	for key, score := range possibleAddressIds {
		// fmt.Printf("found %v last addresses\n", len(possibleAddressIds))
		possibleMatches[key] = possibleMatches[key] + score
	}

	var personKey int
	isNew = true

	if len(possibleMatches) > 0 {
		// fmt.Printf("found %v potential matches", len(possibleAddressIds))
		sortedMatches := sortMatchedPersons(possibleMatches)

		// for _, match := range sortedMatches {
		// 	fmt.Printf("key: %v, score: %.2f\n", match.key, match.score)
		// }

		personKey = sortedMatches[0].key
		highScore := sortedMatches[0].score

		if lastName == "Kudo" {
			fmt.Printf("%v, %v MATCHED with score: %1.2f\n", lastName, firstName, highScore)
			for key, score := range possibleMatches {
				mp := index.people[key]
				fmt.Printf("id: %v, score: %1.2f [%1.2f, %1.2f, %1.2f]\n", key, score, possibleLastNameIds[key], possibleFirstNameIds[key], possibleAddressIds[key])
				fmt.Printf("person: %v, %v\n", mp.LastName, mp.FirstName)
				for _, address := range mp.addresses {
					fmt.Printf("%3.3f, %3.3f, bearing: %3.3f, distance %.1f\n", address.lat, address.lon, address.bearingToFixedReference, address.distanceToFixedReference)
				}
			}
		}

		if highScore > 1 {
			isNew = false
			person = index.people[personKey]
			// return existing person
		}

	} else {
		// fmt.Printf("No Possible Matches for: %v, %v\n", lastName, firstName)
	}

	if isNew {
		person = NewPerson(firstName, lastName)
		if lat != 0 {
			person.addresses = append(person.addresses, NewAddress(lat, lon, distance, bearing))
		}

		person.Id = index.lastId
		personKey = index.lastId
		index.people[personKey] = person

		index.lastId++
		index.count++

		if len(lastName) > 0 {
			index.lastNameTree.Put(normalizeNameString(lastName), personKey)
		}
		if len(firstName) > 0 {
			index.firstNameTree.Put(normalizeNameString(firstName), personKey)
		}
	} else if _, isSet := possibleAddressIds[personKey]; !isSet && lat != 0 {
		// we didn't match on address, so add this address to the person
		// fmt.Printf("no address match on: %v, got: %v\n", personKey, key)
		// for akey, value := range possibleAddressIds {
		// 	fmt.Printf("key: %v, value: %v\n", akey, value)
		// }
		person.addresses = append(person.addresses, NewAddress(lat, lon, distance, bearing))
	}

	return
}
