package bed

import (
	"testing"
	//"fmt"
)

func TestBinaryInsertABCD(t *testing.T) {

	binaryBTree := NewTree(2)
	binaryBTree.Insert("a")
	if binaryBTree.ToString() != "a\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("b")
	if binaryBTree.ToString() != "b0 <-- a\nb1 <-- b\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("c")
	if binaryBTree.ToString() != "c0:b0 <-- a\nc0:b1 <-- b\nc1:-0 <-- c\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("d")
	if binaryBTree.ToString() != "c0:b0 <-- a\nc0:b1 <-- b\nc1:d0 <-- c\nc1:d1 <-- d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}
}

func TestBinaryInsertABDC(t *testing.T) {

	binaryBTree := NewTree(2)
	binaryBTree.Insert("a")
	if binaryBTree.ToString() != "a\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("b")
	if binaryBTree.ToString() != "b0 <-- a\nb1 <-- b\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("d")
	if binaryBTree.ToString() != "d0:b0 <-- a\nd0:b1 <-- b\nd1:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("c")
	if binaryBTree.ToString() != "d0:c0:b0 <-- a\nd0:c0:b1 <-- b\nd0:c1:-0 <-- c\nd1:-0:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}
}

func TestBinaryInsertDCBA(t *testing.T) {
	binaryBTree := NewTree(2)
	binaryBTree.Insert("d")
	if binaryBTree.ToString() != "d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("c")
	if binaryBTree.ToString() != "d0 <-- c\nd1 <-- d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("b")
	if binaryBTree.ToString() != "d0:c0 <-- b\nd0:c1 <-- c\nd1:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}

	binaryBTree.Insert("a")
	if binaryBTree.ToString() != "d0:c0:b0 <-- a\nd0:c0:b1 <-- b\nd0:c1:-0 <-- c\nd1:-0:-0 <-- d\n" {
		t.Error("Not expected:", binaryBTree.ToString())
	}
}

func TestTertiaryInsertABCD(t *testing.T) {
	binaryBTree := NewTree(3)
	binaryBTree.Insert("a")
	if binaryBTree.ToString() != "a\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("b")
	if binaryBTree.ToString() != "ab\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("c")
	if binaryBTree.ToString() != "b0 <-- a\nb1 <-- bc\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("d")
	if binaryBTree.ToString() != "bc0 <-- a\nbc1 <-- b\nbc2 <-- cd\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}
}

func TestTertiaryInsertABDC(t *testing.T) {
	binaryBTree := NewTree(3)
	binaryBTree.Insert("a")
	if binaryBTree.ToString() != "a\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("b")
	if binaryBTree.ToString() != "ab\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("d")
	if binaryBTree.ToString() != "b0 <-- a\nb1 <-- bd\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("c")
	if binaryBTree.ToString() != "bd0 <-- a\nbd1 <-- bc\nbd2 <-- d\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}
}

func TestTertiaryInsertDCBA(t *testing.T) {
	binaryBTree := NewTree(3)

	binaryBTree.Insert("d")
	if binaryBTree.ToString() != "d\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("c")
	if binaryBTree.ToString() != "cd\n" {
		t.Error("Expected: cd, but was:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("b")
	if binaryBTree.ToString() != "d0 <-- bc\nd1 <-- d\n" {
		t.Error("Expected: d0 <-- bc, d1 <-- d, but was:\n", binaryBTree.ToString())
	}

	binaryBTree.Insert("a")
	if binaryBTree.ToString() != "cd0 <-- ab\ncd1 <-- c\ncd2 <-- d\n" {
		t.Error("Not expected:\n", binaryBTree.ToString())
	}
}
