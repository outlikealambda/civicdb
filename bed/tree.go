package bed

import (
	"fmt"
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
}

func intMin(a int, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func VerifyED(s1 string, s2 string, distanceThreshold int) bool {

	if int(math.Abs(float64(len(s1)-len(s2)))) > distanceThreshold {
		return false
	}

	// construct tabel of 2 rows and len(s2) + 1 columns
	table := make([][]int, 2)
	for i := 0; i < len(table); i++ {
		table[i] = make([]int, len(s2)+1)
	}

	for j, n := 1, intMin(len(s2)+1, 1+distanceThreshold); j <= n; j++ {
		table[0][j] = j - 1
	}

	m := distanceThreshold + 1
	for i := 2; i <= len(s1)+1; i++ {
		for j, n := intMin(1, i-distanceThreshold), intMin(len(s2)+1, i+distanceThreshold); j <= n; j++ {
			var d1 int
			if j < i+distanceThreshold {
				d1 = table[0][j]
			} else {
				d1 = distanceThreshold + 1
			}

			var d2 int
			var d3 int
			if j > 1 {
				d2 = table[1][j-1] + 1
				d3 = table[0][j-1]
				if s1[i-1] == s2[j-1] {
					d3 += 1
				}
			} else {
				d2 = distanceThreshold + 1
				d3 = distanceThreshold + 1
			}
			table[1][j] = intMin(intMin(d1, d2), d3)
			m = intMin(m, table[1][j])
		}
		if m > distanceThreshold {
			return false
		}
		for j, n := 0, len(s2)+1; j <= n; j++ {
			table[0][j] = table[1][j]
		}
	}
	return true
}

func NewTree(b int) *BPlusTree {
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
	for i, n := 0, len(nodeToInsert.splits); i < n; i++ {
		//fmt.Printf("Comparing '%s' to '%s'\n", q, nodeToInsert.splits[i])
		if compare(q, nodeToInsert.splits[i]) < 0 {
			insertIdx = i
			break
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
			if compare(q, nodeWithSplitValue.splits[0]) >= 0 {
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

	if insertIdx >= 0 {
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

func lowerBound(q string, smin string, smax string) int {
	//fmt.Println("TODO", q, smin, smax)
	return 0
}

func RangeQuery(q string, node *BPlusTreeNode, distanceThreshold int, smin string, smax string) {
	if node.isLeafNode() {
		for j := 0; j < len(node.splits); j++ {
			sj := node.splits[j]
			if VerifyED(q, sj, distanceThreshold) {
				//fmt.Println("Include", sj)
			}
		}
	} else {

	}
}

func compare(a string, b string) int {
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
