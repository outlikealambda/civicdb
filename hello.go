package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

type Person struct {
	names []string
}

func main() {
	fmt.Printf("hello, world\n")

	file, err := os.Open("Campaign_Contributions_Made_To_Candidates_By_Hawaii_Noncandidate_Committees_From_January_1__2008_Through_December_31__2013.csv")
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)

	m := make(map[string]Person)
	for i := 0; i < 10; i++ {
		fields, err := csvReader.Read()
		if err != nil {
			log.Fatal(err)
		}

		ncCommitteeName := fields[0]
		candidateName := fields[1]
		fmt.Printf("committee name: %s, candidate name: %s\n", ncCommitteeName, candidateName)
	}
}
