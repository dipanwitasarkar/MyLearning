package easymedium

func removeDuplicates(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	// 2 pointers approach
	indexpos := 1
	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1] {
			nums[indexpos] = nums[i]
			indexpos++
		}
	}
	return indexpos
}
