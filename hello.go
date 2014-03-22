package main

import (
	"encoding/csv"
	"fmt"
	"github.com/megesdal/melodispurences/address"
	"github.com/megesdal/melodispurences/bed"
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

	groupByAddress()
	simpleDamerau()

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

func simpleDamerau() {
	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)

	startTS := time.Now()
	countTotal := 0
	countUnique := 0

	//var queryDurationSum time.Duration
	var insertDurationSum time.Duration

	//m := []string{}
	//normThreshold := 0.2
	branchFactor := 32
	tree := bed.New(branchFactor, bed.CreateCompareEditDistance(0.1))

	for true {
		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		found := false
		toCheck := fields[2]
		//lenToCheck := float64(len(toCheck))

		//beforeQueryTS := time.Now()
		//results := tree.RangeQuery(toCheck, 5)
		//queryDurationSum += time.Now().Sub(beforeQueryTS)
		//if len(results) > 0 {
		//fmt.Printf("%v\n", results)
		//found = true
		//}

		/*for j := 0; j < len(m); j++ {

			existing := m[j]
			lenExisting := float64(len(existing))
			distMax := math.Max(lenExisting, lenToCheck)

			// if the minimum norm distance is greater than the threshold, don't bother
			if math.Abs(lenExisting-lenToCheck)/distMax > normThreshold {
				continue
			}

			normDist := float64(damerau.DamerauLevenshteinDistance(m[j], fields[2])) / distMax
			if normDist <= normThreshold {
				//if normDist > 0 {
				//	fmt.Printf("  %f is value: %s vs. %s\n", normDist, existing, toCheck)
				//}
				found = true
				break
			}
		}*/

		if !found {
			countUnique++
			beforeInsertTS := time.Now()
			tree.Insert(toCheck)
			insertDurationSum += time.Now().Sub(beforeInsertTS)
			//m = append(m, toCheck)
		}
		countTotal++

		//if countTotal%1000 == 0 {
		//	durationSoFarNano := time.Now().Sub(startTS)
		//	fmt.Printf("%d unique out of %d processed so far [%dms total, %dms inserting]\n", tree.Size(), countTotal, durationSoFarNano/time.Millisecond, insertDurationSum/time.Millisecond)
		//}
	}

	durationNano := time.Now().Sub(startTS)
	fmt.Printf("%d unique out of %d processed\n", tree.Size(), countTotal)
	fmt.Printf("%d branch factor: %d num nodes in b-tree index, of which %d are leaf nodes [%dms total, %dms inserting]\n", branchFactor, tree.NumTotalNodes(), tree.NumLeafNodes(), durationNano/time.Millisecond, insertDurationSum/time.Millisecond)
	//sort.Strings(m)
	//for i := 0; i < len(m); i++ {
	//	fmt.Printf("contributor name: %s\n", m[i])
	//}
	//fmt.Printf("\n%d unique contributors out of %d entries", len(m), count)
}
