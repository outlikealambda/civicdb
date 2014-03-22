package bed

import (
	//"github.com/megesdal/melodispurences/damerau"
	"math"
)

type BPlusTree struct {
	branchFactor int
	root         *bPlusTreeNode
	compare      func(string, string) int
}

type bPlusTreeNode struct {
	//tree           *BPlusTree
	parent         *bPlusTreeNode
	parentChildIdx int
	splits         []string         // size m
	children       []*bPlusTreeNode // size m + 1
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

func New(b int, compare func(string, string) int) *BPlusTree {
	if b < 2 {
		return nil
	}

	tree := new(BPlusTree)
	tree.branchFactor = b
	tree.compare = compare
	tree.root = tree.createTreeNode()
	return tree
}

func (tree *BPlusTree) createTreeNode() *bPlusTreeNode {
	node := new(bPlusTreeNode)
	//node.tree = tree
	node.splits = make([]string, 0, tree.branchFactor-1)
	node.children = make([]*bPlusTreeNode, 0, tree.branchFactor)
	return node
}

func (tree *BPlusTree) Insert(q string) {
	nodeToInsert := tree.recFindNode(q, tree.root)
	tree.recInsert(q, nodeToInsert, nil)
}

func (tree *BPlusTree) addToParentNode(parent *bPlusTreeNode, child *bPlusTreeNode) {
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
			newRoot := tree.createTreeNode()
			newRoot.children = append(newRoot.children, tree.root)
			newRoot.children = append(newRoot.children, child)
			newRoot.splits = append(newRoot.splits, splitValue)
			tree.root.parent = newRoot
			tree.root.parentChildIdx = 0
			child.parent = newRoot
			child.parentChildIdx = 1
			tree.root = newRoot
			//fmt.Println("new root...", revToString(newRoot.children[0]), revToString(newRoot.children[1]))
		} else {
			tree.recInsert(splitValue, parent, child)
		}
	}
}

// TODO: figure out why this isn't working
// TODO: I'd like to get this so it didn't depend on q and addToNew is computed outside
/*func splitIfNecessary(parent *bPlusTreeNode, q string) (*bPlusTreeNode, bool) {

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
		newNode.children = make([]*bPlusTreeNode, len(parent.children)-childSplitIdx)
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

func (tree *BPlusTree) recInsert(q string, parent *bPlusTreeNode, child *bPlusTreeNode) {

	indent := ""
	node := parent.parent
	for node != nil {
		indent += "  "
		node = node.parent
	}

	nodeToInsert := parent
	//fmt.Println(indent, "Parent node", revToString(parent))
	var newNode *bPlusTreeNode

	shouldSplit := false
	if len(nodeToInsert.splits) == tree.branchFactor-1 {
		shouldSplit = true
	}

	if shouldSplit {

		//fmt.Println(indent, "Splitting before adding child node")
		newNode = tree.createTreeNode()
		//childSplitIdx := int(math.Ceil(float64(parent.tree.branchFactor) / 2))
		splitIdx := int(math.Floor(float64(tree.branchFactor) / 2))

		addToNew := false
		if tree.branchFactor == 2 {
			// only one of the two nodes will have room...
			if tree.compare(q, parent.splits[0]) < 0 {
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

			if tree.compare(q, nodeToInsert.splits[splitIdx]) >= 0 {
				addToNew = true
			}
			nodeToInsert.splits = nodeToInsert.splits[:splitIdx]
		} // else is branchFactor = 2 (binary tree)

		//fmt.Printf("%s split: '%s' %d | '%s' %d \n", indent, strings.Join(nodeToInsert.splits, ""), len(nodeToInsert.splits), strings.Join(newNode.splits, ""), len(newNode.splits))

		if child != nil {
			childSplitIdx := splitIdx + 1
			//fmt.Println(indent, "child split index", childSplitIdx)
			newNode.children = make([]*bPlusTreeNode, len(nodeToInsert.children)-childSplitIdx)
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

		cmpVal := tree.compare(q, nodeToInsert.splits[i])
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
			cmpVal := tree.compare(q, nodeWithSplitValue.splits[0])
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
		tree.addToParentNode(parent.parent, newNode)
	} else {
		//fmt.Println(recToString(nodeToInsert))
	}
}

func (node *bPlusTreeNode) isLeafNode() bool {
	return len(node.children) == 0
}

func (tree *BPlusTree) recFindNode(q string, node *bPlusTreeNode) *bPlusTreeNode {

	if node.isLeafNode() {
		return node
	}

	for j := 0; j < len(node.splits); j++ {
		sj := node.splits[j]
		if tree.compare(q, sj) < 0 {
			return tree.recFindNode(q, node.children[j])
		}
	}
	return tree.recFindNode(q, node.children[len(node.splits)])
}
