package main

import (
	"encoding/csv"
	"fmt"
	"github.com/megesdal/melodispurences/bed"
	//"github.com/megesdal/melodispurences/damerau"
	"io"
	"log"
	//"math"
	"os"
	//"sort"
	"time"
)

type Person struct {
	names []string
}

func main() {
	fmt.Printf("hello, world\n")

	file, err := os.Open("data/Campaign_Contributions_Received_By_Hawaii_State_and_County_Candidates_From_November_8__2006_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)

	startTS := time.Now()
	countTotal := 0
	countUnique := 0

	//m := []string{}
	//normThreshold := 0.2
	tree := bed.New(2000)

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

		results := tree.RangeQuery(toCheck, 5)
		if len(results) > 0 {
			//fmt.Printf("%v\n", results)
			found = true
		}

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
			tree.Insert(toCheck)
			//m = append(m, toCheck)
		}
		countTotal++

		if countTotal%1000 == 0 {
			durationSoFarNano := time.Now().Sub(startTS)
			fmt.Printf("%d unique out of %d processed so far [%dms]\n", countUnique, countTotal, durationSoFarNano/time.Millisecond)
		}
	}

	fmt.Printf("%d unique out of %d processed FINAL\n", countUnique, countTotal)
	//sort.Strings(m)
	//for i := 0; i < len(m); i++ {
	//	fmt.Printf("contributor name: %s\n", m[i])
	//}
	//fmt.Printf("\n%d unique contributors out of %d entries", len(m), count)
}
