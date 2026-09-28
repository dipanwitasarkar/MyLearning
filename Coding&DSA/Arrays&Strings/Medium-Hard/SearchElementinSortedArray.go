package mediumhard

func search(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left <= right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			return mid
		}

		// left half sorted
		if nums[left] <= nums[mid] {

			if target >= nums[left] &&
				target < nums[mid] {
				right = mid - 1
			} else {
				left = mid + 1
			}

		} else { // right half sorted

			if target > nums[mid] &&
				target <= nums[right] {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
	}

	return -1
}
