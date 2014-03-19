package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"io"
	"sort"
	"github.com/megesdal/melodispurences/damerau"
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

	m := []string{}
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
		found := false
		for j := 0; j < len(m); j++ {

			existing := m[j]
			toCheck := fields[2]
			distMin := float64(len(existing) + len(toCheck)) / float64(2.0)

			value := float64(damerau.DamerauLevenshteinDistance(m[j], fields[2])) / distMin
			if value < 0.2 {
			//if m[j] == fields[2] {
				if value > 0 {
					fmt.Printf("  %f is value: %s vs. %s\n", value, existing, toCheck)
				}
				found = true
				break
			}
		}

		if !found {
			m = append(m, fields[2])
		}
	}

	sort.Strings(m)
	for i := 0; i < len(m); i++ {
		fmt.Printf("contributor name: %s\n", m[i])
	}
	fmt.Printf("\n%d unique contributors out of %d entries", len(m), count)
}
