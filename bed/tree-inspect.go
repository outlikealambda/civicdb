package bed

func (tree *BPlusTree) NumTotalNodes() int {
	return recNumNodes(tree.root, func(*bPlusTreeNode) int {
		return 1
	})
}

func (tree *BPlusTree) NumLeafNodes() int {
	return recNumNodes(tree.root, func(node *bPlusTreeNode) int {
		if node.isLeafNode() {
			return 1
		}
		return 0
	})
}

func (tree *BPlusTree) Size() int {
	return recNumNodes(tree.root, func(node *bPlusTreeNode) int {
		if node.isLeafNode() {
			return len(node.splits)
		}
		return 0
	})
}

func recNumNodes(parent *bPlusTreeNode, count func(*bPlusTreeNode) int) int {
	total := count(parent)
	for _, child := range parent.children {
		total += recNumNodes(child, count)
	}
	return total
}
