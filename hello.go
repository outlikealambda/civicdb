package main

import (
	"encoding/csv"
	"fmt"
	"github.com/peg-one/civicdb/data"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	start := time.Now()

	peopleIdx := data.NewPersonIndex()
	candidateCommitteesByRegNo := make(map[string]*data.CandidateCommittee)
	candidateCommitteesByName := make(map[string]*data.CandidateCommittee)
	races := make(map[string]*data.Office)
	nContributions := 0

	log.Println("Connecting to db...")
	graph := data.ConnectGraphDb()

	log.Println("Cleaning old data...")
	graph.Clean()

	log.Println("Populating candidate committees...")
	populateCandidateCommitees(peopleIdx, candidateCommitteesByRegNo, candidateCommitteesByName, races, graph)

	log.Println("Populating races...")
	populateCandidacies(candidateCommitteesByRegNo, graph)

	log.Println("Populating candidate contributions by people...")
	nContributions = populateCandidateContributionsByPeople(peopleIdx, candidateCommitteesByRegNo, races, graph)

	nonCandidateCommittees := make(map[string]*data.NonCandidateCommittee)

	log.Println("Populating non-candidate committees...")
	populateNonCandidateCommittees(peopleIdx, nonCandidateCommittees, graph)

	log.Println("Populating non-candidate contributions by people...")
	nContributions = populateNonCandidateContributionsByPeople(peopleIdx, nonCandidateCommittees, nContributions, graph)

	// TODO: too many candidate committee name errors... need to line up with received list
	//log.Println("Population candidate contributions by non-candidate committees...")
	//nContributions = populateCandidateContributionsByNonCandidateCommittees(nonCandidateCommittees, candidateCommitteesByName, 0, graph)

	log.Printf("Found %d people and %d contributions\n", peopleIdx.Size(), nContributions)
	log.Printf("FINISHED [%dms]\n", time.Now().Sub(start)/time.Millisecond)
	return
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

func parseDollars(dollarString string) int {
	if strings.HasPrefix(dollarString, "$") {
		dollarString = dollarString[1:]
	}

	dollarString = strings.Replace(dollarString, ".", "", -1)
	cents, err := strconv.Atoi(dollarString)
	if err != nil {
		log.Println("Could not parse: ", dollarString)
		cents = 0
	}

	// forget the cents?
	return cents / 100
}

func populateCandidateContributionsByPeople(personIdx *data.PersonIndex, candidateCommittees map[string]*data.CandidateCommittee, races map[string]*data.Office, graph *data.Neo4jConnection) int {

	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	startTS := time.Now()
	countPersonTotal := 0
	countPersonUnique := 0

	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		contributorType := fields[1]
		isPerson := false
		if contributorType == "Immediate Family" || contributorType == "Candidate" || contributorType == "Individual" {
			isPerson = true
		}

		if !isPerson {
			continue
		}

		amount := parseDollars(strings.Trim(fields[4], " "))
		aggregate := parseDollars(strings.Trim(fields[5], " "))
		regNo := fields[20]
		committee := candidateCommittees[regNo]

		//var office *data.Office
		officeName := strings.Trim(fields[16], " ")
		district := strings.Trim(fields[17], " ")
		county := strings.Trim(fields[18], " ")
		period := strings.Trim(fields[21], " ")

		if IsOfficeInfoValid(officeName, district, county) {
			raceKey := officeName + ":" + district + ":" + county
			office, exists := races[raceKey]
			if !exists {
				office = data.NewOffice(len(races), officeName, "HI", district, county)
				races[raceKey] = office
			}

			if committee.Race == nil || committee.Race.Title != office.Title || committee.Race.District != office.District || committee.Race.County != office.County {
				log.Println("CONTRIBUTION TO DIFFERENT OFFICE AS CAND COMMITTEE ON FILE")
				log.Printf("%v\n", committee.Race)
				log.Println("    ", officeName, district, county)
				//if committee.GetOtherOfficeForTerm(period) == office.Title {
				//	log.Println("FOUND MISSING INFO:", committee.Candidate.Name(), period, office.Title, office.District, office.County)
				//}
			}
		} else {
			log.Println("INVALID office for CONTRIBUTION", committee.Name(), regNo, officeName, district, county)
		}

		contribution := data.NewContribution(countPersonTotal, &committee.Committee, amount, aggregate, period)

		toCheck := fields[2]
		parsedAddress := strings.Split(fields[22], "\n")
		if len(parsedAddress) < 3 {
			continue
		}
		person, isNew := personIdx.ExtractAndGetOrCreatePerson(toCheck, parsedAddress[2])
		if isNew {
			countPersonUnique++
		}

		contribution.SetContributor(person, contributorType)
		if graph != nil {
			graph.AddContribution(contribution)
		}
		countPersonTotal++

		if countPersonTotal%1000 == 0 {
			durationSoFarNano := time.Now().Sub(startTS)
			log.Printf("%d unique out of %d processed so far [%dms total, %dms inserting]\n", countPersonUnique, countPersonTotal, durationSoFarNano/time.Millisecond, personIdx.InsertTimeSpent()/time.Millisecond)
		}
	}

	durationNano := time.Now().Sub(startTS)
	log.Printf("%d unique out of %d processed TOTAL [%dms total, %dms inserting]\n", countPersonUnique, countPersonTotal, durationNano/time.Millisecond, personIdx.InsertTimeSpent()/time.Millisecond)
	//fmt.Printf("%d branch factor: %d num nodes in firstName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, firstNameTree.NumTotalNodes(), firstNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)
	//fmt.Printf("%d branch factor: %d num nodes in lastName b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, lastNameTree.NumTotalNodes(), lastNameTree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)

	//fmt.Println(firstNameTree.String())
	//fmt.Println(lastNameTree.String())
	return countPersonTotal
}

