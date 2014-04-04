package main

import (
	"encoding/csv"
	"fmt"
	"github.com/megesdal/melodispurences/address"
	//"github.com/megesdal/melodispurences/bed"
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

	peopleIdx := data.NewPersonIndex()
	//TODO: candidatesByRegNo := make(map[string]*Person)
	populateCandidacies(peopleIdx /*, candidatesByRegNo*/)
	//populateCandidateCommitees(peopleIdx)
	//groupByAddress()
	runRangeQuery(peopleIdx)
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

func runRangeQuery(personIdx *data.PersonIndex) {

	fmt.Println("Populating contributions...")

	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	startTS := time.Now()
	countPersonTotal := 0
	countPersonUnique := 0

	contributions := make([]*data.Contribution, 0)
	//organizations := make(map[int]*data.Organization)
	//pacs := make(map[int]*data.Committee)
	peopleIdx := data.NewPersonIndex()
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

		contributorType := fields[1]
		isPerson := false
		if contributorType == "Immediate Family" || contributorType == "Candidate" || contributorType == "Individual" {
			isPerson = true
		}

		if !isPerson {
			continue
		}

		toCheck := fields[2]
		person, isNew := peopleIdx.GetOrCreatePerson(toCheck)
		if isNew {
			countPersonUnique++
			contribution.SetContributor(person, contributorType)
			graph.PopulateGraphWithPersonContribution(person, contribution)
		} else {
			contribution.SetContributor(person, contributorType)
		}
		contributions = append(contributions, contribution)
		countPersonTotal++

		if countPersonTotal%1000 == 0 {
			durationSoFarNano := time.Now().Sub(startTS)
			fmt.Printf("%d unique out of %d processed so far [%dms total, %dms inserting]\n", countPersonUnique, countPersonTotal, durationSoFarNano/time.Millisecond, peopleIdx.InsertTimeSpent()/time.Millisecond)
		}
	}

	durationNano := time.Now().Sub(startTS)
	fmt.Printf("%d unique out of %d processed TOTAL [%dms total, %dms inserting]\n", countPersonUnique, countPersonTotal, durationNano/time.Millisecond, peopleIdx.InsertTimeSpent()/time.Millisecond)
	//fmt.Printf("%d branch factor: %d num nodes in firstName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, firstNameTree.NumTotalNodes(), firstNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)
	//fmt.Printf("%d branch factor: %d num nodes in lastName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, lastNameTree.NumTotalNodes(), lastNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)

	//fmt.Println(firstNameTree.String())
	//fmt.Println(lastNameTree.String())
}

func populateCandidateCommitees(index *data.PersonIndex) {
	file, err := os.Open("data/Organizational_Reports_For_Hawaii_State_and_County_Candidates.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	count := 0
	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		committeeRegNo := strings.Trim(fields[0], " ")
		candidateName := strings.Trim(fields[1], " ")
		committeeName := strings.Trim(fields[2], " ")
		person, isNew := index.GetOrCreatePerson(candidateName)
		if isNew {
			fmt.Println("NEW", person.Name(), committeeRegNo, committeeName)
		} else {
			fmt.Println("OLD", person.Name(), committeeRegNo, committeeName)
		}
		count++
	}
	fmt.Printf("%d candidates added to people index out of %d candidate commitees\n", index.Size(), count)
}

func populateCandidacies(index *data.PersonIndex) {

	fmt.Println("Populating candidacies...")

	file, err := os.Open("data/Profiles_For_Hawaii_State_and_County_Candidates.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	count := 0
	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		count++
		//if count < 1000 {
		//	continue
		//}
		committeeRegNo := strings.Trim(fields[0], " ")
		candidateName := strings.Trim(fields[1], " ")
		office := strings.Trim(fields[2], " ")
		period := strings.Trim(fields[3], " ")
		person, isNew := index.GetOrCreatePerson(candidateName)
		//if isNew {
		//	fmt.Println(count, "NEW", person.Name(), committeeRegNo, office, period)
		//} else {
		//	fmt.Println(count, "OLD", person.Name(), committeeRegNo, office, period)
		//}
		if count > 10000 {
			fmt.Println(count, isNew, person.Name(), committeeRegNo, office, period)
		}
	}

	fmt.Printf("%d candidates added to people index out of %d registration entries\n\n", index.Size(), count)
}
