package bed

func CompareDictionaryOrder(a string, b string) int {
	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}

	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}

	return 0
}

func CreateCompareEditDistance(normDist float64) func(string, string) int {

	return func(a string, b string) int {
		//if float64(damerau.DamerauLevenshteinDistance(a, b))/math.Max(float64(len(a)), float64(len(b))) < normDist {
		if VerifyEditDistance(a, b, int(normDist*float64(intMax(len(a), len(b))))) {
			return 0
		}

		return CompareDictionaryOrder(a, b)
	}
}

func VerifyEditDistance(s1 string, s2 string, distanceThreshold int) bool {

	if intAbs(len(s1)-len(s2)) > distanceThreshold {
		return false
	}

	_, result := createVerificationTable(len(s1), len(s2), distanceThreshold, func(rowIdx int, colIdx int) bool {
		return s1[rowIdx] == s2[colIdx]
	})

	return result
}

func VerifyLowerBound(q string, smin string, smax string, distanceThreshold int) bool {
	lcp := longestCommonPrefix(smin, smax)
	nlcp := len(lcp)
	var cmin uint8
	if len(smin) == nlcp {
		cmin = '!' // \u0021... if I trim, do I need to go lower?
	} else {
		cmin = smin[nlcp]
	}

	if len(smax) == nlcp {
		return intMax(nlcp, len(q)) <= distanceThreshold
	}
	cmax := smax[nlcp]

	lastRow, verified := createVerificationTable(nlcp+1, len(q), distanceThreshold, func(rowIdx int, colIdx int) bool {
		if rowIdx < len(lcp) {
			if lcp[rowIdx] != q[colIdx] {
				return false
			}
		} else {
			if q[colIdx] < cmin || q[colIdx] > cmax {
				return false
			}
		}
		return true
	})

	if !verified {
		return verified
	}

	min := intMax(len(q), nlcp+1)
	for k, n := 0, len(lastRow); k < n; k++ {
		if lastRow[k] < min {
			min = lastRow[k]
		}
	}

	return min <= distanceThreshold
}

func createVerificationTable(nrows int, ncols int, distanceThreshold int, compare func(int, int) bool) ([]int, bool) {

	// construct tabel of 2 rows and len(s2) + 1 columns
	table := make([][]int, 2)
	for i := 0; i < len(table); i++ {
		table[i] = make([]int, ncols+1)
	}
	// table[0][0] = 0 // empty string is 0 edits from itself

	firstEnd := intMin(ncols+1, 1+distanceThreshold)
	for j := 0; j < firstEnd; j++ {
		table[0][j] = j
	}
	/*
		headerRow := "  \u2205 "
		for k := 0; k < len(s2); k++ {
			headerRow += string(s2[k]) + " "
		}
		//fmt.Println(headerRow)
		row := fmt.Sprintf("\u2205 ")
		for k := 0; k < len(table[0]); k++ {
			if k < firstEnd {
				row += fmt.Sprintf("%d ", table[0][k])
			} else {
				row += "- "
			}
		}
		//fmt.Println(row)
	*/
	// i == 0 is the empty string... handled by init above
	for i := 1; i < nrows+1; i++ {
		start := intMax(0, i-distanceThreshold)
		end := intMin(ncols+1, i+distanceThreshold+1)
		m := distanceThreshold + 1
		//fmt.Println(start, end)
		for j := start; j < end; j++ {
			var d1 int
			if j < i+distanceThreshold {
				d1 = table[0][j] + 1
			} else {
				d1 = distanceThreshold + 1
			}

			var d2 int
			var d3 int
			if j > 0 {
				d2 = table[1][j-1] + 1
				d3 = table[0][j-1]
				//fmt.Printf("comparing %s vs. %s (prev d=%d [%d])\n", string(s1[i-1]), string(s2[j-1]), d3, j)
				if !compare(i-1, j-1) {
					d3 += 1
				}
			} else {
				d2 = distanceThreshold + 1
				d3 = distanceThreshold + 1
			}
			//fmt.Printf("%d %d %d\n", d1, d2, d3)
			table[1][j] = intMin(intMin(d1, d2), d3)

			m = intMin(m, table[1][j])
		}
		/*
			row := fmt.Sprintf("%v ", string(s1[i-1]))
			for k := 0; k < len(table[0]); k++ {
				if k >= start && k < end {
					row += fmt.Sprintf("%d ", table[1][k])
				} else {
					row += "- "
				}
			}
			//fmt.Println(row, "|", m)
		*/
		if m > distanceThreshold {
			return table[1], false
		}
		for j, n := 0, ncols+1; j < n; j++ {
			table[0][j] = table[1][j]
		}
	}
	return table[0], true
}

// longest common prefix
func longestCommonPrefix(smin string, smax string) string {
	lcp := make([]uint8, intMax(len(smin), len(smax)))
	n := 0
	for k, nmin, nmax := 0, len(smin), len(smax); k < nmin && k < nmax; k++ {
		if smin[k] == smax[k] {
			lcp[k] = smin[k]
			n++
		} else {
			break
		}
	}
	return string(lcp[:n])
}