func populateCandidateCommitees(index *data.PersonIndex, candidateCommitteesByRegNo map[string]*data.CandidateCommittee, candidateCommitteesByName map[string]*data.CandidateCommittee, races map[string]*data.Office, graph *data.Neo4jConnection) {

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

		committeeName := strings.Trim(fields[2], " ")
		committeeRegNo := strings.Trim(fields[0], " ")
		officeName := strings.Trim(fields[23], " ")
		district := strings.Trim(fields[24], " ")
		county := strings.Trim(fields[25], " ")

		var office *data.Office
		if IsOfficeInfoValid(officeName, district, county) {
			raceKey := officeName + ":" + district + ":" + county
			var exists bool
			office, exists = races[raceKey]
			if !exists {
				office = data.NewOffice(len(races), officeName, "HI", district, county)
				races[raceKey] = office
			}
		} else {
			log.Println("INVALID office description", committeeName, committeeRegNo, officeName, district, county)
		}

		candidateName := strings.Trim(fields[1], " ")
		candidate, _ := index.ExtractAndGetOrCreatePerson(candidateName, "")

		chairpersonName := strings.Trim(fields[9], " ")
		chairperson, _ := index.ExtractAndGetOrCreatePerson(chairpersonName, "")

		treasurerName := strings.Trim(fields[16], " ")
		treasurer, _ := index.ExtractAndGetOrCreatePerson(treasurerName, "")

		party := strings.Trim(fields[26], " ")
		terminated := strings.Trim(fields[27], " ") == "Y"
		inOffice := strings.Trim(fields[28], " ") == "1"

		committee := data.NewCandidateCommittee(committeeRegNo, committeeName, candidate, chairperson, treasurer, office, party, terminated, inOffice)
		candidateCommitteesByRegNo[committeeRegNo] = committee

		if old, exists := candidateCommitteesByName[committeeName+":"+officeName+":"+district+":"+county]; exists {
			log.Printf("NON-UNIQUE Commitee name: %v:%v:%v:%v (new: %v, existing: %v)", committeeName, officeName, district, county, committeeRegNo, old.RegNo)
		}

		candidateCommitteesByName[committeeName+":"+officeName+":"+district+":"+county] = committee

		if graph != nil {
			graph.AddCandidateCommittee(committee)
		}
		count++
	}
	log.Printf("%d candidates added to people index out of %d candidate commitees\n", index.Size(), count)
}

func IsOfficeInfoValid(title string, district string, county string) bool {
	if title != "Mayor" && title != "Prosecuting Attorney" && title != "Governor" && title != "Lt. Governor" {
		if district == "" {
			return false
		}
	} else if title != "Governor" && title != "Lt. Governor" {
		if county == "" {
			return false
		}
	}
	return true
}

func populateCandidacies(candidateCommittees map[string]*data.CandidateCommittee, graph *data.Neo4jConnection) {

	log.Println("Populating candidacies...")

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
			log.Println("Candidate committee not found!", committeeRegNo, candidateName, officeName, period)
			continue
		}
		//candidacy := data.NewCandidacy(candidateCommittee.Candidate, candidateCommittee.Race)

		// should I verify the name?
		if candidateCommittee.Candidate.Name() != candidateName {
			log.Println(candidateName, "ne committee filing", candidateCommittee.Candidate.Name())
		}

		var office *data.Office
		if candidateCommittee.Race != nil {
			if officeName != candidateCommittee.Race.Title {
				log.Println(candidateName, "office diff than committee filing", officeName, "ne", candidateCommittee.Race.Title)
				//candidateCommittee.AddOtherOffice(period, officeName)
			} else {
				office = candidateCommittee.Race
			}
		} else if officeName != "" {
			//fmt.Println(candidateName, "office diff than committee filing", officeName, "not nil")
			//candidateCommittee.AddOtherOffice(period, officeName)
		}

		if graph != nil {
			graph.AddCandidacy(candidateCommittee, office, period)
		}
	}

	log.Printf("%d candidacies loaded\n\n", count)
}

