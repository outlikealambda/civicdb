package data

import (
	"testing"
)

func TestScoreLastNameDistance(t *testing.T) {
	score := scoreLastNameDistance(0.1)
	if score != 0.5 {
		t.Errorf("Expecting score of 1.0, got: %v", score)
	}
}

func TestScoreFirstNameDistance(t *testing.T) {
	score := scoreFirstNameDistance(0.1)
	if score != 0.25 {
		t.Errorf("Expecting score of 1.0, got: %v ", score)
	}
}

func TestNormalizeNameString(t *testing.T) {
	testString := "A very long name with initial K."
	testString = normalizeNameString(testString)

	if testString != "averylongnamewithinitialk" {
		t.Errorf("didn't normalize string correctly: %v", testString)
	}
}

func TestSortMatchedPersons(t *testing.T) {
	matchedScores := make(map[int]float64)

	matchedScores[1] = 0.1
	matchedScores[2] = 0.3
	matchedScores[3] = 0.2

	sortedScores := sortMatchedPersons(matchedScores)

	if sortedScores[0].score != 0.3 || sortedScores[0].key != 2 || sortedScores[1].score != 0.2 {
		t.Errorf("%v, %v, %v", sortedScores[0], sortedScores[1], sortedScores[2])
	}

}
