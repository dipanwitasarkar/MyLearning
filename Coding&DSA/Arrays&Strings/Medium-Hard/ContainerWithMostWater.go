package mediumhard

func maxArea(height []int) int {
	i := 0
	j := len(height) - 1
	result := 0
	for j > i {
		result = max(result, min(height[i], height[j])*(j-i))
		if height[i] < height[j] {
			i++
		} else {
			j--
		}
	}
	return result
}

func min(i, j int) int {
	if i <= j {
		return i
	}
	return j
}

func max(result int, height int) int {
	if result >= height {
		return result
	}
	return height
}
