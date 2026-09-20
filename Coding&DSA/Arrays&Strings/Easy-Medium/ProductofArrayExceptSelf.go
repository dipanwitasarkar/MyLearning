package easymedium

func productExceptSelf(nums []int) []int {
	n := len(nums)
	answer := make([]int, n)

	// find prefix products (product of the numbers before nums[i])
	answer[0] = 1
	for i := 1; i < len(nums); i++ {
		answer[i] = answer[i-1] * nums[i-1]
	}

	// find suffix products(product of the numbers after nums[i]) and multiply with prefix products
	right := 1
	for i := n - 1; i >= 0; i-- {
		answer[i] = answer[i] * right
		right = right * nums[i]
	}
	return answer
}
