package bed

import "testing"

func TestRangeQuery(t *testing.T) {

	tertiaryBTree := New(3, CompareDictionaryOrder)
	tertiaryBTree.Insert("Jim Gray")
	tertiaryBTree.Insert("Jim Grey")

	results := tertiaryBTree.RangeQuery("Jam Gray", 0.25)
	if len(results) != 2 {
		t.Error("Expect: 2 query results for distance threshold of 0.26", len(results))
	}

	// TODO
	if len(tertiaryBTree.RangeQuery("Jam Gray", 0.125)) != 1 {
		t.Error("Expect: 1 query result for distance threshold of 0.13", len(results))
	}

	if len(tertiaryBTree.RangeQuery("Jim Gray", 0)) != 1 {
		t.Error("Expect: 1 query result for exact match query with threshold of 1", len(results))
	}
}

func TestVerify(t *testing.T) {

	if success, distance := VerifyEditDistance("Jim Grey", "Jim Gay", 2); !success || distance != 2 {
		t.Error("Expect: Jim Grey and Jim Gay to be within 2 edits.  Distance was: %v", distance)
	}

	if success, distance := VerifyEditDistance("Jim Grey", "Jim Grey", 0); !success || distance != 0 {
		t.Error("Expect: Jim Grey to be within 0 edits of itself.  Distance was: %v", distance)
	}

	if success, distance := VerifyEditDistance("Jim Grey", "Jim Gay", 1); success || distance != -1 {
		t.Error("Expect: Jim Grey and Jim Gay to not be within 1 edit")
	}
}

func TestLowerBound(t *testing.T) {
	//if LowerBoundEst("Jam Gray", "Jim Gray", "Jim Grey") != 1 {
	//	t.Error("Expect: Jam Grey compared to interval [Jim Gray, Jim Grey] to have a lower bound estimate of 1")
	//}

	if !VerifyLowerBound("Jam Gray", "Jim Gray", "Jim Grey", 1) {
		t.Error("Expect: Jam Grey compared to interval [Jim Gray, Jim Grey] to PASS a lower bound estimate verification")
	}

	if VerifyLowerBound("Mark Egesdal", "Jim Gray", "Jim Grey", 1) {
		t.Error("Expect: Mark Egesdal compared to interval [Jim Gray, Jim Grey] to FAIL a lower bound estimate verification")
	}
}

func TestBinaryInsertABCD(t *testing.T) {

	binaryBTree := New(2, CompareDictionaryOrder)
	binaryBTree.Insert("a")
	if binaryBTree.String() != "a\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("b")
	if binaryBTree.String() != "b0 <-- a\nb1 <-- b\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("c")
	if binaryBTree.String() != "c0:b0 <-- a\nc0:b1 <-- b\nc1:-0 <-- c\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("d")
	if binaryBTree.String() != "c0:b0 <-- a\nc0:b1 <-- b\nc1:d0 <-- c\nc1:d1 <-- d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}
}

func TestBinaryInsertABDC(t *testing.T) {

	binaryBTree := New(2, CompareDictionaryOrder)
	binaryBTree.Insert("a")
	if binaryBTree.String() != "a\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("b")
	if binaryBTree.String() != "b0 <-- a\nb1 <-- b\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("d")
	if binaryBTree.String() != "d0:b0 <-- a\nd0:b1 <-- b\nd1:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("c")
	if binaryBTree.String() != "d0:c0:b0 <-- a\nd0:c0:b1 <-- b\nd0:c1:-0 <-- c\nd1:-0:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}
}

func TestBinaryInsertDCBA(t *testing.T) {
	binaryBTree := New(2, CompareDictionaryOrder)
	binaryBTree.Insert("d")
	if binaryBTree.String() != "d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("c")
	if binaryBTree.String() != "d0 <-- c\nd1 <-- d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("b")
	if binaryBTree.String() != "d0:c0 <-- b\nd0:c1 <-- c\nd1:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}

	binaryBTree.Insert("a")
	if binaryBTree.String() != "d0:c0:b0 <-- a\nd0:c0:b1 <-- b\nd0:c1:-0 <-- c\nd1:-0:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.String())
	}
}

