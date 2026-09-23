package mediumhard

func maxSlidingWindow(nums []int, k int) []int {
	// sliding window + monotonic queue
	dq := make([]int, 0) // store index
	result := make([]int, 0)

	for i := 0; i < len(nums); i++ {
		// remove expired items for window ending at index i, the start would be i-k+1.
		// anything smaller than i-k is expired
		if len(dq) > 0 && dq[0] <= i-k {
			dq = dq[1:]
		}
		// decreasing order (remove all numbers lesser than nums[i])
		for len(dq) > 0 && nums[dq[len(dq)-1]] < nums[i] {
			dq = dq[:len(dq)-1]
		}
		dq = append(dq, i)
		// window reached
		if i >= k-1 {
			result = append(result, nums[dq[0]])
		}
	}
	return result
}
