package bed

import (
	"fmt"
	//"github.com/megesdal/melodispurences/damerau"
	"math"
	"strings"
)

type BPlusTree struct {
	branchFactor int
	root         *BPlusTreeNode
}

type BPlusTreeNode struct {
	tree           *BPlusTree
	parent         *BPlusTreeNode
	parentChildIdx int
	splits         []string         // size m
	children       []*BPlusTreeNode // size m + 1
	data           [][]string       // size m x n (leaf only)
}

func intMax(a int, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func intMin(a int, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func intAbs(a int) int {
	return int(math.Abs(float64(a)))
}

func (tree *BPlusTree) NumTotalNodes() int {
	return recNumNodes(tree.root, func(*BPlusTreeNode) int {
		return 1
	})
}

func (tree *BPlusTree) NumLeafNodes() int {
	return recNumNodes(tree.root, func(node *BPlusTreeNode) int {
		if node.isLeafNode() {
			return 1
		}
		return 0
	})
}

func (tree *BPlusTree) Size() int {
	return recNumNodes(tree.root, func(node *BPlusTreeNode) int {
		if node.isLeafNode() {
			return len(node.splits)
		}
		return 0
	})
}

func recNumNodes(parent *BPlusTreeNode, count func(*BPlusTreeNode) int) int {
	total := count(parent)
	for _, child := range parent.children {
		total += recNumNodes(child, count)
	}
	return total
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

func New(b int) *BPlusTree {
	if b < 2 {
		return nil
	}

	tree := new(BPlusTree)
	tree.branchFactor = b
	tree.root = tree.createTreeNode()
	return tree
}

func (tree *BPlusTree) ToString() string {
	return recToString(tree.root)
}

func recToString(node *BPlusTreeNode) string {
	if node == nil {
		return ""
	}

	parentStr := ""
	if node.parent != nil {
		parentStr += fmt.Sprintf("%s%d <-- ", revToString(node.parent), node.parentChildIdx)
	}

	childStr := ""
	if node.children != nil && len(node.children) > 0 {
		childStr = " --> " + childrenStr(node)
	}

	rv := ""
	if node.isLeafNode() {
		rv += parentStr + strings.Join(node.splits, "") + childStr + "\n"
	} else {
		for i := 0; i < len(node.children); i++ {
			rv += recToString(node.children[i])
		}
	}
	return rv
}

func (tree *BPlusTree) createTreeNode() *BPlusTreeNode {
	node := new(BPlusTreeNode)
	node.tree = tree
	node.splits = make([]string, 0, tree.branchFactor-1)
	node.children = make([]*BPlusTreeNode, 0, tree.branchFactor)
	return node
}

func revToString(node *BPlusTreeNode) string {
	rv := ""
	first := true
	lastIdx := -1
	for node != nil {
		if first {
			first = false
		} else {
			rv = ":" + rv
		}

		if node.isLeafNode() {
			rv = "*" + rv
		}

		if lastIdx >= 0 {
			rv = fmt.Sprintf("%d", lastIdx) + rv
		}
		if len(node.splits) > 0 {
			rv = strings.Join(node.splits, "") + rv
		} else {
			rv = "-" + rv
		}

		lastIdx = node.parentChildIdx
		node = node.parent
	}
	return rv
}

func (tree *BPlusTree) Insert(q string) {
	nodeToInsert := tree.FindNode(q)
	// TODO: the find step should be able to say if we are inserting a new node or merging with an existing
	// can take in a distance threshold on insert to
	//fmt.Printf("Found node where value %s belongs: %s\n", q, revToString(nodeToInsert))
	recInsert(q, nodeToInsert, nil)
}

func (parent *BPlusTreeNode) AddChildNode(q string, child *BPlusTreeNode) {
	recInsert(q, parent, child)
}

func addToParentNode(parent *BPlusTreeNode, child *BPlusTreeNode) {
	if child != nil {
		nodeWithSplitValue := child
		for len(nodeWithSplitValue.splits) == 0 {
			//fmt.Printf("looking for splitValue %s\n", childrenStr(nodeWithSplitValue))
			nodeWithSplitValue = nodeWithSplitValue.children[0]
		}

		splitValue := nodeWithSplitValue.splits[0]
		//fmt.Printf("found split value at %s\n", revToString(nodeWithSplitValue))
		//fmt.Printf("adding newNode %s to parent %s\n", splitValue, revToString(parent))
		if parent == nil {
			//fmt.Println("old root...", revToString(child.tree.root))
			newRoot := child.tree.createTreeNode()
			newRoot.children = append(newRoot.children, child.tree.root)
			newRoot.children = append(newRoot.children, child)
			newRoot.splits = append(newRoot.splits, splitValue)
			child.tree.root.parent = newRoot
			child.tree.root.parentChildIdx = 0
			child.parent = newRoot
			child.parentChildIdx = 1
			child.tree.root = newRoot
			//fmt.Println("new root...", revToString(newRoot.children[0]), revToString(newRoot.children[1]))
		} else {
			parent.AddChildNode(splitValue, child)
		}
	}
}

func childrenStr(node *BPlusTreeNode) string {
	rv := make([]string, len(node.children))
	for i := 0; i < len(node.children); i++ {
		childSplits := node.children[i].splits
		if len(childSplits) > 0 {
			rv[i] = strings.Join(childSplits, "")
		} else {
			rv[i] = "-"
		}
	}
	return strings.Join(rv, ",")
}

// TODO: figure out why this isn't working
// TODO: I'd like to get this so it didn't depend on q and addToNew is computed outside
/*func splitIfNecessary(parent *BPlusTreeNode, q string) (*BPlusTreeNode, bool) {

	shouldSplit := false
	if len(parent.splits) == parent.tree.branchFactor-1 {
		shouldSplit = true
	}

	if shouldSplit {
		return nil, false
	}

	//fmt.Println(indent, "Splitting before adding child node")
	newNode := parent.tree.createTreeNode()
	addToNew := false

	//childSplitIdx := int(math.Ceil(float64(parent.tree.branchFactor) / 2))
	splitIdx := int(math.Floor(float64(parent.tree.branchFactor) / 2))
	if parent.tree.branchFactor == 2 {
		// only one of the two nodes will have room...
		if compare(q, parent.splits[0]) < 0 {
			//fmt.Println(indent, "decrementing splitIdx", splitIdx, splitIdx - 1)
			splitIdx--
		} else {
			addToNew = true
		}
	}

	if splitIdx < len(parent.splits) {
		//fmt.Printf("%s splitting '%s' at %d\n", indent, strings.Join(parent.splits, ""), splitIdx)

		if parent.isLeafNode() {
			newNode.splits = make([]string, len(parent.splits)-splitIdx)
			copy(newNode.splits, parent.splits[splitIdx:])
		} else {
			newNode.splits = make([]string, len(parent.splits)-splitIdx-1)
			copy(newNode.splits, parent.splits[splitIdx+1:])
		}

		if compare(q, parent.splits[splitIdx]) >= 0 {
			addToNew = true
		}
		parent.splits = parent.splits[:splitIdx]
	} // else is branchFactor = 2 (binary tree)

	//fmt.Printf("%s split: '%s' %d | '%s' %d \n", indent, strings.Join(parent.splits, ""), len(parent.splits), strings.Join(newNode.splits, ""), len(newNode.splits))

	if len(parent.children) > 0 {
		childSplitIdx := splitIdx + 1
		//fmt.Println(indent, "child split index", childSplitIdx)
		newNode.children = make([]*BPlusTreeNode, len(parent.children)-childSplitIdx)
		copy(newNode.children, parent.children[childSplitIdx:])
		for i := 0; i < len(newNode.children); i++ {
			newNode.children[i].parent = newNode
			newNode.children[i].parentChildIdx = i
		}
		parent.children = parent.children[:childSplitIdx]
		//fmt.Println(indent, "children:", childrenStr(parent), len(parent.children), "|", childrenStr(newNode), len(newNode.children))
	}

	return newNode, addToNew
}*/

func recInsert(q string, parent *BPlusTreeNode, child *BPlusTreeNode) {

	indent := ""
	node := parent.parent
	for node != nil {
		indent += "  "
		node = node.parent
	}

	nodeToInsert := parent
	//fmt.Println(indent, "Parent node", revToString(parent))
	var newNode *BPlusTreeNode

	shouldSplit := false
	if len(nodeToInsert.splits) == parent.tree.branchFactor-1 {
		shouldSplit = true
	}

	if shouldSplit {

		//fmt.Println(indent, "Splitting before adding child node")
		newNode = parent.tree.createTreeNode()
		//childSplitIdx := int(math.Ceil(float64(parent.tree.branchFactor) / 2))
		splitIdx := int(math.Floor(float64(parent.tree.branchFactor) / 2))

		addToNew := false
		if parent.tree.branchFactor == 2 {
			// only one of the two nodes will have room...
			if compare(q, parent.splits[0]) < 0 {
				//fmt.Println(indent, "decrementing splitIdx", splitIdx, splitIdx - 1)
				splitIdx--
			} else {
				addToNew = true
			}
		}

		if splitIdx < len(nodeToInsert.splits) {
			//fmt.Printf("%s splitting '%s' at %d\n", indent, strings.Join(nodeToInsert.splits, ""), splitIdx)

			if nodeToInsert.isLeafNode() {
				newNode.splits = make([]string, len(nodeToInsert.splits)-splitIdx)
				copy(newNode.splits, nodeToInsert.splits[splitIdx:])
			} else {
				newNode.splits = make([]string, len(nodeToInsert.splits)-splitIdx-1)
				copy(newNode.splits, nodeToInsert.splits[splitIdx+1:])
			}

			if compare(q, nodeToInsert.splits[splitIdx]) >= 0 {
				addToNew = true
			}
			nodeToInsert.splits = nodeToInsert.splits[:splitIdx]
		} // else is branchFactor = 2 (binary tree)

		//fmt.Printf("%s split: '%s' %d | '%s' %d \n", indent, strings.Join(nodeToInsert.splits, ""), len(nodeToInsert.splits), strings.Join(newNode.splits, ""), len(newNode.splits))

		if child != nil {
			childSplitIdx := splitIdx + 1
			//fmt.Println(indent, "child split index", childSplitIdx)
			newNode.children = make([]*BPlusTreeNode, len(nodeToInsert.children)-childSplitIdx)
			copy(newNode.children, nodeToInsert.children[childSplitIdx:])
			for i := 0; i < len(newNode.children); i++ {
				newNode.children[i].parent = newNode
				newNode.children[i].parentChildIdx = i
			}
			nodeToInsert.children = nodeToInsert.children[:childSplitIdx]
			//fmt.Println(indent, "children:", childrenStr(parent), len(parent.children), "|", childrenStr(newNode), len(newNode.children))
		}

		if addToNew {
			nodeToInsert = newNode
			//fmt.Printf("Now adding %s to new node\n", q)
		}
	}

	//fmt.Printf("%s Finding insertIdx for %s among %s\n", indent, q, revToString(nodeToInsert))
	if child != nil {
		//fmt.Printf("%s   Child is %s\n", indent, revToString(child))
	} else {
		//fmt.Printf("%s   No Child\n", indent)
	}

	insertIdx := len(nodeToInsert.splits)
	//fmt.Printf("Initial insert index to insert %s is %d\n", q, insertIdx)

	performInsert := true
	for i, n := 0, len(nodeToInsert.splits); i < n; i++ {
		//fmt.Printf("Comparing '%s' to '%s'\n", q, nodeToInsert.splits[i])

		cmpVal := compare(q, nodeToInsert.splits[i])
		if cmpVal < 0 {
			insertIdx = i
			break
		} else if cmpVal == 0 {
			if child == nil {
				// duplicate strategy for leaf nodes?
				// 1. just insert anyway
				// 2. ignore
				// 3. merge into data (if it is a leaf node)
				performInsert = false
				break
			}
		}
	}
	//fmt.Printf("Planning to insert %s at idx %d\n", q, insertIdx)

	if child != nil {

		// now I need to compare and see which one promotes the split
		if len(nodeToInsert.children) > 0 {
			nodeWithSplitValue := nodeToInsert.children[insertIdx]
			for len(nodeWithSplitValue.splits) == 0 {
				nodeWithSplitValue = nodeWithSplitValue.children[0]
			}

			//fmt.Println("comparing", q, nodeWithSplitValue.splits[0])
			cmpVal := compare(q, nodeWithSplitValue.splits[0])
			if cmpVal >= 0 {
				insertIdx++
			}
		}

		if insertIdx < len(nodeToInsert.children) {
			//fmt.Println("Inserting CHILD at index:", insertIdx, revToString(child))
			nodeToInsert.children = append(nodeToInsert.children, nil)
			copy(nodeToInsert.children[insertIdx+1:], nodeToInsert.children[insertIdx:])
			nodeToInsert.children[insertIdx] = child
			child.parent = nodeToInsert
			child.parentChildIdx = insertIdx
			for j := insertIdx + 1; j < len(nodeToInsert.children); j++ {
				nodeToInsert.children[j].parentChildIdx = j
			}
		} else {
			//fmt.Println("Appending CHILD to end:", revToString(child))
			nodeToInsert.children = append(nodeToInsert.children, child)
			child.parent = nodeToInsert
			child.parentChildIdx = insertIdx
		}
		insertIdx--
	}

	if performInsert && insertIdx >= 0 {
		if insertIdx < len(nodeToInsert.splits) {
			//fmt.Printf("Modifiying SPLIT %s with len=%d and cap=%d\n", strings.Join(nodeToInsert.splits, ""), len(nodeToInsert.splits), cap(nodeToInsert.splits))
			//fmt.Printf("Moving SPLIT from %s with len=%d and cap=%d\n", strings.Join(nodeToInsert.splits[insertIdx:], ""), len(nodeToInsert.splits[insertIdx:]), cap(nodeToInsert.splits[insertIdx:]))
			//fmt.Printf("Moving SPLIT onto %s with len=%d and cap=%d\n", strings.Join(nodeToInsert.splits[insertIdx+1:], ""), len(nodeToInsert.splits[insertIdx+1:]), cap(nodeToInsert.splits[insertIdx+1:]))
			nodeToInsert.splits = append(nodeToInsert.splits, " ")
			copy(nodeToInsert.splits[insertIdx+1:], nodeToInsert.splits[insertIdx:])
			//fmt.Printf("After copy SPLIT %s with len=%d and cap=%d\n", strings.Join(nodeToInsert.splits, ""), len(nodeToInsert.splits), cap(nodeToInsert.splits))
			nodeToInsert.splits[insertIdx] = q
			//fmt.Printf("Inserting SPLIT %s at %d: %s\n", q, insertIdx, strings.Join(nodeToInsert.splits, ""))
			if insertIdx == 0 && nodeToInsert.parent != nil && nodeToInsert.parentChildIdx > 0 {
				// update parent split for this node in nodeToInsert.parent.splits
				nodeToInsert.parent.splits[nodeToInsert.parentChildIdx-1] = q
				//fmt.Println("Updating parent split to", q, revToString(nodeToInsert))
			}
		} else {
			nodeToInsert.splits = append(nodeToInsert.splits, q)
			//fmt.Printf("Inserting SPLIT %s at end (%d): %s\n", q, insertIdx, strings.Join(nodeToInsert.splits, ""))
		}
	} else {
		//fmt.Println("Skipping SPLIT")
	}

	if newNode != nil {
		//fmt.Println("children after:", childrenStr(parent), len(parent.children), "|", childrenStr(newNode), len(newNode.children))
		//fmt.Println("adding newNode to parent: ", recToString(parent.parent))
		addToParentNode(parent.parent, newNode)
	} else {
		//fmt.Println(recToString(nodeToInsert))
	}
}

func (node *BPlusTreeNode) isLeafNode() bool {
	return len(node.children) == 0
}

func (tree *BPlusTree) FindNode(q string) *BPlusTreeNode {
	return recFindNode(q, tree.root)
}

func recFindNode(q string, node *BPlusTreeNode) *BPlusTreeNode {

	if node.isLeafNode() {
		return node
	}

	for j := 0; j < len(node.splits); j++ {
		sj := node.splits[j]
		if compare(q, sj) < 0 {
			return recFindNode(q, node.children[j])
		}
	}
	return recFindNode(q, node.children[len(node.splits)])
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

/*
// TODO Consolidate this with the Verify fn and add in distance Threshold
// Above is already DONE, leaving here in case bugs in the are found in the short term
func LowerBoundEst(q string, smin string, smax string) int {
	// TODO: how is distance threshold calculated?
	//fmt.Println("TODO", q, smin, smax)
	lcp := longestCommonPrefix(smin, smax)
	nlcp := len(lcp)
	var cmin uint8
	if len(smin) == nlcp {
		cmin = '!' // \u0021... if I trim, do I need to go lower?
	} else {
		cmin = smin[nlcp]
	}

	if len(smax) == nlcp {
		return intMax(nlcp, len(q))
	}
	cmax := smax[nlcp]

	//fmt.Println(string(cmin), string(cmax))
	distanceThreshold := 2 // TODO
	distanceThreshold = len(q)

	table := make([][]int, 2)
	for i := 0; i < len(table); i++ {
		table[i] = make([]int, len(q)+1)
	}

	firstEnd := intMin(len(q)+1, 1+distanceThreshold)
	for j := 0; j < firstEnd; j++ {
		table[0][j] = j
	}

	headerRow := "   \u2205 "
	for k := 0; k < len(q); k++ {
		headerRow += string(q[k]) + " "
	}
	//fmt.Println(headerRow)

	row := fmt.Sprintf("\u2205  ")
	for k := 0; k < len(table[0]); k++ {
		if k < firstEnd {
			row += fmt.Sprintf("%d ", table[0][k])
		} else {
			row += "- "
		}
	}
	//fmt.Println(row)

	// i == 0 is the empty string... handled by init above
	for i := 1; i < len(lcp)+2; i++ {
		start := intMax(0, i-distanceThreshold)
		end := intMin(len(q)+1, i+distanceThreshold+1)
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

				if i < len(lcp)+1 {
					if lcp[i-1] != q[j-1] {
						d3 += 1
					}
				} else {
					if q[j-1] < cmin || q[j-1] > cmax {
						d3 += 1
					}
				}
			} else {
				d2 = distanceThreshold + 1
				d3 = distanceThreshold + 1
			}
			//fmt.Printf("%d %d %d\n", d1, d2, d3)
			table[1][j] = intMin(intMin(d1, d2), d3)
		}

		var row string
		if i < len(lcp)+1 {
			row = fmt.Sprintf("%v  ", string(lcp[i-1]))
		} else {
			row = fmt.Sprintf("%v%v ", string(cmin), string(cmax))
		}
		for k := 0; k < len(table[0]); k++ {
			if k >= start && k < end {
				row += fmt.Sprintf("%d ", table[1][k])
			} else {
				row += "- "
			}
		}
		//fmt.Println(row)

		for j, n := 0, len(q)+1; j < n; j++ {
			table[0][j] = table[1][j]
		}
	}

	min := intMax(len(q), nlcp+1)
	row = "   "
	for k := 0; k < len(table[0]); k++ {
		row += fmt.Sprintf("%d ", table[0][k])
		if table[0][k] < min {
			min = table[0][k]
		}
	}
	//fmt.Println(row)

	return min
}
*/

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

func (tree *BPlusTree) RangeQuery(q string, distanceThreshold int) []string {
	results := make([]string, 0) // length?
	//resultsChan := make(chan string)
	// TODO: change to channels to parallelize the next bit
	results = recRangeQuery(q, tree.root, distanceThreshold, "", "", results)
	//fmt.Printf("%v\n", results)
	return results
}

func recRangeQuery(q string, node *BPlusTreeNode, distanceThreshold int, smin string, smax string, results []string) []string {
	if node.isLeafNode() {
		for j := 0; j < len(node.splits); j++ {
			sj := node.splits[j]
			if VerifyEditDistance(q, sj, distanceThreshold) {
				results = append(results, sj)
			}
		}
	} else {
		if len(node.splits) > 0 {
			if VerifyLowerBound(q, smin, node.splits[0], distanceThreshold) {
				results = recRangeQuery(q, node.children[0], distanceThreshold, smin, node.splits[0], results)
			}

			for j, m := 1, len(node.splits); j < m; j++ {
				if VerifyLowerBound(q, node.splits[j-1], node.splits[j], distanceThreshold) {
					results = recRangeQuery(q, node.children[j], distanceThreshold, node.splits[j-1], node.splits[j], results)
				}
			}

			// I want smax == "" to be interpretted like the last possible word in the alphabet?
			// which would pretty much guarantee an lcp of 0, which the empty string achieves
			if VerifyLowerBound(q, node.splits[len(node.splits)-1], smax, distanceThreshold) {
				results = recRangeQuery(q, node.children[len(node.splits)], distanceThreshold, node.splits[len(node.splits)-1], smax, results)
			}
		} else {
			if len(node.children) > 0 { // should only ever be one...
				results = recRangeQuery(q, node.children[0], distanceThreshold, smin, smax, results)
			}
		}
	}
	return results
}

func compareDictionaryOrder(a string, b string) int {
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

func compare(a string, b string) int {

	//if float64(damerau.DamerauLevenshteinDistance(a, b))/math.Max(float64(len(a)), float64(len(b))) < 0.1 {
	//if VerifyEditDistance(a, b, int(0.1*float64(intMax(len(a), len(b))))) {
	//	return 0
	//}

	return compareDictionaryOrder(a, b)
}
