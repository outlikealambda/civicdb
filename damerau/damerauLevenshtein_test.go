package damerau

import "testing"

func TestDamerauLevenshteinDistanceReplace(t *testing.T) {
	const inS, inT, outN = "cat", "hat", 1

	if x := DamerauLevenshteinDistance(inS, inT); x != outN {
		t.Errorf("DamerauLevenshteinDistance(%s, %s) = %v, want %v", inS, inT, x, outN)
	}
}

func TestDamerauLevenshteinDistanceAdd(t *testing.T) {
	const inS, inT, outN = "", "hat", 3

	if x := DamerauLevenshteinDistance(inS, inT); x != outN {
		t.Errorf("DamerauLevenshteinDistance(%s, %s) = %v, want %v", inS, inT, x, outN)
	}
}

func TestDamerauLevenshteinDistanceSwap(t *testing.T) {
	const inS, inT, outN = "cat", "act", 1

	if x := DamerauLevenshteinDistance(inS, inT); x != outN {
		t.Errorf("DamerauLevenshteinDistance(%s, %s) = %v, want %v", inS, inT, x, outN)
	}
}

func TestDamerauLevenshteinDistanceSwapTwo(t *testing.T) {
	const inS, inT, outN = "cat", "atc", 2

	if x := DamerauLevenshteinDistance(inS, inT); x != outN {
		t.Errorf("DamerauLevenshteinDistance(%s, %s) = %v, want %v", inS, inT, x, outN)
	}
}
