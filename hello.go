package main

import (
	"encoding/csv"
	"fmt"
	"github.com/megesdal/melodispurences/address"
	"github.com/megesdal/melodispurences/bed"
	"github.com/megesdal/melodispurences/data"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

type Person struct {
	location   address.Coordinates
	names      []string
	toBoxJelly float64
	address    string
}

func main() {
	fmt.Printf("hello, world\n")

	runRangeQuery()
	//groupByAddress()
	//findContributorTypes()

	return
}

func groupByAddress() {
	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)

	people := []Person{}
	boxJellyCoordinates := address.New(21.296834, -157.85665)

	count := 0

	for true {
		count++
		donation, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		personName := donation[2]
		parsedAddress := strings.Split(donation[22], "\n")
		// fmt.Printf("%v\n", addressSplit)
		if len(parsedAddress) < 3 {
			continue
		}

		var location address.Coordinates
		location, err = address.ExtractCoordinates(parsedAddress[2])

		if err != nil {
			log.Printf("%v\n", err)
			continue
		}

		toBoxJelly := address.CalculateDistance(location, boxJellyCoordinates)

		mergedPerson := false

		// fmt.Printf("Processed %v records, with %v uniques\n", count, len(people))
		for j := 0; j < len(people); j++ {
			if math.Abs(toBoxJelly-people[j].toBoxJelly) < 2 && address.CalculateDistance(location, people[j].location) < 2 {
				people[j].names = append(people[j].names, personName)
				mergedPerson = true
				fmt.Printf("Merging %s into %s @ %v vs %v\n", personName, people[j].names[0], donation[8], people[j].address)
				break
			}
		}

		if !mergedPerson {
			// fmt.Printf("Attaching new candidate: %s\n", personName)
			people = append(people, Person{location, []string{personName}, toBoxJelly, donation[8]})
		}
	}
	for i := 0; i < len(people); i++ {
		fmt.Printf("%v\n", strings.Join(people[i].names, " | "))
	}
}

func findContributorTypes() {
	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	contributorTypes := make(map[string]bool)

	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		contributorTypes[fields[1]] = true
	}

	fmt.Println(contributorTypes)
}

func runRangeQuery() {

	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	startTS := time.Now()
	countPersonTotal := 0
	countPersonUnique := 0

	var queryDurationSum time.Duration
	var insertDurationSum time.Duration

	branchFactor := 32
	lastNameTree := bed.New(branchFactor, bed.CompareDictionaryOrder)
	firstNameTree := bed.New(branchFactor, bed.CompareDictionaryOrder)
	lastId := 0
	people := make(map[int]*data.Person)
	contributions := make([]*data.Contribution, 0)
	//organizations := make(map[int]*data.Organization)
	//pacs := make(map[int]*data.Committee)

	graph := data.ConnectGraphDb()
	graph.Clean()
	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		contribution := data.NewContribution()

		foundId := -1
		exactMatchLN := false
		exactMatchFN := false
		contributorType := fields[1]
		isPerson := false
		if contributorType == "Immediate Family" || contributorType == "Candidate" || contributorType == "Individual" {
			isPerson = true
		}

		if !isPerson {
			continue
		}

		toCheck := fields[2]
		firstLast := strings.SplitN(toCheck, ",", 2)

		lastName := strings.Trim(firstLast[0], "\r\n\t ")
		var firstName string
		if len(firstLast) == 2 {
			firstName = strings.Trim(firstLast[1], "\r\n\t ")
		}

		beforeLastQueryTS := time.Now()
		lastNameResults := lastNameTree.RangeQuery(lastName, 0.2)
		queryDurationSum += time.Now().Sub(beforeLastQueryTS)

		if len(lastNameResults) > 0 {

			var possibleIds map[int]bool
			possibleLastNameResultIds := make(map[int]bool)
			for _, lastNameResult := range lastNameResults {
				for _, lastNameResultId := range lastNameResult.Values {
					possibleLastNameResultIds[lastNameResultId.(int)] = lastNameResult.Key == lastName
				}
			}

			if firstName != "" {

				beforeFirstQueryTS := time.Now()
				firstNameResults := firstNameTree.RangeQuery(firstName, 0.2)
				queryDurationSum += time.Now().Sub(beforeFirstQueryTS)

				if len(firstNameResults) > 0 {
					possibleIds = make(map[int]bool)
					for _, firstNameResult := range firstNameResults {
						for _, firstNameResultId := range firstNameResult.Values {
							// find intersection...
							_, existsInLastName := possibleLastNameResultIds[firstNameResultId.(int)]
							if existsInLastName {
								possibleIds[firstNameResultId.(int)] = firstNameResult.Key == firstName
							}
						}
					}
				}
			} else {
				possibleIds = possibleLastNameResultIds
			}

			if len(possibleIds) > 0 {
				// pick a better id...
				for possibleId, exactMatch := range possibleIds {
					foundId = possibleId
					if firstName != "" {
						exactMatchFN = exactMatch
						exactMatchLN = possibleLastNameResultIds[possibleId]
					} else {
						exactMatchLN = exactMatch
					}
					break
				}
			}
		}

		if foundId < 0 {
			foundId = lastId
			newPerson := data.NewPerson(firstName, lastName)
			contribution.SetContributor(newPerson, contributorType)
			people[foundId] = newPerson
			graph.PopulateGraphWithPersonContribution(newPerson, contribution)
			lastId++
			countPersonUnique++
		} else {
			contribution.SetContributor(people[foundId], contributorType)
		}
		contributions = append(contributions, contribution)

		if firstName != "" && !exactMatchFN {
			beforeInsertTS := time.Now()
			firstNameTree.Put(firstName, foundId)
			insertDurationSum += time.Now().Sub(beforeInsertTS)
		}

		if !exactMatchLN {
			beforeInsertTS := time.Now()
			lastNameTree.Put(lastName, foundId)
			insertDurationSum += time.Now().Sub(beforeInsertTS)
		}

		countPersonTotal++

		if countPersonTotal%1000 == 0 {
			durationSoFarNano := time.Now().Sub(startTS)
			fmt.Printf("%d unique out of %d processed so far [%dms total, %dms inserting]\n", lastNameTree.Size(), countPersonTotal, durationSoFarNano/time.Millisecond, insertDurationSum/time.Millisecond)
		}
	}

	durationNano := time.Now().Sub(startTS)
	fmt.Printf("%d unique out of %d processed\n", countPersonUnique, countPersonTotal)
	fmt.Printf("%d branch factor: %d num nodes in firstName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, firstNameTree.NumTotalNodes(), firstNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)
	fmt.Printf("%d branch factor: %d num nodes in lastName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, lastNameTree.NumTotalNodes(), lastNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)

	//fmt.Println(firstNameTree.String())
	//fmt.Println(lastNameTree.String())
}