func TestTertiaryInsertABCD(t *testing.T) {
	tertiaryBTree := New(3, CompareDictionaryOrder)
	tertiaryBTree.Insert("a")
	if tertiaryBTree.String() != "a\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("b")
	if tertiaryBTree.String() != "ab\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("c")
	if tertiaryBTree.String() != "b0 <-- a\nb1 <-- bc\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("d")
	if tertiaryBTree.String() != "bc0 <-- a\nbc1 <-- b\nbc2 <-- cd\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}
}

func TestTertiaryInsertABDC(t *testing.T) {
	tertiaryBTree := New(3, CompareDictionaryOrder)
	tertiaryBTree.Insert("a")
	if tertiaryBTree.String() != "a\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("b")
	if tertiaryBTree.String() != "ab\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("d")
	if tertiaryBTree.String() != "b0 <-- a\nb1 <-- bd\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("c")
	if tertiaryBTree.String() != "bd0 <-- a\nbd1 <-- bc\nbd2 <-- d\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}
}

func TestTertiaryInsertDCBA(t *testing.T) {
	tertiaryBTree := New(3, CompareDictionaryOrder)

	tertiaryBTree.Insert("d")
	if tertiaryBTree.String() != "d\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("c")
	if tertiaryBTree.String() != "cd\n" {
		t.Error("Expected: cd, but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("b")
	if tertiaryBTree.String() != "d0 <-- bc\nd1 <-- d\n" {
		t.Error("Expected: d0 <-- bc, d1 <-- d, but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Insert("a")
	if tertiaryBTree.String() != "cd0 <-- ab\ncd1 <-- c\ncd2 <-- d\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}
}

func TestTertiaryInsertDCBAWithData(t *testing.T) {
	tertiaryBTree := New(3, CompareDictionaryOrder)

	tertiaryBTree.Put("d", 1)
	if tertiaryBTree.String() != "d[1]\n" {
		t.Error("Expected: d[1], but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("c", 2)
	if tertiaryBTree.String() != "c[2]d[1]\n" {
		t.Error("Expected: c[2]d[1], but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("b", 3)
	if tertiaryBTree.String() != "d0 <-- b[3]c[2]\nd1 <-- d[1]\n" {
		t.Error("Expected: d0 <-- b[3]c[2], d1 <-- d[1], but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("a", 4)
	if tertiaryBTree.String() != "cd0 <-- a[4]b[3]\ncd1 <-- c[2]\ncd2 <-- d[1]\n" {
		t.Error("Not expected:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("a", 5)
	if tertiaryBTree.String() != "cd0 <-- a[4,5]b[3]\ncd1 <-- c[2]\ncd2 <-- d[1]\n" {
		t.Error("Expected: cd0 <-- a[4,5]b[3], cd1 <-- c[2], cd2 <-- d[1], but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("b", 6)
	if tertiaryBTree.String() != "cd0 <-- a[4,5]b[3,6]\ncd1 <-- c[2]\ncd2 <-- d[1]\n" {
		t.Error("Expected: cd0 <-- a[4,5]b[3,6], cd1 <-- c[2], cd2 <-- d[1]\n, but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("d", 7)
	if tertiaryBTree.String() != "cd0 <-- a[4,5]b[3,6]\ncd1 <-- c[2]\ncd2 <-- d[1,7]\n" {
		t.Error("Expected: cd0 <-- a[4,5]b[3,6], cd1 <-- c[2], cd2 <-- d[1,7], but was:\n", tertiaryBTree.String())
	}

	tertiaryBTree.Put("c", 8)
	if tertiaryBTree.String() != "cd0 <-- a[4,5]b[3,6]\ncd1 <-- c[2,8]\ncd2 <-- d[1,7]\n" {
		t.Error("Expected: cd0 <-- a[4,5]b[3,6], cd1 <-- c[2,8], cd2 <-- d[1,7], but was:\n", tertiaryBTree.String())
	}
}
