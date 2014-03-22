package bed

import (
	"fmt"
	"strings"
)

func (tree *BPlusTree) String() string {
	return recToString(tree.root)
}

func recToString(node *bPlusTreeNode) string {
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

func revToString(node *bPlusTreeNode) string {
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

func childrenStr(node *bPlusTreeNode) string {
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
