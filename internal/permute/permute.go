package permute

type PermutationIterator struct {
	lists    [][]string
	indices  []int
	finished bool
}

func NewPermutationIterator(lists [][]string) *PermutationIterator {
	return &PermutationIterator{
		lists:   lists,
		indices: make([]int, len(lists)),
	}
}

func (pi *PermutationIterator) Next() ([]string, bool) {
	if pi.finished {
		return nil, false
	}
	result := make([]string, len(pi.lists))
	for i, list := range pi.lists {
		result[i] = list[pi.indices[i]]
	}
	// Update indices
	for i := len(pi.indices) - 1; i >= 0; i-- {
		pi.indices[i]++
		if pi.indices[i] < len(pi.lists[i]) {
			break
		}
		pi.indices[i] = 0
		if i == 0 {
			pi.finished = true
		}
	}
	return result, true
}

func ShardLists(listOfLists [][]string, n int, numOfShards int) [][][]string {
	if len(listOfLists) == 0 || n < 0 || n >= len(listOfLists) || numOfShards <= 0 {
		return [][][]string{}
	}
	// Determine the size of each shard for the nth list
	nthListLen := len(listOfLists[n])
	shardSize := (nthListLen + numOfShards - 1) / numOfShards // Round up division
	// Create the shards
	shards := make([][][]string, numOfShards)
	for i := range shards {
		shards[i] = make([][]string, len(listOfLists))
	}
	// Fill the shards
	for i := 0; i < len(listOfLists); i++ {
		if i == n {
			// Shard the nth list
			for j := 0; j < numOfShards; j++ {
				start := j * shardSize
				end := start + shardSize
				if end > nthListLen {
					end = nthListLen
				}
				shards[j][i] = listOfLists[i][start:end]
			}
		} else {
			// For other lists, just assign the reference to all shards
			for j := 0; j < numOfShards; j++ {
				shards[j][i] = listOfLists[i]
			}
		}
	}
	return shards
}

func CalculateTotalPermutations(shardedLists [][][]string) int {
	if len(shardedLists) == 0 || len(shardedLists[0]) == 0 {
		return 0
	}
	total := 1
	for _, list := range shardedLists[0] {
		total *= len(list)
	}
	return total * len(shardedLists)
}

func IteratePermutations(permuter *PermutationIterator, results chan<- []string) {
	for {
		perm, ok := permuter.Next()
		if !ok {
			return
		}
		results <- perm
	}
}
