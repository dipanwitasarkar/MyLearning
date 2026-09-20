package easymedium

func moveZeroes(nums []int) {
	indexpos := 0
	for i := 0; i < len(nums); i++ {
		if nums[i] != 0 {
			nums[indexpos], nums[i] = nums[i], nums[indexpos]
			indexpos++
		}
	}
}
