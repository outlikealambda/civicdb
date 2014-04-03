package bed

type QueryResult struct {
	Key    string
	Values []interface{}
}

func (tree *BPlusTree) RangeQuery(q string, distanceThreshold int) []QueryResult {
	results := make([]QueryResult, 0)
	results = recRangeQuery(q, tree.root, distanceThreshold, "", "", results)
	return results
}

/*
func (tree *BPlusTree) RangeQuery(q string, distanceThreshold int) []string {
	results := make([]string, 0) // length?
	//resultsChan := make(chan string)
	// TODO: change to channels to parallelize the next bit
	results = recRangeQuery(q, tree.root, distanceThreshold, "", "", results)
	//fmt.Printf("%v\n", results)
	return results
}*/

func recRangeQuery(q string, node *bPlusTreeNode, distanceThreshold int, smin string, smax string, results []QueryResult) []QueryResult {
	if node.isLeafNode() {
		for j, split := range node.splits {
			if VerifyEditDistance(q, split, distanceThreshold) {
				results = append(results, QueryResult{split, node.data[j]})
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