func populateNonCandidateCommittees(index *data.PersonIndex, nonCandidateCommittees map[string]*data.NonCandidateCommittee, graph *data.Neo4jConnection) {

	file, err := os.Open("data/Organizational_Reports_For_Hawaii_Noncandidate_Committees.csv")
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

		committeeName := strings.Trim(fields[1], " ")
		committeeRegNo := strings.Trim(fields[0], " ")

		chairpersonName := strings.Trim(fields[7], " ")
		chairperson, _ := index.ExtractAndGetOrCreatePerson(chairpersonName, "")

		treasurerName := strings.Trim(fields[16], " ")
		treasurer, _ := index.ExtractAndGetOrCreatePerson(treasurerName, "")

		nctype := strings.Trim(fields[25], " ")
		terminated := strings.Trim(fields[30], " ") == "Y"

		committee := data.NewNonCandidateCommittee(committeeRegNo, committeeName, chairperson, treasurer, nctype, terminated)
		nonCandidateCommittees[committeeRegNo] = committee

		area := strings.Trim(fields[26], " ")
		party := strings.Trim(fields[27], " ")
		issue := strings.Trim(fields[28], " ")
		committee.SetFocus(area, party, issue)

		singleCandidateName := strings.Trim(fields[29], " ")
		if singleCandidateName != "" {
			singleCandidate, _ := index.ExtractAndGetOrCreatePerson(singleCandidateName, "")
			committee.SetSingleCandidate(singleCandidate)
		}

		if graph != nil {
			graph.AddNonCandidateCommittee(committee)
		}

		count++
	}
	log.Printf("%d non-candidate commitees found\n", count)
}

func populateNonCandidateContributionsByPeople(personIdx *data.PersonIndex, nonCandidateCommittees map[string]*data.NonCandidateCommittee, startId int, graph *data.Neo4jConnection) int {

	file, err := os.Open("data/Contributions_Received_By_Hawaii_Noncandidate_Committees_From_January_1__2008_Through_December_31__2013.csv")
	// file, err := os.Open("data/kudo.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	startTS := time.Now()
	contributionsProcessed := startId

	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		contributorType := fields[1]
		isPerson := false
		if contributorType == "Immediate Family" || contributorType == "Candidate" || contributorType == "Individual" {
			isPerson = true
		}

		if !isPerson {
			continue
		}

		amount := parseDollars(strings.Trim(fields[4], " "))
		aggregate := parseDollars(strings.Trim(fields[5], " "))
		regNo := strings.Trim(fields[16], " ")
		period := strings.Trim(fields[17], " ")
		committee := nonCandidateCommittees[regNo]

		contribution := data.NewContribution(contributionsProcessed, &committee.Committee, amount, aggregate, period)

		toCheck := fields[2]
		parsedAddress := strings.Split(fields[18], "\n")
		if len(parsedAddress) < 3 {
			continue
		}

		person, _ := personIdx.ExtractAndGetOrCreatePerson(toCheck, parsedAddress[2])
		contribution.SetContributor(person, contributorType)

		if graph != nil {
			graph.AddContribution(contribution)
		}

		contributionsProcessed++
		if (contributionsProcessed-startId)%1000 == 0 {
			log.Printf("%d non-candidate contributions processed so far [%dms]\n", (contributionsProcessed - startId), time.Now().Sub(startTS)/time.Millisecond)
		}
	}

	log.Printf("%d non-candidate contributions processed TOTAL [%dms]\n", (contributionsProcessed - startId), time.Now().Sub(startTS)/time.Millisecond)
	return contributionsProcessed
}

func populateCandidateContributionsByNonCandidateCommittees(nonCandidateCommittees map[string]*data.NonCandidateCommittee, candidateCommitteesByName map[string]*data.CandidateCommittee, startId int, graph *data.Neo4jConnection) int {

	file, err := os.Open("data/Campaign_Contributions_Made_To_Candidates_By_Hawaii_Noncandidate_Committees_From_January_1__2008_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	startTS := time.Now()
	contributionsProcessed := startId

	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		amount := parseDollars(strings.Trim(fields[4], " "))
		aggregate := parseDollars(strings.Trim(fields[5], " "))
		period := strings.Trim(fields[15], " ")

		nonCandidateRegNo := strings.Trim(fields[14], " ")

		nonCandidateCommittee := nonCandidateCommittees[nonCandidateRegNo]

		candidateCommitteeName := strings.Trim(fields[2], " ")
		officeName := strings.Trim(fields[17], " ")
		district := strings.Trim(fields[18], " ")
		county := strings.Trim(fields[19], " ")
		candidateCommittee := candidateCommitteesByName[candidateCommitteeName+":"+officeName+":"+district+":"+county]
		if candidateCommittee == nil {
			log.Println("Could not find candidate committee by name:", candidateCommitteeName, officeName, district, county)
			continue
		}

		contribution := data.NewContribution(contributionsProcessed, &candidateCommittee.Committee, amount, aggregate, period)
		contribution.SetContributor(nonCandidateCommittee, "Noncandidate Committee")

		if graph != nil {
			graph.AddContribution(contribution)
		}

		contributionsProcessed++
		if (contributionsProcessed-startId)%1000 == 0 {
			log.Printf("%d non-candidate contributions processed so far [%dms]\n", (contributionsProcessed - startId), time.Now().Sub(startTS)/time.Millisecond)
		}
	}

	log.Printf("%d non-candidate contributions processed TOTAL [%dms]\n", (contributionsProcessed - startId), time.Now().Sub(startTS)/time.Millisecond)
	return contributionsProcessed
}
