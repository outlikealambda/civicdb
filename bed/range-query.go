package bed

type RangeResult struct {
	Key      string
	Values   []interface{}
	Distance float64
}

// TODO: parallelize this...
func (tree *BPlusTree) RangeQuery(q string, distanceThreshold float64) []RangeResult {
	results := make([]RangeResult, 0)
	results = recRangeQuery(q, tree.root, distanceThreshold, "", "", results)
	return results
}

func recRangeQuery(q string, node *bPlusTreeNode, distanceThreshold float64, smin string, smax string, results []RangeResult) []RangeResult {
	if node.isLeafNode() {
		for j, leaf := range node.splits {
			if success, editDistance := VerifyEditDistance(q, leaf, denormalizeDistance(distanceThreshold, q, leaf)); success {
				results = append(results, RangeResult{leaf, node.data[j], normalizeDistance(editDistance, q, leaf)})
			}
		}
	} else {
		if len(node.splits) > 0 {
			if VerifyLowerBound(q, smin, node.splits[0], denormalizeDistance(distanceThreshold, q, smin)) {
				results = recRangeQuery(q, node.children[0], distanceThreshold, smin, node.splits[0], results)
			}

			for j, m := 1, len(node.splits); j < m; j++ {
				if VerifyLowerBound(q, node.splits[j-1], node.splits[j], denormalizeDistance(distanceThreshold, q, smin)) {
					results = recRangeQuery(q, node.children[j], distanceThreshold, node.splits[j-1], node.splits[j], results)
				}
			}

			// I want smax == "" to be interpretted like the last possible word in the alphabet?
			// which would pretty much guarantee an lcp of 0, which the empty string achieves
			if VerifyLowerBound(q, node.splits[len(node.splits)-1], smax, denormalizeDistance(distanceThreshold, q, smin)) {
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

func denormalizeDistance(threshold float64, si string, sj string) int {
	return int(threshold * float64(intMax(len(si), len(sj))))
}

func normalizeDistance(editDistance int, si string, sj string) float64 {
	return float64(editDistance) / float64(intMax(len(si), len(sj)))
}
