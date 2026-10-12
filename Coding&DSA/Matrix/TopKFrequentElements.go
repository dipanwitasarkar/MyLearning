package Matrix

import "container/heap"

func topKFrequent(nums []int, k int) []int {
	// o(n)
	// // Step 1: Count frequencies
	// freq := make(map[int]int)

	// for _, num := range nums {
	// 	freq[num]++
	// }

	// // Step 2: Create buckets
	// // bucket[i] contains numbers that appear i times
	// buckets := make([][]int, len(nums)+1)

	// for num, count := range freq {
	// 	buckets[count] = append(buckets[count], num)
	// }

	// // Step 3: Collect top k frequent elements
	// result := make([]int, 0, k)

	// for i := len(buckets) - 1; i >= 0; i-- {
	// 	for _, num := range buckets[i] {
	// 		result = append(result, num)

	// 		if len(result) == k {
	// 			return result
	// 		}
	// 	}
	// }

	// return result

	// Approach 2 - heap
	// O(n + m log k)
	freq := make(map[int]int)

	for _, num := range nums {
		freq[num]++
	}

	h := &Minheap{}
	heap.Init(h)
	for num, count := range freq {
		heap.Push(
			h, Pair{
				num:  num,
				freq: count,
			})
		if h.Len() > k {
			heap.Pop(h)
		}
	}

	result := []int{}
	for h.Len() > 0 {
		p := heap.Pop(h).(Pair)
		result = append(result, p.num)
	}

	return result

}

type Pair struct {
	num  int
	freq int
}

type Minheap []Pair

func (h Minheap) Len() int {
	return len(h)
}

func (h Minheap) Less(i, j int) bool {
	return h[i].freq < h[j].freq
}

func (h Minheap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *Minheap) Push(x interface{}) {
	*h = append(*h, x.(Pair))
}

func (h *Minheap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
