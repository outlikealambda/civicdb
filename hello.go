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
	"strconv"
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
	candidateCommittees := make(map[string]*data.CandidateCommittee)

	log.Println("Connecting to db...")
	graph := data.ConnectGraphDb()

	log.Println("Cleaning old data...")
	graph.Clean()

	log.Println("Populating candidate committees...")
	populateCandidateCommitees(peopleIdx, candidateCommittees, graph)

	log.Println("Populating races...")
	populateCandidacies(candidateCommittees, graph)

	log.Println("Populating contributions...")
	populateContributions(peopleIdx, candidateCommittees, graph)
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

func populateContributions(personIdx *data.PersonIndex, candidateCommittees map[string]*data.CandidateCommittee, graph *data.Neo4jConnection) {

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

	//contributions := make([]*data.Contribution, 0)
	//organizations := make(map[int]*data.Organization)
	//pacs := make(map[int]*data.Committee)
	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		//officeName := fields[16]
		//if officeName != "Governor" {
		//	continue
		//}

		contributorType := fields[1]
		isPerson := false
		if contributorType == "Immediate Family" || contributorType == "Candidate" || contributorType == "Individual" {
			isPerson = true
		}

		if !isPerson {
			continue
		}

		amount, err := strconv.Atoi(fields[4])
		if err != nil {
			amount = 0
		}
		regNo := fields[20]
		committee := candidateCommittees[regNo]
		period := fields[21]
		contribution := data.NewContribution(&committee.Committee, amount, period)

		toCheck := fields[2]
		person, isNew := personIdx.GetOrCreatePerson(toCheck)
		if isNew {
			countPersonUnique++
		}

		contribution.SetContributor(person, contributorType)
		graph.PopulateGraphWithPersonContribution(contribution)
		//contributions = append(contributions, contribution)
		countPersonTotal++

		if countPersonTotal%1000 == 0 {
			durationSoFarNano := time.Now().Sub(startTS)
			fmt.Printf("%d unique out of %d processed so far [%dms total, %dms inserting]\n", countPersonUnique, countPersonTotal, durationSoFarNano/time.Millisecond, personIdx.InsertTimeSpent()/time.Millisecond)
		}
	}

	durationNano := time.Now().Sub(startTS)
	fmt.Printf("%d unique out of %d processed TOTAL [%dms total, %dms inserting]\n", countPersonUnique, countPersonTotal, durationNano/time.Millisecond, personIdx.InsertTimeSpent()/time.Millisecond)
	//fmt.Printf("%d branch factor: %d num nodes in firstName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, firstNameTree.NumTotalNodes(), firstNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)
	//fmt.Printf("%d branch factor: %d num nodes in lastName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, lastNameTree.NumTotalNodes(), lastNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)

	//fmt.Println(firstNameTree.String())
	//fmt.Println(lastNameTree.String())
}

func populateCandidateCommitees(index *data.PersonIndex, candidateCommittees map[string]*data.CandidateCommittee, graph *data.Neo4jConnection) {

	file, err := os.Open("data/Organizational_Reports_For_Hawaii_State_and_County_Candidates.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	races := make(map[string]*data.Office)

	count := 0
	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		committeeName := strings.Trim(fields[2], " ")
		committeeRegNo := strings.Trim(fields[0], " ")
		officeName := strings.Trim(fields[23], " ")
		district := strings.Trim(fields[24], " ")
		county := strings.Trim(fields[25], " ")
		valid := true
		if officeName != "Mayor" && officeName != "Prosecuting Attorney" && officeName != "Governor" && officeName != "Lt. Governor" {
			if district == "" {
				log.Println(committeeName, committeeRegNo, officeName, district, county)
				valid = false
			}
		} else if officeName != "Governor" && officeName != "Lt. Governor" {
			if county == "" {
				log.Println(committeeName, committeeRegNo, officeName, district, county)
				valid = false
			}
		}

		var office *data.Office
		if valid {
			raceKey := officeName + ":" + district + ":" + county
			office, exists := races[raceKey]
			if !exists {
				office = data.NewOffice(officeName, "HI", district, county)
				races[raceKey] = office
			}
		}

		candidateName := strings.Trim(fields[1], " ")
		candidate, _ := index.GetOrCreatePerson(candidateName)

		chairpersonName := strings.Trim(fields[9], " ")
		chairperson, _ := index.GetOrCreatePerson(chairpersonName)

		treasurerName := strings.Trim(fields[16], " ")
		treasurer, _ := index.GetOrCreatePerson(treasurerName)

		committee := data.NewCandidateCommittee(committeeRegNo, committeeName, candidate, chairperson, treasurer, office)
		candidateCommittees[committeeRegNo] = committee

		if graph != nil {
			graph.AddCandidateCommittee(committee)
		}
		//if isNew {
		//	fmt.Println("NEW", candidate.Name(), committeeRegNo, committeeName)
		//} else {
		//	fmt.Println("OLD", candidate.Name(), committeeRegNo, committeeName)
		//}
		count++
	}
	fmt.Printf("%d candidates added to people index out of %d candidate commitees\n", index.Size(), count)
}

func populateCandidacies(candidateCommittees map[string]*data.CandidateCommittee, graph *data.Neo4jConnection) {

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

		candidateCommittee := candidateCommittees[committeeRegNo]
		officeName := strings.Trim(fields[2], " ")
		period := strings.Trim(fields[3], " ")
		if candidateCommittee == nil {
			log.Println(committeeRegNo, candidateName, officeName, period)
			continue
		}
		//candidacy := data.NewCandidacy(candidateCommittee.Candidate, candidateCommittee.Race)

		// should I verify the name?
		if candidateCommittee.Candidate.Name() != candidateName {
			fmt.Println(candidateName, "ne committee filing", candidateCommittee.Candidate.Name())
		}

		var office *data.Office
		if candidateCommittee.Race != nil {
			if officeName != candidateCommittee.Race.Title {
				//fmt.Println(candidateName, "office diff than committee filing", officeName, "ne", candidateCommittee.Race.Title)
				office = data.NewOffice(officeName, "", "", "")
			}
		} else if officeName != "" {
			//fmt.Println(candidateName, "office nil in committee filing", officeName)
			office = data.NewOffice(officeName, "", "", "")
		}

		if graph != nil {
			graph.AddCandidacy(candidateCommittee, office, period)
		}
	}

	fmt.Printf("%d candidacies loaded\n\n", count)
}
