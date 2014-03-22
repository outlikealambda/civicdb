package damerau

import (
	"math"
	"strings"
)

// DamerauLevenshteinDistance is damerau levenshtein distance, it computes the
// edit distance between two strings with transposition
// the measurements are: deletion, insertion, substituion, transposition
func DamerauLevenshteinDistance(s, t string) int {

	// convert two strings to arrays
	source := strings.Split(s, "")
	target := strings.Split(t, "")

	// obtain the length of the strings
	sourcelen, targetlen := len(source), len(target)

	// check for zero length strings
	if sourcelen == 0 {
		if targetlen == 0 {
			return 0 // if the length of both strings are zero
		} else {
			return targetlen // If only the length of source string is zero
		}
	} else if targetlen == 0 {
		return sourcelen // if only the length of target string is zero
	}

	// score is a two dimensional slice that keeps track of the distances
	score := make([][]int, sourcelen+2)
	for i := range score {
		score[i] = make([]int, targetlen+2) // making slice two dimensional
	}
	sumlen := sourcelen + targetlen // sum of lengths of two strings
	score[0][0] = sumlen
	// initialize all the score values in rows
	for i := 0; i <= sourcelen; i++ {
		score[i+1][1] = i
		score[i+1][0] = sumlen
	}
	// initialize all the score values in columns
	for j := 0; j <= targetlen; j++ {
		score[1][j+1] = j
		score[0][j+1] = sumlen
	}

	// sd is a map with keys as every letter (unique) in both strings
	sd := make(map[string]int)
	temp := source
	temp = append(temp, target...)
	// initialize the values for each key
	for i := range temp {
		_, ok := sd[temp[i]]
		if !ok {
			sd[temp[i]] = 0
		}
	}

	// loop through source and target strings and calculate distances
	for i := 1; i <= sourcelen; i++ {
		db := 0
		for j := 1; j <= targetlen; j++ {
			i1 := sd[target[j-1]]
			j1 := db

			if source[i-1] == target[j-1] {
				score[i+1][j+1] = score[i][j]
				db = j
			} else {
				score[i+1][j+1] = int(math.Min(float64(score[i][j]),
					math.Min(float64(score[i+1][j]),
						float64(score[i][j+1])))) + 1
			}
			score[i+1][j+1] = int(math.Min(float64(score[i+1][j+1]),
				float64(score[i1][j1]+(i-i1-1)+1+(j-j1-1))))
		}
		sd[source[i-1]] = i
	}
	return score[sourcelen+1][targetlen+1]
}
