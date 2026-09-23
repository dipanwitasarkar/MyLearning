package mediumhard

func trap(height []int) int {
	// 2 pointers
	left := 0
	leftMax := 0
	right := len(height) - 1
	rightMax := 0
	water := 0
	for left < right {
		if height[left] < height[right] {
			if height[left] > leftMax {
				leftMax = height[left]
			} else {
				water += leftMax - height[left]
			}
			left++
		} else {
			if height[right] > rightMax {
				rightMax = height[right]
			} else {
				water += rightMax - height[right]
			}
			right--
		}
	}
	return water
}
